// Command history-service records price snapshots into TimescaleDB, runs a
// scheduled heartbeat that re-checks followed products and fires price-drop
// alerts on a downward threshold crossing, and serves the product price-history
// API plus GET /health and /stats.
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

	"github.com/Vislaren/MergeMarket/services/history/internal/alert"
	"github.com/Vislaren/MergeMarket/services/history/internal/config"
	"github.com/Vislaren/MergeMarket/services/history/internal/pricesource"
	"github.com/Vislaren/MergeMarket/services/history/internal/runner"
	"github.com/Vislaren/MergeMarket/services/history/internal/server"
	"github.com/Vislaren/MergeMarket/services/history/internal/store"
)

// version is overridable at build time with -ldflags "-X main.version=...".
var version = "0.1.0"

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)

	if err := run(log); err != nil {
		log.Error("history-service exited with error", "error", err)
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

	// PostgreSQL/TimescaleDB is a hard dependency — it is both the source of
	// followed products and the sink for snapshots. Fail fast if unreachable.
	repo, err := store.NewPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := repo.Close(); cerr != nil {
			log.Error("closing db pool", "error", cerr)
		}
	}()

	// Redis carries fired alert events to the Notification Worker.
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
		log.Warn("redis not reachable at startup; alert publishing will retry per cycle", "error", perr)
	}

	var prices pricesource.Source
	switch cfg.HeartbeatMode {
	case config.ModeHTTP:
		prices = pricesource.NewHTTP(cfg.HeartbeatTimeout)
	default:
		prices = pricesource.DBSource{}
	}

	publisher := alert.NewRedisPublisher(rdb, cfg.AlertQueueKey)
	run := runner.New(repo, prices, publisher, cfg.SnapshotInterval, cfg.HeartbeatInterval, log)

	log.Info("starting history-service",
		"port", cfg.Port, "db", "configured", "redis", cfg.RedisAddr,
		"snapshot_interval", cfg.SnapshotInterval.String(),
		"heartbeat_interval", cfg.HeartbeatInterval.String(),
		"heartbeat_mode", string(cfg.HeartbeatMode), "alert_queue", cfg.AlertQueueKey)

	// Scheduled jobs.
	runnerDone := make(chan struct{})
	go func() {
		run.Run(ctx, cfg.SnapshotOnStart, cfg.HeartbeatOnStart)
		close(runnerDone)
	}()

	// HTTP server (API + health + stats).
	srv := server.New(":"+cfg.Port, version, repo, run)
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
		<-runnerDone
		return err
	}

	if err := server.Shutdown(srv); err != nil {
		return err
	}
	<-runnerDone
	log.Info("history-service stopped cleanly")
	return nil
}
