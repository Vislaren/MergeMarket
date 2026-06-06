// Package config loads all runtime configuration for the history-service from
// environment variables. Per the MergeMarket Go standards no configuration value
// is hardcoded at a call site: everything funnels through Load so the service is
// fully driven by the environment / .env file.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// HeartbeatMode selects how the heartbeat determines a followed product's
// current price.
type HeartbeatMode string

const (
	// ModeDB reads the freshest price the scraper/normalization pipeline already
	// persisted (products.last_price). Reliable and the default.
	ModeDB HeartbeatMode = "db"
	// ModeHTTP re-fetches the product URL and best-effort extracts an embedded
	// price (JSON-LD / price meta tags). Honours "scrape followed product URLs"
	// but is inherently fragile; opt in explicitly.
	ModeHTTP HeartbeatMode = "http"
)

// Config holds every tunable parameter for the history-service.
type Config struct {
	// Port is the HTTP port the API + /health + /stats server listens on
	// (HISTORY_PORT, default 8085).
	Port string

	// DatabaseURL is the PostgreSQL/TimescaleDB DSN. Assembled from DB_* vars
	// when DATABASE_URL is not set explicitly.
	DatabaseURL string

	// Redis connection used to publish fired price-drop alert events.
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	// AlertQueueKey is the Redis list price-drop alert events are pushed to for
	// the Notification Worker (ARCHITECTURE §1) to consume.
	AlertQueueKey string

	// SnapshotInterval is how often a full daily price snapshot is taken.
	SnapshotInterval time.Duration
	// SnapshotOnStart runs one snapshot immediately at startup when true.
	SnapshotOnStart bool

	// HeartbeatInterval is how often followed products are re-checked for alerts.
	HeartbeatInterval time.Duration
	// HeartbeatOnStart runs one heartbeat immediately at startup when true.
	HeartbeatOnStart bool
	// HeartbeatMode selects the price source (db or http).
	HeartbeatMode HeartbeatMode
	// HeartbeatTimeout bounds a single HTTP product fetch (ModeHTTP only).
	HeartbeatTimeout time.Duration
}

// Load reads configuration from the environment, applies sensible defaults, and
// validates the result.
func Load() (*Config, error) {
	cfg := &Config{
		Port:          getEnv("HISTORY_PORT", "8085"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		AlertQueueKey: getEnv("HISTORY_ALERT_QUEUE", "price_alert_events"),
		HeartbeatMode: HeartbeatMode(getEnv("HISTORY_HEARTBEAT_MODE", string(ModeDB))),
	}

	cfg.RedisAddr = getEnv("REDIS_HOST", "localhost") + ":" + getEnv("REDIS_PORT", "6379")
	cfg.DatabaseURL = databaseURL()

	var err error
	if cfg.RedisDB, err = getEnvInt("REDIS_DB", 0); err != nil {
		return nil, err
	}
	if cfg.SnapshotInterval, err = getEnvDuration("HISTORY_SNAPSHOT_INTERVAL", 24*time.Hour); err != nil {
		return nil, err
	}
	if cfg.HeartbeatInterval, err = getEnvDuration("HISTORY_HEARTBEAT_INTERVAL", time.Hour); err != nil {
		return nil, err
	}
	if cfg.HeartbeatTimeout, err = getEnvDuration("HISTORY_HEARTBEAT_TIMEOUT", 10*time.Second); err != nil {
		return nil, err
	}
	if cfg.SnapshotOnStart, err = getEnvBool("HISTORY_SNAPSHOT_ON_START", false); err != nil {
		return nil, err
	}
	if cfg.HeartbeatOnStart, err = getEnvBool("HISTORY_HEARTBEAT_ON_START", true); err != nil {
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
		return fmt.Errorf("config: HISTORY_PORT must not be empty")
	}
	if c.DatabaseURL == "" {
		return fmt.Errorf("config: DATABASE_URL could not be determined")
	}
	if c.AlertQueueKey == "" {
		return fmt.Errorf("config: HISTORY_ALERT_QUEUE must not be empty")
	}
	if c.SnapshotInterval <= 0 {
		return fmt.Errorf("config: HISTORY_SNAPSHOT_INTERVAL must be > 0")
	}
	if c.HeartbeatInterval <= 0 {
		return fmt.Errorf("config: HISTORY_HEARTBEAT_INTERVAL must be > 0")
	}
	if c.HeartbeatTimeout <= 0 {
		return fmt.Errorf("config: HISTORY_HEARTBEAT_TIMEOUT must be > 0")
	}
	switch c.HeartbeatMode {
	case ModeDB, ModeHTTP:
	default:
		return fmt.Errorf("config: HISTORY_HEARTBEAT_MODE must be %q or %q, got %q", ModeDB, ModeHTTP, c.HeartbeatMode)
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
		return 0, fmt.Errorf("config: %s must be a Go duration (e.g. 1h, 30m): %w", key, err)
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
