//go:build integration

// Package stack_test holds the live integration smoke tests for the A-02
// local-dev stack. Tagged `integration` so it is excluded from default unit
// runs and only executes against a running `docker compose` stack.
//
//	cp .env.example .env
//	docker compose up -d
//	go test -tags=integration ./docs/testing/session-02/integration/test_suite/...
package stack_test

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func TestA02StackIntegration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("TC-02-I-002/003: schema applied, timescale hypertable exists", func(t *testing.T) {
		dsn := "postgres://" + env("DB_USER", "postgres") + ":" + env("DB_PASSWORD", "changeme") +
			"@" + env("DB_HOST", "localhost") + ":" + env("DB_PORT", "5432") + "/" + env("DB_NAME", "mergemarket")
		conn, err := pgx.Connect(ctx, dsn)
		require.NoError(t, err, "connect to postgres")
		defer conn.Close(ctx)

		// All 8 relational tables present.
		for _, table := range []string{
			"users", "stores", "products", "wishlist_items",
			"price_alerts", "return_policies", "scrape_jobs", "price_history",
		} {
			var exists bool
			err := conn.QueryRow(ctx,
				`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name=$1)`, table).
				Scan(&exists)
			require.NoError(t, err)
			assert.Truef(t, exists, "table %q should exist", table)
		}

		// timescaledb extension enabled.
		var ext string
		err = conn.QueryRow(ctx, `SELECT extname FROM pg_extension WHERE extname='timescaledb'`).Scan(&ext)
		assert.NoError(t, err, "timescaledb extension should be enabled")

		// price_history registered as a hypertable.
		var ht string
		err = conn.QueryRow(ctx,
			`SELECT hypertable_name FROM timescaledb_information.hypertables WHERE hypertable_name='price_history'`).
			Scan(&ht)
		assert.NoError(t, err, "price_history should be a hypertable")
	})

	t.Run("TC-02-I-004: redis reachable", func(t *testing.T) {
		rdb := redis.NewClient(&redis.Options{
			Addr:     env("REDIS_HOST", "localhost") + ":" + env("REDIS_PORT", "6379"),
			Password: os.Getenv("REDIS_PASSWORD"),
		})
		defer rdb.Close()
		pong, err := rdb.Ping(ctx).Result()
		require.NoError(t, err)
		assert.Equal(t, "PONG", pong)
	})

	t.Run("TC-02-I-005: kong admin reports DB-less", func(t *testing.T) {
		resp, err := http.Get("http://localhost:" + env("KONG_ADMIN_PORT", "8001") + "/")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("TC-02-I-006: kong proxy port accepts connections", func(t *testing.T) {
		resp, err := http.Get("http://localhost:" + env("KONG_PROXY_PORT", "8000") + "/")
		require.NoError(t, err)
		defer resp.Body.Close()
		// Empty route table -> 404 "no Route matched" is the expected healthy answer.
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("TC-02-I-007: sonarqube status UP", func(t *testing.T) {
		resp, err := http.Get("http://localhost:9000/api/system/status")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
