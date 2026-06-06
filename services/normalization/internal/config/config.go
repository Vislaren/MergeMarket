// Package config loads all runtime configuration for the normalization-service
// from environment variables. Per the MergeMarket Go standards no configuration
// value is hardcoded at a call site: everything funnels through Load so the
// service is fully driven by the environment / .env file.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds every tunable parameter for the normalization-service.
type Config struct {
	// Port is the HTTP port the /health + /stats server listens on
	// (NORMALIZATION_PORT, default 8084).
	Port string

	// RedisAddr is the host:port of the Redis instance holding the queues.
	RedisAddr string
	// RedisPassword is optional; empty means no authentication.
	RedisPassword string
	// RedisDB is the logical Redis database index.
	RedisDB int

	// NormalizeQueueKey is the Redis list raw scrape results are consumed from
	// (the scraper-service A-05 pushes RawResult JSON here).
	NormalizeQueueKey string
	// QueuePollTimeout bounds a single blocking dequeue so workers stay
	// responsive to shutdown even when the queue is idle.
	QueuePollTimeout time.Duration

	// Workers is the number of concurrent worker goroutines consuming results.
	Workers int

	// DatabaseURL is the PostgreSQL connection string (DSN) products are upserted
	// into. Assembled from DB_* vars when DATABASE_URL is not set explicitly.
	DatabaseURL string

	// AffiliateConfigPath points at the JSON file mapping store ids to affiliate
	// link templates / parameters. Empty (or a missing file) disables injection
	// beyond any configured default parameters.
	AffiliateConfigPath string
}

// Load reads configuration from the environment, applies sensible defaults, and
// validates the result. It returns a named error describing the first invalid
// value encountered.
func Load() (*Config, error) {
	cfg := &Config{
		Port:                getEnv("NORMALIZATION_PORT", "8084"),
		RedisPassword:       getEnv("REDIS_PASSWORD", ""),
		NormalizeQueueKey:   getEnv("SCRAPER_NORMALIZE_QUEUE", "normalize_queue"),
		AffiliateConfigPath: getEnv("NORMALIZATION_AFFILIATE_CONFIG", ""),
	}

	cfg.RedisAddr = getEnv("REDIS_HOST", "localhost") + ":" + getEnv("REDIS_PORT", "6379")
	cfg.DatabaseURL = databaseURL()

	var err error
	if cfg.RedisDB, err = getEnvInt("REDIS_DB", 0); err != nil {
		return nil, err
	}
	if cfg.Workers, err = getEnvInt("NORMALIZATION_WORKERS", 5); err != nil {
		return nil, err
	}
	if cfg.QueuePollTimeout, err = getEnvDuration("NORMALIZATION_QUEUE_POLL_TIMEOUT", 5*time.Second); err != nil {
		return nil, err
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// databaseURL prefers an explicit DATABASE_URL, otherwise assembles a DSN from
// the standard DB_* variables (DATABASE_SCHEMA.md §4).
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

// validate enforces invariants the rest of the service relies on.
func (c *Config) validate() error {
	if c.Port == "" {
		return fmt.Errorf("config: NORMALIZATION_PORT must not be empty")
	}
	if c.NormalizeQueueKey == "" {
		return fmt.Errorf("config: SCRAPER_NORMALIZE_QUEUE must not be empty")
	}
	if c.Workers < 1 {
		return fmt.Errorf("config: NORMALIZATION_WORKERS must be >= 1, got %d", c.Workers)
	}
	if c.QueuePollTimeout <= 0 {
		return fmt.Errorf("config: NORMALIZATION_QUEUE_POLL_TIMEOUT must be > 0")
	}
	if c.DatabaseURL == "" {
		return fmt.Errorf("config: DATABASE_URL could not be determined")
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
		return 0, fmt.Errorf("config: %s must be a Go duration (e.g. 5s, 1m): %w", key, err)
	}
	return d, nil
}
