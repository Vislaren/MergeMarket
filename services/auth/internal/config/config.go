// Package config loads every auth-service setting from environment variables.
package config

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all runtime settings for the auth-service.
type Config struct {
	Port            string
	DatabaseURL     string
	RedisAddr       string
	RedisPassword   string
	RedisDB         int
	JWTSecret       []byte
	JWTTTL          time.Duration
	RefreshTokenTTL time.Duration
	SessionTTL      time.Duration
	EncryptionKey   []byte
	TLSCertFile     string
	TLSKeyFile      string
}

// Load reads and validates auth-service configuration from the process env.
func Load() (*Config, error) {
	cfg := &Config{
		Port:          getEnv("AUTH_PORT", "8091"),
		DatabaseURL:   databaseURL(),
		RedisAddr:     getEnv("REDIS_HOST", "localhost") + ":" + getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		TLSCertFile:   getEnv("AUTH_TLS_CERT_FILE", ""),
		TLSKeyFile:    getEnv("AUTH_TLS_KEY_FILE", ""),
	}

	var err error
	if cfg.RedisDB, err = getEnvInt("REDIS_DB", 0); err != nil {
		return nil, err
	}
	if cfg.JWTSecret, err = secretBytes("JWT_SECRET", 32); err != nil {
		return nil, err
	}
	if cfg.EncryptionKey, err = secretBytes("AUTH_ENCRYPTION_KEY", 32); err != nil {
		return nil, err
	}
	if cfg.JWTTTL, err = getEnvHours("JWT_EXPIRY_HOURS", time.Hour); err != nil {
		return nil, err
	}
	if cfg.RefreshTokenTTL, err = getEnvDays("REFRESH_TOKEN_EXPIRY_DAYS", 30*24*time.Hour); err != nil {
		return nil, err
	}
	if cfg.SessionTTL, err = getEnvDuration("AUTH_SESSION_TTL", time.Hour); err != nil {
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
		return fmt.Errorf("config: AUTH_PORT must not be empty")
	}
	if c.DatabaseURL == "" {
		return fmt.Errorf("config: database URL could not be determined")
	}
	if c.JWTTTL <= 0 {
		return fmt.Errorf("config: JWT_EXPIRY_HOURS must be > 0")
	}
	if c.RefreshTokenTTL <= 0 {
		return fmt.Errorf("config: REFRESH_TOKEN_EXPIRY_DAYS must be > 0")
	}
	if c.SessionTTL != time.Hour {
		return fmt.Errorf("config: AUTH_SESSION_TTL must remain 1h to match the service contract")
	}
	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("config: JWT_SECRET must decode to at least 32 bytes")
	}
	if len(c.EncryptionKey) != 32 {
		return fmt.Errorf("config: AUTH_ENCRYPTION_KEY must decode to exactly 32 bytes")
	}
	if (c.TLSCertFile == "") != (c.TLSKeyFile == "") {
		return fmt.Errorf("config: AUTH_TLS_CERT_FILE and AUTH_TLS_KEY_FILE must be set together")
	}
	if c.TLSCertFile == "" {
		return fmt.Errorf("config: TLS 1.3 is required; set AUTH_TLS_CERT_FILE and AUTH_TLS_KEY_FILE")
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

func getEnvHours(key string, fallback time.Duration) (time.Duration, error) {
	n, err := getEnvInt(key, int(fallback/time.Hour))
	if err != nil {
		return 0, err
	}
	return time.Duration(n) * time.Hour, nil
}

func getEnvDays(key string, fallback time.Duration) (time.Duration, error) {
	n, err := getEnvInt(key, int(fallback/(24*time.Hour)))
	if err != nil {
		return 0, err
	}
	return time.Duration(n) * 24 * time.Hour, nil
}

func getEnvDuration(key string, fallback time.Duration) (time.Duration, error) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("config: %s must be a Go duration: %w", key, err)
	}
	return d, nil
}

func secretBytes(key string, min int) ([]byte, error) {
	raw := getEnv(key, "")
	if raw == "" {
		return nil, fmt.Errorf("config: %s must be set", key)
	}
	if b, err := hex.DecodeString(raw); err == nil && len(b) >= min {
		return b, nil
	}
	if b, err := base64.StdEncoding.DecodeString(raw); err == nil && len(b) >= min {
		return b, nil
	}
	if len(raw) >= min {
		return []byte(raw), nil
	}
	return nil, fmt.Errorf("config: %s must be hex, base64, or raw text with at least %d bytes", key, min)
}
