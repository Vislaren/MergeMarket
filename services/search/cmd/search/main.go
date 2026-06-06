// Command search-service serves GET /api/v1/search from normalized products in
// PostgreSQL with a stale-while-revalidate Redis cache, plus GET /health.
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

	"github.com/Vislaren/MergeMarket/services/search/internal/cache"
	"github.com/Vislaren/MergeMarket/services/search/internal/config"
	"github.com/Vislaren/MergeMarket/services/search/internal/search"
	"github.com/Vislaren/MergeMarket/services/search/internal/server"
	"github.com/Vislaren/MergeMarket/services/search/internal/store"
)

// version is overridable at build time with -ldflags "-X main.version=...".
var version = "0.1.0"

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)

	if err := run(log); err != nil {
		log.Error("search-service exited with error", "error", err)
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

	// PostgreSQL is the source of normalized products — fail fast if unreachable.
	repo, err := store.NewPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := repo.Close(); cerr != nil {
			log.Error("closing db pool", "error", cerr)
		}
	}()

	// Redis caches aggregated results. It is best-effort: if it is down the
	// service degrades to always querying the database.
	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr, Password: cfg.RedisPassword, DB: cfg.RedisDB})
	defer func() {
		if cerr := rdb.Close(); cerr != nil {
			log.Error("closing redis client", "error", cerr)
		}
	}()
	if perr := rdb.Ping(ctx).Err(); perr != nil {
		log.Warn("redis not reachable at startup; serving uncached until it recovers", "error", perr)
	}

	svc := search.New(repo, cache.NewRedis(rdb), search.Config{
		CachePrefix:     cfg.CachePrefix,
		CacheTTL:        cfg.CacheTTL,
		CacheStaleAfter: cfg.CacheStaleAfter,
		MaxResults:      cfg.MaxResults,
	}, log)

	srv := server.New(":"+cfg.Port, version, svc)
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
	log.Info("search-service stopped cleanly")
	return nil
}
