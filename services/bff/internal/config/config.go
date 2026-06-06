// Package config loads the BFF's runtime configuration from the environment.
// Every value has a safe default so the service boots for local development
// with no configuration, pointing at the B-02 mock server.
package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
)

// Service identity returned by /health and used in logs.
const (
	// ServiceName is the canonical name of this service.
	ServiceName = "bff"
	// Version is the BFF's semantic version.
	Version = "0.1.0"
)

// defaultPort is the BFF listen port when BFF_PORT / PORT are unset (8082 per
// the project port map).
const defaultPort = 8082

// defaultUpstream is the API the BFF forwards to when BFF_UPSTREAM_URL is unset.
// Locally that is the B-02 mock server (8089); in production it is the Kong
// gateway / real services.
const defaultUpstream = "http://localhost:8089"

// Config holds all runtime configuration for the BFF.
type Config struct {
	// Addr is the TCP address the HTTP server listens on, e.g. ":8082".
	Addr string
	// Port is the numeric listen port.
	Port int
	// UpstreamURL is the base URL of the API the BFF forwards/aggregates from.
	UpstreamURL string
}

// Load reads configuration from the environment and validates it.
//
// BFF_PORT takes precedence over PORT; if neither is set the default port
// (8082) is used. BFF_UPSTREAM_URL overrides the default upstream. An invalid
// port or upstream URL is a fatal error so the problem surfaces at startup.
func Load() (Config, error) {
	port := defaultPort
	if raw := firstNonEmpty(os.Getenv("BFF_PORT"), os.Getenv("PORT")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return Config{}, fmt.Errorf("config: invalid port %q: %w", raw, err)
		}
		if parsed < 1 || parsed > 65535 {
			return Config{}, fmt.Errorf("config: port %d out of range (1-65535)", parsed)
		}
		port = parsed
	}

	upstream := firstNonEmpty(os.Getenv("BFF_UPSTREAM_URL"), defaultUpstream)
	u, err := url.Parse(upstream)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return Config{}, fmt.Errorf("config: invalid BFF_UPSTREAM_URL %q", upstream)
	}

	return Config{
		Addr:        fmt.Sprintf(":%d", port),
		Port:        port,
		UpstreamURL: upstream,
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
