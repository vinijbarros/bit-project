package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"portal-solicitacoes/internal/domain"
)

var ErrNotFound = errors.New("repository record not found")

type Auth struct {
	db *sql.DB
}

func NewAuth(db *sql.DB) *Auth {
	return &Auth{db: db}
}

func (repository *Auth) FindUserCredentials(ctx context.Context, username string) (domain.UserCredentials, error) {
	var credentials domain.UserCredentials
	err := repository.db.QueryRowContext(ctx, `
		SELECT id, username, display_name, password_hash
		FROM users
		WHERE username = $1
	`, username).Scan(
		&credentials.ID,
		&credentials.Username,
		&credentials.DisplayName,
		&credentials.PasswordHash,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.UserCredentials{}, ErrNotFound
	}
	if err != nil {
		return domain.UserCredentials{}, fmt.Errorf("find user credentials: %w", err)
	}
	return credentials, nil
}

func (repository *Auth) RotateSession(ctx context.Context, previousTokenHash string, session domain.Session) (err error) {
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin session rotation: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if previousTokenHash != "" {
		if _, err = tx.ExecContext(ctx, `DELETE FROM sessions WHERE token_hash = $1`, previousTokenHash); err != nil {
			return fmt.Errorf("revoke previous session: %w", err)
		}
	}
	if _, err = tx.ExecContext(ctx, `
		INSERT INTO sessions (token_hash, user_id, created_at, expires_at)
		VALUES ($1, $2, $3, $4)
	`, session.TokenHash, session.UserID, session.CreatedAt, session.ExpiresAt); err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit session rotation: %w", err)
	}
	return nil
}

func (repository *Auth) FindUserBySession(ctx context.Context, tokenHash string) (domain.User, error) {
	var user domain.User
	err := repository.db.QueryRowContext(ctx, `
		SELECT users.id, users.username, users.display_name
		FROM sessions
		JOIN users ON users.id = sessions.user_id
		WHERE sessions.token_hash = $1
		  AND sessions.expires_at > CURRENT_TIMESTAMP
	`, tokenHash).Scan(&user.ID, &user.Username, &user.DisplayName)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.User{}, ErrNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("find session user: %w", err)
	}
	return user, nil
}

func (repository *Auth) RevokeSession(ctx context.Context, tokenHash string) error {
	if _, err := repository.db.ExecContext(ctx, `DELETE FROM sessions WHERE token_hash = $1`, tokenHash); err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
}
