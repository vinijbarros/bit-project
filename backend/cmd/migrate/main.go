package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/pressly/goose/v3"

	"portal-solicitacoes/internal/config"
	"portal-solicitacoes/internal/repository/database"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger, os.Args[1:]); err != nil {
		logger.Error("migration command failed", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: go run ./cmd/migrate [up|down|status|version]")
	}

	action := strings.ToLower(args[0])
	if action != "up" && action != "down" && action != "status" && action != "version" {
		return fmt.Errorf("unsupported migration action %q", action)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("invalid application configuration: %w", err)
	}

	db, err := database.Open(cfg.Database)
	if err != nil {
		return fmt.Errorf("invalid database configuration")
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Database.ConnectTimeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("database is unavailable")
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("configure goose dialect: %w", err)
	}

	directory := cfg.MigrationsDirectory
	switch action {
	case "up":
		err = goose.Up(db, directory)
	case "down":
		err = goose.Down(db, directory)
	case "status":
		err = goose.Status(db, directory)
	case "version":
		err = goose.Version(db, directory)
	}
	if err != nil {
		return fmt.Errorf("goose %s failed: %w", action, err)
	}

	logger.Info("migration command completed", "action", action)
	return nil
}
