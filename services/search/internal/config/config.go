// Package config loads all runtime configuration for the search-service from
// environment variables. Per the MergeMarket Go standards (INSTRUCTIONS §3) no
// configuration value is hardcoded at a call site: everything funnels through
// Load so the service is fully driven by the environment / .env file.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds every tunable parameter for the search-service.
type Config struct {
	// Port is the HTTP port the API + /health server listens on
	// (SEARCH_PORT, default 8087).
	Port string

	// DatabaseURL is the PostgreSQL DSN. Assembled from DB_* vars when
	// DATABASE_URL is not set explicitly.
	DatabaseURL string

	// Redis connection used for the search:{query_hash} result cache.
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// CachePrefix is prepended to the query hash to form the Redis key
	// (DATABASE_SCHEMA.md §3: search:{query_hash}).
	CachePrefix string
	// CacheTTL is how long a cached result set lives before Redis evicts it.
	CacheTTL time.Duration
	// CacheStaleAfter is the age past which a cache hit is still served but a
	// background refresh is triggered (stale-while-revalidate, ARCHITECTURE §10).
	CacheStaleAfter time.Duration

	// MaxResults caps the number of products returned per query.
	MaxResults int
}

// Load reads configuration from the environment, applies sensible defaults, and
// validates the result.
func Load() (*Config, error) {
	cfg := &Config{
		Port:          getEnv("SEARCH_PORT", "8087"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		CachePrefix:   getEnv("SEARCH_CACHE_PREFIX", "search:"),
	}
	cfg.RedisAddr = getEnv("REDIS_HOST", "localhost") + ":" + getEnv("REDIS_PORT", "6379")
	cfg.DatabaseURL = databaseURL()

	var err error
	if cfg.RedisDB, err = getEnvInt("REDIS_DB", 0); err != nil {
		return nil, err
	}
	if cfg.MaxResults, err = getEnvInt("SEARCH_MAX_RESULTS", 50); err != nil {
		return nil, err
	}
	if cfg.CacheTTL, err = getEnvDuration("SEARCH_CACHE_TTL", 15*time.Minute); err != nil {
		return nil, err
	}
	if cfg.CacheStaleAfter, err = getEnvDuration("SEARCH_CACHE_STALE_AFTER", 5*time.Minute); err != nil {
		return nil, err
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func databaseURL() string {
	if v := getEnv("DATABASE_URL", ""); v != "" {
		return v
	}
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	name := getEnv("DB_NAME", "mergemarket")
	user := getEnv("DB_USER", "postgres")
	pass := getEnv("DB_PASSWORD", "")
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, name)
}

func (c *Config) validate() error {
	if c.Port == "" {
		return fmt.Errorf("config: SEARCH_PORT must not be empty")
	}
	if c.DatabaseURL == "" {
		return fmt.Errorf("config: DATABASE_URL could not be determined")
	}
	if c.CachePrefix == "" {
		return fmt.Errorf("config: SEARCH_CACHE_PREFIX must not be empty")
	}
	if c.CacheTTL <= 0 {
		return fmt.Errorf("config: SEARCH_CACHE_TTL must be > 0")
	}
	if c.CacheStaleAfter <= 0 {
		return fmt.Errorf("config: SEARCH_CACHE_STALE_AFTER must be > 0")
	}
	if c.CacheStaleAfter > c.CacheTTL {
		return fmt.Errorf("config: SEARCH_CACHE_STALE_AFTER (%s) must not exceed SEARCH_CACHE_TTL (%s)", c.CacheStaleAfter, c.CacheTTL)
	}
	if c.MaxResults <= 0 {
		return fmt.Errorf("config: SEARCH_MAX_RESULTS must be > 0")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) (int, error) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("config: %s must be an integer: %w", key, err)
	}
	return n, nil
}

func getEnvDuration(key string, fallback time.Duration) (time.Duration, error) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("config: %s must be a Go duration (e.g. 15m, 1h): %w", key, err)
	}
	return d, nil
}
