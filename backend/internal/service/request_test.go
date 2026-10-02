package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"portal-solicitacoes/internal/domain"
	"portal-solicitacoes/internal/repository"
)

type requestRepositoryStub struct {
	createdRequesterID int64
	createdTitle       string
	createdDescription string
	createdCategory    string
	createResult       domain.Request
	createErr          error
	findResult         domain.Request
	findErr            error
	updateChanges      repository.RequestChanges
	updateResult       domain.Request
	updateErr          error
	deleteErr          error
	statusValue        string
	statusResult       domain.Request
	statusErr          error
}

func (stub *requestRepositoryStub) Create(_ context.Context, requesterID int64, title, description, category string) (domain.Request, error) {
	stub.createdRequesterID = requesterID
	stub.createdTitle = title
	stub.createdDescription = description
	stub.createdCategory = category
	return stub.createResult, stub.createErr
}

func (stub *requestRepositoryStub) FindByID(context.Context, int64) (domain.Request, error) {
	return stub.findResult, stub.findErr
}

func (stub *requestRepositoryStub) UpdateOwnedOpen(_ context.Context, _, _ int64, changes repository.RequestChanges) (domain.Request, error) {
	stub.updateChanges = changes
	return stub.updateResult, stub.updateErr
}

func (stub *requestRepositoryStub) DeleteOwnedOpen(context.Context, int64, int64) error {
	return stub.deleteErr
}

func (stub *requestRepositoryStub) UpdateStatus(_ context.Context, _ int64, status string) (domain.Request, error) {
	stub.statusValue = status
	if stub.statusErr != nil {
		return domain.Request{}, stub.statusErr
	}
	if stub.statusResult.Status != status {
		stub.statusResult.Status = status
		stub.statusResult.UpdatedAt = stub.statusResult.UpdatedAt.Add(time.Second)
	}
	return stub.statusResult, nil
}

func exampleRequest(ownerID int64, status string) domain.Request {
	return domain.Request{
		ID:          12,
		Title:       "Título atual",
		Description: "Descrição atual suficientemente longa.",
		Category:    domain.CategoryTI,
		Status:      status,
		Requester:   domain.User{ID: ownerID, Username: "autor", DisplayName: "Autor"},
		CreatedAt:   time.Date(2026, 10, 2, 10, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 10, 2, 10, 0, 0, 0, time.UTC),
	}
}

func TestCreateRequestTrimsAndCountsUnicodeRunes(t *testing.T) {
	created := exampleRequest(7, domain.StatusOpen)
	repositoryStub := &requestRepositoryStub{createResult: created}
	requestService := NewRequest(repositoryStub)

	result, err := requestService.Create(context.Background(), 7, CreateRequestInput{
		Title:       "  日本語  ",
		Description: "  áéíóúç🙂abc  ",
		Category:    " ti ",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if repositoryStub.createdRequesterID != 7 || repositoryStub.createdTitle != "日本語" || repositoryStub.createdDescription != "áéíóúç🙂abc" || repositoryStub.createdCategory != "ti" {
		t.Fatalf("repository input = %d %q %q %q", repositoryStub.createdRequesterID, repositoryStub.createdTitle, repositoryStub.createdDescription, repositoryStub.createdCategory)
	}
	if !result.CanEdit || !result.CanDelete {
		t.Fatalf("permissions = %+v", result)
	}
}

func TestCreateRequestRejectsUnicodeOutsideLimits(t *testing.T) {
	requestService := NewRequest(&requestRepositoryStub{})
	_, err := requestService.Create(context.Background(), 1, CreateRequestInput{
		Title:       strings.Repeat("🙂", 151),
		Description: strings.Repeat("界", 9),
		Category:    "juridico",
	})
	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("error = %v, want ValidationError", err)
	}
	for _, field := range []string{"title", "description", "category"} {
		if len(validationErr.Fields[field]) == 0 {
			t.Errorf("missing validation for %s: %v", field, validationErr.Fields)
		}
	}
}

func TestUpdateRequestDistinguishesAbsentNullEmptyAndPartial(t *testing.T) {
	tests := []struct {
		name      string
		input     UpdateRequestInput
		wantField string
	}{
		{name: "all absent", input: UpdateRequestInput{}, wantField: "body"},
		{name: "null", input: UpdateRequestInput{Title: OptionalString{Present: true, Null: true}}, wantField: "title"},
		{name: "empty", input: UpdateRequestInput{Description: OptionalString{Present: true, Value: "   "}}, wantField: "description"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			requestService := NewRequest(&requestRepositoryStub{})
			_, err := requestService.Update(context.Background(), 1, 1, test.input)
			var validationErr *ValidationError
			if !errors.As(err, &validationErr) || len(validationErr.Fields[test.wantField]) == 0 {
				t.Fatalf("error = %v fields=%v", err, validationErr)
			}
		})
	}

	repositoryStub := &requestRepositoryStub{updateResult: exampleRequest(1, domain.StatusOpen)}
	requestService := NewRequest(repositoryStub)
	_, err := requestService.Update(context.Background(), 1, 1, UpdateRequestInput{
		Category: OptionalString{Present: true, Value: " rh "},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !repositoryStub.updateChanges.CategorySet || repositoryStub.updateChanges.Category != "rh" || repositoryStub.updateChanges.TitleSet || repositoryStub.updateChanges.DescriptionSet {
		t.Fatalf("changes = %+v", repositoryStub.updateChanges)
	}
}

func TestRequestMutationErrorsAreMapped(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want error
	}{
		{name: "missing", err: repository.ErrNotFound, want: ErrRequestNotFound},
		{name: "other author", err: repository.ErrForbidden, want: ErrRequestForbidden},
		{name: "closed", err: repository.ErrConflict, want: ErrRequestConflict},
		{name: "unchanged", err: repository.ErrNoChanges, want: ErrRequestNoChanges},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repositoryStub := &requestRepositoryStub{updateErr: test.err, deleteErr: test.err}
			requestService := NewRequest(repositoryStub)
			_, updateErr := requestService.Update(context.Background(), 1, 1, UpdateRequestInput{
				Title: OptionalString{Present: true, Value: "Novo título"},
			})
			if !errors.Is(updateErr, test.want) {
				t.Errorf("Update() error = %v, want %v", updateErr, test.want)
			}
			if deleteErr := requestService.Delete(context.Background(), 1, 1); !errors.Is(deleteErr, test.want) {
				t.Errorf("Delete() error = %v, want %v", deleteErr, test.want)
			}
		})
	}
}

func TestGetRequestPermissionsDependOnAuthorAndOpenStatus(t *testing.T) {
	tests := []struct {
		name      string
		ownerID   int64
		status    string
		currentID int64
		allowed   bool
	}{
		{name: "open author", ownerID: 1, status: domain.StatusOpen, currentID: 1, allowed: true},
		{name: "other user", ownerID: 1, status: domain.StatusOpen, currentID: 2},
		{name: "closed author", ownerID: 1, status: domain.StatusCompleted, currentID: 1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			requestService := NewRequest(&requestRepositoryStub{findResult: exampleRequest(test.ownerID, test.status)})
			result, err := requestService.Get(context.Background(), 12, test.currentID)
			if err != nil {
				t.Fatal(err)
			}
			if result.CanEdit != test.allowed || result.CanDelete != test.allowed {
				t.Fatalf("permissions = edit:%v delete:%v", result.CanEdit, result.CanDelete)
			}
		})
	}
}

func TestUpdateRequestStatusAcceptsEveryTransitionAndRecalculatesPermissions(t *testing.T) {
	statuses := []string{domain.StatusOpen, domain.StatusInProgress, domain.StatusCompleted}
	for _, from := range statuses {
		for _, to := range statuses {
			t.Run(from+"_to_"+to, func(t *testing.T) {
				initial := exampleRequest(1, from)
				repositoryStub := &requestRepositoryStub{statusResult: initial}
				result, err := NewRequest(repositoryStub).UpdateStatus(context.Background(), 12, 1, UpdateRequestStatusInput{
					Status: OptionalString{Present: true, Value: to},
				})
				if err != nil {
					t.Fatalf("transition %s -> %s: %v", from, to, err)
				}
				if repositoryStub.statusValue != to || result.Request.Status != to {
					t.Fatalf("status sent/result = %q/%q, want %q", repositoryStub.statusValue, result.Request.Status, to)
				}
				wantUpdatedAt := initial.UpdatedAt
				if from != to {
					wantUpdatedAt = wantUpdatedAt.Add(time.Second)
				}
				if !result.Request.UpdatedAt.Equal(wantUpdatedAt) {
					t.Fatalf("updated_at = %s, want %s", result.Request.UpdatedAt, wantUpdatedAt)
				}
				wantEditable := to == domain.StatusOpen
				if result.CanEdit != wantEditable || result.CanDelete != wantEditable {
					t.Fatalf("permissions for %s = edit:%v delete:%v", to, result.CanEdit, result.CanDelete)
				}
			})
		}
	}

	otherUserRepository := &requestRepositoryStub{statusResult: exampleRequest(1, domain.StatusOpen)}
	result, err := NewRequest(otherUserRepository).UpdateStatus(context.Background(), 12, 2, UpdateRequestStatusInput{
		Status: OptionalString{Present: true, Value: domain.StatusCompleted},
	})
	if err != nil || result.Request.Status != domain.StatusCompleted {
		t.Fatalf("other user status result/error = %+v/%v", result, err)
	}
	if result.CanEdit || result.CanDelete {
		t.Fatalf("other user permissions = %+v", result)
	}
}

func TestUpdateRequestStatusValidatesPayloadAndMapsMissingResource(t *testing.T) {
	tests := []struct {
		name  string
		input UpdateRequestStatusInput
	}{
		{name: "missing"},
		{name: "null", input: UpdateRequestStatusInput{Status: OptionalString{Present: true, Null: true}}},
		{name: "unknown", input: UpdateRequestStatusInput{Status: OptionalString{Present: true, Value: "cancelado"}}},
		{name: "whitespace", input: UpdateRequestStatusInput{Status: OptionalString{Present: true, Value: " aberto "}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewRequest(&requestRepositoryStub{}).UpdateStatus(context.Background(), 12, 1, test.input)
			var validationErr *ValidationError
			if !errors.As(err, &validationErr) || len(validationErr.Fields["status"]) == 0 {
				t.Fatalf("error = %v, want status validation", err)
			}
		})
	}

	_, err := NewRequest(&requestRepositoryStub{statusErr: repository.ErrNotFound}).UpdateStatus(context.Background(), 999, 1, UpdateRequestStatusInput{
		Status: OptionalString{Present: true, Value: domain.StatusOpen},
	})
	if !errors.Is(err, ErrRequestNotFound) {
		t.Fatalf("error = %v, want ErrRequestNotFound", err)
	}
}
