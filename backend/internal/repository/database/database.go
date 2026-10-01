package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"

	"portal-solicitacoes/internal/config"
)

var ErrInvalidConfig = errors.New("invalid database configuration")

func Open(cfg config.DatabaseConfig) (*sql.DB, error) {
	connectionConfig, err := pgx.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("%w: DATABASE_URL cannot be parsed", ErrInvalidConfig)
	}

	db := stdlib.OpenDB(*connectionConfig)
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	return db, nil
}
