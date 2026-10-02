package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"portal-solicitacoes/internal/domain"
	"portal-solicitacoes/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthenticated    = errors.New("unauthenticated")
)

var (
	usernamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]*[a-z0-9]$`)
	dummyBcryptHash = func() []byte {
		hash, err := bcrypt.GenerateFromPassword([]byte("dummy-password-for-timing"), bcrypt.DefaultCost)
		if err != nil {
			panic("initialize bcrypt comparison")
		}
		return hash
	}()
)

type AuthRepository interface {
	FindUserCredentials(context.Context, string) (domain.UserCredentials, error)
	RotateSession(context.Context, string, domain.Session) error
	FindUserBySession(context.Context, string) (domain.User, error)
	RevokeSession(context.Context, string) error
}

type AuthConfig struct {
	SessionDuration time.Duration
	TokenBytes      int
	Now             func() time.Time
	Random          func([]byte) (int, error)
}

type Auth struct {
	repository      AuthRepository
	sessionDuration time.Duration
	tokenBytes      int
	now             func() time.Time
	random          func([]byte) (int, error)
}

type LoginInput struct {
	Username string
	Password string
}

type LoginResult struct {
	User      domain.User
	Token     string
	ExpiresAt time.Time
}

type ValidationError struct {
	Fields map[string][]string
}

func (err *ValidationError) Error() string {
	return "validation failed"
}

func NewAuth(repository AuthRepository, cfg AuthConfig) *Auth {
	now := cfg.Now
	if now == nil {
		now = time.Now
	}
	random := cfg.Random
	if random == nil {
		random = rand.Read
	}
	return &Auth{
		repository:      repository,
		sessionDuration: cfg.SessionDuration,
		tokenBytes:      cfg.TokenBytes,
		now:             now,
		random:          random,
	}
}

func (service *Auth) Login(ctx context.Context, input LoginInput, previousToken string) (LoginResult, error) {
	username, validationErr := validateLogin(input)
	if validationErr != nil {
		return LoginResult{}, validationErr
	}

	credentials, err := service.repository.FindUserCredentials(ctx, username)
	if errors.Is(err, repository.ErrNotFound) {
		_ = bcrypt.CompareHashAndPassword(dummyBcryptHash, []byte(input.Password))
		return LoginResult{}, ErrInvalidCredentials
	}
	if err != nil {
		return LoginResult{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(credentials.PasswordHash), []byte(input.Password)) != nil {
		return LoginResult{}, ErrInvalidCredentials
	}

	var token string
	for attempt := 0; attempt < 3; attempt++ {
		token, err = service.generateToken()
		if err != nil {
			return LoginResult{}, err
		}
		if token != previousToken {
			break
		}
	}
	if token == previousToken {
		return LoginResult{}, errors.New("could not rotate session token")
	}
	now := service.now().UTC()
	expiresAt := now.Add(service.sessionDuration)
	previousHash := ""
	if service.validToken(previousToken) {
		previousHash = hashToken(previousToken)
	}
	if err := service.repository.RotateSession(ctx, previousHash, domain.Session{
		TokenHash: hashToken(token),
		UserID:    credentials.ID,
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}); err != nil {
		return LoginResult{}, err
	}

	return LoginResult{
		User:      credentials.User,
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

func (service *Auth) Authenticate(ctx context.Context, token string) (domain.User, error) {
	if !service.validToken(token) {
		return domain.User{}, ErrUnauthenticated
	}
	user, err := service.repository.FindUserBySession(ctx, hashToken(token))
	if errors.Is(err, repository.ErrNotFound) {
		return domain.User{}, ErrUnauthenticated
	}
	if err != nil {
		return domain.User{}, err
	}
	return user, nil
}

func (service *Auth) Logout(ctx context.Context, token string) error {
	if !service.validToken(token) {
		return nil
	}
	return service.repository.RevokeSession(ctx, hashToken(token))
}

func validateLogin(input LoginInput) (string, error) {
	username := strings.ToLower(strings.TrimSpace(input.Username))
	fields := make(map[string][]string)
	if len(username) < 3 || len(username) > 50 || !usernamePattern.MatchString(username) {
		fields["username"] = []string{"Informe um usuário válido com 3 a 50 caracteres."}
	}
	passwordBytes := len([]byte(input.Password))
	if passwordBytes == 0 || passwordBytes > 72 {
		fields["password"] = []string{"Informe uma senha com no máximo 72 bytes."}
	}
	if len(fields) > 0 {
		return "", &ValidationError{Fields: fields}
	}
	return username, nil
}

func (service *Auth) generateToken() (string, error) {
	data := make([]byte, service.tokenBytes)
	read, err := service.random(data)
	if err != nil {
		return "", err
	}
	if read != len(data) {
		return "", io.ErrUnexpectedEOF
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}

func (service *Auth) validToken(token string) bool {
	decoded, err := base64.RawURLEncoding.DecodeString(token)
	return err == nil && len(decoded) == service.tokenBytes
}

func hashToken(token string) string {
	digest := sha256.Sum256([]byte(token))
	return hex.EncodeToString(digest[:])
}
