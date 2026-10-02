package service

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"portal-solicitacoes/internal/domain"
	"portal-solicitacoes/internal/repository"
)

var (
	ErrRequestNotFound  = errors.New("request not found")
	ErrRequestForbidden = errors.New("request operation forbidden")
	ErrRequestConflict  = errors.New("request state conflict")
	ErrRequestNoChanges = errors.New("request has no changes")
)

type RequestRepository interface {
	Create(context.Context, int64, string, string, string) (domain.Request, error)
	FindByID(context.Context, int64) (domain.Request, error)
	UpdateOwnedOpen(context.Context, int64, int64, repository.RequestChanges) (domain.Request, error)
	DeleteOwnedOpen(context.Context, int64, int64) error
	UpdateStatus(context.Context, int64, string) (domain.Request, error)
}

type Request struct {
	repository RequestRepository
}

type CreateRequestInput struct {
	Title       string
	Description string
	Category    string
}

type OptionalString struct {
	Present bool
	Null    bool
	Value   string
}

type UpdateRequestInput struct {
	Title       OptionalString
	Description OptionalString
	Category    OptionalString
}

type UpdateRequestStatusInput struct {
	Status OptionalString
}

type RequestResult struct {
	Request   domain.Request
	CanEdit   bool
	CanDelete bool
}

func NewRequest(repository RequestRepository) *Request {
	return &Request{repository: repository}
}

func (service *Request) Create(ctx context.Context, requesterID int64, input CreateRequestInput) (RequestResult, error) {
	title, description, category, validationErr := validateRequestFields(input.Title, input.Description, input.Category)
	if validationErr != nil {
		return RequestResult{}, validationErr
	}
	created, err := service.repository.Create(ctx, requesterID, title, description, category)
	if err != nil {
		return RequestResult{}, err
	}
	return presentRequestResult(created, requesterID), nil
}

func (service *Request) Get(ctx context.Context, id, currentUserID int64) (RequestResult, error) {
	found, err := service.repository.FindByID(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return RequestResult{}, ErrRequestNotFound
	}
	if err != nil {
		return RequestResult{}, err
	}
	return presentRequestResult(found, currentUserID), nil
}

func (service *Request) Update(ctx context.Context, id, currentUserID int64, input UpdateRequestInput) (RequestResult, error) {
	changes, validationErr := validateRequestChanges(input)
	if validationErr != nil {
		return RequestResult{}, validationErr
	}
	updated, err := service.repository.UpdateOwnedOpen(ctx, id, currentUserID, changes)
	if mapped := mapRequestMutationError(err); mapped != nil {
		return RequestResult{}, mapped
	}
	return presentRequestResult(updated, currentUserID), nil
}

func (service *Request) Delete(ctx context.Context, id, currentUserID int64) error {
	return mapRequestMutationError(service.repository.DeleteOwnedOpen(ctx, id, currentUserID))
}

func (service *Request) UpdateStatus(ctx context.Context, id, currentUserID int64, input UpdateRequestStatusInput) (RequestResult, error) {
	fields := make(map[string][]string)
	switch {
	case !input.Status.Present:
		fields["status"] = []string{"Informe o status."}
	case input.Status.Null:
		fields["status"] = []string{"O status não pode ser nulo."}
	default:
		if _, ok := domain.StatusLabel(input.Status.Value); !ok {
			fields["status"] = []string{"Escolha um status válido."}
		}
	}
	if len(fields) > 0 {
		return RequestResult{}, &ValidationError{Fields: fields}
	}

	updated, err := service.repository.UpdateStatus(ctx, id, input.Status.Value)
	if errors.Is(err, repository.ErrNotFound) {
		return RequestResult{}, ErrRequestNotFound
	}
	if err != nil {
		return RequestResult{}, err
	}
	return presentRequestResult(updated, currentUserID), nil
}

func validateRequestFields(rawTitle, rawDescription, rawCategory string) (string, string, string, error) {
	title := strings.TrimSpace(rawTitle)
	description := strings.TrimSpace(rawDescription)
	category := strings.TrimSpace(rawCategory)
	fields := make(map[string][]string)
	validateTextLength(fields, "title", title, 3, 150)
	validateTextLength(fields, "description", description, 10, 5000)
	if _, ok := domain.CategoryLabel(category); !ok {
		fields["category"] = []string{"Escolha uma categoria válida."}
	}
	if len(fields) > 0 {
		return "", "", "", &ValidationError{Fields: fields}
	}
	return title, description, category, nil
}

func validateRequestChanges(input UpdateRequestInput) (repository.RequestChanges, error) {
	fields := make(map[string][]string)
	changes := repository.RequestChanges{}
	provided := 0

	if input.Title.Present {
		provided++
		if input.Title.Null {
			fields["title"] = []string{"O título não pode ser nulo."}
		} else {
			changes.TitleSet = true
			changes.Title = strings.TrimSpace(input.Title.Value)
			validateTextLength(fields, "title", changes.Title, 3, 150)
		}
	}
	if input.Description.Present {
		provided++
		if input.Description.Null {
			fields["description"] = []string{"A descrição não pode ser nula."}
		} else {
			changes.DescriptionSet = true
			changes.Description = strings.TrimSpace(input.Description.Value)
			validateTextLength(fields, "description", changes.Description, 10, 5000)
		}
	}
	if input.Category.Present {
		provided++
		if input.Category.Null {
			fields["category"] = []string{"A categoria não pode ser nula."}
		} else {
			changes.CategorySet = true
			changes.Category = strings.TrimSpace(input.Category.Value)
			if _, ok := domain.CategoryLabel(changes.Category); !ok {
				fields["category"] = []string{"Escolha uma categoria válida."}
			}
		}
	}
	if provided == 0 {
		return repository.RequestChanges{}, &ValidationError{Fields: map[string][]string{
			"body": {"Informe ao menos um entre title, description e category."},
		}}
	}
	if len(fields) > 0 {
		return repository.RequestChanges{}, &ValidationError{Fields: fields}
	}
	return changes, nil
}

func validateTextLength(fields map[string][]string, name, value string, minimum, maximum int) {
	length := utf8.RuneCountInString(value)
	if length < minimum || length > maximum {
		fields[name] = []string{requestLengthMessage(name, minimum, maximum)}
	}
}

func requestLengthMessage(name string, minimum, maximum int) string {
	if name == "title" {
		return "Informe um título entre 3 e 150 caracteres."
	}
	return "Informe uma descrição entre 10 e 5000 caracteres."
}

func mapRequestMutationError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, repository.ErrNotFound):
		return ErrRequestNotFound
	case errors.Is(err, repository.ErrForbidden):
		return ErrRequestForbidden
	case errors.Is(err, repository.ErrConflict):
		return ErrRequestConflict
	case errors.Is(err, repository.ErrNoChanges):
		return ErrRequestNoChanges
	default:
		return err
	}
}

func presentRequestResult(request domain.Request, currentUserID int64) RequestResult {
	allowed := request.Requester.ID == currentUserID && request.Status == domain.StatusOpen
	return RequestResult{Request: request, CanEdit: allowed, CanDelete: allowed}
}
