package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

type databasePinger interface {
	PingContext(context.Context) error
}

type RouterConfig struct {
	ReadinessTimeout time.Duration
	Logger           *slog.Logger
	AuthService      authService
	TrustedOrigins   []string
	SessionCookie    sessionCookieConfig
	LoginRateLimit   loginRateLimitConfig
}

func NewRouter(db databasePinger, cfg RouterConfig) http.Handler {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(discardWriter{}, nil))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", health)
	mux.HandleFunc("GET /readyz", ready(db, cfg.ReadinessTimeout))

	auth := &authHandler{
		service: cfg.AuthService,
		cookie:  cfg.SessionCookie,
		limiter: newLoginRateLimiter(cfg.LoginRateLimit),
		logger:  logger,
	}
	originGuard := func(next http.Handler) http.Handler {
		return requireTrustedOrigin(trustedOriginSet(cfg.TrustedOrigins), next)
	}
	mux.Handle("POST /api/v1/auth/login", originGuard(http.HandlerFunc(auth.login)))
	mux.Handle("POST /api/v1/auth/logout", originGuard(http.HandlerFunc(auth.logout)))
	mux.Handle("GET /api/v1/auth/me", auth.requireAuthentication(http.HandlerFunc(auth.me)))
	mux.Handle("/api/", auth.requireAuthentication(http.HandlerFunc(apiNotFound)))

	return requestID(requestLog(logger, recoverPanics(logger, mux)))
}

func apiNotFound(response http.ResponseWriter, request *http.Request) {
	writeError(response, request, http.StatusNotFound, "route_not_found", "Rota não encontrada.", nil)
}

type discardWriter struct{}

func (discardWriter) Write(data []byte) (int, error) {
	return len(data), nil
}
