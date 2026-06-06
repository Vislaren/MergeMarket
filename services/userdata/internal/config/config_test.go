package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("USERDATA_PORT", "")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("USERDATA_JWT_ISSUER", "")
	t.Setenv("JWT_SECRET", "supersecret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Port != "8090" {
		t.Errorf("Port = %q, want 8090", cfg.Port)
	}
	if cfg.JWTIssuer != "mergemarket-auth" {
		t.Errorf("JWTIssuer = %q, want mergemarket-auth", cfg.JWTIssuer)
	}
	if string(cfg.JWTSecret) != "supersecret" {
		t.Errorf("JWTSecret not loaded")
	}
	if cfg.DatabaseURL == "" {
		t.Error("DatabaseURL should be assembled from DB_* defaults")
	}
}

func TestLoadRequiresJWTSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	if _, err := Load(); err == nil {
		t.Fatal("expected error when JWT_SECRET is unset")
	}
}

func TestLoadOverrides(t *testing.T) {
	t.Setenv("JWT_SECRET", "x")
	t.Setenv("USERDATA_PORT", "9001")
	t.Setenv("USERDATA_JWT_ISSUER", "custom-iss")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Port != "9001" || cfg.JWTIssuer != "custom-iss" {
		t.Errorf("overrides not applied: %+v", cfg)
	}
}
