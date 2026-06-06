// Package config loads all runtime configuration for the proxy-validator
// service from environment variables. Per the MergeMarket Go standards, no
// configuration value is ever hardcoded at a call site: everything funnels
// through Load so the service is fully driven by the environment / .env file.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds every tunable parameter for the proxy-validator service.
type Config struct {
	// Port is the HTTP port the /health server listens on (PROXY_VALIDATOR_PORT).
	Port string

	// RedisAddr is the host:port of the Redis instance holding the proxy pool.
	RedisAddr string
	// RedisPassword is optional; empty means no authentication.
	RedisPassword string
	// RedisDB is the logical Redis database index.
	RedisDB int

	// ProxyPoolKey is the Redis Set key that holds working "ip:port" members.
	// Per DATABASE_SCHEMA.md §3 this is "proxy_pool" with a 5-minute TTL.
	ProxyPoolKey string
	// ProxyPoolTTL is how long the written proxy_pool Set survives in Redis.
	ProxyPoolTTL time.Duration

	// Sources is the list of public proxy-list URLs to scrape (plaintext
	// "ip:port" per line). Override with PROXY_SOURCES (comma-separated).
	Sources []string

	// TestURL is the real endpoint each proxy is tested against.
	TestURL string
	// TestTimeout is the per-proxy validation timeout.
	TestTimeout time.Duration

	// Concurrency caps how many proxies are validated simultaneously.
	Concurrency int
	// RefreshInterval is how often the full scrape→validate→write cycle runs.
	// It should be shorter than ProxyPoolTTL so the pool never fully expires.
	RefreshInterval time.Duration

	// PolitenessMin / PolitenessMax bound the adaptive random delay inserted
	// between outbound validation requests (the politeness protocol).
	PolitenessMin time.Duration
	PolitenessMax time.Duration
}

// defaultSources are well-known free, frequently-updated public proxy lists.
// They are plaintext, one "ip:port" per line.
var defaultSources = []string{
	"https://raw.githubusercontent.com/TheSpeedX/PROXY-List/master/http.txt",
	"https://raw.githubusercontent.com/clarketm/proxy-list/master/proxy-list-raw.txt",
	"https://api.proxyscrape.com/v2/?request=getproxies&protocol=http&timeout=10000",
}

// Load reads configuration from the environment, applies sensible defaults,
// and validates the result. It returns a named error describing the first
// invalid value encountered.
func Load() (*Config, error) {
	cfg := &Config{
		Port:          getEnv("PROXY_VALIDATOR_PORT", "8086"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		ProxyPoolKey:  getEnv("PROXY_POOL_KEY", "proxy_pool"),
		TestURL:       getEnv("PROXY_TEST_URL", "http://www.google.com/generate_204"),
	}

	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")
	cfg.RedisAddr = redisHost + ":" + redisPort

	var err error
	if cfg.RedisDB, err = getEnvInt("REDIS_DB", 0); err != nil {
		return nil, err
	}
	if cfg.ProxyPoolTTL, err = getEnvDuration("PROXY_POOL_TTL", 5*time.Minute); err != nil {
		return nil, err
	}
	if cfg.TestTimeout, err = getEnvDuration("PROXY_TEST_TIMEOUT", 8*time.Second); err != nil {
		return nil, err
	}
	if cfg.RefreshInterval, err = getEnvDuration("PROXY_REFRESH_INTERVAL", 4*time.Minute); err != nil {
		return nil, err
	}
	if cfg.PolitenessMin, err = getEnvDuration("PROXY_POLITENESS_MIN", 50*time.Millisecond); err != nil {
		return nil, err
	}
	if cfg.PolitenessMax, err = getEnvDuration("PROXY_POLITENESS_MAX", 400*time.Millisecond); err != nil {
		return nil, err
	}
	if cfg.Concurrency, err = getEnvInt("PROXY_VALIDATOR_CONCURRENCY", 50); err != nil {
		return nil, err
	}

	if sources := getEnv("PROXY_SOURCES", ""); sources != "" {
		cfg.Sources = splitAndTrim(sources)
	} else {
		cfg.Sources = defaultSources
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// validate enforces invariants the rest of the service relies on.
func (c *Config) validate() error {
	if c.Port == "" {
		return fmt.Errorf("config: PROXY_VALIDATOR_PORT must not be empty")
	}
	if c.ProxyPoolKey == "" {
		return fmt.Errorf("config: PROXY_POOL_KEY must not be empty")
	}
	if c.TestURL == "" {
		return fmt.Errorf("config: PROXY_TEST_URL must not be empty")
	}
	if len(c.Sources) == 0 {
		return fmt.Errorf("config: at least one proxy source is required")
	}
	if c.Concurrency < 1 {
		return fmt.Errorf("config: PROXY_VALIDATOR_CONCURRENCY must be >= 1, got %d", c.Concurrency)
	}
	if c.PolitenessMin < 0 || c.PolitenessMax < 0 {
		return fmt.Errorf("config: politeness delays must not be negative")
	}
	if c.PolitenessMax < c.PolitenessMin {
		return fmt.Errorf("config: PROXY_POLITENESS_MAX (%s) must be >= PROXY_POLITENESS_MIN (%s)",
			c.PolitenessMax, c.PolitenessMin)
	}
	if c.RefreshInterval >= c.ProxyPoolTTL {
		return fmt.Errorf("config: PROXY_REFRESH_INTERVAL (%s) must be < PROXY_POOL_TTL (%s) so the pool never fully expires",
			c.RefreshInterval, c.ProxyPoolTTL)
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
		return 0, fmt.Errorf("config: %s must be a Go duration (e.g. 5m, 400ms): %w", key, err)
	}
	return d, nil
}

func splitAndTrim(csv string) []string {
	parts := strings.Split(csv, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
