package httpapi

import (
	"context"
	"encoding/json"
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
			writeJSON(response, http.StatusServiceUnavailable, map[string]any{
				"status": "unavailable",
				"error": map[string]string{
					"code":    "database_unavailable",
					"message": "database is unavailable",
				},
			})
			return
		}

		writeJSON(response, http.StatusOK, map[string]string{"status": "ready"})
	}
}

func writeJSON(response http.ResponseWriter, status int, payload any) {
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	response.Header().Set("Cache-Control", "no-store")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(payload)
}
