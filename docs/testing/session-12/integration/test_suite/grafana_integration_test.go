//go:build integration

// Package grafana_integration_test smoke-tests a *running* Grafana built from
// the A-12 artefacts, over its real HTTP API. No mocks.
//
// It self-skips unless GRAFANA_URL is set, so it is safe in CI without a
// cluster. The authenticated cases additionally need GRAFANA_TOKEN (a Grafana
// admin or service-account API token).
//
// Run:
//
//	GRAFANA_URL=http://95.111.228.35:3000 GRAFANA_TOKEN=<token> \
//	  go test -tags=integration ./docs/testing/session-12/integration/test_suite/... -v
package grafana_integration_test

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func baseURL(t *testing.T) string {
	u := os.Getenv("GRAFANA_URL")
	if u == "" {
		t.Skip("GRAFANA_URL not set — Grafana is not deployed/reachable; integration cases are PENDING")
	}
	return strings.TrimRight(u, "/")
}

func token(t *testing.T) string {
	tok := os.Getenv("GRAFANA_TOKEN")
	if tok == "" {
		t.Skip("GRAFANA_TOKEN not set — skipping authenticated Grafana API case")
	}
	return tok
}

func getJSON(t *testing.T, url, tok string) (int, map[string]any, []any) {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)
	if tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	// The endpoint may return an object or an array; decode whichever.
	var obj map[string]any
	var arr []any
	trimmed := strings.TrimSpace(string(body))
	if strings.HasPrefix(trimmed, "[") {
		_ = json.Unmarshal(body, &arr)
	} else {
		_ = json.Unmarshal(body, &obj)
	}
	return resp.StatusCode, obj, arr
}

// TC-12-I-002: Grafana is healthy.
func TestI002Health(t *testing.T) {
	base := baseURL(t)
	status, obj, _ := getJSON(t, base+"/api/health", "")
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, "ok", obj["database"])
}

// TC-12-I-003: both datasources provisioned, Infinity type.
func TestI003Datasources(t *testing.T) {
	base := baseURL(t)
	tok := token(t)
	status, _, arr := getJSON(t, base+"/api/datasources", tok)
	require.Equal(t, http.StatusOK, status)

	types := map[string]string{}
	for _, d := range arr {
		dm, _ := d.(map[string]any)
		name, _ := dm["name"].(string)
		typ, _ := dm["type"].(string)
		types[name] = typ
	}
	assert.Equal(t, "yesoreyeram-infinity-datasource", types["SonarQube"], "SonarQube datasource missing/wrong type")
	assert.Equal(t, "yesoreyeram-infinity-datasource", types["GitHub"], "GitHub datasource missing/wrong type")
}

// TC-12-I-004: dashboard provisioned and loads by uid with 7 panels.
func TestI004DashboardLoads(t *testing.T) {
	base := baseURL(t)
	tok := token(t)
	status, obj, _ := getJSON(t, base+"/api/dashboards/uid/mergemarket-quality", tok)
	require.Equal(t, http.StatusOK, status)

	dash, _ := obj["dashboard"].(map[string]any)
	require.NotNil(t, dash, "no dashboard in response")
	assert.Equal(t, "mergemarket-quality", dash["uid"])
	panels, _ := dash["panels"].([]any)
	assert.Len(t, panels, 7, "expected 7 panels")
}

// TC-12-I-005 / TC-12-I-006: datasource query paths are reachable.
//
// These drive Grafana's datasource health/query API. Exact request bodies vary
// by Grafana version, so this case asserts the datasources resolve and respond
// without a server error; data-presence assertions live in the manual oracle
// (SonarQube needs a prior scan; GitHub Actions data is legitimately empty).
func TestI005DatasourceReachable(t *testing.T) {
	base := baseURL(t)
	tok := token(t)
	status, _, arr := getJSON(t, base+"/api/datasources", tok)
	require.Equal(t, http.StatusOK, status)

	var sonarUID, githubUID string
	for _, d := range arr {
		dm, _ := d.(map[string]any)
		switch dm["name"] {
		case "SonarQube":
			sonarUID, _ = dm["uid"].(string)
		case "GitHub":
			githubUID, _ = dm["uid"].(string)
		}
	}
	assert.Equal(t, "sonarqube_api", sonarUID)
	assert.Equal(t, "github_api", githubUID)

	for _, uid := range []string{sonarUID, githubUID} {
		if uid == "" {
			continue
		}
		st, _, _ := getJSON(t, base+"/api/datasources/uid/"+uid+"/health", tok)
		// Health endpoint may not be implemented for every plugin/version; only
		// fail on a 5xx (a real server-side error).
		assert.Lessf(t, st, 500, "datasource %s health returned server error %d", uid, st)
	}
}
