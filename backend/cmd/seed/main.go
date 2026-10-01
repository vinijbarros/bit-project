package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"portal-solicitacoes/internal/config"
	"portal-solicitacoes/internal/repository/database"
	demoseed "portal-solicitacoes/internal/seed"
)

type options struct {
	resetPasswords bool
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger, os.Args[1:]); err != nil {
		logger.Error("demo seed failed", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger, args []string) error {
	options, err := parseOptions(args)
	if err != nil {
		return err
	}

	appConfig, err := config.Load()
	if err != nil {
		return fmt.Errorf("invalid application configuration: %w", err)
	}
	if appConfig.Environment == "production" {
		return fmt.Errorf("demo seed is disabled when APP_ENV=production")
	}
	seedConfig, err := demoseed.LoadConfigFromEnv()
	if err != nil {
		return fmt.Errorf("invalid demo seed configuration: %w", err)
	}

	db, err := database.Open(appConfig.Database)
	if err != nil {
		return fmt.Errorf("invalid database configuration")
	}
	defer db.Close()

	connectCtx, cancelConnect := context.WithTimeout(context.Background(), appConfig.Database.ConnectTimeout)
	if err := db.PingContext(connectCtx); err != nil {
		cancelConnect()
		return fmt.Errorf("database is unavailable")
	}
	cancelConnect()

	seedCtx, cancelSeed := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelSeed()
	result, err := demoseed.Run(seedCtx, db, seedConfig, demoseed.Options{ResetPasswords: options.resetPasswords})
	if err != nil {
		return err
	}

	logger.Info("demo seed completed",
		"users_created", result.UsersCreated,
		"users_existing", result.UsersExisting,
		"passwords_reset", result.PasswordsReset,
		"requests_created", result.RequestsCreated,
		"requests_existing", result.RequestsExisting,
	)
	return nil
}

func parseOptions(args []string) (options, error) {
	switch {
	case len(args) == 0:
		return options{}, nil
	case len(args) == 1 && args[0] == "--reset-passwords":
		return options{resetPasswords: true}, nil
	default:
		return options{}, fmt.Errorf("usage: go run ./cmd/seed [--reset-passwords]")
	}
}
