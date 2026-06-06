// Package config loads all runtime configuration for the user-data service from
// environment variables. Per the MergeMarket Go standards (INSTRUCTIONS §3) no
// configuration value is hardcoded at a call site.
package config

import (
	"fmt"
	"os"
	"time"
)

// Config holds every tunable parameter for the user-data service.
type Config struct {
	// Port is the HTTP port the API + /health server listens on
	// (USERDATA_PORT, default 8090).
	Port string

	// DatabaseURL is the PostgreSQL DSN. Assembled from DB_* vars when
	// DATABASE_URL is not set explicitly.
	DatabaseURL string

	// JWTSecret is the HS256 key. It MUST equal the auth service's JWT_SECRET
	// (A-08) so tokens it issues validate here.
	JWTSecret []byte
	// JWTIssuer is the required "iss" claim ("mergemarket-auth").
	JWTIssuer string
	// JWTLeeway tolerates small clock skew when checking token expiry.
	JWTLeeway time.Duration
}

// Load reads configuration from the environment, applies defaults, and validates.
func Load() (*Config, error) {
	cfg := &Config{
		Port:      getEnv("USERDATA_PORT", "8090"),
		JWTSecret: []byte(getEnv("JWT_SECRET", "")),
		JWTIssuer: getEnv("USERDATA_JWT_ISSUER", "mergemarket-auth"),
		JWTLeeway: 30 * time.Second,
	}
	cfg.DatabaseURL = databaseURL()

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
		return fmt.Errorf("config: USERDATA_PORT must not be empty")
	}
	if c.DatabaseURL == "" {
		return fmt.Errorf("config: DATABASE_URL could not be determined")
	}
	if len(c.JWTSecret) == 0 {
		return fmt.Errorf("config: JWT_SECRET must be set (shared with the auth service)")
	}
	if c.JWTIssuer == "" {
		return fmt.Errorf("config: USERDATA_JWT_ISSUER must not be empty")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
