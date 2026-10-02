package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type jsonPayload struct {
	Name string `json:"name"`
}

func TestReadJSONAcceptsSingleKnownObject(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"Portal"}`))
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	response := httptest.NewRecorder()
	var payload jsonPayload

	if err := readJSON(response, request, &payload); err != nil {
		t.Fatalf("readJSON() error = %v", err)
	}
	if payload.Name != "Portal" {
		t.Fatalf("Name = %q, want Portal", payload.Name)
	}
}

func TestReadJSONRejectsInvalidInputs(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		body        string
		wantStatus  int
		wantCode    string
	}{
		{name: "missing content type", body: `{}`, wantStatus: 415, wantCode: "unsupported_media_type"},
		{name: "wrong content type", contentType: "text/plain", body: `{}`, wantStatus: 415, wantCode: "unsupported_media_type"},
		{name: "unknown field", contentType: "application/json", body: `{"unknown":true}`, wantStatus: 400, wantCode: "invalid_json"},
		{name: "extra value", contentType: "application/json", body: `{"name":"one"} {"name":"two"}`, wantStatus: 400, wantCode: "invalid_json"},
		{name: "empty", contentType: "application/json", wantStatus: 400, wantCode: "invalid_json"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.body))
			request.Header.Set("Content-Type", test.contentType)
			response := httptest.NewRecorder()
			var payload jsonPayload

			err := readJSON(response, request, &payload)
			var requestErr *requestError
			if !errors.As(err, &requestErr) {
				t.Fatalf("error = %v, want requestError", err)
			}
			if requestErr.Status != test.wantStatus || requestErr.Code != test.wantCode {
				t.Fatalf("error = %d/%q, want %d/%q", requestErr.Status, requestErr.Code, test.wantStatus, test.wantCode)
			}
		})
	}
}

func TestReadJSONRejectsOversizedBody(t *testing.T) {
	body := `{"name":"` + strings.Repeat("a", int(maxJSONBodyBytes)) + `"}`
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	var payload jsonPayload

	err := readJSON(response, request, &payload)
	var requestErr *requestError
	if !errors.As(err, &requestErr) || requestErr.Code != "request_body_too_large" {
		t.Fatalf("error = %v, want request_body_too_large", err)
	}
}

func TestParsePositiveID(t *testing.T) {
	if got, err := parsePositiveID("42"); err != nil || got != 42 {
		t.Fatalf("parsePositiveID(42) = %d, %v", got, err)
	}
	for _, raw := range []string{"", "0", "-1", "1.5", "9223372036854775808"} {
		if _, err := parsePositiveID(raw); err == nil {
			t.Errorf("parsePositiveID(%q) expected error", raw)
		}
	}
}

func TestParsePagination(t *testing.T) {
	tests := []struct {
		name      string
		values    url.Values
		want      pagination
		wantError bool
	}{
		{name: "defaults", values: url.Values{}, want: pagination{Page: 1, PageSize: 20}},
		{name: "custom", values: url.Values{"page": {"3"}, "page_size": {"50"}}, want: pagination{Page: 3, PageSize: 50}},
		{name: "zero page", values: url.Values{"page": {"0"}}, wantError: true},
		{name: "oversized", values: url.Values{"page_size": {"101"}}, wantError: true},
		{name: "duplicate", values: url.Values{"page": {"1", "2"}}, wantError: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parsePagination(test.values)
			if (err != nil) != test.wantError {
				t.Fatalf("parsePagination() error = %v, wantError = %v", err, test.wantError)
			}
			if got != test.want {
				t.Fatalf("parsePagination() = %+v, want %+v", got, test.want)
			}
		})
	}
}

func TestWriteRequestErrorUsesFieldEnvelope(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/requests?page=zero", nil)
	request = request.WithContext(context.WithValue(request.Context(), requestIDContextKey, "req-test"))
	response := httptest.NewRecorder()
	_, err := parsePagination(url.Values{"page": {"zero"}})

	if !writeRequestError(response, request, err) {
		t.Fatal("writeRequestError() = false, want true")
	}
	body := response.Body.String()
	for _, expected := range []string{`"code":"invalid_query_parameter"`, `"fields":{"page":`, `"request_id":"req-test"`} {
		if !strings.Contains(body, expected) {
			t.Fatalf("body = %q, missing %q", body, expected)
		}
	}
}
