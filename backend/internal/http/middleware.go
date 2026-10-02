package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type contextKey string

const requestIDContextKey contextKey = "request_id"

var requestIDPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		id := request.Header.Get("X-Request-ID")
		if !requestIDPattern.MatchString(id) {
			id = newRequestID()
		}
		response.Header().Set("X-Request-ID", id)
		ctx := context.WithValue(request.Context(), requestIDContextKey, id)
		next.ServeHTTP(response, request.WithContext(ctx))
	})
}

func newRequestID() string {
	data := make([]byte, 16)
	if _, err := rand.Read(data); err != nil {
		return time.Now().UTC().Format("20060102T150405.000000000")
	}
	return hex.EncodeToString(data)
}

func requestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDContextKey).(string)
	return id
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (recorder *statusRecorder) Unwrap() http.ResponseWriter {
	return recorder.ResponseWriter
}

func (recorder *statusRecorder) WriteHeader(status int) {
	if recorder.status != 0 {
		return
	}
	recorder.status = status
	recorder.ResponseWriter.WriteHeader(status)
}

func (recorder *statusRecorder) Write(data []byte) (int, error) {
	if recorder.status == 0 {
		recorder.WriteHeader(http.StatusOK)
	}
	return recorder.ResponseWriter.Write(data)
}

func requestLog(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		startedAt := time.Now()
		recorder := &statusRecorder{ResponseWriter: response}
		next.ServeHTTP(recorder, request)
		status := recorder.status
		if status == 0 {
			status = http.StatusOK
		}
		logger.Info("http request completed",
			"request_id", requestIDFromContext(request.Context()),
			"method", request.Method,
			"route", routeName(request),
			"status", status,
			"duration_ms", time.Since(startedAt).Milliseconds(),
		)
	})
}

func recoverPanics(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error("panic recovered",
					"request_id", requestIDFromContext(request.Context()),
					"method", request.Method,
					"route", routeName(request),
				)
				writeError(response, request, http.StatusInternalServerError, "internal_error", "Erro interno inesperado.", nil)
			}
		}()
		next.ServeHTTP(response, request)
	})
}

func routeName(request *http.Request) string {
	if request.Method == http.MethodGet && request.URL.Path == "/healthz" {
		return "GET /healthz"
	}
	if request.Method == http.MethodGet && request.URL.Path == "/readyz" {
		return "GET /readyz"
	}
	if request.Method == http.MethodPost && request.URL.Path == "/api/v1/auth/login" {
		return "POST /api/v1/auth/login"
	}
	if request.Method == http.MethodPost && request.URL.Path == "/api/v1/auth/logout" {
		return "POST /api/v1/auth/logout"
	}
	if request.Method == http.MethodGet && request.URL.Path == "/api/v1/auth/me" {
		return "GET /api/v1/auth/me"
	}
	if request.Method == http.MethodGet && request.URL.Path == "/api/v1/metadata" {
		return "GET /api/v1/metadata"
	}
	if request.Method == http.MethodGet && request.URL.Path == "/api/v1/dashboard" {
		return "GET /api/v1/dashboard"
	}
	if request.Method == http.MethodPost && request.URL.Path == "/api/v1/requests" {
		return "POST /api/v1/requests"
	}
	if request.Method == http.MethodGet && request.URL.Path == "/api/v1/requests" {
		return "GET /api/v1/requests"
	}
	requestSuffix := strings.TrimPrefix(request.URL.Path, "/api/v1/requests/")
	if request.Method == http.MethodPatch && strings.HasSuffix(requestSuffix, "/status") {
		requestID := strings.TrimSuffix(requestSuffix, "/status")
		if requestID != "" && !strings.Contains(requestID, "/") {
			return "PATCH /api/v1/requests/{id}/status"
		}
	}
	if requestSuffix != request.URL.Path && requestSuffix != "" && !strings.Contains(requestSuffix, "/") {
		switch request.Method {
		case http.MethodGet:
			return "GET /api/v1/requests/{id}"
		case http.MethodPatch:
			return "PATCH /api/v1/requests/{id}"
		case http.MethodDelete:
			return "DELETE /api/v1/requests/{id}"
		}
	}
	if strings.HasPrefix(request.URL.Path, "/api/") {
		return "/api/*"
	}
	return "unmatched"
}
