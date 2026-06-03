// Package compose_test holds the static (unit) validation of the A-02
// local-dev infrastructure artefacts. These tests run without a Docker
// daemon: they parse/read the committed files and assert their structure.
//
// Run from the repo with module downloads available (e.g. CI / A-10):
//
//	go test ./docs/testing/session-02/unit/test_suite/...
package compose_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// findRepoRoot walks up from the test's working directory until it finds the
// repository root (identified by docker-compose.yml).
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

func readFile(t *testing.T, root, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, rel))
	require.NoError(t, err, "reading %s", rel)
	return string(b)
}

func TestA02LocalDevArtefacts(t *testing.T) {
	root := findRepoRoot(t)

	t.Run("TC-02-U-001: docker-compose.yml parses", func(t *testing.T) {
		// Arrange: docker compose reads ./.env automatically; ensure one exists.
		// (CI provisions .env from .env.example before this runs.)
		cmd := exec.Command("docker", "compose", "config")
		cmd.Dir = root
		out, err := cmd.CombinedOutput()
		assert.NoErrorf(t, err, "docker compose config failed: %s", out)
	})

	t.Run("TC-02-U-002: required services defined", func(t *testing.T) {
		compose := readFile(t, root, "docker-compose.yml")
		for _, svc := range []string{"postgres:", "redis:", "kong:", "sonarqube:", "sonar-db:"} {
			assert.Containsf(t, compose, svc, "missing service %q", svc)
		}
	})

	t.Run("TC-02-U-003: every service declares a healthcheck", func(t *testing.T) {
		compose := readFile(t, root, "docker-compose.yml")
		// One healthcheck test line per service (5 services).
		assert.Equal(t, 5, strings.Count(compose, "test: ["), "expected 5 healthcheck definitions")
	})

	t.Run("TC-02-U-004: all named volumes declared", func(t *testing.T) {
		compose := readFile(t, root, "docker-compose.yml")
		for _, v := range []string{"pgdata:", "redisdata:", "sonar_db_data:", "sonarqube_data:", "sonarqube_extensions:", "sonarqube_logs:"} {
			assert.Containsf(t, compose, v, "missing volume %q", v)
		}
	})

	t.Run("TC-02-U-005: expected published ports present", func(t *testing.T) {
		compose := readFile(t, root, "docker-compose.yml")
		for _, p := range []string{"5432", "6379", "8000", "8443", "8001", "8444", "9000"} {
			assert.Containsf(t, compose, p, "missing port %q", p)
		}
	})

	t.Run("TC-02-U-006: .env.example completeness", func(t *testing.T) {
		env := readFile(t, root, ".env.example")
		required := []string{
			"DB_HOST", "DB_PORT", "DB_NAME", "DB_USER", "DB_PASSWORD",
			"TIMESCALE_HOST", "TIMESCALE_PORT",
			"REDIS_HOST", "REDIS_PORT", "REDIS_PASSWORD",
			"JWT_SECRET", "JWT_EXPIRY_HOURS", "REFRESH_TOKEN_EXPIRY_DAYS",
			"AUTH_PORT", "BFF_PORT", "SCRAPER_PORT", "NORMALIZATION_PORT",
			"HISTORY_PORT", "PROXY_VALIDATOR_PORT",
			"SONAR_HOST_URL", "SONAR_TOKEN", "FIREBASE_SERVER_KEY",
		}
		for _, k := range required {
			assert.Containsf(t, env, k+"=", "missing env key %q", k)
		}
	})

	t.Run("TC-02-U-007: kong.yml is a valid declarative doc", func(t *testing.T) {
		kong := readFile(t, root, "api-gateway/kong.yml")
		assert.Contains(t, kong, "_format_version")
		assert.Contains(t, kong, "services:")
		assert.Contains(t, kong, "routes:")
	})

	t.Run("TC-02-U-008: bootstrap schema declares all objects", func(t *testing.T) {
		sql := readFile(t, root, "infra/db/init/01-schema.sql")
		for _, obj := range []string{
			"CREATE EXTENSION IF NOT EXISTS timescaledb",
			"TABLE IF NOT EXISTS users", "TABLE IF NOT EXISTS stores",
			"TABLE IF NOT EXISTS products", "TABLE IF NOT EXISTS wishlist_items",
			"TABLE IF NOT EXISTS price_alerts", "TABLE IF NOT EXISTS return_policies",
			"TABLE IF NOT EXISTS scrape_jobs", "TABLE IF NOT EXISTS price_history",
			"job_status", "create_hypertable",
		} {
			assert.Containsf(t, sql, obj, "schema missing %q", obj)
		}
	})

	t.Run("TC-02-U-009: single Postgres+Timescale connection", func(t *testing.T) {
		env := readFile(t, root, ".env.example")
		// Both DB_* and TIMESCALE_* must point at the same host:port.
		assert.Contains(t, env, "DB_HOST=localhost")
		assert.Contains(t, env, "TIMESCALE_HOST=localhost")
		assert.Contains(t, env, "DB_PORT=5432")
		assert.Contains(t, env, "TIMESCALE_PORT=5432")
	})
}
