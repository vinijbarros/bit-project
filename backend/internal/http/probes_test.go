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

type stubPinger struct {
	err error
}

func newTestRouter(db databasePinger) http.Handler {
	return NewRouter(db, RouterConfig{
		ReadinessTimeout: time.Second,
		Logger:           slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil)),
	})
}

func (p stubPinger) PingContext(context.Context) error {
	return p.err
}

func TestHealthIsIndependentFromDatabase(t *testing.T) {
	router := newTestRouter(stubPinger{err: errors.New("offline")})
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if !strings.Contains(response.Body.String(), `"status":"ok"`) {
		t.Fatalf("body = %q, want healthy status", response.Body.String())
	}
}

func TestReadyWhenDatabaseIsAvailable(t *testing.T) {
	router := newTestRouter(stubPinger{})
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
}

func TestReadyWhenDatabaseIsUnavailable(t *testing.T) {
	router := newTestRouter(stubPinger{err: errors.New("offline")})
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
	body := response.Body.String()
	if !strings.Contains(body, `"code":"database_unavailable"`) {
		t.Fatalf("body = %q, want safe database error", body)
	}
	if strings.Contains(body, "offline") {
		t.Fatalf("body = %q, must not expose internal error", body)
	}
}

func TestProbeRejectsUnsupportedMethod(t *testing.T) {
	router := newTestRouter(stubPinger{})
	request := httptest.NewRequest(http.MethodPost, "/healthz", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
}

func TestUnimplementedAPIRouteReturnsStructuredNotFound(t *testing.T) {
	router := newTestRouter(stubPinger{})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"username":"user","password":"secret"}`))
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
	body := response.Body.String()
	if !strings.Contains(body, `"code":"route_not_found"`) {
		t.Fatalf("body = %q, want structured route_not_found", body)
	}
	if strings.Contains(body, "secret") {
		t.Fatalf("body = %q, must not echo request body", body)
	}
}
