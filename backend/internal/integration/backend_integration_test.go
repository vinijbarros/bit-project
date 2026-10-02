package integration_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pressly/goose/v3"
	"golang.org/x/crypto/bcrypt"

	"portal-solicitacoes/internal/config"
	"portal-solicitacoes/internal/domain"
	httpapi "portal-solicitacoes/internal/http"
	"portal-solicitacoes/internal/repository"
	"portal-solicitacoes/internal/repository/database"
	"portal-solicitacoes/internal/service"
)

const trustedOrigin = "http://127.0.0.1:5173"

var migrationOnce sync.Once
var migrationErr error

type testApp struct {
	db      *sql.DB
	server  *httptest.Server
	client1 *http.Client
	client2 *http.Client
}

type apiResponse struct {
	Data       json.RawMessage `json:"data"`
	Items      json.RawMessage `json:"items"`
	Pagination json.RawMessage `json:"pagination"`
	Error      *struct {
		Code string `json:"code"`
	} `json:"error"`
}

func openTestDatabase(t *testing.T) *sql.DB {
	t.Helper()
	rawURL := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL"))
	if rawURL == "" {
		t.Skip("TEST_DATABASE_URL ausente; PostgreSQL de integração não está disponível")
	}
	if err := validateTestDatabase(rawURL, os.Getenv("TEST_DATABASE_ALLOW_RESET")); err != nil {
		t.Fatal(err)
	}

	db, err := database.Open(config.DatabaseConfig{
		URL: rawURL, ConnectTimeout: 3 * time.Second, MaxOpenConns: 12, MaxIdleConns: 6,
		ConnMaxLifetime: 10 * time.Minute, ConnMaxIdleTime: time.Minute,
	})
	if err != nil {
		t.Fatalf("abrir banco de teste: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		t.Fatalf("conectar ao banco de teste: %v", err)
	}

	migrationOnce.Do(func() {
		if err := goose.SetDialect("postgres"); err != nil {
			migrationErr = err
			return
		}
		migrationDir, err := filepath.Abs("../../db/migrations")
		if err != nil {
			migrationErr = err
			return
		}
		migrationErr = goose.Up(db, migrationDir)
	})
	if migrationErr != nil {
		db.Close()
		t.Fatalf("aplicar migrations reais: %v", migrationErr)
	}
	return db
}

func validateTestDatabase(rawURL, allowReset string) error {
	if strings.TrimSpace(rawURL) == "" {
		return errors.New("TEST_DATABASE_URL é obrigatória")
	}
	if allowReset != "yes" {
		return errors.New("TEST_DATABASE_ALLOW_RESET deve ser exatamente 'yes' para autorizar limpeza")
	}
	parsed, err := pgx.ParseConfig(rawURL)
	if err != nil {
		return fmt.Errorf("TEST_DATABASE_URL inválida: %w", err)
	}
	if !strings.HasSuffix(parsed.Database, "_test") {
		return fmt.Errorf("banco recusado: nome %q não termina em _test", parsed.Database)
	}
	return nil
}

func TestDatabaseSafetyGuard(t *testing.T) {
	validURL := "postgres://tester:tester@127.0.0.1:5432/portal_test?sslmode=disable"
	tests := []struct {
		name, databaseURL, confirmation string
		wantError                       bool
	}{
		{name: "missing url", confirmation: "yes", wantError: true},
		{name: "missing confirmation", databaseURL: validURL, wantError: true},
		{name: "development database name", databaseURL: "postgres://tester:tester@127.0.0.1:5432/portal?sslmode=disable", confirmation: "yes", wantError: true},
		{name: "explicit test database", databaseURL: validURL, confirmation: "yes"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateTestDatabase(test.databaseURL, test.confirmation)
			if (err != nil) != test.wantError {
				t.Fatalf("validateTestDatabase() error = %v, wantError=%v", err, test.wantError)
			}
		})
	}
}

func newTestApp(t *testing.T) *testApp {
	t.Helper()
	db := openTestDatabase(t)
	if _, err := db.Exec(`TRUNCATE requests, sessions, users RESTART IDENTITY CASCADE`); err != nil {
		db.Close()
		t.Fatalf("limpar banco de teste: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(`TRUNCATE requests, sessions, users RESTART IDENTITY CASCADE`)
		db.Close()
	})

	insertUser(t, db, "colaborador1", "Colaborador 1", "SenhaLocal-1")
	insertUser(t, db, "colaborador2", "Colaborador 2", "SenhaLocal-2")

	authService := service.NewAuth(repository.NewAuth(db), service.AuthConfig{SessionDuration: 8 * time.Hour, TokenBytes: 32})
	requestService := service.NewRequest(repository.NewRequest(db))
	dashboardService := service.NewDashboard(repository.NewDashboard(db))
	router := httpapi.NewRouter(db, httpapi.RouterConfig{
		ReadinessTimeout: time.Second,
		Logger:           slog.New(slog.NewTextHandler(io.Discard, nil)),
		AuthService:      authService,
		RequestService:   requestService,
		DashboardService: dashboardService,
		TrustedOrigins:   []string{trustedOrigin},
		SessionCookie:    httpapi.SessionCookieConfig("portal_session", false, 8*time.Hour),
		LoginRateLimit:   httpapi.LoginRateLimitConfig(20, time.Minute, 100),
	})
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)
	return &testApp{db: db, server: server, client1: newClient(t), client2: newClient(t)}
}

func newClient(t *testing.T) *http.Client {
	t.Helper()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	return &http.Client{Jar: jar, Timeout: 5 * time.Second}
}

func insertUser(t *testing.T, db *sql.DB, username, displayName, password string) {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO users (username, display_name, password_hash) VALUES ($1, $2, $3)`, username, displayName, string(hash)); err != nil {
		t.Fatalf("inserir usuário de teste: %v", err)
	}
}

func (app *testApp) do(t *testing.T, client *http.Client, method, path string, body any, withOrigin bool) (int, http.Header, apiResponse) {
	t.Helper()
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		reader = bytes.NewReader(encoded)
	}
	request, err := http.NewRequest(method, app.server.URL+path, reader)
	if err != nil {
		t.Fatal(err)
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if withOrigin {
		request.Header.Set("Origin", trustedOrigin)
	}
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer response.Body.Close()
	var decoded apiResponse
	if response.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(response.Body).Decode(&decoded); err != nil {
			t.Fatalf("decodificar %s %s (status %d): %v", method, path, response.StatusCode, err)
		}
	}
	return response.StatusCode, response.Header, decoded
}

func (app *testApp) login(t *testing.T, client *http.Client, username, password string) {
	t.Helper()
	status, _, response := app.do(t, client, http.MethodPost, "/api/v1/auth/login", map[string]string{"username": username, "password": password}, true)
	if status != http.StatusOK || response.Error != nil {
		t.Fatalf("login de %s: status=%d error=%+v", username, status, response.Error)
	}
}

func (app *testApp) createRequest(t *testing.T, client *http.Client, title, description, category string) int64 {
	t.Helper()
	status, headers, response := app.do(t, client, http.MethodPost, "/api/v1/requests", map[string]string{
		"title": title, "description": description, "category": category,
	}, true)
	if status != http.StatusCreated {
		t.Fatalf("criar solicitação: status=%d error=%+v", status, response.Error)
	}
	var data struct {
		ID     int64  `json:"id"`
		Title  string `json:"title"`
		Status struct {
			Value string `json:"value"`
		} `json:"status"`
		Requester struct {
			Username string `json:"username"`
		} `json:"requester"`
		CreatedAt time.Time `json:"created_at"`
	}
	if err := json.Unmarshal(response.Data, &data); err != nil {
		t.Fatal(err)
	}
	if data.ID <= 0 || headers.Get("Location") != fmt.Sprintf("/api/v1/requests/%d", data.ID) || data.Status.Value != domain.StatusOpen || data.Requester.Username == "" || data.CreatedAt.IsZero() {
		t.Fatalf("campos automáticos/Location inválidos: headers=%v data=%+v", headers, data)
	}
	return data.ID
}

func TestAuthenticationPersistenceAndFailures(t *testing.T) {
	app := newTestApp(t)

	status, _, invalid := app.do(t, app.client1, http.MethodPost, "/api/v1/auth/login", map[string]string{"username": "colaborador1", "password": "incorreta"}, true)
	if status != http.StatusUnauthorized || invalid.Error == nil || invalid.Error.Code != "invalid_credentials" {
		t.Fatalf("credencial inválida: status=%d error=%+v", status, invalid.Error)
	}
	app.login(t, app.client1, "colaborador1", "SenhaLocal-1")
	status, _, me := app.do(t, app.client1, http.MethodGet, "/api/v1/auth/me", nil, false)
	if status != http.StatusOK || bytes.Contains(me.Data, []byte("password")) || bytes.Contains(me.Data, []byte("token")) {
		t.Fatalf("/me inseguro: status=%d data=%s", status, me.Data)
	}

	app.server.Close()
	reconnectedDB := openTestDatabase(t)
	t.Cleanup(func() { _ = reconnectedDB.Close() })
	authService := service.NewAuth(repository.NewAuth(reconnectedDB), service.AuthConfig{SessionDuration: 8 * time.Hour, TokenBytes: 32})
	app.server = httptest.NewServer(httpapi.NewRouter(reconnectedDB, httpapi.RouterConfig{
		ReadinessTimeout: time.Second, AuthService: authService,
		RequestService:   service.NewRequest(repository.NewRequest(reconnectedDB)),
		DashboardService: service.NewDashboard(repository.NewDashboard(reconnectedDB)),
		TrustedOrigins:   []string{trustedOrigin},
		SessionCookie:    httpapi.SessionCookieConfig("portal_session", false, 8*time.Hour),
		LoginRateLimit:   httpapi.LoginRateLimitConfig(20, time.Minute, 100),
	}))
	t.Cleanup(app.server.Close)
	status, _, _ = app.do(t, app.client1, http.MethodGet, "/api/v1/auth/me", nil, false)
	if status != http.StatusOK {
		t.Fatalf("sessão não persistiu após reinício HTTP: status=%d", status)
	}

	status, _, _ = app.do(t, app.client1, http.MethodPost, "/api/v1/auth/logout", nil, true)
	if status != http.StatusNoContent {
		t.Fatalf("logout: status=%d", status)
	}
	status, _, _ = app.do(t, app.client1, http.MethodGet, "/api/v1/auth/me", nil, false)
	if status != http.StatusUnauthorized {
		t.Fatalf("sessão revogada ainda aceita: status=%d", status)
	}

	expiredToken := strings.Repeat("a", 43)
	// O token precisa ter 32 bytes ao decodificar; usa uma sessão válida obtida por login e expira seu registro no banco.
	app.login(t, app.client1, "colaborador1", "SenhaLocal-1")
	serverURL, _ := url.Parse(app.server.URL)
	cookies := app.client1.Jar.Cookies(serverURL)
	for _, cookie := range cookies {
		if cookie.Name == "portal_session" {
			expiredToken = cookie.Value
		}
	}
	digest := serviceTokenHash(expiredToken)
	if _, err := app.db.Exec(`UPDATE sessions SET expires_at = created_at + interval '1 microsecond' WHERE token_hash = $1`, digest); err != nil {
		t.Fatal(err)
	}
	status, _, _ = app.do(t, app.client1, http.MethodGet, "/api/v1/auth/me", nil, false)
	if status != http.StatusUnauthorized {
		t.Fatalf("sessão expirada ainda aceita: status=%d", status)
	}
}

func TestRequestsAuthorizationStateFiltersAndDashboard(t *testing.T) {
	app := newTestApp(t)
	app.login(t, app.client1, "colaborador1", "SenhaLocal-1")
	app.login(t, app.client2, "colaborador2", "SenhaLocal-2")

	for _, route := range []struct{ method, path string }{
		{http.MethodGet, "/api/v1/requests"}, {http.MethodGet, "/api/v1/requests/1"}, {http.MethodGet, "/api/v1/dashboard"},
		{http.MethodPost, "/api/v1/requests"}, {http.MethodPatch, "/api/v1/requests/1"}, {http.MethodDelete, "/api/v1/requests/1"}, {http.MethodPatch, "/api/v1/requests/1/status"},
	} {
		status, _, _ := app.do(t, newClient(t), route.method, route.path, nil, route.method != http.MethodGet)
		if status != http.StatusUnauthorized {
			t.Fatalf("rota anônima %s %s: status=%d", route.method, route.path, status)
		}
	}

	id := app.createRequest(t, app.client1, "  Título com espaços  ", "  Descrição criada pelo primeiro usuário.  ", "ti")
	status, _, detail := app.do(t, app.client1, http.MethodGet, fmt.Sprintf("/api/v1/requests/%d", id), nil, false)
	if status != http.StatusOK || !bytes.Contains(detail.Data, []byte(`"title":"Título com espaços"`)) || !bytes.Contains(detail.Data, []byte(`"username":"colaborador1"`)) {
		t.Fatalf("criação/trim/autor: status=%d data=%s", status, detail.Data)
	}

	status, _, massAssignment := app.do(t, app.client1, http.MethodPost, "/api/v1/requests", map[string]any{
		"title": "Tentativa indevida", "description": "Payload tenta definir o autor.", "category": "rh", "requester_id": 2,
	}, true)
	if status != http.StatusBadRequest || massAssignment.Error == nil || massAssignment.Error.Code != "invalid_json" {
		t.Fatalf("mass assignment: status=%d error=%+v", status, massAssignment.Error)
	}

	status, _, _ = app.do(t, app.client2, http.MethodPatch, fmt.Sprintf("/api/v1/requests/%d", id), map[string]string{"title": "Alteração de outro autor"}, true)
	if status != http.StatusForbidden {
		t.Fatalf("edição por outro autor: status=%d", status)
	}
	status, _, _ = app.do(t, app.client2, http.MethodDelete, fmt.Sprintf("/api/v1/requests/%d", id), nil, true)
	if status != http.StatusForbidden {
		t.Fatalf("exclusão por outro autor: status=%d", status)
	}

	status, _, _ = app.do(t, app.client2, http.MethodPatch, fmt.Sprintf("/api/v1/requests/%d/status", id), map[string]string{"status": domain.StatusCompleted}, true)
	if status != http.StatusOK {
		t.Fatalf("mudança de status por outro usuário: status=%d", status)
	}
	status, _, _ = app.do(t, app.client1, http.MethodPatch, fmt.Sprintf("/api/v1/requests/%d", id), map[string]string{"title": "Bloqueada"}, true)
	if status != http.StatusConflict {
		t.Fatalf("edição concluída: status=%d", status)
	}
	status, _, _ = app.do(t, app.client1, http.MethodDelete, fmt.Sprintf("/api/v1/requests/%d", id), nil, true)
	if status != http.StatusConflict {
		t.Fatalf("exclusão concluída: status=%d", status)
	}
	status, _, _ = app.do(t, app.client2, http.MethodPatch, fmt.Sprintf("/api/v1/requests/%d/status", id), map[string]string{"status": domain.StatusOpen}, true)
	if status != http.StatusOK {
		t.Fatalf("reabertura: status=%d", status)
	}

	maliciousTitle := `Relatório 100%_d'água; DROP TABLE users; --`
	maliciousID := app.createRequest(t, app.client2, maliciousTitle, "Texto literal, sem interpretação como comando SQL.", "financeiro")
	if maliciousID == id {
		t.Fatal("IDs de solicitações diferentes coincidiram")
	}
	if _, err := app.db.Exec(`UPDATE requests SET created_at = '2026-10-02T02:59:59Z', updated_at = '2026-10-02T02:59:59Z' WHERE id = $1`, maliciousID); err != nil {
		t.Fatal(err)
	}
	status, _, list := app.do(t, app.client1, http.MethodGet, "/api/v1/requests?date_from=2026-10-01&date_to=2026-10-01&category=financeiro&status=aberto&q=100%25_d%27%C3%A1gua&page=1&page_size=1", nil, false)
	if status != http.StatusOK || !bytes.Contains(list.Items, []byte(maliciousTitle)) || !bytes.Contains(list.Pagination, []byte(`"total_items":1`)) {
		t.Fatalf("filtros/LIKE/período/paginação: status=%d items=%s pagination=%s", status, list.Items, list.Pagination)
	}
	var users int
	if err := app.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&users); err != nil || users != 2 {
		t.Fatalf("entrada textual afetou users: count=%d err=%v", users, err)
	}

	status, _, dashboard := app.do(t, app.client1, http.MethodGet, "/api/v1/dashboard", nil, false)
	if status != http.StatusOK || !bytes.Contains(dashboard.Data, []byte(`"total":2`)) || !bytes.Contains(dashboard.Data, []byte(`"abertas":2`)) {
		t.Fatalf("dashboard após reabertura/criação: status=%d data=%s", status, dashboard.Data)
	}
	status, _, _ = app.do(t, app.client1, http.MethodDelete, fmt.Sprintf("/api/v1/requests/%d", id), nil, true)
	if status != http.StatusNoContent {
		t.Fatalf("exclusão do autor após reabertura: status=%d", status)
	}
}

func TestConditionalMutationLosesRaceToStatusChange(t *testing.T) {
	app := newTestApp(t)
	app.login(t, app.client1, "colaborador1", "SenhaLocal-1")
	id := app.createRequest(t, app.client1, "Concorrência controlada", "Solicitação usada para validar atomicidade.", "infraestrutura")

	tx, err := app.db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`SELECT id FROM requests WHERE id = $1 FOR UPDATE`, id); err != nil {
		t.Fatal(err)
	}
	result := make(chan error, 1)
	go func() {
		_, updateErr := repository.NewRequest(app.db).UpdateOwnedOpen(context.Background(), id, 1, repository.RequestChanges{TitleSet: true, Title: "Não deve vencer"})
		result <- updateErr
	}()
	time.Sleep(100 * time.Millisecond)
	if _, err := tx.Exec(`UPDATE requests SET status = 'concluido', updated_at = CURRENT_TIMESTAMP WHERE id = $1`, id); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-result:
		if !errors.Is(err, repository.ErrConflict) {
			t.Fatalf("edição concorrente = %v; esperado conflito", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("edição concorrente não terminou")
	}
	var title, status string
	if err := app.db.QueryRow(`SELECT title, status FROM requests WHERE id = $1`, id).Scan(&title, &status); err != nil {
		t.Fatal(err)
	}
	if title != "Concorrência controlada" || status != domain.StatusCompleted {
		t.Fatalf("estado final inesperado: title=%q status=%q", title, status)
	}
}

func TestDatabaseFailuresKeepTheirMeaning(t *testing.T) {
	db := openTestDatabase(t)
	authService := service.NewAuth(repository.NewAuth(db), service.AuthConfig{SessionDuration: time.Hour, TokenBytes: 32})
	requestService := service.NewRequest(repository.NewRequest(db))
	dashboardService := service.NewDashboard(repository.NewDashboard(db))
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	_, authErr := authService.Login(context.Background(), service.LoginInput{Username: "colaborador1", Password: "qualquer"}, "")
	if authErr == nil || errors.Is(authErr, service.ErrInvalidCredentials) {
		t.Fatalf("falha de banco virou credencial inválida: %v", authErr)
	}
	_, requestErr := requestService.Get(context.Background(), 1, 1)
	if requestErr == nil || errors.Is(requestErr, service.ErrRequestNotFound) {
		t.Fatalf("falha de banco virou solicitação inexistente: %v", requestErr)
	}
	counts, dashboardErr := dashboardService.Get(context.Background())
	if dashboardErr == nil || counts != (domain.DashboardCounts{}) {
		t.Fatalf("falha de banco virou dashboard válido: counts=%+v err=%v", counts, dashboardErr)
	}
}

func serviceTokenHash(token string) string {
	// Replica somente a representação pública do contrato para localizar o registro de teste.
	return fmt.Sprintf("%x", sha256Sum([]byte(token)))
}

func sha256Sum(value []byte) [32]byte {
	return sha256.Sum256(value)
}
