package httpapi

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

type fieldErrors map[string][]string

type errorBody struct {
	Error     errorDetail `json:"error"`
	RequestID string      `json:"request_id,omitempty"`
}

type errorDetail struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Fields  fieldErrors `json:"fields,omitempty"`
}

func writeJSON(response http.ResponseWriter, status int, payload any) {
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	response.Header().Set("Cache-Control", "no-store")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(payload)
}

func writeError(response http.ResponseWriter, request *http.Request, status int, code, message string, fields fieldErrors) {
	writeJSON(response, status, errorBody{
		Error: errorDetail{
			Code:    code,
			Message: message,
			Fields:  fields,
		},
		RequestID: requestIDFromContext(request.Context()),
	})
}

func writeInternalError(logger *slog.Logger, response http.ResponseWriter, request *http.Request, err error) {
	logger.Error("request failed",
		"request_id", requestIDFromContext(request.Context()),
		"method", request.Method,
		"route", routeName(request),
		"error_type", fmt.Sprintf("%T", err),
	)
	writeError(response, request, http.StatusInternalServerError, "internal_error", "Erro interno inesperado.", nil)
}
