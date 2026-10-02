package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"portal-solicitacoes/internal/domain"
)

var (
	ErrForbidden = errors.New("repository operation forbidden")
	ErrConflict  = errors.New("repository state conflict")
	ErrNoChanges = errors.New("repository no changes")
)

type Request struct {
	db *sql.DB
}

type RequestChanges struct {
	TitleSet       bool
	Title          string
	DescriptionSet bool
	Description    string
	CategorySet    bool
	Category       string
}

func NewRequest(db *sql.DB) *Request {
	return &Request{db: db}
}

func (repository *Request) Create(ctx context.Context, requesterID int64, title, description, category string) (domain.Request, error) {
	row := repository.db.QueryRowContext(ctx, `
		WITH inserted AS (
			INSERT INTO requests (
				title, description, category, status, requester_id, created_at, updated_at
			)
			VALUES ($1, $2, $3, 'aberto', $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			RETURNING id, title, description, category, status, requester_id, created_at, updated_at
		)
		SELECT
			inserted.id, inserted.title, inserted.description, inserted.category, inserted.status,
			users.id, users.username, users.display_name,
			inserted.created_at, inserted.updated_at
		FROM inserted
		JOIN users ON users.id = inserted.requester_id
	`, title, description, category, requesterID)

	created, err := scanRequest(row)
	if err != nil {
		return domain.Request{}, fmt.Errorf("create request: %w", err)
	}
	return created, nil
}

func (repository *Request) FindByID(ctx context.Context, id int64) (domain.Request, error) {
	row := repository.db.QueryRowContext(ctx, `
		SELECT
			requests.id, requests.title, requests.description, requests.category, requests.status,
			users.id, users.username, users.display_name,
			requests.created_at, requests.updated_at
		FROM requests
		JOIN users ON users.id = requests.requester_id
		WHERE requests.id = $1
	`, id)

	found, err := scanRequest(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Request{}, ErrNotFound
	}
	if err != nil {
		return domain.Request{}, fmt.Errorf("find request: %w", err)
	}
	return found, nil
}

func (repository *Request) UpdateOwnedOpen(ctx context.Context, id, requesterID int64, changes RequestChanges) (domain.Request, error) {
	row := repository.db.QueryRowContext(ctx, `
		WITH updated AS (
			UPDATE requests
			SET
				title = CASE WHEN $3 THEN $4 ELSE title END,
				description = CASE WHEN $5 THEN $6 ELSE description END,
				category = CASE WHEN $7 THEN $8 ELSE category END,
				updated_at = CURRENT_TIMESTAMP
			WHERE id = $1
			  AND requester_id = $2
			  AND status = 'aberto'
			  AND (
				($3 AND title IS DISTINCT FROM $4)
				OR ($5 AND description IS DISTINCT FROM $6)
				OR ($7 AND category IS DISTINCT FROM $8)
			  )
			RETURNING id, title, description, category, status, requester_id, created_at, updated_at
		)
		SELECT
			updated.id, updated.title, updated.description, updated.category, updated.status,
			users.id, users.username, users.display_name,
			updated.created_at, updated.updated_at
		FROM updated
		JOIN users ON users.id = updated.requester_id
	`, id, requesterID,
		changes.TitleSet, changes.Title,
		changes.DescriptionSet, changes.Description,
		changes.CategorySet, changes.Category,
	)

	updated, err := scanRequest(row)
	if err == nil {
		return updated, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return domain.Request{}, fmt.Errorf("update request: %w", err)
	}
	return domain.Request{}, repository.classifyMutationFailure(ctx, id, requesterID, changes)
}

func (repository *Request) DeleteOwnedOpen(ctx context.Context, id, requesterID int64) error {
	result, err := repository.db.ExecContext(ctx, `
		DELETE FROM requests
		WHERE id = $1 AND requester_id = $2 AND status = 'aberto'
	`, id, requesterID)
	if err != nil {
		return fmt.Errorf("delete request: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete request result: %w", err)
	}
	if rows == 1 {
		return nil
	}
	return repository.classifyMutationFailure(ctx, id, requesterID, RequestChanges{})
}

func (repository *Request) UpdateStatus(ctx context.Context, id int64, status string) (result domain.Request, err error) {
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Request{}, fmt.Errorf("begin request status update: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	row := tx.QueryRowContext(ctx, `
		SELECT
			requests.id, requests.title, requests.description, requests.category, requests.status,
			users.id, users.username, users.display_name,
			requests.created_at, requests.updated_at
		FROM requests
		JOIN users ON users.id = requests.requester_id
		WHERE requests.id = $1
		FOR UPDATE OF requests
	`, id)
	result, err = scanRequest(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Request{}, ErrNotFound
	}
	if err != nil {
		return domain.Request{}, fmt.Errorf("lock request for status update: %w", err)
	}

	if result.Status != status {
		err = tx.QueryRowContext(ctx, `
			UPDATE requests
			SET status = $2, updated_at = CURRENT_TIMESTAMP
			WHERE id = $1
			RETURNING status, updated_at
		`, id, status).Scan(&result.Status, &result.UpdatedAt)
		if err != nil {
			return domain.Request{}, fmt.Errorf("update request status: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return domain.Request{}, fmt.Errorf("commit request status update: %w", err)
	}
	return result, nil
}

func (repository *Request) classifyMutationFailure(ctx context.Context, id, requesterID int64, changes RequestChanges) error {
	var ownerID int64
	var status, title, description, category string
	err := repository.db.QueryRowContext(ctx, `
		SELECT requester_id, status, title, description, category
		FROM requests
		WHERE id = $1
	`, id).Scan(&ownerID, &status, &title, &description, &category)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("classify request mutation: %w", err)
	}
	if ownerID != requesterID {
		return ErrForbidden
	}
	if status != domain.StatusOpen {
		return ErrConflict
	}
	if changes.TitleSet || changes.DescriptionSet || changes.CategorySet {
		return ErrNoChanges
	}
	return ErrConflict
}

type rowScanner interface {
	Scan(...any) error
}

func scanRequest(row rowScanner) (domain.Request, error) {
	var request domain.Request
	err := row.Scan(
		&request.ID,
		&request.Title,
		&request.Description,
		&request.Category,
		&request.Status,
		&request.Requester.ID,
		&request.Requester.Username,
		&request.Requester.DisplayName,
		&request.CreatedAt,
		&request.UpdatedAt,
	)
	return request, err
}
