package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisQueue is a Redis-list-backed Queue + Sink. Jobs are popped from jobKey
// with a blocking BLPOP (bounded by pollTimeout so workers stay shutdown-aware);
// raw results are pushed to normalizeKey with RPUSH for the normalization
// service (A-06) to consume in FIFO order.
type RedisQueue struct {
	client       *redis.Client
	jobKey       string
	normalizeKey string
	pollTimeout  time.Duration
}

// NewRedis connects to Redis and returns a RedisQueue. pollTimeout bounds each
// blocking Dequeue call.
func NewRedis(addr, password string, db int, jobKey, normalizeKey string, pollTimeout time.Duration) *RedisQueue {
	client := redis.NewClient(&redis.Options{Addr: addr, Password: password, DB: db})
	return &RedisQueue{
		client:       client,
		jobKey:       jobKey,
		normalizeKey: normalizeKey,
		pollTimeout:  pollTimeout,
	}
}

// NewRedisWithClient builds a RedisQueue around an existing client (lets the
// proxy pool reader and queue share one connection pool).
func NewRedisWithClient(client *redis.Client, jobKey, normalizeKey string, pollTimeout time.Duration) *RedisQueue {
	return &RedisQueue{client: client, jobKey: jobKey, normalizeKey: normalizeKey, pollTimeout: pollTimeout}
}

// Dequeue performs a blocking left-pop bounded by pollTimeout. On timeout it
// returns ErrEmpty; on a cancelled context it returns the context error.
func (q *RedisQueue) Dequeue(ctx context.Context) (Job, error) {
	res, err := q.client.BLPop(ctx, q.pollTimeout, q.jobKey).Result()
	if err != nil {
		if err == redis.Nil {
			return Job{}, ErrEmpty
		}
		if ctx.Err() != nil {
			return Job{}, ctx.Err()
		}
		return Job{}, fmt.Errorf("queue: blpop %s: %w", q.jobKey, err)
	}
	// BLPop returns [key, value].
	if len(res) != 2 {
		return Job{}, fmt.Errorf("queue: unexpected blpop result length %d", len(res))
	}
	var j Job
	if err := json.Unmarshal([]byte(res[1]), &j); err != nil {
		return Job{}, fmt.Errorf("queue: decode job: %w", err)
	}
	return j, nil
}

// Publish marshals the result and RPUSHes it onto the normalization queue.
func (q *RedisQueue) Publish(ctx context.Context, result RawResult) error {
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("queue: encode result: %w", err)
	}
	if err := q.client.RPush(ctx, q.normalizeKey, data).Err(); err != nil {
		return fmt.Errorf("queue: rpush %s: %w", q.normalizeKey, err)
	}
	return nil
}

// Ping verifies Redis connectivity.
func (q *RedisQueue) Ping(ctx context.Context) error {
	if err := q.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("queue: ping: %w", err)
	}
	return nil
}

// Close releases the Redis client.
func (q *RedisQueue) Close() error { return q.client.Close() }
