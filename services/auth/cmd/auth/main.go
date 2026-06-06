// Command auth-service manages MergeMarket user registration, login, refresh
// tokens, JWT issuance, Redis-backed sessions, and GET /health.
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

	"github.com/Vislaren/MergeMarket/services/auth/internal/config"
	"github.com/Vislaren/MergeMarket/services/auth/internal/server"
	"github.com/Vislaren/MergeMarket/services/auth/internal/service"
	"github.com/Vislaren/MergeMarket/services/auth/internal/session"
	"github.com/Vislaren/MergeMarket/services/auth/internal/store"
	"github.com/Vislaren/MergeMarket/services/auth/internal/token"
)

// version is overridable at build time with -ldflags "-X main.version=...".
var version = "0.1.0"

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)
	if err := run(log); err != nil {
		log.Error("auth-service exited with error", "error", err)
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

	repo, err := store.NewPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer repo.Close()

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr, Password: cfg.RedisPassword, DB: cfg.RedisDB})
	defer func() {
		if cerr := rdb.Close(); cerr != nil {
			log.Error("closing redis client", "error", cerr)
		}
	}()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Warn("redis not reachable at startup; auth requests will retry per request", "error", err)
	}

	issuer := token.NewIssuer(cfg.JWTSecret, cfg.JWTTTL)
	sessions, err := session.NewRedis(rdb, cfg.SessionTTL, cfg.RefreshTokenTTL, cfg.EncryptionKey)
	if err != nil {
		return err
	}
	auth := service.New(repo, sessions, issuer, cfg.RefreshTokenTTL, log)

	srv := server.New(":"+cfg.Port, version, auth)
	srvErr := make(chan error, 1)
	go func() {
		log.Info("http server listening", "addr", srv.Addr)
		if err := srv.ListenAndServeTLS(cfg.TLSCertFile, cfg.TLSKeyFile); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
	log.Info("auth-service stopped cleanly")
	return nil
}
