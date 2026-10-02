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

type stubRequestService struct {
	create func(context.Context, int64, service.CreateRequestInput) (service.RequestResult, error)
	get    func(context.Context, int64, int64) (service.RequestResult, error)
	update func(context.Context, int64, int64, service.UpdateRequestInput) (service.RequestResult, error)
	delete func(context.Context, int64, int64) error
	status func(context.Context, int64, int64, service.UpdateRequestStatusInput) (service.RequestResult, error)
	list   func(context.Context, service.ListRequestsInput) (service.RequestListResult, error)
}

func (stub stubRequestService) Create(ctx context.Context, userID int64, input service.CreateRequestInput) (service.RequestResult, error) {
	return stub.create(ctx, userID, input)
}

func (stub stubRequestService) List(ctx context.Context, input service.ListRequestsInput) (service.RequestListResult, error) {
	return stub.list(ctx, input)
}

func (stub stubRequestService) Get(ctx context.Context, id, userID int64) (service.RequestResult, error) {
	return stub.get(ctx, id, userID)
}

func (stub stubRequestService) Update(ctx context.Context, id, userID int64, input service.UpdateRequestInput) (service.RequestResult, error) {
	return stub.update(ctx, id, userID, input)
}

func (stub stubRequestService) Delete(ctx context.Context, id, userID int64) error {
	return stub.delete(ctx, id, userID)
}

func (stub stubRequestService) UpdateStatus(ctx context.Context, id, userID int64, input service.UpdateRequestStatusInput) (service.RequestResult, error) {
	return stub.status(ctx, id, userID, input)
}

func requestHTTPResult(ownerID int64, status string) service.RequestResult {
	localTime := time.FixedZone("America/Sao_Paulo", -3*60*60)
	request := domain.Request{
		ID:          42,
		Title:       "Acesso ao sistema",
		Description: "Solicitação detalhada para acesso ao sistema.",
		Category:    domain.CategoryTI,
		Status:      status,
		Requester:   domain.User{ID: ownerID, Username: "colaborador1", DisplayName: "Colaborador 1"},
		CreatedAt:   time.Date(2026, 10, 2, 12, 0, 0, 0, localTime),
		UpdatedAt:   time.Date(2026, 10, 2, 12, 5, 0, 0, localTime),
	}
	allowed := ownerID == 1 && status == domain.StatusOpen
	return service.RequestResult{Request: request, CanEdit: allowed, CanDelete: allowed}
}

func requestRouter(requests requestService) http.Handler {
	auth := stubAuthService{authenticate: func(_ context.Context, token string) (domain.User, error) {
		if token == "" {
			return domain.User{}, service.ErrUnauthenticated
		}
		return domain.User{ID: 1, Username: "colaborador1", DisplayName: "Colaborador 1"}, nil
	}}
	return NewRouter(stubPinger{}, RouterConfig{
		ReadinessTimeout: time.Second,
		AuthService:      auth,
		RequestService:   requests,
		TrustedOrigins:   []string{"http://127.0.0.1:5173"},
		SessionCookie:    SessionCookieConfig("portal_session", false, 8*time.Hour),
		LoginRateLimit:   LoginRateLimitConfig(5, 15*time.Minute, 100),
	})
}

func authenticatedRequest(method, target, body string) *http.Request {
	request := httptest.NewRequest(method, target, strings.NewReader(body))
	request.AddCookie(&http.Cookie{Name: "portal_session", Value: "token"})
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	if method == http.MethodPost || method == http.MethodPatch || method == http.MethodDelete {
		request.Header.Set("Origin", "http://127.0.0.1:5173")
	}
	return request
}

func TestCreateRequestReturnsSafeRepresentationAndLocation(t *testing.T) {
	requests := stubRequestService{create: func(_ context.Context, userID int64, input service.CreateRequestInput) (service.RequestResult, error) {
		if userID != 1 || input.Title != "Acesso" {
			t.Fatalf("user/input = %d %+v", userID, input)
		}
		return requestHTTPResult(1, domain.StatusOpen), nil
	}}
	router := requestRouter(requests)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, authenticatedRequest(http.MethodPost, "/api/v1/requests", `{"title":"Acesso","description":"Descrição suficientemente longa.","category":"ti"}`))

	if response.Code != http.StatusCreated || response.Header().Get("Location") != "/api/v1/requests/42" {
		t.Fatalf("status/location = %d %q body=%s", response.Code, response.Header().Get("Location"), response.Body.String())
	}
	body := response.Body.String()
	for _, expected := range []string{`"code":"SOL-000042"`, `"category":{"value":"ti","label":"TI"}`, `"status":{"value":"aberto","label":"Aberto"}`, `"created_at":"2026-10-02T15:00:00Z"`, `"can_edit":true`} {
		if !strings.Contains(body, expected) {
			t.Fatalf("body = %s, missing %s", body, expected)
		}
	}
	for _, forbidden := range []string{"password_hash", "portal_session", "token_hash"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("body leaked %s: %s", forbidden, body)
		}
	}
}

func TestCreateRequestRejectsAutomaticOrUnknownFields(t *testing.T) {
	called := false
	requests := stubRequestService{create: func(context.Context, int64, service.CreateRequestInput) (service.RequestResult, error) {
		called = true
		return service.RequestResult{}, nil
	}}
	router := requestRouter(requests)
	for _, field := range []string{"id", "requester_id", "created_at", "updated_at", "status"} {
		body := `{"title":"Título","description":"Descrição válida aqui.","category":"ti","` + field + `":1}`
		response := httptest.NewRecorder()
		router.ServeHTTP(response, authenticatedRequest(http.MethodPost, "/api/v1/requests", body))
		if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), `"code":"invalid_json"`) {
			t.Fatalf("field=%s status/body=%d %s", field, response.Code, response.Body.String())
		}
	}
	if called {
		t.Fatal("service must not be called for forbidden payload fields")
	}
}

func TestGetRequestAllowsAnotherAuthenticatedUserWithSafePermissions(t *testing.T) {
	requests := stubRequestService{get: func(_ context.Context, id, userID int64) (service.RequestResult, error) {
		if id != 42 || userID != 1 {
			t.Fatalf("id/user = %d/%d", id, userID)
		}
		return requestHTTPResult(2, domain.StatusOpen), nil
	}}
	response := httptest.NewRecorder()
	requestRouter(requests).ServeHTTP(response, authenticatedRequest(http.MethodGet, "/api/v1/requests/42", ""))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"can_edit":false`) {
		t.Fatalf("status/body = %d %s", response.Code, response.Body.String())
	}
}

func TestPatchPreservesAbsentAndRecognizesNull(t *testing.T) {
	requests := stubRequestService{update: func(_ context.Context, _, _ int64, input service.UpdateRequestInput) (service.RequestResult, error) {
		if input.Title.Present || !input.Description.Present || !input.Description.Null || !input.Category.Present || input.Category.Null || input.Category.Value != "rh" {
			t.Fatalf("input = %+v", input)
		}
		return service.RequestResult{}, &service.ValidationError{Fields: map[string][]string{"description": {"A descrição não pode ser nula."}}}
	}}
	response := httptest.NewRecorder()
	requestRouter(requests).ServeHTTP(response, authenticatedRequest(http.MethodPatch, "/api/v1/requests/42", `{"description":null,"category":"rh"}`))
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), `"fields":{"description":`) {
		t.Fatalf("status/body = %d %s", response.Code, response.Body.String())
	}
}

func TestRequestMutationErrorsUseDocumentedStatuses(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{name: "missing", err: service.ErrRequestNotFound, status: 404, code: "request_not_found"},
		{name: "other author", err: service.ErrRequestForbidden, status: 403, code: "request_forbidden"},
		{name: "closed", err: service.ErrRequestConflict, status: 409, code: "request_not_open"},
		{name: "unchanged", err: service.ErrRequestNoChanges, status: 400, code: "no_changes"},
		{name: "infrastructure", err: errors.New("SQL secret"), status: 500, code: "internal_error"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			requests := stubRequestService{update: func(context.Context, int64, int64, service.UpdateRequestInput) (service.RequestResult, error) {
				return service.RequestResult{}, test.err
			}}
			response := httptest.NewRecorder()
			requestRouter(requests).ServeHTTP(response, authenticatedRequest(http.MethodPatch, "/api/v1/requests/42", `{"title":"Título novo"}`))
			if response.Code != test.status || !strings.Contains(response.Body.String(), `"code":"`+test.code+`"`) || strings.Contains(response.Body.String(), "SQL") {
				t.Fatalf("status/body = %d %s", response.Code, response.Body.String())
			}
		})
	}
}

func TestDeleteReturnsNoContentWithoutBody(t *testing.T) {
	requests := stubRequestService{delete: func(context.Context, int64, int64) error { return nil }}
	response := httptest.NewRecorder()
	requestRouter(requests).ServeHTTP(response, authenticatedRequest(http.MethodDelete, "/api/v1/requests/42", ""))
	if response.Code != http.StatusNoContent || response.Body.Len() != 0 {
		t.Fatalf("status/body = %d %q", response.Code, response.Body.String())
	}
}

func TestRequestMutationsRequireAuthenticationAndTrustedOrigin(t *testing.T) {
	requests := stubRequestService{create: func(context.Context, int64, service.CreateRequestInput) (service.RequestResult, error) {
		t.Fatal("request service must not run before authentication and origin checks")
		return service.RequestResult{}, nil
	}}
	router := requestRouter(requests)
	body := `{"title":"Acesso","description":"Descrição suficientemente longa.","category":"ti"}`

	unauthenticated := httptest.NewRequest(http.MethodPost, "/api/v1/requests", strings.NewReader(body))
	unauthenticated.Header.Set("Content-Type", "application/json")
	unauthenticated.Header.Set("Origin", "http://127.0.0.1:5173")
	unauthenticatedResponse := httptest.NewRecorder()
	router.ServeHTTP(unauthenticatedResponse, unauthenticated)
	if unauthenticatedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status/body = %d %s", unauthenticatedResponse.Code, unauthenticatedResponse.Body.String())
	}

	untrusted := authenticatedRequest(http.MethodPost, "/api/v1/requests", body)
	untrusted.Header.Set("Origin", "https://hostil.example")
	untrustedResponse := httptest.NewRecorder()
	router.ServeHTTP(untrustedResponse, untrusted)
	if untrustedResponse.Code != http.StatusForbidden || !strings.Contains(untrustedResponse.Body.String(), `"code":"origin_not_allowed"`) {
		t.Fatalf("untrusted origin status/body = %d %s", untrustedResponse.Code, untrustedResponse.Body.String())
	}
}

func TestUpdateStatusAcceptsAnotherUserAndReturnsLabels(t *testing.T) {
	requests := stubRequestService{status: func(_ context.Context, id, userID int64, input service.UpdateRequestStatusInput) (service.RequestResult, error) {
		if id != 42 || userID != 1 || !input.Status.Present || input.Status.Null || input.Status.Value != domain.StatusInProgress {
			t.Fatalf("id/user/input = %d/%d/%+v", id, userID, input)
		}
		return requestHTTPResult(2, domain.StatusInProgress), nil
	}}
	response := httptest.NewRecorder()
	requestRouter(requests).ServeHTTP(response, authenticatedRequest(http.MethodPatch, "/api/v1/requests/42/status", `{"status":"em_atendimento"}`))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"status":{"value":"em_atendimento","label":"Em Atendimento"}`) || !strings.Contains(response.Body.String(), `"can_edit":false`) {
		t.Fatalf("status/body = %d %s", response.Code, response.Body.String())
	}
}

func TestUpdateStatusRejectsInvalidPayloadsBeforeService(t *testing.T) {
	requests := stubRequestService{status: func(_ context.Context, _, _ int64, input service.UpdateRequestStatusInput) (service.RequestResult, error) {
		if !input.Status.Present || input.Status.Null || input.Status.Value == "cancelado" {
			return service.RequestResult{}, &service.ValidationError{Fields: map[string][]string{"status": {"Escolha um status válido."}}}
		}
		t.Fatalf("service called with unexpected input: %+v", input)
		return service.RequestResult{}, nil
	}}
	tests := []struct {
		name string
		body string
		code string
	}{
		{name: "missing", body: `{}`, code: "validation_failed"},
		{name: "null", body: `{"status":null}`, code: "validation_failed"},
		{name: "unknown", body: `{"status":"cancelado"}`, code: "validation_failed"},
		{name: "invalid type", body: `{"status":1}`, code: "invalid_json"},
		{name: "extra", body: `{"status":"aberto","title":"indevido"}`, code: "invalid_json"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			requestRouter(requests).ServeHTTP(response, authenticatedRequest(http.MethodPatch, "/api/v1/requests/42/status", test.body))
			if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), `"code":"`+test.code+`"`) {
				t.Fatalf("status/body = %d %s", response.Code, response.Body.String())
			}
		})
	}
}

func TestUpdateStatusRequiresAuthenticationOriginAndExistingID(t *testing.T) {
	called := 0
	requests := stubRequestService{status: func(context.Context, int64, int64, service.UpdateRequestStatusInput) (service.RequestResult, error) {
		called++
		return service.RequestResult{}, service.ErrRequestNotFound
	}}
	router := requestRouter(requests)
	body := `{"status":"aberto"}`

	unauthenticated := httptest.NewRequest(http.MethodPatch, "/api/v1/requests/42/status", strings.NewReader(body))
	unauthenticated.Header.Set("Content-Type", "application/json")
	unauthenticated.Header.Set("Origin", "http://127.0.0.1:5173")
	unauthenticatedResponse := httptest.NewRecorder()
	router.ServeHTTP(unauthenticatedResponse, unauthenticated)
	if unauthenticatedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status = %d", unauthenticatedResponse.Code)
	}

	untrusted := authenticatedRequest(http.MethodPatch, "/api/v1/requests/42/status", body)
	untrusted.Header.Set("Origin", "https://hostil.example")
	untrustedResponse := httptest.NewRecorder()
	router.ServeHTTP(untrustedResponse, untrusted)
	if untrustedResponse.Code != http.StatusForbidden {
		t.Fatalf("untrusted status = %d", untrustedResponse.Code)
	}

	missingResponse := httptest.NewRecorder()
	router.ServeHTTP(missingResponse, authenticatedRequest(http.MethodPatch, "/api/v1/requests/999/status", body))
	if missingResponse.Code != http.StatusNotFound || called != 1 {
		t.Fatalf("missing status/calls = %d/%d body=%s", missingResponse.Code, called, missingResponse.Body.String())
	}
}

func TestListRequestsReturnsExactSummaryAndPagination(t *testing.T) {
	requests := stubRequestService{list: func(_ context.Context, input service.ListRequestsInput) (service.RequestListResult, error) {
		if input.DateFrom != "2026-10-01" || input.DateTo != "2026-10-02" || input.Category != "ti" || input.Status != "aberto" || input.Query != "ácesso_%" || input.Page != 2 || input.PageSize != 1 {
			t.Fatalf("input = %+v", input)
		}
		item := requestHTTPResult(2, domain.StatusOpen).Request
		return service.RequestListResult{Items: []domain.Request{item}, Page: 2, PageSize: 1, TotalItems: 3, TotalPages: 3}, nil
	}}
	response := httptest.NewRecorder()
	requestRouter(requests).ServeHTTP(response, authenticatedRequest(http.MethodGet, "/api/v1/requests?date_from=2026-10-01&date_to=2026-10-02&category=ti&status=aberto&q=%C3%A1cesso_%25&page=2&page_size=1", ""))
	if response.Code != http.StatusOK {
		t.Fatalf("status/body = %d %s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	for _, expected := range []string{`"items":[`, `"code":"SOL-000042"`, `"requester":{"id":2,"username":"colaborador1","display_name":"Colaborador 1"}`, `"created_at":"2026-10-02T15:00:00Z"`, `"pagination":{"page":2,"page_size":1,"total_items":3,"total_pages":3}`} {
		if !strings.Contains(body, expected) {
			t.Fatalf("body = %s, missing %s", body, expected)
		}
	}
	for _, forbidden := range []string{"description", "updated_at", "permissions", "password_hash", "token_hash"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("body contains forbidden list field %q: %s", forbidden, body)
		}
	}
}

func TestListRequestsUsesDefaultsAndReturnsEmptyItems(t *testing.T) {
	requests := stubRequestService{list: func(_ context.Context, input service.ListRequestsInput) (service.RequestListResult, error) {
		if input.Page != 1 || input.PageSize != 20 {
			t.Fatalf("input = %+v", input)
		}
		return service.RequestListResult{Items: nil, Page: 1, PageSize: 20}, nil
	}}
	response := httptest.NewRecorder()
	requestRouter(requests).ServeHTTP(response, authenticatedRequest(http.MethodGet, "/api/v1/requests?category=&q=%20%20", ""))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"items":[]`) {
		t.Fatalf("status/body = %d %s", response.Code, response.Body.String())
	}
}

func TestListRequestsRejectsAmbiguousAndInvalidQuery(t *testing.T) {
	requests := stubRequestService{list: func(_ context.Context, input service.ListRequestsInput) (service.RequestListResult, error) {
		if input.Category == "invalida" {
			return service.RequestListResult{}, &service.ValidationError{Fields: map[string][]string{"category": {"Escolha uma categoria válida."}}}
		}
		t.Fatal("service must not run for syntactically invalid query")
		return service.RequestListResult{}, nil
	}}
	tests := []struct {
		name   string
		target string
		field  string
	}{
		{name: "repeated", target: "/api/v1/requests?status=aberto&status=concluido", field: "status"},
		{name: "invalid integer", target: "/api/v1/requests?page=um", field: "page"},
		{name: "page size limit", target: "/api/v1/requests?page_size=101", field: "page_size"},
		{name: "unknown", target: "/api/v1/requests?sort=title", field: "query"},
		{name: "service validation", target: "/api/v1/requests?category=invalida", field: "category"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			requestRouter(requests).ServeHTTP(response, authenticatedRequest(http.MethodGet, test.target, ""))
			if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), `"code":"invalid_query_parameter"`) || !strings.Contains(response.Body.String(), `"`+test.field+`":`) {
				t.Fatalf("status/body = %d %s", response.Code, response.Body.String())
			}
		})
	}
}

func TestListRequestsRequiresAuthentication(t *testing.T) {
	requests := stubRequestService{list: func(context.Context, service.ListRequestsInput) (service.RequestListResult, error) {
		t.Fatal("service must not run anonymously")
		return service.RequestListResult{}, nil
	}}
	response := httptest.NewRecorder()
	requestRouter(requests).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/requests", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status/body = %d %s", response.Code, response.Body.String())
	}
}
