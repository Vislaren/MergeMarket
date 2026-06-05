package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisSource is a Redis-list-backed Source. Raw results are popped from
// normalizeKey with a blocking BLPOP (bounded by pollTimeout so workers stay
// shutdown-aware). The scraper-service pushes results with RPUSH, so BLPOP here
// consumes them in FIFO order.
type RedisSource struct {
	client       *redis.Client
	normalizeKey string
	pollTimeout  time.Duration
}

// NewRedis connects to Redis and returns a RedisSource. pollTimeout bounds each
// blocking Dequeue call.
func NewRedis(addr, password string, db int, normalizeKey string, pollTimeout time.Duration) *RedisSource {
	client := redis.NewClient(&redis.Options{Addr: addr, Password: password, DB: db})
	return &RedisSource{client: client, normalizeKey: normalizeKey, pollTimeout: pollTimeout}
}

// NewRedisWithClient builds a RedisSource around an existing client.
func NewRedisWithClient(client *redis.Client, normalizeKey string, pollTimeout time.Duration) *RedisSource {
	return &RedisSource{client: client, normalizeKey: normalizeKey, pollTimeout: pollTimeout}
}

// Dequeue performs a blocking left-pop bounded by pollTimeout. On timeout it
// returns ErrEmpty; on a cancelled context it returns the context error.
func (s *RedisSource) Dequeue(ctx context.Context) (RawResult, error) {
	res, err := s.client.BLPop(ctx, s.pollTimeout, s.normalizeKey).Result()
	if err != nil {
		if err == redis.Nil {
			return RawResult{}, ErrEmpty
		}
		if ctx.Err() != nil {
			return RawResult{}, ctx.Err()
		}
		return RawResult{}, fmt.Errorf("queue: blpop %s: %w", s.normalizeKey, err)
	}
	// BLPop returns [key, value].
	if len(res) != 2 {
		return RawResult{}, fmt.Errorf("queue: unexpected blpop result length %d", len(res))
	}
	var r RawResult
	if err := json.Unmarshal([]byte(res[1]), &r); err != nil {
		return RawResult{}, fmt.Errorf("queue: decode raw result: %w", err)
	}
	return r, nil
}

// Ping verifies Redis connectivity.
func (s *RedisSource) Ping(ctx context.Context) error {
	if err := s.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("queue: ping: %w", err)
	}
	return nil
}

// Close releases the Redis client.
func (s *RedisSource) Close() error { return s.client.Close() }
