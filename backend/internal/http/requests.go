package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"portal-solicitacoes/internal/domain"
	"portal-solicitacoes/internal/service"
)

type requestService interface {
	Create(context.Context, int64, service.CreateRequestInput) (service.RequestResult, error)
	Get(context.Context, int64, int64) (service.RequestResult, error)
	Update(context.Context, int64, int64, service.UpdateRequestInput) (service.RequestResult, error)
	Delete(context.Context, int64, int64) error
	UpdateStatus(context.Context, int64, int64, service.UpdateRequestStatusInput) (service.RequestResult, error)
}

type requestHandler struct {
	service requestService
	logger  *slog.Logger
}

type createRequestBody struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

type optionalString struct {
	Present bool
	Null    bool
	Value   string
}

func (value *optionalString) UnmarshalJSON(data []byte) error {
	value.Present = true
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		value.Null = true
		return nil
	}
	if err := json.Unmarshal(data, &value.Value); err != nil {
		return err
	}
	return nil
}

type updateRequestBody struct {
	Title       optionalString `json:"title"`
	Description optionalString `json:"description"`
	Category    optionalString `json:"category"`
}

type updateRequestStatusBody struct {
	Status optionalString `json:"status"`
}

type requestResponse struct {
	Data requestData `json:"data"`
}

type requestData struct {
	ID          int64             `json:"id"`
	Code        string            `json:"code"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Category    labeledValue      `json:"category"`
	Status      labeledValue      `json:"status"`
	Requester   userData          `json:"requester"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Permissions requestPermission `json:"permissions"`
}

type labeledValue struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type requestPermission struct {
	CanEdit   bool `json:"can_edit"`
	CanDelete bool `json:"can_delete"`
}

func (handler *requestHandler) create(response http.ResponseWriter, request *http.Request) {
	user, ok := authenticatedUser(request.Context())
	if !ok {
		writeInternalError(handler.logger, response, request, errors.New("authenticated user missing from context"))
		return
	}
	var body createRequestBody
	if err := readJSON(response, request, &body); err != nil {
		writeRequestError(response, request, err)
		return
	}

	result, err := handler.service.Create(request.Context(), user.ID, service.CreateRequestInput{
		Title: body.Title, Description: body.Description, Category: body.Category,
	})
	if err != nil {
		handler.writeServiceError(response, request, err)
		return
	}
	response.Header().Set("Location", "/api/v1/requests/"+strconv.FormatInt(result.Request.ID, 10))
	writeJSON(response, http.StatusCreated, requestResponse{Data: presentRequest(result)})
}

func (handler *requestHandler) get(response http.ResponseWriter, request *http.Request) {
	user, ok := authenticatedUser(request.Context())
	if !ok {
		writeInternalError(handler.logger, response, request, errors.New("authenticated user missing from context"))
		return
	}
	id, err := parsePositiveID(request.PathValue("id"))
	if err != nil {
		writeRequestError(response, request, err)
		return
	}

	result, err := handler.service.Get(request.Context(), id, user.ID)
	if err != nil {
		handler.writeServiceError(response, request, err)
		return
	}
	writeJSON(response, http.StatusOK, requestResponse{Data: presentRequest(result)})
}

func (handler *requestHandler) update(response http.ResponseWriter, request *http.Request) {
	user, ok := authenticatedUser(request.Context())
	if !ok {
		writeInternalError(handler.logger, response, request, errors.New("authenticated user missing from context"))
		return
	}
	id, err := parsePositiveID(request.PathValue("id"))
	if err != nil {
		writeRequestError(response, request, err)
		return
	}
	var body updateRequestBody
	if err := readJSON(response, request, &body); err != nil {
		writeRequestError(response, request, err)
		return
	}

	result, err := handler.service.Update(request.Context(), id, user.ID, service.UpdateRequestInput{
		Title:       toServiceOptional(body.Title),
		Description: toServiceOptional(body.Description),
		Category:    toServiceOptional(body.Category),
	})
	if err != nil {
		handler.writeServiceError(response, request, err)
		return
	}
	writeJSON(response, http.StatusOK, requestResponse{Data: presentRequest(result)})
}

func (handler *requestHandler) delete(response http.ResponseWriter, request *http.Request) {
	user, ok := authenticatedUser(request.Context())
	if !ok {
		writeInternalError(handler.logger, response, request, errors.New("authenticated user missing from context"))
		return
	}
	id, err := parsePositiveID(request.PathValue("id"))
	if err != nil {
		writeRequestError(response, request, err)
		return
	}
	if err := handler.service.Delete(request.Context(), id, user.ID); err != nil {
		handler.writeServiceError(response, request, err)
		return
	}
	response.Header().Set("Cache-Control", "no-store")
	response.WriteHeader(http.StatusNoContent)
}

func (handler *requestHandler) updateStatus(response http.ResponseWriter, request *http.Request) {
	user, ok := authenticatedUser(request.Context())
	if !ok {
		writeInternalError(handler.logger, response, request, errors.New("authenticated user missing from context"))
		return
	}
	id, err := parsePositiveID(request.PathValue("id"))
	if err != nil {
		writeRequestError(response, request, err)
		return
	}
	var body updateRequestStatusBody
	if err := readJSON(response, request, &body); err != nil {
		writeRequestError(response, request, err)
		return
	}

	result, err := handler.service.UpdateStatus(request.Context(), id, user.ID, service.UpdateRequestStatusInput{
		Status: toServiceOptional(body.Status),
	})
	if err != nil {
		handler.writeServiceError(response, request, err)
		return
	}
	writeJSON(response, http.StatusOK, requestResponse{Data: presentRequest(result)})
}

func (handler *requestHandler) writeServiceError(response http.ResponseWriter, request *http.Request, err error) {
	var validationErr *service.ValidationError
	switch {
	case errors.As(err, &validationErr):
		writeError(response, request, http.StatusBadRequest, "validation_failed", "Corrija os campos informados.", fieldErrors(validationErr.Fields))
	case errors.Is(err, service.ErrRequestNotFound):
		writeError(response, request, http.StatusNotFound, "request_not_found", "Solicitação não encontrada.", nil)
	case errors.Is(err, service.ErrRequestForbidden):
		writeError(response, request, http.StatusForbidden, "request_forbidden", "Você não pode alterar esta solicitação.", nil)
	case errors.Is(err, service.ErrRequestConflict):
		writeError(response, request, http.StatusConflict, "request_not_open", "Somente solicitações abertas podem ser alteradas ou excluídas.", nil)
	case errors.Is(err, service.ErrRequestNoChanges):
		writeError(response, request, http.StatusBadRequest, "no_changes", "Nenhuma alteração útil foi informada.", nil)
	default:
		writeInternalError(handler.logger, response, request, err)
	}
}

func presentRequest(result service.RequestResult) requestData {
	request := result.Request
	categoryLabel, _ := domain.CategoryLabel(request.Category)
	statusLabel, _ := domain.StatusLabel(request.Status)
	return requestData{
		ID:          request.ID,
		Code:        fmt.Sprintf("SOL-%06d", request.ID),
		Title:       request.Title,
		Description: request.Description,
		Category:    labeledValue{Value: request.Category, Label: categoryLabel},
		Status:      labeledValue{Value: request.Status, Label: statusLabel},
		Requester:   presentUser(request.Requester),
		CreatedAt:   request.CreatedAt.UTC(),
		UpdatedAt:   request.UpdatedAt.UTC(),
		Permissions: requestPermission{CanEdit: result.CanEdit, CanDelete: result.CanDelete},
	}
}

func toServiceOptional(value optionalString) service.OptionalString {
	return service.OptionalString{Present: value.Present, Null: value.Null, Value: value.Value}
}
