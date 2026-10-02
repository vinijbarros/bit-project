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

type stubAuthService struct {
	login        func(context.Context, service.LoginInput, string) (service.LoginResult, error)
	authenticate func(context.Context, string) (domain.User, error)
	logout       func(context.Context, string) error
}

func (stub stubAuthService) Login(ctx context.Context, input service.LoginInput, previous string) (service.LoginResult, error) {
	if stub.login == nil {
		return service.LoginResult{}, service.ErrInvalidCredentials
	}
	return stub.login(ctx, input, previous)
}

func (stub stubAuthService) Authenticate(ctx context.Context, token string) (domain.User, error) {
	if stub.authenticate == nil {
		return domain.User{}, service.ErrUnauthenticated
	}
	return stub.authenticate(ctx, token)
}

func (stub stubAuthService) Logout(ctx context.Context, token string) error {
	if stub.logout == nil {
		return nil
	}
	return stub.logout(ctx, token)
}

func authTestRouter(auth authService, maxAttempts int) http.Handler {
	return NewRouter(stubPinger{}, RouterConfig{
		ReadinessTimeout: time.Second,
		Logger:           slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil)),
		AuthService:      auth,
		TrustedOrigins:   []string{"http://127.0.0.1:5173"},
		SessionCookie:    SessionCookieConfig("portal_session", false, 8*time.Hour),
		LoginRateLimit:   LoginRateLimitConfig(maxAttempts, 15*time.Minute, 100),
	})
}

func TestLoginSetsSecureSessionShapeWithoutExposingToken(t *testing.T) {
	expiresAt := time.Date(2026, 10, 2, 1, 0, 0, 0, time.UTC)
	auth := stubAuthService{login: func(_ context.Context, input service.LoginInput, previous string) (service.LoginResult, error) {
		if input.Username != "colaborador1" || input.Password != "senha-correta" || previous != "old-token" {
			t.Fatalf("unexpected login input: %+v previous=%q", input, previous)
		}
		return service.LoginResult{
			User:      domain.User{ID: 7, Username: "colaborador1", DisplayName: "Colaborador 1"},
			Token:     "new-secret-token",
			ExpiresAt: expiresAt,
		}, nil
	}}
	router := authTestRouter(auth, 5)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"username":"colaborador1","password":"senha-correta"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "http://127.0.0.1:5173")
	request.AddCookie(&http.Cookie{Name: "portal_session", Value: "old-token"})
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %d, want 1", len(cookies))
	}
	cookie := cookies[0]
	if cookie.Name != "portal_session" || cookie.Value != "new-secret-token" || !cookie.HttpOnly || cookie.SameSite != http.SameSiteLaxMode || cookie.Path != "/" {
		t.Fatalf("unexpected cookie: %+v", cookie)
	}
	body := response.Body.String()
	if strings.Contains(body, "new-secret-token") || strings.Contains(body, "senha-correta") || strings.Contains(body, "password_hash") {
		t.Fatalf("body leaked secret: %s", body)
	}
	if !strings.Contains(body, `"username":"colaborador1"`) {
		t.Fatalf("body = %s", body)
	}
}

func TestLoginUsesGenericInvalidCredentialError(t *testing.T) {
	auth := stubAuthService{login: func(context.Context, service.LoginInput, string) (service.LoginResult, error) {
		return service.LoginResult{}, service.ErrInvalidCredentials
	}}
	router := authTestRouter(auth, 5)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"username":"inexistente","password":"segredo"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "http://127.0.0.1:5173")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized || !strings.Contains(response.Body.String(), `"code":"invalid_credentials"`) {
		t.Fatalf("status/body = %d %s", response.Code, response.Body.String())
	}
	if strings.Contains(response.Body.String(), "inexistente") || strings.Contains(response.Body.String(), "segredo") {
		t.Fatalf("body leaked credential: %s", response.Body.String())
	}
}

func TestLoginRejectsUntrustedOriginBeforeService(t *testing.T) {
	called := false
	auth := stubAuthService{login: func(context.Context, service.LoginInput, string) (service.LoginResult, error) {
		called = true
		return service.LoginResult{}, nil
	}}
	router := authTestRouter(auth, 5)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"username":"colaborador1","password":"senha"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "https://attacker.example")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden || called {
		t.Fatalf("status=%d called=%v", response.Code, called)
	}
}

func TestMeRequiresAndReturnsAuthenticatedIdentity(t *testing.T) {
	auth := stubAuthService{authenticate: func(_ context.Context, token string) (domain.User, error) {
		if token != "valid" {
			return domain.User{}, service.ErrUnauthenticated
		}
		return domain.User{ID: 9, Username: "colaborador2", DisplayName: "Colaborador 2"}, nil
	}}
	router := authTestRouter(auth, 5)

	anonymous := httptest.NewRecorder()
	router.ServeHTTP(anonymous, httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil))
	if anonymous.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous status = %d", anonymous.Code)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	request.AddCookie(&http.Cookie{Name: "portal_session", Value: "valid"})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"id":9`) {
		t.Fatalf("status/body = %d %s", response.Code, response.Body.String())
	}
}

func TestLogoutIsIdempotentAndClearsCookie(t *testing.T) {
	var tokens []string
	auth := stubAuthService{logout: func(_ context.Context, token string) error {
		tokens = append(tokens, token)
		return nil
	}}
	router := authTestRouter(auth, 5)

	for _, token := range []string{"valid-token", ""} {
		request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
		request.Header.Set("Origin", "http://127.0.0.1:5173")
		if token != "" {
			request.AddCookie(&http.Cookie{Name: "portal_session", Value: token})
		}
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusNoContent {
			t.Fatalf("status = %d", response.Code)
		}
		cookies := response.Result().Cookies()
		if len(cookies) != 1 || cookies[0].MaxAge != -1 || cookies[0].Path != "/" {
			t.Fatalf("clear cookie = %+v", cookies)
		}
	}
	if len(tokens) != 2 || tokens[0] != "valid-token" || tokens[1] != "" {
		t.Fatalf("logout tokens = %v", tokens)
	}
}

func TestAuthenticationInfrastructureErrorIsNotUnauthorizedOrLeaked(t *testing.T) {
	auth := stubAuthService{authenticate: func(context.Context, string) (domain.User, error) {
		return domain.User{}, errors.New("password=secret SELECT * FROM sessions")
	}}
	router := authTestRouter(auth, 5)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	request.AddCookie(&http.Cookie{Name: "portal_session", Value: "token"})
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError || strings.Contains(response.Body.String(), "secret") || strings.Contains(response.Body.String(), "SELECT") {
		t.Fatalf("status/body = %d %s", response.Code, response.Body.String())
	}
}

func TestLoginRateLimitIsScopedAndIgnoresForwardedIP(t *testing.T) {
	auth := stubAuthService{login: func(context.Context, service.LoginInput, string) (service.LoginResult, error) {
		return service.LoginResult{}, service.ErrInvalidCredentials
	}}
	router := authTestRouter(auth, 2)
	login := func(username, forwarded string) *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"username":"`+username+`","password":"errada"}`))
		request.RemoteAddr = "192.0.2.10:4321"
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Origin", "http://127.0.0.1:5173")
		request.Header.Set("X-Forwarded-For", forwarded)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		return response
	}

	if login("colaborador1", "198.51.100.1").Code != 401 || login("colaborador1", "198.51.100.2").Code != 401 {
		t.Fatal("first two failures should reach authentication")
	}
	limited := login("colaborador1", "203.0.113.9")
	if limited.Code != http.StatusTooManyRequests || limited.Header().Get("Retry-After") == "" {
		t.Fatalf("limited = %d headers=%v", limited.Code, limited.Header())
	}
	if got := login("colaborador2", "203.0.113.9").Code; got != http.StatusUnauthorized {
		t.Fatalf("different username status = %d, want 401", got)
	}
}

func TestLoginLimiterCapsStoredEntries(t *testing.T) {
	limiter := newLoginRateLimiter(loginRateLimitConfig{MaxAttempts: 2, Window: time.Hour, MaxEntries: 2})
	limiter.failure("one")
	limiter.failure("two")
	limiter.failure("three")
	if len(limiter.entries) != 2 {
		t.Fatalf("entries = %d, want 2", len(limiter.entries))
	}
}
