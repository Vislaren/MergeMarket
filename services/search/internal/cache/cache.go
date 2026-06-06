// Package cache is the search-service's Redis result cache. It stores aggregated
// search results under search:{query_hash} (DATABASE_SCHEMA.md §3) as a JSON
// payload that also records when the entry was written, so the orchestrator can
// implement stale-while-revalidate (ARCHITECTURE.md §10). Access is behind the
// Cache interface so the orchestrator can be unit-tested with an in-memory fake.
package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Vislaren/MergeMarket/services/search/internal/store"
)

// Entry is one cached result set plus the time it was written.
type Entry struct {
	Results  []store.Product `json:"results"`
	CachedAt time.Time       `json:"cached_at"`
}

// Cache is the search result cache surface.
type Cache interface {
	// Get returns the cached entry for key and whether it was present. A cache
	// miss is (Entry{}, false, nil) — not an error.
	Get(ctx context.Context, key string) (Entry, bool, error)
	// Set stores entry under key with the given TTL.
	Set(ctx context.Context, key string, entry Entry, ttl time.Duration) error
}

// Key builds the canonical cache key for a query+location pair. The query and
// location are normalized (trimmed, lower-cased) before hashing so logically
// identical searches share a cache entry.
func Key(prefix, query, location string) string {
	norm := strings.ToLower(strings.TrimSpace(query)) + "|" + strings.ToLower(strings.TrimSpace(location))
	sum := sha256.Sum256([]byte(norm))
	return prefix + hex.EncodeToString(sum[:])
}

// Redis implements Cache over go-redis.
type Redis struct {
	client *redis.Client
}

// NewRedis wraps a go-redis client as a Cache.
func NewRedis(client *redis.Client) *Redis {
	return &Redis{client: client}
}

// Get loads and decodes a cached entry; a missing key is a clean miss.
func (r *Redis) Get(ctx context.Context, key string) (Entry, bool, error) {
	raw, err := r.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return Entry{}, false, nil
	}
	if err != nil {
		return Entry{}, false, err
	}
	var e Entry
	if err := json.Unmarshal(raw, &e); err != nil {
		// A corrupt entry is treated as a miss so the caller recomputes.
		return Entry{}, false, nil
	}
	return e, true, nil
}

// Set encodes and stores an entry with the supplied TTL.
func (r *Redis) Set(ctx context.Context, key string, entry Entry, ttl time.Duration) error {
	raw, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, raw, ttl).Err()
}
