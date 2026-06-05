// Package config loads all runtime configuration for the scraper-service from
// environment variables. Per the MergeMarket Go standards no configuration
// value is hardcoded at a call site: everything funnels through Load so the
// service is fully driven by the environment / .env file.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds every tunable parameter for the scraper-service.
type Config struct {
	// Port is the HTTP port the /health + /stats server listens on (SCRAPER_PORT).
	Port string

	// ConfigDir is the directory holding StoreConfig JSON/YAML files, one per
	// store (ARCHITECTURE.md §5: services/scraper-service/configs/).
	ConfigDir string

	// RedisAddr is the host:port of the Redis instance holding the queues and
	// the proxy pool.
	RedisAddr string
	// RedisPassword is optional; empty means no authentication.
	RedisPassword string
	// RedisDB is the logical Redis database index.
	RedisDB int

	// JobQueueKey is the Redis list the workers pop scrape jobs from.
	JobQueueKey string
	// NormalizeQueueKey is the Redis list raw scrape results are pushed to for
	// the normalization service (A-06) to consume.
	NormalizeQueueKey string
	// ProxyPoolKey is the Redis Set of working "ip:port" proxies written by the
	// proxy-validator service (A-04). DATABASE_SCHEMA.md §3: proxy_pool.
	ProxyPoolKey string
	// CircuitKeyPrefix is prepended to a store id to form the per-store circuit
	// breaker state key. DATABASE_SCHEMA.md §3: circuit:{store_id}.
	CircuitKeyPrefix string

	// Workers is the number of concurrent worker goroutines consuming jobs.
	Workers int
	// QueuePollTimeout bounds a single blocking dequeue so workers stay
	// responsive to shutdown even when the queue is idle.
	QueuePollTimeout time.Duration
	// ScrapeTimeout is the per-request timeout for a single store scrape.
	ScrapeTimeout time.Duration

	// CircuitThreshold is the number of consecutive 403/429 responses from a
	// store that trips its circuit breaker open.
	CircuitThreshold int
	// CircuitCooldown is how long a tripped breaker stays open before allowing
	// a single half-open trial request.
	CircuitCooldown time.Duration

	// UseProxy enables routing scrape requests through a proxy drawn from the
	// proxy pool. When false (or the pool is empty) scrapes go direct.
	UseProxy bool
}

// Load reads configuration from the environment, applies sensible defaults, and
// validates the result. It returns a named error describing the first invalid
// value encountered.
func Load() (*Config, error) {
	cfg := &Config{
		Port:              getEnv("SCRAPER_PORT", "8083"),
		ConfigDir:         getEnv("SCRAPER_CONFIG_DIR", "configs"),
		RedisPassword:     getEnv("REDIS_PASSWORD", ""),
		JobQueueKey:       getEnv("SCRAPER_JOB_QUEUE", "scrape_queue"),
		NormalizeQueueKey: getEnv("SCRAPER_NORMALIZE_QUEUE", "normalize_queue"),
		ProxyPoolKey:      getEnv("PROXY_POOL_KEY", "proxy_pool"),
		CircuitKeyPrefix:  getEnv("SCRAPER_CIRCUIT_PREFIX", "circuit:"),
	}

	cfg.RedisAddr = getEnv("REDIS_HOST", "localhost") + ":" + getEnv("REDIS_PORT", "6379")

	var err error
	if cfg.RedisDB, err = getEnvInt("REDIS_DB", 0); err != nil {
		return nil, err
	}
	if cfg.Workers, err = getEnvInt("SCRAPER_WORKERS", 10); err != nil {
		return nil, err
	}
	if cfg.CircuitThreshold, err = getEnvInt("SCRAPER_CIRCUIT_THRESHOLD", 5); err != nil {
		return nil, err
	}
	if cfg.QueuePollTimeout, err = getEnvDuration("SCRAPER_QUEUE_POLL_TIMEOUT", 5*time.Second); err != nil {
		return nil, err
	}
	if cfg.ScrapeTimeout, err = getEnvDuration("SCRAPER_SCRAPE_TIMEOUT", 15*time.Second); err != nil {
		return nil, err
	}
	if cfg.CircuitCooldown, err = getEnvDuration("SCRAPER_CIRCUIT_COOLDOWN", 60*time.Second); err != nil {
		return nil, err
	}
	if cfg.UseProxy, err = getEnvBool("SCRAPER_USE_PROXY", true); err != nil {
		return nil, err
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// validate enforces invariants the rest of the service relies on.
func (c *Config) validate() error {
	if c.Port == "" {
		return fmt.Errorf("config: SCRAPER_PORT must not be empty")
	}
	if c.ConfigDir == "" {
		return fmt.Errorf("config: SCRAPER_CONFIG_DIR must not be empty")
	}
	if c.JobQueueKey == "" {
		return fmt.Errorf("config: SCRAPER_JOB_QUEUE must not be empty")
	}
	if c.NormalizeQueueKey == "" {
		return fmt.Errorf("config: SCRAPER_NORMALIZE_QUEUE must not be empty")
	}
	if c.JobQueueKey == c.NormalizeQueueKey {
		return fmt.Errorf("config: SCRAPER_JOB_QUEUE and SCRAPER_NORMALIZE_QUEUE must differ")
	}
	if c.Workers < 1 {
		return fmt.Errorf("config: SCRAPER_WORKERS must be >= 1, got %d", c.Workers)
	}
	if c.CircuitThreshold < 1 {
		return fmt.Errorf("config: SCRAPER_CIRCUIT_THRESHOLD must be >= 1, got %d", c.CircuitThreshold)
	}
	if c.QueuePollTimeout <= 0 {
		return fmt.Errorf("config: SCRAPER_QUEUE_POLL_TIMEOUT must be > 0")
	}
	if c.ScrapeTimeout <= 0 {
		return fmt.Errorf("config: SCRAPER_SCRAPE_TIMEOUT must be > 0")
	}
	if c.CircuitCooldown <= 0 {
		return fmt.Errorf("config: SCRAPER_CIRCUIT_COOLDOWN must be > 0")
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

func getEnvBool(key string, fallback bool) (bool, error) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback, nil
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return false, fmt.Errorf("config: %s must be a boolean (true/false): %w", key, err)
	}
	return b, nil
}
