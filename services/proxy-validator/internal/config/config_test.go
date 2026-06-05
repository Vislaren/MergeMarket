package config

import (
	"testing"
	"time"
)

// setEnv sets the given env vars for the duration of the test and clears any
// others this package reads, so each case starts from a known baseline.
func withEnv(t *testing.T, vars map[string]string) {
	t.Helper()
	keys := []string{
		"PROXY_VALIDATOR_PORT", "REDIS_HOST", "REDIS_PORT", "REDIS_PASSWORD",
		"REDIS_DB", "PROXY_POOL_KEY", "PROXY_POOL_TTL", "PROXY_TEST_URL",
		"PROXY_TEST_TIMEOUT", "PROXY_REFRESH_INTERVAL", "PROXY_POLITENESS_MIN",
		"PROXY_POLITENESS_MAX", "PROXY_VALIDATOR_CONCURRENCY", "PROXY_SOURCES",
	}
	for _, k := range keys {
		t.Setenv(k, "")
	}
	for k, v := range vars {
		t.Setenv(k, v)
	}
}

func TestLoadDefaults(t *testing.T) {
	withEnv(t, nil)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Port != "8086" {
		t.Errorf("Port = %q, want 8086", cfg.Port)
	}
	if cfg.RedisAddr != "localhost:6379" {
		t.Errorf("RedisAddr = %q, want localhost:6379", cfg.RedisAddr)
	}
	if cfg.ProxyPoolKey != "proxy_pool" {
		t.Errorf("ProxyPoolKey = %q, want proxy_pool", cfg.ProxyPoolKey)
	}
	if cfg.ProxyPoolTTL != 5*time.Minute {
		t.Errorf("ProxyPoolTTL = %s, want 5m", cfg.ProxyPoolTTL)
	}
	if len(cfg.Sources) == 0 {
		t.Error("Sources should default to the built-in list")
	}
	if cfg.RefreshInterval >= cfg.ProxyPoolTTL {
		t.Error("default RefreshInterval must be < ProxyPoolTTL")
	}
}

func TestLoadOverrides(t *testing.T) {
	withEnv(t, map[string]string{
		"PROXY_VALIDATOR_PORT":        "9099",
		"REDIS_HOST":                  "redis.internal",
		"REDIS_PORT":                  "6380",
		"REDIS_PASSWORD":              "s3cret",
		"PROXY_SOURCES":               "http://a.test/list , http://b.test/list",
		"PROXY_POLITENESS_MIN":        "10ms",
		"PROXY_POLITENESS_MAX":        "20ms",
		"PROXY_VALIDATOR_CONCURRENCY": "5",
	})

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Port != "9099" {
		t.Errorf("Port = %q, want 9099", cfg.Port)
	}
	if cfg.RedisAddr != "redis.internal:6380" {
		t.Errorf("RedisAddr = %q", cfg.RedisAddr)
	}
	if cfg.RedisPassword != "s3cret" {
		t.Errorf("RedisPassword = %q", cfg.RedisPassword)
	}
	if len(cfg.Sources) != 2 || cfg.Sources[0] != "http://a.test/list" {
		t.Errorf("Sources = %#v, want trimmed 2-element list", cfg.Sources)
	}
	if cfg.Concurrency != 5 {
		t.Errorf("Concurrency = %d, want 5", cfg.Concurrency)
	}
}

func TestLoadValidationErrors(t *testing.T) {
	cases := map[string]map[string]string{
		"max < min politeness": {"PROXY_POLITENESS_MIN": "100ms", "PROXY_POLITENESS_MAX": "10ms"},
		"zero concurrency":     {"PROXY_VALIDATOR_CONCURRENCY": "0"},
		"refresh >= ttl":       {"PROXY_REFRESH_INTERVAL": "10m", "PROXY_POOL_TTL": "5m"},
		"bad duration":         {"PROXY_POOL_TTL": "not-a-duration"},
		"bad int":              {"PROXY_VALIDATOR_CONCURRENCY": "abc"},
	}
	for name, env := range cases {
		t.Run(name, func(t *testing.T) {
			withEnv(t, env)
			if _, err := Load(); err == nil {
				t.Errorf("Load() expected an error for %s, got nil", name)
			}
		})
	}
}
