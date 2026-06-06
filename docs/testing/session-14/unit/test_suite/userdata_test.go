package unit_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUserdataServiceUnit verifies Agent A's A-16..A-18 consolidated userdata
// service (wishlist + alerts + savings).
func TestUserdataServiceUnit(t *testing.T) {
	dir := serviceDir(t, "userdata")

	t.Run("TC-14-U-014: service compiles (go build ./...)", func(t *testing.T) {
		out, err := runTool(t, dir, "go", "build", "./...")
		require.NoError(t, err, "go build failed:\n%s", out)
	})

	t.Run("TC-14-U-015: developer in-package tests pass (go test ./...)", func(t *testing.T) {
		out, err := runTool(t, dir, "go", "test", "./...")
		require.NoError(t, err, "go test failed:\n%s", out)
	})

	t.Run("TC-14-U-016: go vet clean", func(t *testing.T) {
		out, err := runTool(t, dir, "go", "vet", "./...")
		require.NoError(t, err, "go vet reported issues:\n%s", out)
	})

	t.Run("TC-14-U-017: gofmt clean", func(t *testing.T) {
		out, err := runTool(t, dir, "gofmt", "-l", ".")
		require.NoError(t, err)
		assert.Empty(t, strings.TrimSpace(out), "gofmt lists unformatted files:\n%s", out)
	})

	t.Run("TC-14-U-018: env-driven (USERDATA_PORT 8090) and requires JWT_SECRET", func(t *testing.T) {
		cfg := sourceOf(t, dir+"/internal/config")
		assert.Contains(t, cfg, "USERDATA_PORT")
		assert.Contains(t, cfg, "8090")
		assert.Contains(t, cfg, "JWT_SECRET")
		assert.Contains(t, cfg, "JWT_SECRET must be set", "config must reject an empty JWT_SECRET")
	})

	t.Run("TC-14-U-019: all 7 API routes + /health are registered", func(t *testing.T) {
		srv := sourceOf(t, dir+"/internal/server")
		for _, route := range []string{
			`"GET /health"`,
			`"GET /api/v1/wishlist"`,
			`"POST /api/v1/wishlist"`,
			`"DELETE /api/v1/wishlist/{wishlist_id}"`,
			`"GET /api/v1/alerts"`,
			`"POST /api/v1/alerts"`,
			`"DELETE /api/v1/alerts/{alert_id}"`,
			`"GET /api/v1/savings"`,
		} {
			assert.Contains(t, srv, route, "route %s not registered", route)
		}
	})

	t.Run("TC-14-U-020: every API route is JWT-protected; /health is not", func(t *testing.T) {
		srv := sourceOf(t, dir+"/internal/server")
		// API routes go through the authed() wrapper.
		assert.Contains(t, srv, "h.authed(h.listWishlist)")
		assert.Contains(t, srv, "h.authed(h.addWishlist)")
		assert.Contains(t, srv, "h.authed(h.createAlert)")
		assert.Contains(t, srv, "h.authed(h.savings)")
		// health is registered directly (no auth wrapper).
		assert.Contains(t, srv, `"GET /health", healthHandler(version)`)
	})

	t.Run("TC-14-U-021: token verifier checks HS256 signature, expiry, and issuer", func(t *testing.T) {
		tok := sourceOf(t, dir+"/internal/token")
		assert.Contains(t, tok, "HS256")
		assert.Contains(t, tok, "hmac.Equal")
		assert.Contains(t, tok, "ErrExpired")
		assert.Contains(t, tok, ".Iss")
	})

	t.Run("TC-14-U-022: persistence is scoped by user_id (no cross-user access)", func(t *testing.T) {
		pg := sourceOf(t, dir+"/internal/store")
		assert.Contains(t, pg, "user_id = $", "reads/writes must filter by user_id")
		// Deletes require BOTH the row id and the owning user.
		assert.Contains(t, pg, "WHERE id = $1 AND user_id = $2")
	})

	t.Run("TC-14-U-023: contract error codes are emitted", func(t *testing.T) {
		srv := sourceOf(t, dir+"/internal/server")
		for _, code := range []string{"already_in_wishlist", "not_found", "invalid_input", "token_expired", "unauthorized"} {
			assert.Contains(t, srv, code, "missing error code %q", code)
		}
	})

	t.Run("TC-14-U-024: purchases table is defined (EnsureSchema + canonical init SQL)", func(t *testing.T) {
		pg := sourceOf(t, dir+"/internal/store")
		assert.Contains(t, pg, "CREATE TABLE IF NOT EXISTS purchases")
		sql := readFile(t, "infra/db/init/01-schema.sql")
		assert.Contains(t, sql, "CREATE TABLE IF NOT EXISTS purchases")
		assert.Contains(t, sql, "idx_purchases_user_id")
	})

	t.Run("TC-14-U-025: every internal package has a GoDoc comment", func(t *testing.T) {
		eachInternalPackageHasGoDoc(t, dir)
	})
}
