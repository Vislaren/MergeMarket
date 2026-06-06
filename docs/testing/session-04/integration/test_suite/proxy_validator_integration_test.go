//go:build integration

// Package proxyvalidator_it is the Session-04 QE integration suite for Agent A's
// A-04 proxy-validator service. Tagged `integration` so it is excluded from
// default unit runs.
//
// It exercises the REAL compiled binary end-to-end:
//
//	source(httptest) ──▶ proxy-validator ──validate via fake proxy──▶ Redis proxy_pool
//
// A local httptest server plays the public proxy LIST; a second local httptest
// server plays a WORKING PROXY (it answers 204 to whatever it is asked to relay,
// so the validator accepts it). The service then writes the working proxy to a
// live Redis, which the test reads back to assert Set membership + TTL.
//
// Prereqs: a Go toolchain (to build the binary) and a live Redis. Without Redis
// the live-pipeline cases skip; the /health case needs only the binary.
//
//	REDIS_TEST_ADDR=localhost:6379 \
//	  go test -tags=integration ./docs/testing/session-04/integration/test_suite/...
package proxyvalidator_it

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const poolKey = "proxy_pool:it-session04"

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

// buildBinary compiles the proxy-validator command to a temp path.
func buildBinary(t *testing.T) string {
	t.Helper()
	root := findRepoRoot(t)
	svc := filepath.Join(root, "services", "proxy-validator")
	bin := filepath.Join(t.TempDir(), "proxy-validator-it")
	if os.PathSeparator == '\\' {
		bin += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/proxy-validator")
	cmd.Dir = svc
	out, err := cmd.CombinedOutput()
	require.NoErrorf(t, err, "build binary:\n%s", out)
	return bin
}

func redisAddr() string {
	if a := os.Getenv("REDIS_TEST_ADDR"); a != "" {
		return a
	}
	return "localhost:6379"
}

// liveRedis returns a client if Redis is reachable, else skips the test.
func liveRedis(t *testing.T) *redis.Client {
	t.Helper()
	c := redis.NewClient(&redis.Options{Addr: redisAddr(), Password: os.Getenv("REDIS_TEST_PASSWORD")})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := c.Ping(ctx).Err(); err != nil {
		_ = c.Close()
		t.Skipf("Redis not reachable at %s: %v — skipping live-pipeline case", redisAddr(), err)
	}
	return c
}

// startService launches the binary with the given env and returns its base URL
// plus a cancel func. host:port for /health is fixed by the caller via env.
func startService(t *testing.T, bin string, env map[string]string) (*exec.Cmd, string) {
	t.Helper()
	cmd := exec.Command(bin)
	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	require.NoError(t, cmd.Start())
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	})
	return cmd, "http://127.0.0.1:" + env["PROXY_VALIDATOR_PORT"]
}

func getJSON(t *testing.T, url string, into any) bool {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false
	}
	return json.NewDecoder(resp.Body).Decode(into) == nil
}

func waitUntil(t *testing.T, d time.Duration, cond func() bool) bool {
	t.Helper()
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return cond()
}

// TC-04-I-001: the binary serves the /health contract.
func TestHealthEndpoint(t *testing.T) {
	bin := buildBinary(t)
	_, base := startService(t, bin, map[string]string{
		"PROXY_VALIDATOR_PORT":   "18186",
		"PROXY_SOURCES":          "http://127.0.0.1:9/none", // unreachable; health must still serve
		"PROXY_REFRESH_INTERVAL": "4m",
		"PROXY_POOL_TTL":         "5m",
		"REDIS_HOST":             "127.0.0.1",
		"REDIS_PORT":             "9",
	})

	var health struct {
		Status, Service, Version string
	}
	ok := waitUntil(t, 10*time.Second, func() bool { return getJSON(t, base+"/health", &health) })
	require.True(t, ok, "service did not answer /health in time")
	assert.Equal(t, "ok", health.Status)
	assert.Equal(t, "proxy-validator", health.Service)
	assert.NotEmpty(t, health.Version)
}

// TC-04-I-002/003/004: full pipeline — fetch → validate via fake proxy →
// write working proxy to Redis proxy_pool with a TTL ≤ 5m.
func TestPipelineWritesProxyPool(t *testing.T) {
	rdb := liveRedis(t)
	ctx := context.Background()
	t.Cleanup(func() { _ = rdb.Del(ctx, poolKey, poolKey+":staging").Err(); _ = rdb.Close() })

	// Fake working proxy: answers 204 to any relayed request.
	proxySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer proxySrv.Close()
	proxyHostPort := proxySrv.Listener.Addr().String()

	// Fake source: a proxy list returning exactly the fake proxy's host:port.
	srcSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(proxyHostPort + "\n"))
	}))
	defer srcSrv.Close()

	host, port := splitHostPort(redisAddr())
	bin := buildBinary(t)
	_, base := startService(t, bin, map[string]string{
		"PROXY_VALIDATOR_PORT":   "18187",
		"PROXY_SOURCES":          srcSrv.URL,
		"PROXY_TEST_URL":         "http://proxy-validator.invalid/check", // fake proxy short-circuits to 204
		"PROXY_POOL_KEY":         poolKey,
		"PROXY_POOL_TTL":         "5m",
		"PROXY_REFRESH_INTERVAL": "30s",
		"PROXY_POLITENESS_MIN":   "1ms",
		"PROXY_POLITENESS_MAX":   "2ms",
		"REDIS_HOST":             host,
		"REDIS_PORT":             port,
		"REDIS_PASSWORD":         os.Getenv("REDIS_TEST_PASSWORD"),
	})

	// Wait for the first cycle to complete and report a working proxy.
	var stats struct {
		Working int  `json:"working"`
		HasRun  bool `json:"has_run"`
	}
	ran := waitUntil(t, 20*time.Second, func() bool {
		return getJSON(t, base+"/stats", &stats) && stats.HasRun
	})
	require.True(t, ran, "service never completed a validation cycle")
	assert.GreaterOrEqual(t, stats.Working, 1, "the fake working proxy should validate")

	// TC-04-I-003: proxy_pool is a populated Set containing the fake proxy.
	members, err := rdb.SMembers(ctx, poolKey).Result()
	require.NoError(t, err)
	assert.Contains(t, members, proxyHostPort, "pool must contain the validated proxy")

	// TC-04-I-004: the pool carries a positive TTL no greater than 5m.
	ttl, err := rdb.TTL(ctx, poolKey).Result()
	require.NoError(t, err)
	assert.Greater(t, ttl, time.Duration(0), "pool must have a TTL")
	assert.LessOrEqual(t, ttl, 5*time.Minute, "pool TTL must not exceed 5m")
}

// splitHostPort splits "host:port" without importing net for one call.
func splitHostPort(addr string) (host, port string) {
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return addr[:i], addr[i+1:]
		}
	}
	return addr, "6379"
}
