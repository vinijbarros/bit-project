package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"portal-solicitacoes/internal/config"
	httpapi "portal-solicitacoes/internal/http"
	"portal-solicitacoes/internal/repository/database"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("invalid application configuration", "error", err)
		os.Exit(1)
	}

	db, err := database.Open(cfg.Database)
	if err != nil {
		logger.Error("invalid database configuration")
		os.Exit(1)
	}
	defer db.Close()

	startupContext, cancelStartup := context.WithTimeout(context.Background(), cfg.Database.ConnectTimeout)
	if err := db.PingContext(startupContext); err != nil {
		logger.Warn("database unavailable at startup; readiness will remain unavailable")
	}
	cancelStartup()

	server := &http.Server{
		Addr:              cfg.HTTP.Address,
		Handler:           httpapi.NewRouter(db, cfg.Database.ConnectTimeout),
		ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("http server starting", "address", cfg.HTTP.Address, "environment", cfg.Environment)
		serverErrors <- server.ListenAndServe()
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signals)

	select {
	case received := <-signals:
		logger.Info("shutdown signal received", "signal", received.String())
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
		return
	}

	shutdownContext, cancelShutdown := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancelShutdown()

	if err := server.Shutdown(shutdownContext); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("http server stopped")
}
