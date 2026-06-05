// Command normalization-service consumes raw scrape results from the
// normalization queue, converts each product to the canonical MergeMarket
// schema, injects retailer-specific affiliate links, and upserts the products
// into PostgreSQL. It serves GET /health and GET /stats.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"

	"github.com/Vislaren/MergeMarket/services/normalization/internal/affiliate"
	"github.com/Vislaren/MergeMarket/services/normalization/internal/config"
	"github.com/Vislaren/MergeMarket/services/normalization/internal/queue"
	"github.com/Vislaren/MergeMarket/services/normalization/internal/server"
	"github.com/Vislaren/MergeMarket/services/normalization/internal/store"
	"github.com/Vislaren/MergeMarket/services/normalization/internal/worker"
)

// version is overridable at build time with -ldflags "-X main.version=...".
var version = "0.1.0"

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)

	if err := run(log); err != nil {
		log.Error("normalization-service exited with error", "error", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	injector, err := affiliate.Load(cfg.AffiliateConfigPath)
	if err != nil {
		return err
	}

	// Root context cancelled on SIGINT/SIGTERM for graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// PostgreSQL is a hard dependency for this service — without it there is
	// nowhere to write normalized products, so fail fast if it is unreachable.
	repo, err := store.NewPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := repo.Close(); cerr != nil {
			log.Error("closing db pool", "error", cerr)
		}
	}()
	if err := repo.EnsureSchema(ctx); err != nil {
		return err
	}

	// Redis client for the input queue.
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	defer func() {
		if cerr := rdb.Close(); cerr != nil {
			log.Error("closing redis client", "error", cerr)
		}
	}()
	if perr := rdb.Ping(ctx).Err(); perr != nil {
		// Log but continue: Redis may come up moments after this service.
		log.Warn("redis not reachable at startup; workers will retry", "error", perr)
	}

	src := queue.NewRedisWithClient(rdb, cfg.NormalizeQueueKey, cfg.QueuePollTimeout)
	pool := worker.New(src, repo, injector, cfg.Workers, log)

	log.Info("starting normalization-service",
		"port", cfg.Port, "workers", cfg.Workers, "redis", cfg.RedisAddr,
		"normalize_queue", cfg.NormalizeQueueKey, "affiliate_config", cfg.AffiliateConfigPath)

	// Worker pool.
	poolDone := make(chan struct{})
	go func() {
		pool.Run(ctx)
		close(poolDone)
	}()

	// HTTP server for /health and /stats.
	srv := server.New(":"+cfg.Port, version, pool)
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
		<-poolDone
		return err
	}

	if err := server.Shutdown(srv); err != nil {
		return err
	}
	<-poolDone
	log.Info("normalization-service stopped cleanly")
	return nil
}
