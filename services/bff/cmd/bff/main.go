// Command bff runs the MergeMarket Backend-for-Frontend (B-09). It forwards most
// client read endpoints to the upstream API and serves one aggregated
// product-detail view shaped for the Flutter app.
//
// Configuration is environment-driven (see internal/config). Run with:
//
//	go run ./cmd/bff                                    # listens on :8082
//	BFF_PORT=9000 BFF_UPSTREAM_URL=http://kong:8088 go run ./cmd/bff
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

	"github.com/Vislaren/MergeMarket/services/bff/internal/config"
	"github.com/Vislaren/MergeMarket/services/bff/internal/server"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	srv, err := server.New(cfg, logger)
	if err != nil {
		logger.Error("failed to build server", "error", err)
		os.Exit(1)
	}

	httpServer := &http.Server{
		Addr:              cfg.Addr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("bff listening",
			"addr", cfg.Addr,
			"service", config.ServiceName,
			"version", config.Version,
			"upstream", cfg.UpstreamURL,
		)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

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
	logger.Info("bff stopped cleanly")
}
