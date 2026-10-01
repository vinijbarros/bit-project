package httpapi

import (
	"context"
	"net/http"
	"time"
)

func health(response http.ResponseWriter, _ *http.Request) {
	writeJSON(response, http.StatusOK, map[string]string{"status": "ok"})
}

func ready(db databasePinger, timeout time.Duration) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		ctx, cancel := context.WithTimeout(request.Context(), timeout)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			writeError(response, request, http.StatusServiceUnavailable, "database_unavailable", "Banco de dados indisponível.", nil)
			return
		}

		writeJSON(response, http.StatusOK, map[string]string{"status": "ready"})
	}
}
