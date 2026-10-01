package config

import (
	"errors"
	"testing"
	"time"
)

func TestLoadUsesDefaults(t *testing.T) {
	setValidEnvironment(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Environment != "development" {
		t.Errorf("Environment = %q, want development", cfg.Environment)
	}
	if cfg.HTTP.Address != ":8080" {
		t.Errorf("HTTP.Address = %q, want :8080", cfg.HTTP.Address)
	}
	if cfg.Database.ConnectTimeout != 3*time.Second {
		t.Errorf("ConnectTimeout = %s, want 3s", cfg.Database.ConnectTimeout)
	}
	if cfg.Database.MaxOpenConns != 10 || cfg.Database.MaxIdleConns != 5 {
		t.Errorf("pool = %d/%d, want 10/5", cfg.Database.MaxOpenConns, cfg.Database.MaxIdleConns)
	}
	if len(cfg.TrustedOrigins) != 1 || cfg.TrustedOrigins[0] != "http://127.0.0.1:5173" {
		t.Errorf("TrustedOrigins = %v", cfg.TrustedOrigins)
	}
}

func TestLoadParsesAndDeduplicatesTrustedOrigins(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("TRUSTED_ORIGINS", "https://portal.example, http://127.0.0.1:5173,https://portal.example")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(cfg.TrustedOrigins) != 2 || cfg.TrustedOrigins[0] != "https://portal.example" || cfg.TrustedOrigins[1] != "http://127.0.0.1:5173" {
		t.Fatalf("TrustedOrigins = %v", cfg.TrustedOrigins)
	}
}

func TestLoadRejectsTrustedOriginWithPath(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("TRUSTED_ORIGINS", "https://portal.example/app")

	_, err := Load()
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("Load() error = %v, want ErrInvalid", err)
	}
}

func TestLoadRejectsMissingDatabaseURL(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("DATABASE_URL", "")

	_, err := Load()
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("Load() error = %v, want ErrInvalid", err)
	}
}

func TestLoadRejectsMalformedDatabaseConfiguration(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("DATABASE_URL", "https://localhost/portal")

	_, err := Load()
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("Load() error = %v, want ErrInvalid", err)
	}
}

func TestLoadRejectsInvalidDuration(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("DATABASE_CONNECT_TIMEOUT", "soon")

	_, err := Load()
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("Load() error = %v, want ErrInvalid", err)
	}
}

func TestLoadRejectsPoolWithTooManyIdleConnections(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("DATABASE_MAX_OPEN_CONNS", "2")
	t.Setenv("DATABASE_MAX_IDLE_CONNS", "3")

	_, err := Load()
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("Load() error = %v, want ErrInvalid", err)
	}
}

func setValidEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("APP_ENV", "development")
	t.Setenv("HTTP_ADDR", ":8080")
	t.Setenv("DATABASE_URL", "postgres://portal:portal@localhost:5432/portal?sslmode=disable")
	t.Setenv("DATABASE_CONNECT_TIMEOUT", "3s")
	t.Setenv("DATABASE_MAX_OPEN_CONNS", "10")
	t.Setenv("DATABASE_MAX_IDLE_CONNS", "5")
	t.Setenv("DATABASE_CONN_MAX_LIFETIME", "30m")
	t.Setenv("DATABASE_CONN_MAX_IDLE_TIME", "5m")
	t.Setenv("HTTP_READ_HEADER_TIMEOUT", "5s")
	t.Setenv("HTTP_READ_TIMEOUT", "10s")
	t.Setenv("HTTP_WRITE_TIMEOUT", "15s")
	t.Setenv("HTTP_IDLE_TIMEOUT", "60s")
	t.Setenv("SHUTDOWN_TIMEOUT", "10s")
	t.Setenv("MIGRATIONS_DIR", "db/migrations")
	t.Setenv("TRUSTED_ORIGINS", "http://127.0.0.1:5173")
}
