package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"portal-solicitacoes/internal/domain"
	"portal-solicitacoes/internal/repository"
)

type authRepositoryStub struct {
	credentials  domain.UserCredentials
	findUserErr  error
	rotated      domain.Session
	previousHash string
	sessionUser  domain.User
	sessionErr   error
	revokedHash  string
	revokeErr    error
}

func (stub *authRepositoryStub) FindUserCredentials(context.Context, string) (domain.UserCredentials, error) {
	return stub.credentials, stub.findUserErr
}

func (stub *authRepositoryStub) RotateSession(_ context.Context, previous string, session domain.Session) error {
	stub.previousHash = previous
	stub.rotated = session
	return nil
}

func (stub *authRepositoryStub) FindUserBySession(context.Context, string) (domain.User, error) {
	return stub.sessionUser, stub.sessionErr
}

func (stub *authRepositoryStub) RevokeSession(_ context.Context, tokenHash string) error {
	stub.revokedHash = tokenHash
	return stub.revokeErr
}

func newTestAuth(repositoryStub *authRepositoryStub) *Auth {
	randomCall := byte(0)
	return NewAuth(repositoryStub, AuthConfig{
		SessionDuration: 8 * time.Hour,
		TokenBytes:      32,
		Now: func() time.Time {
			return time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
		},
		Random: func(data []byte) (int, error) {
			randomCall++
			for index := range data {
				data[index] = byte(index + 1 + int(randomCall))
			}
			return len(data), nil
		},
	})
}

func TestLoginCreatesHashedPersistentSessionAndRotatesPrevious(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("senha-correta"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	repositoryStub := &authRepositoryStub{credentials: domain.UserCredentials{
		User:         domain.User{ID: 3, Username: "colaborador1", DisplayName: "Colaborador 1"},
		PasswordHash: string(hash),
	}}
	auth := newTestAuth(repositoryStub)
	oldTokenResult, err := auth.generateToken()
	if err != nil {
		t.Fatal(err)
	}

	result, err := auth.Login(context.Background(), LoginInput{Username: " Colaborador1 ", Password: "senha-correta"}, oldTokenResult)
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if result.Token == "" || result.Token == repositoryStub.rotated.TokenHash {
		t.Fatal("plain token must be returned only to caller and hash stored")
	}
	if len(repositoryStub.rotated.TokenHash) != 64 || repositoryStub.rotated.TokenHash != hashToken(result.Token) {
		t.Fatalf("stored hash = %q", repositoryStub.rotated.TokenHash)
	}
	if repositoryStub.previousHash != hashToken(oldTokenResult) {
		t.Fatalf("previous hash = %q", repositoryStub.previousHash)
	}
	if repositoryStub.rotated.ExpiresAt.Sub(repositoryStub.rotated.CreatedAt) != 8*time.Hour {
		t.Fatalf("session duration = %s", repositoryStub.rotated.ExpiresAt.Sub(repositoryStub.rotated.CreatedAt))
	}
}

func TestLoginRejectsInvalidCredentialsGenerically(t *testing.T) {
	wrongHash, _ := bcrypt.GenerateFromPassword([]byte("outra-senha"), bcrypt.MinCost)
	tests := []authRepositoryStub{
		{findUserErr: repository.ErrNotFound},
		{credentials: domain.UserCredentials{PasswordHash: string(wrongHash)}},
	}
	for index := range tests {
		auth := newTestAuth(&tests[index])
		_, err := auth.Login(context.Background(), LoginInput{Username: "colaborador1", Password: "senha-correta"}, "")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("case %d error = %v", index, err)
		}
	}
}

func TestLoginValidatesUsernameAndBcryptByteLimit(t *testing.T) {
	auth := newTestAuth(&authRepositoryStub{})
	_, err := auth.Login(context.Background(), LoginInput{Username: "inválido", Password: strings.Repeat("á", 37)}, "")
	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("error = %v, want ValidationError", err)
	}
	if len(validationErr.Fields["username"]) == 0 || len(validationErr.Fields["password"]) == 0 {
		t.Fatalf("fields = %v", validationErr.Fields)
	}
}

func TestAuthenticateHandlesMalformedMissingAndInfrastructureErrors(t *testing.T) {
	auth := newTestAuth(&authRepositoryStub{})
	if _, err := auth.Authenticate(context.Background(), "malformed"); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("malformed error = %v", err)
	}

	validToken, _ := auth.generateToken()
	notFound := newTestAuth(&authRepositoryStub{sessionErr: repository.ErrNotFound})
	if _, err := notFound.Authenticate(context.Background(), validToken); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("not found error = %v", err)
	}

	infrastructureErr := errors.New("database unavailable")
	unavailable := newTestAuth(&authRepositoryStub{sessionErr: infrastructureErr})
	if _, err := unavailable.Authenticate(context.Background(), validToken); !errors.Is(err, infrastructureErr) {
		t.Fatalf("infrastructure error = %v", err)
	}
}

func TestLogoutHashesValidTokenAndTreatsMissingAsSuccess(t *testing.T) {
	repositoryStub := &authRepositoryStub{}
	auth := newTestAuth(repositoryStub)
	if err := auth.Logout(context.Background(), ""); err != nil || repositoryStub.revokedHash != "" {
		t.Fatalf("empty logout err=%v hash=%q", err, repositoryStub.revokedHash)
	}
	token, _ := auth.generateToken()
	if err := auth.Logout(context.Background(), token); err != nil {
		t.Fatal(err)
	}
	if repositoryStub.revokedHash != hashToken(token) {
		t.Fatalf("revoked hash = %q", repositoryStub.revokedHash)
	}
}
