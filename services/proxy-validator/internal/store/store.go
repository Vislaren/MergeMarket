// Package store persists the validated proxy pool to Redis. The pool is the
// `proxy_pool` Set defined in DATABASE_SCHEMA.md §3 (TTL 5 minutes); downstream
// the scraper service (A-05) reads working proxies from this Set.
package store

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Store is the persistence surface the runner depends on. Keeping it as an
// interface lets the orchestration logic be unit-tested without a live Redis.
type Store interface {
	// Replace atomically swaps the entire proxy pool with members and sets ttl.
	// An empty members slice clears the pool.
	Replace(ctx context.Context, members []string, ttl time.Duration) error
	// Ping verifies connectivity (used by the readiness side of /health).
	Ping(ctx context.Context) error
	// Close releases the underlying client.
	Close() error
}

// RedisStore is the production Store backed by go-redis.
type RedisStore struct {
	client *redis.Client
	key    string
}

// NewRedis connects to Redis and returns a RedisStore writing to poolKey.
func NewRedis(addr, password string, db int, poolKey string) *RedisStore {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &RedisStore{client: client, key: poolKey}
}

// Replace writes the new working set without ever leaving the pool empty mid
// update: it builds the set under a temporary key, then atomically RENAMEs it
// over the live key. Readers therefore only ever see a complete pool.
func (s *RedisStore) Replace(ctx context.Context, members []string, ttl time.Duration) error {
	if len(members) == 0 {
		// No working proxies — clear the pool rather than serve a stale set.
		if err := s.client.Del(ctx, s.key).Err(); err != nil {
			return fmt.Errorf("store: clear empty pool: %w", err)
		}
		return nil
	}

	tmpKey := s.key + ":staging"
	args := make([]interface{}, len(members))
	for i, m := range members {
		args[i] = m
	}

	pipe := s.client.TxPipeline()
	pipe.Del(ctx, tmpKey)
	pipe.SAdd(ctx, tmpKey, args...)
	pipe.Expire(ctx, tmpKey, ttl)
	pipe.Rename(ctx, tmpKey, s.key)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("store: replace pool: %w", err)
	}
	return nil
}

// Ping verifies the Redis connection is alive.
func (s *RedisStore) Ping(ctx context.Context) error {
	if err := s.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("store: ping: %w", err)
	}
	return nil
}

// Close releases the Redis client.
func (s *RedisStore) Close() error {
	return s.client.Close()
}
