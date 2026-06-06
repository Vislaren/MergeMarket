// Package proxypool reads working proxies from the Redis `proxy_pool` Set that
// the proxy-validator service (A-04) maintains (DATABASE_SCHEMA.md §3). The
// scraper draws a random proxy per request to spread load and avoid being
// fingerprinted by a single IP.
package proxypool

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Provider supplies a proxy address for an outbound request.
type Provider interface {
	// Pick returns a random working "ip:port" proxy. An empty string with a nil
	// error means the pool is currently empty (caller should scrape directly).
	Pick(ctx context.Context) (string, error)
}

// RedisPool reads proxies from the proxy_pool Redis Set.
type RedisPool struct {
	client *redis.Client
	key    string
}

// NewRedis builds a RedisPool reading from poolKey on the given client.
func NewRedis(client *redis.Client, poolKey string) *RedisPool {
	return &RedisPool{client: client, key: poolKey}
}

// Pick returns a random member of the proxy pool via SRANDMEMBER. A missing key
// (redis.Nil) is treated as an empty pool, not an error.
func (p *RedisPool) Pick(ctx context.Context) (string, error) {
	v, err := p.client.SRandMember(ctx, p.key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}
		return "", fmt.Errorf("proxypool: srandmember %s: %w", p.key, err)
	}
	return v, nil
}

// Static is a fixed Provider used for tests and for SCRAPER_USE_PROXY=false
// (returns Addr, which may be empty to force direct connections).
type Static struct{ Addr string }

// Pick returns the static address.
func (s Static) Pick(_ context.Context) (string, error) { return s.Addr, nil }
