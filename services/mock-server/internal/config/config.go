// Package config loads the mock server's runtime configuration from the
// environment. Every value has a safe default so the server boots with no
// configuration at all — the common case for local offline development.
package config

import (
	"fmt"
	"os"
	"strconv"
)

// Service identity returned by the /health endpoint and used in logs.
const (
	// ServiceName is the canonical name of this service.
	ServiceName = "mock-server"
	// Version is the mock server's semantic version.
	Version = "0.1.0"
)

// defaultPort is the listen port when MOCK_SERVER_PORT / PORT are unset.
// 8080 is reserved for the mock server; real services use 8081–8086.
const defaultPort = 8080

// Config holds all runtime configuration for the mock server.
type Config struct {
	// Addr is the TCP address the HTTP server listens on, e.g. ":8080".
	Addr string
	// Port is the numeric listen port.
	Port int
}

// Load reads configuration from the environment and validates it.
//
// MOCK_SERVER_PORT takes precedence over PORT; if neither is set the default
// port (8080) is used. An invalid or out-of-range port is a fatal error so the
// problem surfaces at startup rather than as a silent fallback.
func Load() (Config, error) {
	port := defaultPort

	if raw := firstNonEmpty(os.Getenv("MOCK_SERVER_PORT"), os.Getenv("PORT")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return Config{}, fmt.Errorf("config: invalid port %q: %w", raw, err)
		}
		if parsed < 1 || parsed > 65535 {
			return Config{}, fmt.Errorf("config: port %d out of range (1-65535)", parsed)
		}
		port = parsed
	}

	return Config{
		Addr: fmt.Sprintf(":%d", port),
		Port: port,
	}, nil
}

// firstNonEmpty returns the first argument that is not the empty string.
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
