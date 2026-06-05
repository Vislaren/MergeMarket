package config

import (
	"testing"
	"time"
)

// setEnv sets each key for the duration of the test and clears it afterwards.
func setEnv(t *testing.T, kv map[string]string) {
	t.Helper()
	for k, v := range kv {
		t.Setenv(k, v)
	}
}

func TestLoadDefaults(t *testing.T) {
	// t.Setenv guarantees a clean, restored environment per key we touch; the
	// rest fall back to defaults.
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.Port != "8083" {
		t.Errorf("Port = %q, want 8083", cfg.Port)
	}
	if cfg.RedisAddr != "localhost:6379" {
		t.Errorf("RedisAddr = %q, want localhost:6379", cfg.RedisAddr)
	}
	if cfg.Workers != 10 {
		t.Errorf("Workers = %d, want 10", cfg.Workers)
	}
	if cfg.JobQueueKey != "scrape_queue" || cfg.NormalizeQueueKey != "normalize_queue" {
		t.Errorf("queue keys = %q/%q", cfg.JobQueueKey, cfg.NormalizeQueueKey)
	}
	if cfg.CircuitThreshold != 5 {
		t.Errorf("CircuitThreshold = %d, want 5", cfg.CircuitThreshold)
	}
	if cfg.ScrapeTimeout != 15*time.Second {
		t.Errorf("ScrapeTimeout = %s, want 15s", cfg.ScrapeTimeout)
	}
	if !cfg.UseProxy {
		t.Errorf("UseProxy = false, want true")
	}
}

func TestLoadOverrides(t *testing.T) {
	setEnv(t, map[string]string{
		"SCRAPER_PORT":              "9999",
		"REDIS_HOST":                "redis.internal",
		"REDIS_PORT":                "6380",
		"REDIS_DB":                  "3",
		"SCRAPER_WORKERS":           "25",
		"SCRAPER_CIRCUIT_THRESHOLD": "2",
		"SCRAPER_CIRCUIT_COOLDOWN":  "30s",
		"SCRAPER_USE_PROXY":         "false",
	})
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Port != "9999" {
		t.Errorf("Port = %q", cfg.Port)
	}
	if cfg.RedisAddr != "redis.internal:6380" {
		t.Errorf("RedisAddr = %q", cfg.RedisAddr)
	}
	if cfg.RedisDB != 3 {
		t.Errorf("RedisDB = %d", cfg.RedisDB)
	}
	if cfg.Workers != 25 {
		t.Errorf("Workers = %d", cfg.Workers)
	}
	if cfg.CircuitThreshold != 2 || cfg.CircuitCooldown != 30*time.Second {
		t.Errorf("circuit = %d/%s", cfg.CircuitThreshold, cfg.CircuitCooldown)
	}
	if cfg.UseProxy {
		t.Errorf("UseProxy = true, want false")
	}
}

func TestLoadValidationErrors(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
	}{
		{"zero workers", map[string]string{"SCRAPER_WORKERS": "0"}},
		{"bad workers int", map[string]string{"SCRAPER_WORKERS": "lots"}},
		{"zero threshold", map[string]string{"SCRAPER_CIRCUIT_THRESHOLD": "0"}},
		{"bad duration", map[string]string{"SCRAPER_SCRAPE_TIMEOUT": "soon"}},
		{"bad bool", map[string]string{"SCRAPER_USE_PROXY": "maybe"}},
		{"same queues", map[string]string{"SCRAPER_JOB_QUEUE": "q", "SCRAPER_NORMALIZE_QUEUE": "q"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setEnv(t, tt.env)
			if _, err := Load(); err == nil {
				t.Fatalf("Load() expected error for %s, got nil", tt.name)
			}
		})
	}
}
