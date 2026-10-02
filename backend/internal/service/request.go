package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
	_ "time/tzdata"
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
	List(context.Context, repository.RequestFilters) ([]domain.Request, int64, error)
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

type ListRequestsInput struct {
	DateFrom string
	DateTo   string
	Category string
	Status   string
	Query    string
	Page     int
	PageSize int
}

type RequestListResult struct {
	Items      []domain.Request
	Page       int
	PageSize   int
	TotalItems int64
	TotalPages int64
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

func (service *Request) List(ctx context.Context, input ListRequestsInput) (RequestListResult, error) {
	filters, validationErr := validateRequestList(input)
	if validationErr != nil {
		return RequestListResult{}, validationErr
	}
	items, total, err := service.repository.List(ctx, filters)
	if err != nil {
		return RequestListResult{}, err
	}
	totalPages := total / int64(input.PageSize)
	if total%int64(input.PageSize) != 0 {
		totalPages++
	}
	return RequestListResult{
		Items: items, Page: input.Page, PageSize: input.PageSize,
		TotalItems: total, TotalPages: totalPages,
	}, nil
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

func validateRequestList(input ListRequestsInput) (repository.RequestFilters, error) {
	fields := make(map[string][]string)
	category := strings.TrimSpace(input.Category)
	status := strings.TrimSpace(input.Status)
	query := strings.TrimSpace(input.Query)
	if category != "" {
		if _, ok := domain.CategoryLabel(category); !ok {
			fields["category"] = []string{"Escolha uma categoria válida."}
		}
	}
	if status != "" {
		if _, ok := domain.StatusLabel(status); !ok {
			fields["status"] = []string{"Escolha um status válido."}
		}
	}
	if utf8.RuneCountInString(query) > 150 {
		fields["q"] = []string{"Informe no máximo 150 caracteres."}
	}
	if input.Page <= 0 {
		fields["page"] = []string{"Informe um número inteiro positivo."}
	}
	if input.PageSize <= 0 || input.PageSize > 100 {
		fields["page_size"] = []string{"Informe um número inteiro entre 1 e 100."}
	}
	if input.Page > 0 && input.PageSize > 0 && int64(input.Page-1) > math.MaxInt64/int64(input.PageSize) {
		fields["page"] = []string{"O número da página é muito grande."}
	}

	location, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		return repository.RequestFilters{}, fmt.Errorf("load request filter timezone: %w", err)
	}
	var dateFrom, dateToDate *time.Time
	if raw := strings.TrimSpace(input.DateFrom); raw != "" {
		parsed, parseErr := parseCalendarDate(raw)
		if parseErr != nil {
			fields["date_from"] = []string{"Use uma data real no formato YYYY-MM-DD."}
		} else {
			dateFrom = &parsed
		}
	}
	if raw := strings.TrimSpace(input.DateTo); raw != "" {
		parsed, parseErr := parseCalendarDate(raw)
		if parseErr != nil {
			fields["date_to"] = []string{"Use uma data real no formato YYYY-MM-DD."}
		} else {
			dateToDate = &parsed
		}
	}
	if dateFrom != nil && dateToDate != nil && dateFrom.After(*dateToDate) {
		fields["date_to"] = []string{"A data final deve ser igual ou posterior à data inicial."}
	}
	if len(fields) > 0 {
		return repository.RequestFilters{}, &ValidationError{Fields: fields}
	}

	filters := repository.RequestFilters{
		Category:   category,
		Status:     status,
		TitleQuery: query,
		Limit:      input.PageSize,
		Offset:     int64(input.Page-1) * int64(input.PageSize),
	}
	if dateFrom != nil {
		start, startErr := localDayStart(*dateFrom, location)
		if startErr != nil {
			return repository.RequestFilters{}, startErr
		}
		start = start.UTC()
		filters.DateFrom = &start
	}
	if dateToDate != nil {
		nextDate := dateToDate.AddDate(0, 0, 1)
		end, endErr := localDayStart(nextDate, location)
		if endErr != nil {
			return repository.RequestFilters{}, endErr
		}
		end = end.UTC()
		filters.DateToExclusive = &end
	}
	return filters, nil
}

func parseCalendarDate(value string) (time.Time, error) {
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil || parsed.Year() < 1 {
		return time.Time{}, fmt.Errorf("invalid calendar date")
	}
	return parsed, nil
}

func localDayStart(date time.Time, location *time.Location) (time.Time, error) {
	year, month, day := date.Date()
	for hour := 0; hour < 24; hour++ {
		candidate := time.Date(year, month, day, hour, 0, 0, 0, location)
		candidateYear, candidateMonth, candidateDay := candidate.In(location).Date()
		if candidateYear == year && candidateMonth == month && candidateDay == day {
			return candidate, nil
		}
	}
	return time.Time{}, fmt.Errorf("calendar day %04d-%02d-%02d does not exist in America/Sao_Paulo", year, month, day)
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
