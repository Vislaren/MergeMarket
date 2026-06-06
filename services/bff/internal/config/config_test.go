package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("BFF_PORT", "")
	t.Setenv("PORT", "")
	t.Setenv("BFF_UPSTREAM_URL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Port != 8082 {
		t.Errorf("default port = %d, want 8082", cfg.Port)
	}
	if cfg.Addr != ":8082" {
		t.Errorf("addr = %q, want :8082", cfg.Addr)
	}
	if cfg.UpstreamURL != defaultUpstream {
		t.Errorf("upstream = %q, want %q", cfg.UpstreamURL, defaultUpstream)
	}
}

func TestLoadPortPrecedenceAndOverride(t *testing.T) {
	t.Setenv("BFF_PORT", "9000")
	t.Setenv("PORT", "1234")
	t.Setenv("BFF_UPSTREAM_URL", "http://kong:8088")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Port != 9000 {
		t.Errorf("BFF_PORT should win: port = %d, want 9000", cfg.Port)
	}
	if cfg.UpstreamURL != "http://kong:8088" {
		t.Errorf("upstream = %q, want http://kong:8088", cfg.UpstreamURL)
	}
}

func TestLoadInvalidPort(t *testing.T) {
	t.Setenv("BFF_PORT", "not-a-number")
	if _, err := Load(); err == nil {
		t.Fatal("expected error for non-numeric port")
	}

	t.Setenv("BFF_PORT", "70000")
	if _, err := Load(); err == nil {
		t.Fatal("expected error for out-of-range port")
	}
}

func TestLoadInvalidUpstream(t *testing.T) {
	t.Setenv("BFF_PORT", "")
	t.Setenv("PORT", "")
	t.Setenv("BFF_UPSTREAM_URL", "://missing-scheme")
	if _, err := Load(); err == nil {
		t.Fatal("expected error for invalid upstream URL")
	}
}
