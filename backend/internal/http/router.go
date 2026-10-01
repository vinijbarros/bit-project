package httpapi

import (
	"context"
	"net/http"
	"time"
)

type databasePinger interface {
	PingContext(context.Context) error
}

func NewRouter(db databasePinger, readinessTimeout time.Duration) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", health)
	mux.HandleFunc("GET /readyz", ready(db, readinessTimeout))
	return mux
}
