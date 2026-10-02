package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const maxJSONBodyBytes int64 = 1 << 20

type requestError struct {
	Status  int
	Code    string
	Message string
	Fields  fieldErrors
}

func (err *requestError) Error() string {
	return err.Code
}

type pagination struct {
	Page     int
	PageSize int
}

func readJSON(response http.ResponseWriter, request *http.Request, destination any) error {
	mediaType, _, err := mime.ParseMediaType(request.Header.Get("Content-Type"))
	if err != nil || !strings.EqualFold(mediaType, "application/json") {
		return &requestError{
			Status:  http.StatusUnsupportedMediaType,
			Code:    "unsupported_media_type",
			Message: "Use Content-Type application/json.",
		}
	}

	request.Body = http.MaxBytesReader(response, request.Body, maxJSONBodyBytes)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(destination); err != nil {
		return invalidJSONError(err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return malformedJSON("O corpo deve conter um único valor JSON.")
		}
		return invalidJSONError(err)
	}
	return nil
}

func invalidJSONError(err error) error {
	var maxBytesError *http.MaxBytesError
	if errors.As(err, &maxBytesError) {
		return &requestError{
			Status:  http.StatusBadRequest,
			Code:    "request_body_too_large",
			Message: fmt.Sprintf("O corpo JSON deve ter no máximo %d bytes.", maxJSONBodyBytes),
		}
	}
	if errors.Is(err, io.EOF) {
		return malformedJSON("O corpo JSON é obrigatório.")
	}
	return malformedJSON("O corpo JSON é inválido, contém campos desconhecidos ou dados extras.")
}

func malformedJSON(message string) error {
	return &requestError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_json",
		Message: message,
	}
}

func writeRequestError(response http.ResponseWriter, request *http.Request, err error) bool {
	var inputError *requestError
	if !errors.As(err, &inputError) {
		return false
	}
	writeError(response, request, inputError.Status, inputError.Code, inputError.Message, inputError.Fields)
	return true
}

func parsePositiveID(raw string) (int64, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, &requestError{
			Status:  http.StatusBadRequest,
			Code:    "invalid_id",
			Message: "O identificador deve ser um número inteiro positivo.",
		}
	}
	return id, nil
}

func parsePagination(values url.Values) (pagination, error) {
	page, err := parseSinglePositiveQuery(values, "page", 1, 0)
	if err != nil {
		return pagination{}, err
	}
	pageSize, err := parseSinglePositiveQuery(values, "page_size", 20, 100)
	if err != nil {
		return pagination{}, err
	}
	return pagination{Page: page, PageSize: pageSize}, nil
}

func parseSingleOptionalQuery(values url.Values, name string) (string, error) {
	rawValues, exists := values[name]
	if !exists {
		return "", nil
	}
	if len(rawValues) != 1 {
		return "", &requestError{
			Status:  http.StatusBadRequest,
			Code:    "invalid_query_parameter",
			Message: "Um ou mais parâmetros de consulta são inválidos.",
			Fields:  fieldErrors{name: {"Informe o parâmetro no máximo uma vez."}},
		}
	}
	return rawValues[0], nil
}

func rejectUnknownQueryParameters(values url.Values, allowed ...string) error {
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, name := range allowed {
		allowedSet[name] = struct{}{}
	}
	for name := range values {
		if _, ok := allowedSet[name]; !ok {
			return &requestError{
				Status:  http.StatusBadRequest,
				Code:    "invalid_query_parameter",
				Message: "Um ou mais parâmetros de consulta são inválidos.",
				Fields:  fieldErrors{"query": {"Parâmetro não reconhecido: " + name + "."}},
			}
		}
	}
	return nil
}

func parseSinglePositiveQuery(values url.Values, name string, fallback, maximum int) (int, error) {
	rawValues, exists := values[name]
	if !exists || len(rawValues) == 1 && strings.TrimSpace(rawValues[0]) == "" {
		return fallback, nil
	}
	if len(rawValues) != 1 {
		return 0, invalidQueryParameter(name)
	}
	value, err := strconv.Atoi(rawValues[0])
	if err != nil || value <= 0 || maximum > 0 && value > maximum {
		return 0, invalidQueryParameter(name)
	}
	return value, nil
}

func invalidQueryParameter(name string) error {
	return &requestError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_query_parameter",
		Message: "Um ou mais parâmetros de consulta são inválidos.",
		Fields: fieldErrors{
			name: {"Informe um único número inteiro positivo dentro do limite permitido."},
		},
	}
}
