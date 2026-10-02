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
	RequestService   requestService
	DashboardService dashboardService
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
	requests := &requestHandler{service: cfg.RequestService, logger: logger}
	dashboard := &dashboardHandler{service: cfg.DashboardService, logger: logger}
	originGuard := func(next http.Handler) http.Handler {
		return requireTrustedOrigin(trustedOriginSet(cfg.TrustedOrigins), next)
	}
	mux.Handle("POST /api/v1/auth/login", originGuard(http.HandlerFunc(auth.login)))
	mux.Handle("POST /api/v1/auth/logout", originGuard(http.HandlerFunc(auth.logout)))
	mux.Handle("GET /api/v1/auth/me", auth.requireAuthentication(http.HandlerFunc(auth.me)))
	mux.Handle("GET /api/v1/metadata", auth.requireAuthentication(http.HandlerFunc(metadata)))
	mux.Handle("POST /api/v1/requests", auth.requireAuthentication(originGuard(http.HandlerFunc(requests.create))))
	mux.Handle("GET /api/v1/requests", auth.requireAuthentication(http.HandlerFunc(requests.list)))
	mux.Handle("GET /api/v1/requests/{id}", auth.requireAuthentication(http.HandlerFunc(requests.get)))
	mux.Handle("PATCH /api/v1/requests/{id}", auth.requireAuthentication(originGuard(http.HandlerFunc(requests.update))))
	mux.Handle("DELETE /api/v1/requests/{id}", auth.requireAuthentication(originGuard(http.HandlerFunc(requests.delete))))
	mux.Handle("PATCH /api/v1/requests/{id}/status", auth.requireAuthentication(originGuard(http.HandlerFunc(requests.updateStatus))))
	mux.Handle("GET /api/v1/dashboard", auth.requireAuthentication(http.HandlerFunc(dashboard.get)))
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
