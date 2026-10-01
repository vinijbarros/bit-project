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
)

func TestRouterAddsRequestIDAndStructuredLog(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	router := NewRouter(stubPinger{}, RouterConfig{ReadinessTimeout: time.Second, Logger: logger})
	request := httptest.NewRequest(http.MethodGet, "/healthz?token=must-not-be-logged", nil)
	request.Header.Set("X-Request-ID", "request-123")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if got := response.Header().Get("X-Request-ID"); got != "request-123" {
		t.Fatalf("X-Request-ID = %q, want request-123", got)
	}
	logOutput := logs.String()
	for _, expected := range []string{`"request_id":"request-123"`, `"method":"GET"`, `"route":"GET /healthz"`, `"status":200`, `"duration_ms":`} {
		if !strings.Contains(logOutput, expected) {
			t.Fatalf("log = %q, missing %q", logOutput, expected)
		}
	}
	if strings.Contains(logOutput, "must-not-be-logged") {
		t.Fatalf("log = %q, must not include query values", logOutput)
	}
}

func TestRouterReplacesUnsafeRequestID(t *testing.T) {
	router := newTestRouter(stubPinger{})
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	request.Header.Set("X-Request-ID", "invalid id with spaces")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	got := response.Header().Get("X-Request-ID")
	if got == "" || got == "invalid id with spaces" || !requestIDPattern.MatchString(got) {
		t.Fatalf("generated X-Request-ID = %q", got)
	}
}

func TestPanicReturnsSafeInternalError(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	handler := requestID(requestLog(logger, recoverPanics(logger, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("database password must never leak")
	}))))
	request := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", response.Code)
	}
	if strings.Contains(response.Body.String(), "database password") {
		t.Fatalf("body = %q, leaked panic", response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"code":"internal_error"`) {
		t.Fatalf("body = %q, want internal_error", response.Body.String())
	}
}

func TestWriteInternalErrorDoesNotExposeCause(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	request := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	request = request.WithContext(context.WithValue(request.Context(), requestIDContextKey, "request-safe"))
	response := httptest.NewRecorder()

	writeInternalError(logger, response, request, errors.New("SELECT secret FROM users"))

	if strings.Contains(response.Body.String(), "SELECT") {
		t.Fatalf("body = %q, leaked internal cause", response.Body.String())
	}
	if !strings.Contains(logs.String(), "request-safe") {
		t.Fatalf("log = %q, missing request ID", logs.String())
	}
	if strings.Contains(logs.String(), "SELECT") {
		t.Fatalf("log = %q, leaked internal error detail", logs.String())
	}
}

func TestRequireTrustedOrigin(t *testing.T) {
	trusted := trustedOriginSet([]string{"http://127.0.0.1:5173"})
	next := http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusNoContent)
	})
	handler := requireTrustedOrigin(trusted, next)

	tests := []struct {
		name   string
		method string
		origin string
		want   int
	}{
		{name: "trusted post", method: http.MethodPost, origin: "http://127.0.0.1:5173", want: http.StatusNoContent},
		{name: "missing post", method: http.MethodPost, want: http.StatusForbidden},
		{name: "untrusted patch", method: http.MethodPatch, origin: "https://attacker.example", want: http.StatusForbidden},
		{name: "get does not require origin", method: http.MethodGet, want: http.StatusNoContent},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, "/api/v1/test", nil)
			request.Header.Set("Origin", test.origin)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != test.want {
				t.Fatalf("status = %d, want %d", response.Code, test.want)
			}
		})
	}
}
