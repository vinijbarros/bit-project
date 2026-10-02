package seed

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

const (
	minPasswordBytes = 8
	maxPasswordBytes = 72
)

var usernamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]*[a-z0-9]$`)

type Config struct {
	Enabled bool
	Users   [2]User
}

type User struct {
	Username    string
	DisplayName string
	Password    string
}

type Options struct {
	ResetPasswords bool
	Now            time.Time
}

type Result struct {
	UsersCreated     int
	UsersExisting    int
	PasswordsReset   int
	RequestsCreated  int
	RequestsExisting int
}

type demoRequest struct {
	SeedKey     string
	Title       string
	Description string
	Category    string
	Status      string
	UserIndex   int
	Age         time.Duration
}

var demoRequests = []demoRequest{
	{
		SeedKey:     "demo_ti_aberto",
		Title:       "Notebook não inicializa",
		Description: "Equipamento apresenta tela preta ao ligar e impede o início das atividades.",
		Category:    "ti",
		Status:      "aberto",
		UserIndex:   0,
		Age:         30 * 24 * time.Hour,
	},
	{
		SeedKey:     "demo_rh_em_atendimento",
		Title:       "Atualização de dados cadastrais",
		Description: "Solicitação de atualização do endereço residencial nos registros internos.",
		Category:    "rh",
		Status:      "em_atendimento",
		UserIndex:   1,
		Age:         14 * 24 * time.Hour,
	},
	{
		SeedKey:     "demo_compras_concluido",
		Title:       "Reposição de materiais de escritório",
		Description: "Reposição de canetas, blocos de anotação e pastas para a equipe de atendimento.",
		Category:    "compras",
		Status:      "concluido",
		UserIndex:   0,
		Age:         7 * 24 * time.Hour,
	},
	{
		SeedKey:     "demo_financeiro_aberto",
		Title:       "Dúvida sobre reembolso de viagem",
		Description: "Verificação dos documentos necessários para solicitar reembolso de deslocamento.",
		Category:    "financeiro",
		Status:      "aberto",
		UserIndex:   1,
		Age:         3 * 24 * time.Hour,
	},
	{
		SeedKey:     "demo_infraestrutura_em_atendimento",
		Title:       "Ajuste de iluminação da sala",
		Description: "Uma das luminárias da sala de reuniões está oscilando durante o expediente.",
		Category:    "infraestrutura",
		Status:      "em_atendimento",
		UserIndex:   0,
		Age:         24 * time.Hour,
	},
}

func LoadConfigFromEnv() (Config, error) {
	enabled := strings.EqualFold(strings.TrimSpace(os.Getenv("DEMO_SEED_ENABLED")), "true")
	if !enabled {
		return Config{}, errors.New("demo seed is disabled; set DEMO_SEED_ENABLED=true explicitly")
	}

	cfg := Config{Enabled: true}
	for i := range cfg.Users {
		prefix := fmt.Sprintf("DEMO_USER%d_", i+1)
		cfg.Users[i] = User{
			Username:    strings.TrimSpace(os.Getenv(prefix + "USERNAME")),
			DisplayName: strings.TrimSpace(os.Getenv(prefix + "DISPLAY_NAME")),
			Password:    os.Getenv(prefix + "PASSWORD"),
		}
		if err := validateUser(cfg.Users[i], prefix); err != nil {
			return Config{}, err
		}
	}
	if cfg.Users[0].Username == cfg.Users[1].Username {
		return Config{}, errors.New("demo usernames must be distinct")
	}

	return cfg, nil
}

func Run(ctx context.Context, db *sql.DB, cfg Config, options Options) (result Result, err error) {
	if !cfg.Enabled {
		return Result{}, errors.New("demo seed is disabled")
	}
	for i, user := range cfg.Users {
		if err := validateUser(user, fmt.Sprintf("DEMO_USER%d_", i+1)); err != nil {
			return Result{}, err
		}
	}
	if cfg.Users[0].Username == cfg.Users[1].Username {
		return Result{}, errors.New("demo usernames must be distinct")
	}

	now := options.Now
	if now.IsZero() {
		now = time.Now()
	}
	now = now.UTC()

	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return Result{}, fmt.Errorf("begin seed transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	userIDs := make([]int64, len(cfg.Users))
	for i, user := range cfg.Users {
		id, created, reset, ensureErr := ensureUser(ctx, tx, user, options.ResetPasswords)
		if ensureErr != nil {
			return Result{}, ensureErr
		}
		userIDs[i] = id
		if created {
			result.UsersCreated++
		} else {
			result.UsersExisting++
		}
		if reset {
			result.PasswordsReset++
		}
	}

	for _, request := range demoRequests {
		createdAt := now.Add(-request.Age)
		created, insertErr := ensureRequest(ctx, tx, request, userIDs[request.UserIndex], createdAt)
		if insertErr != nil {
			return Result{}, insertErr
		}
		if created {
			result.RequestsCreated++
		} else {
			result.RequestsExisting++
		}
	}

	if err = tx.Commit(); err != nil {
		return Result{}, fmt.Errorf("commit seed transaction: %w", err)
	}
	return result, nil
}

func validateUser(user User, prefix string) error {
	if user.Username == "" || user.DisplayName == "" || user.Password == "" {
		return fmt.Errorf("%sUSERNAME, %sDISPLAY_NAME and %sPASSWORD are required", prefix, prefix, prefix)
	}
	if user.Username != strings.ToLower(user.Username) || len(user.Username) < 3 || len(user.Username) > 50 || !usernamePattern.MatchString(user.Username) {
		return fmt.Errorf("%sUSERNAME must be 3-50 lowercase ASCII characters using letters, digits, dot, underscore, or hyphen", prefix)
	}
	if user.DisplayName != strings.TrimSpace(user.DisplayName) || utf8.RuneCountInString(user.DisplayName) > 120 {
		return fmt.Errorf("%sDISPLAY_NAME must contain 1-120 trimmed characters", prefix)
	}
	passwordBytes := len([]byte(user.Password))
	if passwordBytes < minPasswordBytes || passwordBytes > maxPasswordBytes {
		return fmt.Errorf("%sPASSWORD must contain %d-%d bytes", prefix, minPasswordBytes, maxPasswordBytes)
	}
	return nil
}

func ensureUser(ctx context.Context, tx *sql.Tx, user User, resetPassword bool) (id int64, created, reset bool, err error) {
	var passwordHash string
	err = tx.QueryRowContext(ctx, `
		SELECT id, password_hash
		FROM users
		WHERE username = $1
		FOR UPDATE
	`, user.Username).Scan(&id, &passwordHash)
	if errors.Is(err, sql.ErrNoRows) {
		hash, hashErr := generateVerifiedHash(user.Password)
		if hashErr != nil {
			return 0, false, false, fmt.Errorf("hash password for demo user %q: %w", user.Username, hashErr)
		}
		if insertErr := tx.QueryRowContext(ctx, `
			INSERT INTO users (username, display_name, password_hash)
			VALUES ($1, $2, $3)
			RETURNING id
		`, user.Username, user.DisplayName, hash).Scan(&id); insertErr != nil {
			return 0, false, false, fmt.Errorf("create demo user %q: %w", user.Username, insertErr)
		}
		return id, true, false, nil
	}
	if err != nil {
		return 0, false, false, fmt.Errorf("find demo user %q: %w", user.Username, err)
	}

	if bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(user.Password)) == nil {
		return id, false, false, nil
	}
	if !resetPassword {
		return 0, false, false, fmt.Errorf("demo user %q already exists with a different password; use --reset-passwords for an explicit reset", user.Username)
	}

	hash, hashErr := generateVerifiedHash(user.Password)
	if hashErr != nil {
		return 0, false, false, fmt.Errorf("hash password for demo user %q: %w", user.Username, hashErr)
	}
	if _, updateErr := tx.ExecContext(ctx, `UPDATE users SET password_hash = $1 WHERE id = $2`, hash, id); updateErr != nil {
		return 0, false, false, fmt.Errorf("reset password for demo user %q: %w", user.Username, updateErr)
	}
	return id, false, true, nil
}

func generateVerifiedHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	if err := bcrypt.CompareHashAndPassword(hash, []byte(password)); err != nil {
		return "", fmt.Errorf("verify generated bcrypt hash: %w", err)
	}
	return string(hash), nil
}

func ensureRequest(ctx context.Context, tx *sql.Tx, request demoRequest, requesterID int64, createdAt time.Time) (bool, error) {
	result, err := tx.ExecContext(ctx, `
		INSERT INTO requests (
			title, description, category, status, requester_id, created_at, updated_at, seed_key
		)
		VALUES ($1, $2, $3, $4, $5, $6, $6, $7)
		ON CONFLICT (seed_key) WHERE seed_key IS NOT NULL DO NOTHING
	`, request.Title, request.Description, request.Category, request.Status, requesterID, createdAt, request.SeedKey)
	if err != nil {
		return false, fmt.Errorf("create demo request %q: %w", request.SeedKey, err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("read result for demo request %q: %w", request.SeedKey, err)
	}
	if rows == 1 {
		return true, nil
	}

	var existingRequesterID int64
	if err := tx.QueryRowContext(ctx, `SELECT requester_id FROM requests WHERE seed_key = $1`, request.SeedKey).Scan(&existingRequesterID); err != nil {
		return false, fmt.Errorf("verify existing demo request %q: %w", request.SeedKey, err)
	}
	if existingRequesterID != requesterID {
		return false, fmt.Errorf("demo request %q belongs to an unexpected requester; no data was changed", request.SeedKey)
	}
	return false, nil
}
