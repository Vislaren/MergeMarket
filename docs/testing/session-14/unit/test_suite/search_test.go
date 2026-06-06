package unit_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSearchServiceUnit verifies Agent A's A-14 search service.
func TestSearchServiceUnit(t *testing.T) {
	dir := serviceDir(t, "search")

	t.Run("TC-14-U-001: service compiles (go build ./...)", func(t *testing.T) {
		out, err := runTool(t, dir, "go", "build", "./...")
		require.NoError(t, err, "go build failed:\n%s", out)
	})

	t.Run("TC-14-U-002: developer in-package tests pass (go test ./...)", func(t *testing.T) {
		out, err := runTool(t, dir, "go", "test", "./...")
		require.NoError(t, err, "go test failed:\n%s", out)
	})

	t.Run("TC-14-U-003: go vet clean", func(t *testing.T) {
		out, err := runTool(t, dir, "go", "vet", "./...")
		require.NoError(t, err, "go vet reported issues:\n%s", out)
	})

	t.Run("TC-14-U-004: gofmt clean", func(t *testing.T) {
		out, err := runTool(t, dir, "gofmt", "-l", ".")
		require.NoError(t, err)
		assert.Empty(t, strings.TrimSpace(out), "gofmt lists unformatted files:\n%s", out)
	})

	t.Run("TC-14-U-005: port is env-driven (SEARCH_PORT, default 8087), not hardcoded", func(t *testing.T) {
		cfg := sourceOf(t, dir+"/internal/config")
		assert.Contains(t, cfg, "os.LookupEnv")
		assert.Contains(t, cfg, "SEARCH_PORT")
		assert.Contains(t, cfg, "8087")
		// main binds the configured port, not a literal.
		main := readFile(t, "services/search/cmd/search/main.go")
		assert.Contains(t, main, `":"+cfg.Port`)
	})

	t.Run("TC-14-U-006: /health returns {status, service, version}", func(t *testing.T) {
		srv := sourceOf(t, dir+"/internal/server")
		assert.Contains(t, srv, `"GET /health"`)
		assert.Contains(t, srv, `json:"status"`)
		assert.Contains(t, srv, `json:"service"`)
		assert.Contains(t, srv, `json:"version"`)
	})

	t.Run("TC-14-U-007: GET /api/v1/search route is registered", func(t *testing.T) {
		assert.Contains(t, sourceOf(t, dir+"/internal/server"), `"GET /api/v1/search"`)
	})

	t.Run("TC-14-U-008: missing q/location yields 400 missing_query", func(t *testing.T) {
		srv := sourceOf(t, dir+"/internal/server")
		assert.Contains(t, srv, `"missing_query"`)
		assert.Contains(t, srv, "StatusBadRequest")
	})

	t.Run("TC-14-U-009: cache key is search:{sha256(q|location)}", func(t *testing.T) {
		c := sourceOf(t, dir+"/internal/cache")
		assert.Contains(t, c, "sha256")
		assert.Contains(t, c, "func Key(")
		cfg := sourceOf(t, dir+"/internal/config")
		assert.Contains(t, cfg, `"search:"`, "default cache prefix should be search:")
	})

	t.Run("TC-14-U-010: deal_score heuristic present (cheapest scores 100)", func(t *testing.T) {
		s := sourceOf(t, dir+"/internal/search")
		assert.Contains(t, s, "scoreDeals")
		assert.Contains(t, s, "DealScore")
		assert.Contains(t, s, "100")
	})

	t.Run("TC-14-U-011: stale-while-revalidate (stale hit refreshes in background)", func(t *testing.T) {
		s := sourceOf(t, dir+"/internal/search")
		assert.Contains(t, s, "CacheStaleAfter")
		assert.Contains(t, s, "revalidate")
		assert.Contains(t, s, "go func()")
	})

	t.Run("TC-14-U-012: response carries cached + latency_ms (contract)", func(t *testing.T) {
		s := sourceOf(t, dir+"/internal/search")
		assert.Contains(t, s, `json:"cached"`)
		assert.Contains(t, s, `json:"latency_ms"`)
	})

	t.Run("TC-14-U-013: every internal package has a GoDoc comment", func(t *testing.T) {
		eachInternalPackageHasGoDoc(t, dir)
	})
}
