package config

import (
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	// Ensure a clean environment so defaults apply.
	for _, k := range []string{
		"SEARCH_PORT", "DATABASE_URL", "SEARCH_CACHE_PREFIX", "SEARCH_CACHE_TTL",
		"SEARCH_CACHE_STALE_AFTER", "SEARCH_MAX_RESULTS", "REDIS_DB",
	} {
		t.Setenv(k, "")
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Port != "8087" {
		t.Errorf("Port = %q, want 8087", cfg.Port)
	}
	if cfg.CacheTTL != 15*time.Minute {
		t.Errorf("CacheTTL = %v, want 15m", cfg.CacheTTL)
	}
	if cfg.CacheStaleAfter != 5*time.Minute {
		t.Errorf("CacheStaleAfter = %v, want 5m", cfg.CacheStaleAfter)
	}
	if cfg.MaxResults != 50 {
		t.Errorf("MaxResults = %d, want 50", cfg.MaxResults)
	}
	if cfg.DatabaseURL == "" {
		t.Error("DatabaseURL should be assembled from DB_* defaults")
	}
}

func TestLoadOverrides(t *testing.T) {
	t.Setenv("SEARCH_PORT", "9999")
	t.Setenv("SEARCH_CACHE_TTL", "1h")
	t.Setenv("SEARCH_CACHE_STALE_AFTER", "30m")
	t.Setenv("SEARCH_MAX_RESULTS", "10")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Port != "9999" || cfg.CacheTTL != time.Hour || cfg.CacheStaleAfter != 30*time.Minute || cfg.MaxResults != 10 {
		t.Errorf("overrides not applied: %+v", cfg)
	}
}

func TestLoadRejectsStaleExceedingTTL(t *testing.T) {
	t.Setenv("SEARCH_CACHE_TTL", "1m")
	t.Setenv("SEARCH_CACHE_STALE_AFTER", "5m")
	if _, err := Load(); err == nil {
		t.Fatal("expected error when stale-after exceeds TTL")
	}
}

func TestLoadRejectsBadDuration(t *testing.T) {
	t.Setenv("SEARCH_CACHE_TTL", "not-a-duration")
	if _, err := Load(); err == nil {
		t.Fatal("expected error for invalid duration")
	}
}
