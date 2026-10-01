package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type stubPinger struct {
	err error
}

func (p stubPinger) PingContext(context.Context) error {
	return p.err
}

func TestHealthIsIndependentFromDatabase(t *testing.T) {
	router := NewRouter(stubPinger{err: errors.New("offline")}, time.Second)
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
	router := NewRouter(stubPinger{}, time.Second)
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
}

func TestReadyWhenDatabaseIsUnavailable(t *testing.T) {
	router := NewRouter(stubPinger{err: errors.New("offline")}, time.Second)
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
	router := NewRouter(stubPinger{}, time.Second)
	request := httptest.NewRequest(http.MethodPost, "/healthz", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
}
