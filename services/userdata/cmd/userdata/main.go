// Command userdata-service serves the JWT-protected wishlist, price-alert, and
// savings-dashboard APIs (API_CONTRACTS.md), plus GET /health.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Vislaren/MergeMarket/services/userdata/internal/config"
	"github.com/Vislaren/MergeMarket/services/userdata/internal/server"
	"github.com/Vislaren/MergeMarket/services/userdata/internal/store"
	"github.com/Vislaren/MergeMarket/services/userdata/internal/token"
)

// version is overridable at build time with -ldflags "-X main.version=...".
var version = "0.1.0"

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)

	if err := run(log); err != nil {
		log.Error("userdata-service exited with error", "error", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// PostgreSQL is a hard dependency — it is the source of truth for all
	// per-user data. Fail fast if unreachable.
	repo, err := store.NewPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := repo.Close(); cerr != nil {
			log.Error("closing db pool", "error", cerr)
		}
	}()

	// Best-effort: create the purchases table this service owns if a pre-existing
	// database predates the savings dashboard. A failure here (e.g. missing
	// privileges) is logged, not fatal — the canonical 01-schema.sql is the
	// primary creator.
	if err := repo.EnsureSchema(ctx); err != nil {
		log.Warn("could not ensure purchases schema; assuming it already exists", "error", err)
	}

	verifier := token.NewVerifier(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTLeeway)

	srv := server.New(":"+cfg.Port, version, repo, verifier)
	srvErr := make(chan error, 1)
	go func() {
		log.Info("http server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			srvErr <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Info("shutdown signal received")
	case err := <-srvErr:
		stop()
		return err
	}

	if err := server.Shutdown(srv); err != nil {
		return err
	}
	log.Info("userdata-service stopped cleanly")
	return nil
}
