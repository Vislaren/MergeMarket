package config

import "testing"

func TestLoad_Defaults(t *testing.T) {
	// Ensure a clean environment for the values we assert on.
	for _, k := range []string{
		"NORMALIZATION_PORT", "NORMALIZATION_WORKERS", "DATABASE_URL",
		"SCRAPER_NORMALIZE_QUEUE", "NORMALIZATION_QUEUE_POLL_TIMEOUT",
		"DB_HOST", "DB_PORT", "DB_NAME", "DB_USER", "DB_PASSWORD",
	} {
		t.Setenv(k, "")
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Port != "8084" {
		t.Errorf("default port = %q, want 8084", cfg.Port)
	}
	if cfg.Workers != 5 {
		t.Errorf("default workers = %d, want 5", cfg.Workers)
	}
	if cfg.NormalizeQueueKey != "normalize_queue" {
		t.Errorf("queue key = %q", cfg.NormalizeQueueKey)
	}
	if cfg.DatabaseURL == "" {
		t.Errorf("database url should be assembled from defaults")
	}
}

func TestLoad_ExplicitDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://u:p@h:5/db")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DatabaseURL != "postgres://u:p@h:5/db" {
		t.Errorf("explicit DATABASE_URL not honoured: %q", cfg.DatabaseURL)
	}
}

func TestLoad_RejectsBadWorkers(t *testing.T) {
	t.Setenv("NORMALIZATION_WORKERS", "0")
	if _, err := Load(); err == nil {
		t.Errorf("expected error for workers=0")
	}
}

func TestLoad_RejectsBadDuration(t *testing.T) {
	t.Setenv("NORMALIZATION_QUEUE_POLL_TIMEOUT", "notaduration")
	if _, err := Load(); err == nil {
		t.Errorf("expected error for bad duration")
	}
}
