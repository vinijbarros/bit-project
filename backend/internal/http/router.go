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
}

func NewRouter(db databasePinger, cfg RouterConfig) http.Handler {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(discardWriter{}, nil))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", health)
	mux.HandleFunc("GET /readyz", ready(db, cfg.ReadinessTimeout))
	mux.HandleFunc("/api/", apiNotFound)

	return requestID(requestLog(logger, recoverPanics(logger, mux)))
}

func apiNotFound(response http.ResponseWriter, request *http.Request) {
	writeError(response, request, http.StatusNotFound, "route_not_found", "Rota não encontrada.", nil)
}

type discardWriter struct{}

func (discardWriter) Write(data []byte) (int, error) {
	return len(data), nil
}
