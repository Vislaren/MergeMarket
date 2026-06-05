// Package proxyvalidator_test is the Session-04 QE unit suite for Agent A's
// A-04 proxy-validator service.
//
// A-04's behaviour lives in `internal/` packages, which Go will not let a
// different module import. So this suite verifies A-04 two ways, neither of
// which needs to import those packages:
//
//  1. Subprocess: it runs `go build`, `go test`, `go vet` and `gofmt` against
//     the service itself — confirming it compiles, the developer's own tests
//     pass, and it is vet/format clean.
//  2. Structural: it reads the committed source and asserts the contract-level
//     facts the oracle requires (env-only config, /health shape, proxy_pool
//     Set + 5m TTL, atomic swap, adaptive politeness, resilience, GoDoc).
//
// Run (module cache must be populated; the service's go.sum is committed):
//
//	go test ./docs/testing/session-04/unit/test_suite/...
package proxyvalidator_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// findRepoRoot walks up from the test working directory until it finds the
// repository root (identified by docker-compose.yml, same marker Session-02 used).
func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err)
	for {
		if _, err := os.Stat(filepath.Join(dir, "docker-compose.yml")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		require.NotEqual(t, parent, dir, "reached filesystem root without finding docker-compose.yml")
		dir = parent
	}
}

func serviceDir(t *testing.T) string {
	t.Helper()
	return filepath.Join(findRepoRoot(t), "services", "proxy-validator")
}

// runGo runs a go-toolchain command in the service directory and returns its
// combined output.
func runGo(t *testing.T, svc string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = svc
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func read(t *testing.T, svc, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(svc, rel))
	require.NoError(t, err, "reading %s", rel)
	return string(b)
}

func TestA04ProxyValidator(t *testing.T) {
	svc := serviceDir(t)

	t.Run("TC-04-U-001: service compiles", func(t *testing.T) {
		out, err := runGo(t, svc, "go", "build", "./...")
		assert.NoErrorf(t, err, "go build failed:\n%s", out)
	})

	t.Run("TC-04-U-002: developer unit tests pass", func(t *testing.T) {
		out, err := runGo(t, svc, "go", "test", "./...")
		assert.NoErrorf(t, err, "go test failed:\n%s", out)
	})

	t.Run("TC-04-U-003: go vet is clean", func(t *testing.T) {
		out, err := runGo(t, svc, "go", "vet", "./...")
		assert.NoErrorf(t, err, "go vet reported issues:\n%s", out)
	})

	t.Run("TC-04-U-004: gofmt is clean", func(t *testing.T) {
		out, err := runGo(t, svc, "gofmt", "-l", ".")
		require.NoError(t, err)
		assert.Empty(t, strings.TrimSpace(out), "gofmt -l listed unformatted files:\n%s", out)
	})

	t.Run("TC-04-U-005: configuration is environment-driven (NFR: no hardcoded config)", func(t *testing.T) {
		cfg := read(t, svc, "internal/config/config.go")
		assert.Contains(t, cfg, "os.LookupEnv", "config must read from the environment")
		for _, key := range []string{"PROXY_VALIDATOR_PORT", "REDIS_HOST", "REDIS_PORT", "PROXY_POOL_TTL"} {
			assert.Containsf(t, cfg, key, "config must honour %s", key)
		}
		// main wires the listen address from config, not a literal.
		main := read(t, svc, "cmd/proxy-validator/main.go")
		assert.Contains(t, main, `":"+cfg.Port`, "server port must come from config, not a literal")
	})

	t.Run("TC-04-U-006: /health matches the API contract shape", func(t *testing.T) {
		srv := read(t, svc, "internal/server/server.go")
		assert.Contains(t, srv, `"/health"`, "must register GET /health")
		for _, f := range []string{"Status", "Service", "Version"} {
			assert.Containsf(t, srv, f, "health response missing %s field", f)
		}
		assert.Contains(t, srv, `"proxy-validator"`, "service name must be proxy-validator")
	})

	t.Run("TC-04-U-007: writes proxy_pool Set with a 5m TTL (DATABASE_SCHEMA §3)", func(t *testing.T) {
		cfg := read(t, svc, "internal/config/config.go")
		assert.Contains(t, cfg, `"proxy_pool"`, "default pool key must be proxy_pool")
		assert.Contains(t, cfg, "5*time.Minute", "default pool TTL must be 5 minutes")
		store := read(t, svc, "internal/store/store.go")
		assert.Contains(t, store, "SAdd", "pool must be a Redis Set (SADD)")
		assert.Contains(t, store, "Expire", "pool must carry a TTL (EXPIRE)")
	})

	t.Run("TC-04-U-008: pool is replaced atomically (no transiently-empty pool)", func(t *testing.T) {
		store := read(t, svc, "internal/store/store.go")
		assert.Contains(t, store, "Rename", "atomic swap must use RENAME of a staging key")
		assert.Contains(t, store, "staging", "must stage the new set under a temp key first")
	})

	t.Run("TC-04-U-009: politeness uses an adaptive random delay", func(t *testing.T) {
		pol := read(t, svc, "internal/politeness/politeness.go")
		assert.Contains(t, pol, "rand", "delay must be randomised")
		assert.Contains(t, pol, "func (l *Limiter) Failure()", "must back off on failure")
		assert.Contains(t, pol, "func (l *Limiter) Success()", "must relax on success")
		// The runner must actually pace dispatch and feed outcomes back.
		run := read(t, svc, "internal/runner/runner.go")
		assert.Contains(t, run, "limiter.Wait", "runner must wait politely between dispatches")
		assert.Contains(t, run, "limiter.Failure", "runner must report failures to the limiter")
		assert.Contains(t, run, "limiter.Success", "runner must report successes to the limiter")
	})

	t.Run("TC-04-U-010: resilient to partial source failure (NFR-2)", func(t *testing.T) {
		f := read(t, svc, "internal/fetcher/fetcher.go")
		// One bad source must not abort the others: only error when ALL fail.
		assert.Contains(t, f, "failures == len(f.sources)", "must only fail when every source fails")
		assert.Contains(t, f, "continue", "a failing source must be skipped, not fatal")
	})

	t.Run("TC-04-U-011: default port is 8086 (ARCHITECTURE §2)", func(t *testing.T) {
		cfg := read(t, svc, "internal/config/config.go")
		assert.Contains(t, cfg, `getEnv("PROXY_VALIDATOR_PORT", "8086")`, "default port must be 8086")
	})

	t.Run("TC-04-U-012: every package carries a GoDoc package comment", func(t *testing.T) {
		pkgFiles := []string{
			"internal/config/config.go", "internal/proxy/proxy.go",
			"internal/fetcher/fetcher.go", "internal/validator/validator.go",
			"internal/politeness/politeness.go", "internal/store/store.go",
			"internal/runner/runner.go", "internal/server/server.go",
		}
		for _, f := range pkgFiles {
			src := read(t, svc, f)
			assert.Containsf(t, src, "// Package ", "%s is missing a // Package doc comment", f)
		}
		// main is a command; Go convention documents it as "// Command ...".
		main := read(t, svc, "cmd/proxy-validator/main.go")
		assert.Contains(t, main, "// Command proxy-validator", "main.go is missing its command doc comment")
	})
}
