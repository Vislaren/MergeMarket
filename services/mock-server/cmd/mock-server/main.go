// Command mock-server runs the MergeMarket local-dev mock API. It serves every
// contract in project_docs/api/API_CONTRACTS.md with static fixtures so the
// Flutter app and BFF can be developed offline. It is not deployed to any
// environment.
//
// Configuration is environment-driven (see internal/config). Run with:
//
//	go run ./cmd/mock-server          # listens on :8080
//	MOCK_SERVER_PORT=9090 go run ./cmd/mock-server
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Vislaren/MergeMarket/services/mock-server/internal/config"
	"github.com/Vislaren/MergeMarket/services/mock-server/internal/server"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	srv := server.New(cfg, logger)
	httpServer := &http.Server{
		Addr:              cfg.Addr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Run the server in the background so main can wait for a shutdown signal.
	serverErr := make(chan error, 1)
	go func() {
		logger.Info("mock-server listening",
			"addr", cfg.Addr,
			"service", config.ServiceName,
			"version", config.Version,
		)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// Wait for SIGINT/SIGTERM or a fatal server error.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		logger.Error("server error", "error", err)
		os.Exit(1)
	case sig := <-stop:
		logger.Info("shutdown signal received", "signal", sig.String())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
	logger.Info("mock-server stopped cleanly")
}
