package httpapi

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"portal-solicitacoes/internal/domain"
	"portal-solicitacoes/internal/service"
)

type stubPinger struct {
	err error
}

func newTestRouter(db databasePinger) http.Handler {
	return NewRouter(db, RouterConfig{
		ReadinessTimeout: time.Second,
		Logger:           slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil)),
		AuthService: stubAuthService{
			login: func(context.Context, service.LoginInput, string) (service.LoginResult, error) {
				return service.LoginResult{}, service.ErrInvalidCredentials
			},
			authenticate: func(_ context.Context, token string) (domain.User, error) {
				if token == "valid-test-token" {
					return domain.User{ID: 1, Username: "teste", DisplayName: "Teste"}, nil
				}
				return domain.User{}, service.ErrUnauthenticated
			},
		},
		TrustedOrigins: []string{"http://127.0.0.1:5173"},
		SessionCookie:  SessionCookieConfig("portal_session", false, 8*time.Hour),
		LoginRateLimit: LoginRateLimitConfig(5, 15*time.Minute, 100),
	})
}

func (p stubPinger) PingContext(context.Context) error {
	return p.err
}

func TestHealthIsIndependentFromDatabase(t *testing.T) {
	router := newTestRouter(stubPinger{err: errors.New("offline")})
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if !strings.Contains(response.Body.String(), `"status":"ok"`) {
		t.Fatalf("body = %q, want healthy status", response.Body.String())
	}
}

func TestReadyWhenDatabaseIsAvailable(t *testing.T) {
	router := newTestRouter(stubPinger{})
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
}

func TestReadyWhenDatabaseIsUnavailable(t *testing.T) {
	router := newTestRouter(stubPinger{err: errors.New("offline")})
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
	body := response.Body.String()
	if !strings.Contains(body, `"code":"database_unavailable"`) {
		t.Fatalf("body = %q, want safe database error", body)
	}
	if strings.Contains(body, "offline") {
		t.Fatalf("body = %q, must not expose internal error", body)
	}
}

func TestProbeRejectsUnsupportedMethod(t *testing.T) {
	router := newTestRouter(stubPinger{})
	request := httptest.NewRequest(http.MethodPost, "/healthz", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
}

func TestMetadataReturnsStableLabelsToAuthenticatedUser(t *testing.T) {
	router := newTestRouter(stubPinger{})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/metadata", nil)
	request.AddCookie(&http.Cookie{Name: "portal_session", Value: "valid-test-token"})
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	body := response.Body.String()
	for _, expected := range []string{
		`"value":"ti","label":"TI"`,
		`"value":"infraestrutura","label":"Infraestrutura"`,
		`"value":"aberto","label":"Aberto"`,
		`"value":"em_atendimento","label":"Em Atendimento"`,
		`"value":"concluido","label":"Concluído"`,
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("body = %q, missing %q", body, expected)
		}
	}
}

func TestMetadataRequiresAuthentication(t *testing.T) {
	router := newTestRouter(stubPinger{})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/metadata", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized || !strings.Contains(response.Body.String(), `"code":"authentication_required"`) {
		t.Fatalf("status/body = %d %s", response.Code, response.Body.String())
	}
}
