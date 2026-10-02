package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"portal-solicitacoes/internal/domain"
	"portal-solicitacoes/internal/service"
)

type stubDashboardService struct {
	counts domain.DashboardCounts
	err    error
}

func (stub stubDashboardService) Get(context.Context) (domain.DashboardCounts, error) {
	return stub.counts, stub.err
}

func dashboardRouter(dashboard dashboardService) http.Handler {
	auth := stubAuthService{authenticate: func(_ context.Context, token string) (domain.User, error) {
		if token == "" {
			return domain.User{}, service.ErrUnauthenticated
		}
		return domain.User{ID: 1, Username: "colaborador1", DisplayName: "Colaborador 1"}, nil
	}}
	return NewRouter(stubPinger{}, RouterConfig{
		ReadinessTimeout: time.Second,
		AuthService:      auth,
		DashboardService: dashboard,
		TrustedOrigins:   []string{"http://127.0.0.1:5173"},
		SessionCookie:    SessionCookieConfig("portal_session", false, 8*time.Hour),
		LoginRateLimit:   LoginRateLimitConfig(5, 15*time.Minute, 100),
	})
}

func TestDashboardReturnsPopulatedAndEmptyCounts(t *testing.T) {
	tests := []struct {
		name   string
		counts domain.DashboardCounts
		body   string
	}{
		{name: "empty", body: `{"data":{"total":0,"abertas":0,"em_atendimento":0,"concluidas":0}}`},
		{name: "populated", counts: domain.DashboardCounts{Total: 6, Open: 1, InProgress: 2, Completed: 3}, body: `{"data":{"total":6,"abertas":1,"em_atendimento":2,"concluidas":3}}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			dashboardRouter(stubDashboardService{counts: test.counts}).ServeHTTP(response, authenticatedRequest(http.MethodGet, "/api/v1/dashboard", ""))
			if response.Code != http.StatusOK || strings.TrimSpace(response.Body.String()) != test.body {
				t.Fatalf("status/body = %d %s; want 200 %s", response.Code, response.Body.String(), test.body)
			}
		})
	}
}

func TestDashboardRequiresAuthentication(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard", nil)
	dashboardRouter(stubDashboardService{}).ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized || !strings.Contains(response.Body.String(), `"code":"authentication_required"`) {
		t.Fatalf("status/body = %d %s", response.Code, response.Body.String())
	}
}

func TestDashboardReturnsSafeErrorOnInfrastructureFailure(t *testing.T) {
	response := httptest.NewRecorder()
	dashboardRouter(stubDashboardService{err: errors.New("postgres password=secret")}).ServeHTTP(response, authenticatedRequest(http.MethodGet, "/api/v1/dashboard", ""))
	if response.Code != http.StatusInternalServerError || !strings.Contains(response.Body.String(), `"code":"internal_error"`) {
		t.Fatalf("status/body = %d %s", response.Code, response.Body.String())
	}
	if strings.Contains(response.Body.String(), "password") || strings.Contains(response.Body.String(), "secret") {
		t.Fatalf("response leaked infrastructure details: %s", response.Body.String())
	}
}
