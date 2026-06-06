//go:build integration

package integration_test

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// startService launches a built binary with env, waits for /health on port, and
// returns a cleanup func. Shared by the gated search/userdata live cases.
func startService(t *testing.T, bin string, env []string, port string) func() {
	t.Helper()
	cmd := exec.Command(bin)
	cmd.Env = append(env, "PATH="+os.Getenv("PATH"))
	require.NoError(t, cmd.Start())
	cleanup := func() { _ = cmd.Process.Kill() }
	if !waitHealthy(port, 10*time.Second) {
		cleanup()
		t.Fatalf("service did not become healthy on port %s", port)
	}
	return cleanup
}

// TestSearchServiceIntegration drives the real A-14 search binary.
func TestSearchServiceIntegration(t *testing.T) {
	bin := buildBinary(t, "search", "search")

	t.Run("TC-14-I-001: fails fast when Postgres is unreachable", func(t *testing.T) {
		out := runExpectNonZero(t, bin, []string{
			"SEARCH_PORT=" + freePort(t),
			"DB_HOST=127.0.0.1", "DB_PORT=1", // refused
			"REDIS_HOST=127.0.0.1", "REDIS_PORT=1",
		})
		assert.Contains(t, out, "search-service exited with error",
			"service should log and exit when its hard Postgres dependency is down")
	})

	t.Run("TC-14-I-002: live search returns the contract shape", func(t *testing.T) {
		dsn := requireTestDB(t) // PENDING unless DB_TEST_DSN set
		port := freePort(t)
		env := []string{"SEARCH_PORT=" + port, "DATABASE_URL=" + dsn}
		if r := os.Getenv("REDIS_TEST_ADDR"); r != "" {
			// REDIS_HOST/PORT split is handled by the service; pass through addr parts.
			env = append(env, "REDIS_HOST="+hostOf(r), "REDIS_PORT="+portOf(r))
		}
		defer startService(t, bin, env, port)()

		resp, err := http.Get("http://localhost:" + port + "/api/v1/search?q=phone&location=CM")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var body map[string]any
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
		assert.Contains(t, body, "results")
		assert.Contains(t, body, "cached")
		assert.Contains(t, body, "latency_ms")
	})

	t.Run("TC-14-I-003: missing query parameter yields 400 missing_query", func(t *testing.T) {
		dsn := requireTestDB(t) // PENDING unless DB_TEST_DSN set
		port := freePort(t)
		defer startService(t, bin, []string{"SEARCH_PORT=" + port, "DATABASE_URL=" + dsn}, port)()

		resp, err := http.Get("http://localhost:" + port + "/api/v1/search?location=CM")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		var body map[string]any
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
		assert.Equal(t, "missing_query", body["error"])
	})

	t.Run("TC-14-I-004: identical query is served from cache on the second call", func(t *testing.T) {
		dsn := requireTestDB(t)
		if os.Getenv("REDIS_TEST_ADDR") == "" {
			t.Skip("PENDING: set REDIS_TEST_ADDR to verify the search cache")
		}
		r := os.Getenv("REDIS_TEST_ADDR")
		port := freePort(t)
		env := []string{"SEARCH_PORT=" + port, "DATABASE_URL=" + dsn, "REDIS_HOST=" + hostOf(r), "REDIS_PORT=" + portOf(r)}
		defer startService(t, bin, env, port)()

		url := "http://localhost:" + port + "/api/v1/search?q=cachecheck&location=CM"
		first := decodeSearch(t, url)
		assert.Equal(t, false, first["cached"], "first call should be uncached")
		second := decodeSearch(t, url)
		assert.Equal(t, true, second["cached"], "second identical call should be cached")
	})
}

func decodeSearch(t *testing.T, url string) map[string]any {
	t.Helper()
	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()
	var body map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	return body
}

func hostOf(addr string) string {
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return addr[:i]
		}
	}
	return addr
}

func portOf(addr string) string {
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return addr[i+1:]
		}
	}
	return ""
}
