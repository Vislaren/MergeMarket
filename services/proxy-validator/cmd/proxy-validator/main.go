// Command proxy-validator runs the MergeMarket proxy-validator service: it
// continuously scrapes public proxy lists, validates each proxy against a real
// endpoint with an adaptive politeness delay, and writes the working pool to
// Redis (proxy_pool, 5-minute TTL). It also serves GET /health on its port.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Vislaren/MergeMarket/services/proxy-validator/internal/config"
	"github.com/Vislaren/MergeMarket/services/proxy-validator/internal/fetcher"
	"github.com/Vislaren/MergeMarket/services/proxy-validator/internal/politeness"
	"github.com/Vislaren/MergeMarket/services/proxy-validator/internal/runner"
	"github.com/Vislaren/MergeMarket/services/proxy-validator/internal/server"
	"github.com/Vislaren/MergeMarket/services/proxy-validator/internal/store"
	"github.com/Vislaren/MergeMarket/services/proxy-validator/internal/validator"
)

// version is overridable at build time with -ldflags "-X main.version=...".
var version = "0.1.0"

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)

	if err := run(log); err != nil {
		log.Error("proxy-validator exited with error", "error", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	log.Info("configuration loaded",
		"port", cfg.Port, "redis", cfg.RedisAddr, "sources", len(cfg.Sources),
		"concurrency", cfg.Concurrency, "refresh", cfg.RefreshInterval.String())

	// Root context cancelled on SIGINT/SIGTERM for graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	redisStore := store.NewRedis(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB, cfg.ProxyPoolKey)
	defer func() {
		if cerr := redisStore.Close(); cerr != nil {
			log.Error("closing redis store", "error", cerr)
		}
	}()
	if perr := redisStore.Ping(ctx); perr != nil {
		// Log but continue: Redis may come up moments after this service.
		log.Warn("redis not reachable at startup; will retry on first cycle", "error", perr)
	}

	fetch := fetcher.New(cfg.Sources, nil, log)
	check := validator.New(cfg.TestURL, cfg.TestTimeout, nil)
	limiter := politeness.New(cfg.PolitenessMin, cfg.PolitenessMax, 0)
	r := runner.New(fetch, check, redisStore, limiter, cfg.Concurrency, cfg.ProxyPoolTTL, log)

	// Background validation loop.
	go r.Loop(ctx, cfg.RefreshInterval)

	// HTTP server for /health and /stats.
	srv := server.New(":"+cfg.Port, version, r)
	srvErr := make(chan error, 1)
	go func() {
		log.Info("health server listening", "addr", srv.Addr)
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
	log.Info("proxy-validator stopped cleanly")
	return nil
}
