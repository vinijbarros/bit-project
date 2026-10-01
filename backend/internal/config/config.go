package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var ErrInvalid = errors.New("invalid configuration")

type Config struct {
	Environment         string
	HTTP                HTTPConfig
	Database            DatabaseConfig
	Auth                AuthConfig
	TrustedOrigins      []string
	ShutdownTimeout     time.Duration
	MigrationsDirectory string
}

type HTTPConfig struct {
	Address           string
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
}

type DatabaseConfig struct {
	URL             string
	ConnectTimeout  time.Duration
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type AuthConfig struct {
	SessionDuration     time.Duration
	SessionTokenBytes   int
	CookieName          string
	CookieSecure        bool
	LoginMaxAttempts    int
	LoginWindow         time.Duration
	LoginLimiterEntries int
}

func Load() (Config, error) {
	environment := valueOrDefault("APP_ENV", "development")
	if environment != "development" && environment != "test" && environment != "production" {
		return Config{}, invalid("APP_ENV must be development, test, or production")
	}

	address := valueOrDefault("HTTP_ADDR", ":8080")
	if _, _, err := net.SplitHostPort(address); err != nil {
		return Config{}, invalid("HTTP_ADDR must include a valid host and port")
	}

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return Config{}, invalid("DATABASE_URL is required")
	}
	if err := validateDatabaseURL(databaseURL); err != nil {
		return Config{}, err
	}

	connectTimeout, err := positiveDuration("DATABASE_CONNECT_TIMEOUT", 3*time.Second)
	if err != nil {
		return Config{}, err
	}
	maxOpenConns, err := positiveInt("DATABASE_MAX_OPEN_CONNS", 10)
	if err != nil {
		return Config{}, err
	}
	maxIdleConns, err := nonNegativeInt("DATABASE_MAX_IDLE_CONNS", 5)
	if err != nil {
		return Config{}, err
	}
	if maxIdleConns > maxOpenConns {
		return Config{}, invalid("DATABASE_MAX_IDLE_CONNS cannot exceed DATABASE_MAX_OPEN_CONNS")
	}
	connMaxLifetime, err := positiveDuration("DATABASE_CONN_MAX_LIFETIME", 30*time.Minute)
	if err != nil {
		return Config{}, err
	}
	connMaxIdleTime, err := positiveDuration("DATABASE_CONN_MAX_IDLE_TIME", 5*time.Minute)
	if err != nil {
		return Config{}, err
	}
	readHeaderTimeout, err := positiveDuration("HTTP_READ_HEADER_TIMEOUT", 5*time.Second)
	if err != nil {
		return Config{}, err
	}
	readTimeout, err := positiveDuration("HTTP_READ_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}
	writeTimeout, err := positiveDuration("HTTP_WRITE_TIMEOUT", 15*time.Second)
	if err != nil {
		return Config{}, err
	}
	idleTimeout, err := positiveDuration("HTTP_IDLE_TIMEOUT", 60*time.Second)
	if err != nil {
		return Config{}, err
	}
	shutdownTimeout, err := positiveDuration("SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}

	migrationsDirectory := valueOrDefault("MIGRATIONS_DIR", "db/migrations")
	if strings.TrimSpace(migrationsDirectory) == "" {
		return Config{}, invalid("MIGRATIONS_DIR cannot be empty")
	}
	trustedOrigins, err := parseTrustedOrigins(valueOrDefault("TRUSTED_ORIGINS", "http://127.0.0.1:5173"))
	if err != nil {
		return Config{}, err
	}
	sessionDuration, err := positiveDuration("SESSION_DURATION", 8*time.Hour)
	if err != nil {
		return Config{}, err
	}
	sessionTokenBytes, err := positiveInt("SESSION_TOKEN_BYTES", 32)
	if err != nil || sessionTokenBytes < 32 || sessionTokenBytes > 128 {
		return Config{}, invalid("SESSION_TOKEN_BYTES must be between 32 and 128")
	}
	cookieSecure, err := boolean("SESSION_COOKIE_SECURE", environment == "production")
	if err != nil {
		return Config{}, err
	}
	if environment == "production" && !cookieSecure {
		return Config{}, invalid("SESSION_COOKIE_SECURE must be true in production")
	}
	loginMaxAttempts, err := positiveInt("LOGIN_RATE_LIMIT_MAX_ATTEMPTS", 5)
	if err != nil {
		return Config{}, err
	}
	loginWindow, err := positiveDuration("LOGIN_RATE_LIMIT_WINDOW", 15*time.Minute)
	if err != nil {
		return Config{}, err
	}
	loginLimiterEntries, err := positiveInt("LOGIN_RATE_LIMIT_MAX_ENTRIES", 10000)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Environment: environment,
		HTTP: HTTPConfig{
			Address:           address,
			ReadHeaderTimeout: readHeaderTimeout,
			ReadTimeout:       readTimeout,
			WriteTimeout:      writeTimeout,
			IdleTimeout:       idleTimeout,
		},
		Database: DatabaseConfig{
			URL:             databaseURL,
			ConnectTimeout:  connectTimeout,
			MaxOpenConns:    maxOpenConns,
			MaxIdleConns:    maxIdleConns,
			ConnMaxLifetime: connMaxLifetime,
			ConnMaxIdleTime: connMaxIdleTime,
		},
		Auth: AuthConfig{
			SessionDuration:     sessionDuration,
			SessionTokenBytes:   sessionTokenBytes,
			CookieName:          "portal_session",
			CookieSecure:        cookieSecure,
			LoginMaxAttempts:    loginMaxAttempts,
			LoginWindow:         loginWindow,
			LoginLimiterEntries: loginLimiterEntries,
		},
		TrustedOrigins:      trustedOrigins,
		ShutdownTimeout:     shutdownTimeout,
		MigrationsDirectory: migrationsDirectory,
	}, nil
}

func boolean(name string, fallback bool) (bool, error) {
	raw := valueOrDefault(name, strconv.FormatBool(fallback))
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, invalid(name + " must be true or false")
	}
	return value, nil
}

func parseTrustedOrigins(raw string) ([]string, error) {
	seen := make(map[string]struct{})
	origins := make([]string, 0)
	for _, item := range strings.Split(raw, ",") {
		origin := strings.TrimSpace(item)
		parsed, err := url.Parse(origin)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || parsed.User != nil || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
			return nil, invalid("TRUSTED_ORIGINS must contain comma-separated HTTP(S) origins without paths")
		}
		if _, exists := seen[origin]; exists {
			continue
		}
		seen[origin] = struct{}{}
		origins = append(origins, origin)
	}
	if len(origins) == 0 {
		return nil, invalid("TRUSTED_ORIGINS must contain at least one origin")
	}
	return origins, nil
}

func validateDatabaseURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return invalid("DATABASE_URL is malformed")
	}
	if parsed.Scheme != "postgres" && parsed.Scheme != "postgresql" {
		return invalid("DATABASE_URL must use postgres or postgresql scheme")
	}
	if parsed.Hostname() == "" {
		return invalid("DATABASE_URL must include a host")
	}
	if strings.Trim(parsed.Path, "/") == "" {
		return invalid("DATABASE_URL must include a database name")
	}
	return nil
}

func positiveDuration(name string, fallback time.Duration) (time.Duration, error) {
	raw := valueOrDefault(name, fallback.String())
	value, err := time.ParseDuration(raw)
	if err != nil || value <= 0 {
		return 0, invalid(name + " must be a positive duration")
	}
	return value, nil
}

func positiveInt(name string, fallback int) (int, error) {
	value, err := integer(name, fallback)
	if err != nil || value <= 0 {
		return 0, invalid(name + " must be a positive integer")
	}
	return value, nil
}

func nonNegativeInt(name string, fallback int) (int, error) {
	value, err := integer(name, fallback)
	if err != nil || value < 0 {
		return 0, invalid(name + " must be a non-negative integer")
	}
	return value, nil
}

func integer(name string, fallback int) (int, error) {
	raw := valueOrDefault(name, strconv.Itoa(fallback))
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func valueOrDefault(name, fallback string) string {
	if value, ok := os.LookupEnv(name); ok {
		return strings.TrimSpace(value)
	}
	return fallback
}

func invalid(message string) error {
	return fmt.Errorf("%w: %s", ErrInvalid, message)
}
