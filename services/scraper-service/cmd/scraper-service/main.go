// Command scraper-service runs the MergeMarket config-driven worker-queue
// scraper. It loads per-store StoreConfig files, consumes scrape jobs from a
// Redis list, scrapes each store (optionally through a validated proxy),
// applies a per-store circuit breaker on 403/429 responses, and publishes raw
// results to the normalization queue. It serves GET /health and /stats.
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

	"github.com/Vislaren/MergeMarket/services/scraper-service/internal/circuitbreaker"
	"github.com/Vislaren/MergeMarket/services/scraper-service/internal/config"
	"github.com/Vislaren/MergeMarket/services/scraper-service/internal/proxypool"
	"github.com/Vislaren/MergeMarket/services/scraper-service/internal/queue"
	"github.com/Vislaren/MergeMarket/services/scraper-service/internal/scraper"
	"github.com/Vislaren/MergeMarket/services/scraper-service/internal/server"
	"github.com/Vislaren/MergeMarket/services/scraper-service/internal/storeconfig"
	"github.com/Vislaren/MergeMarket/services/scraper-service/internal/worker"
)

// version is overridable at build time with -ldflags "-X main.version=...".
var version = "0.1.0"

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)

	if err := run(log); err != nil {
		log.Error("scraper-service exited with error", "error", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	registry, err := storeconfig.LoadDir(cfg.ConfigDir)
	if err != nil {
		return err
	}
	log.Info("store configs loaded", "count", registry.Len(), "stores", registry.IDs())

	// Root context cancelled on SIGINT/SIGTERM for graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// One Redis client shared by the queue and the proxy pool reader.
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

	jobQueue := queue.NewRedisWithClient(rdb, cfg.JobQueueKey, cfg.NormalizeQueueKey, cfg.QueuePollTimeout)

	var proxies proxypool.Provider
	if cfg.UseProxy {
		proxies = proxypool.NewRedis(rdb, cfg.ProxyPoolKey)
	} else {
		proxies = proxypool.Static{} // empty addr => direct connections
	}

	engine := scraper.New(proxies, cfg.ScrapeTimeout, nil)
	breakers := circuitbreaker.NewGroup(cfg.CircuitThreshold, cfg.CircuitCooldown)
	pool := worker.New(jobQueue, jobQueue, registry, engine, breakers, cfg.Workers, log)

	log.Info("starting scraper-service",
		"port", cfg.Port, "workers", cfg.Workers, "redis", cfg.RedisAddr,
		"job_queue", cfg.JobQueueKey, "normalize_queue", cfg.NormalizeQueueKey,
		"use_proxy", cfg.UseProxy, "circuit_threshold", cfg.CircuitThreshold)

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
	log.Info("scraper-service stopped cleanly")
	return nil
}
