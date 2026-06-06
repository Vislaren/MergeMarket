//go:build integration

package integration_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUserdataServiceIntegration drives the real A-16..A-18 userdata binary.
func TestUserdataServiceIntegration(t *testing.T) {
	bin := buildBinary(t, "userdata", "userdata")

	t.Run("TC-14-I-005: refuses to start without JWT_SECRET", func(t *testing.T) {
		out := runExpectNonZero(t, bin, []string{
			"USERDATA_PORT=" + freePort(t),
			"DB_HOST=127.0.0.1", "DB_PORT=1",
			// JWT_SECRET deliberately absent — config.Load must fail before any DB work.
		})
		assert.Contains(t, out, "JWT_SECRET",
			"service must refuse to start when the shared JWT_SECRET is unset")
	})

	t.Run("TC-14-I-006: fails fast when Postgres is unreachable", func(t *testing.T) {
		out := runExpectNonZero(t, bin, []string{
			"USERDATA_PORT=" + freePort(t),
			"JWT_SECRET=test-secret",
			"DB_HOST=127.0.0.1", "DB_PORT=1", // refused
		})
		assert.Contains(t, out, "userdata-service exited with error",
			"service should log and exit when its hard Postgres dependency is down")
	})

	t.Run("TC-14-I-007: protected route returns 401 without a bearer token", func(t *testing.T) {
		dsn := requireTestDB(t) // PENDING unless DB_TEST_DSN set
		port := freePort(t)
		defer startService(t, bin, []string{
			"USERDATA_PORT=" + port, "JWT_SECRET=test-secret", "DATABASE_URL=" + dsn,
		}, port)()

		resp, err := http.Get("http://localhost:" + port + "/api/v1/wishlist")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		var body map[string]any
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
		assert.Equal(t, "unauthorized", body["error"])
	})

	t.Run("TC-14-I-008: full live E2E (Kong->BFF->search/userdata; search->wishlist->alert)", func(t *testing.T) {
		// Carries forward the session-13 TC-13-I E2E that was blocked on these
		// services existing. Still PENDING locally: it needs the WHOLE stack up
		// (Kong + BFF + auth + search + userdata + Postgres + Redis), which is not
		// runnable without Docker here. Executable in CI / a deployed environment.
		requireTestDB(t)
		t.Skip("PENDING: full-stack E2E requires Kong+BFF+auth+search+userdata+Postgres+Redis running together")
	})
}
