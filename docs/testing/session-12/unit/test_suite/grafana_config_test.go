// Package grafana_test holds the static (unit) validation of the A-12 Grafana
// artefacts. These tests run without a cluster or network: they parse the
// committed files (K3s manifest, provisioning YAML, dashboard JSON) and assert
// their structure and internal consistency.
//
// Run from the repo:
//
//	go test ./docs/testing/session-12/unit/test_suite/...
package grafana_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// findRepoRoot walks up from the test's working directory until it finds the
// repository root (identified by docker-compose.yml), matching the Session-02
// convention.
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

func readFile(t *testing.T, root, rel string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, rel))
	require.NoError(t, err, "reading %s", rel)
	return b
}

// yamlDocs decodes a multi-document YAML stream into non-nil maps.
func yamlDocs(t *testing.T, b []byte) []map[string]any {
	t.Helper()
	dec := yaml.NewDecoder(strings.NewReader(string(b)))
	var docs []map[string]any
	for {
		var doc map[string]any
		err := dec.Decode(&doc)
		if err != nil {
			break // io.EOF or end of stream
		}
		if doc != nil {
			docs = append(docs, doc)
		}
	}
	return docs
}

// asMap / asSlice coerce decoded YAML/JSON values, tolerating both decoders.
func asMap(v any) map[string]any {
	m, _ := v.(map[string]any)
	return m
}

func asSlice(v any) []any {
	s, _ := v.([]any)
	return s
}

// asInt coerces a YAML int or JSON float64 to int.
func asInt(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	}
	return 0, false
}

const (
	manifestPath    = "infra/k3s/grafana.yml"
	datasourcesPath = "infra/grafana/provisioning/datasources/datasources.yml"
	providerPath    = "infra/grafana/provisioning/dashboards/provider.yml"
	dashboardPath   = "infra/grafana/dashboards/mergemarket-quality.json"
	envPath         = ".env.example"
)

// blockedPorts are never allowed (PORTS_README blocked table).
var blockedPorts = map[int]bool{80: true, 443: true, 8080: true, 8000: true, 8081: true, 9090: true}

// secretLiteralMarkers must never appear in committed config.
var secretLiteralMarkers = []string{"squ_", "ghp_", "github_pat_"}

func TestA12GrafanaArtefacts(t *testing.T) {
	root := findRepoRoot(t)

	manifest := yamlDocs(t, readFile(t, root, manifestPath))
	require.NotEmpty(t, manifest, "manifest decoded to zero documents")

	// byKind indexes the first document of each Kind (Deployment/Service/etc).
	byKind := map[string]map[string]any{}
	var allDocs []map[string]any
	for _, d := range manifest {
		allDocs = append(allDocs, d)
		if k, _ := d["kind"].(string); k != "" {
			if _, seen := byKind[k]; !seen {
				byKind[k] = d
			}
		}
	}

	dep := byKind["Deployment"]
	require.NotNil(t, dep, "no Deployment in manifest")
	containers := asSlice(asMap(asMap(asMap(dep["spec"])["template"])["spec"])["containers"])
	require.NotEmpty(t, containers, "Deployment has no containers")
	grafanaC := asMap(containers[0])

	t.Run("TC-12-U-001: manifest parses as a 6-document set", func(t *testing.T) {
		require.Len(t, manifest, 6, "expected 6 YAML documents")
		var kinds []string
		for _, d := range manifest {
			kinds = append(kinds, d["kind"].(string))
		}
		assert.Equal(t, []string{"Namespace", "PersistentVolumeClaim", "ConfigMap", "ConfigMap", "Deployment", "Service"}, kinds)
	})

	t.Run("TC-12-U-002: namespace reuses mergemarket-observability", func(t *testing.T) {
		ns := byKind["Namespace"]
		require.NotNil(t, ns)
		assert.Equal(t, "mergemarket-observability", asMap(ns["metadata"])["name"])
		for _, d := range allDocs {
			if d["kind"] == "Namespace" {
				continue
			}
			meta := asMap(d["metadata"])
			assert.Equalf(t, "mergemarket-observability", meta["namespace"],
				"%s/%v not in the shared namespace", d["kind"], meta["name"])
		}
	})

	t.Run("TC-12-U-003: deployment installs the Infinity plugin", func(t *testing.T) {
		var got string
		for _, e := range asSlice(grafanaC["env"]) {
			em := asMap(e)
			if em["name"] == "GF_INSTALL_PLUGINS" {
				got, _ = em["value"].(string)
			}
		}
		assert.Equal(t, "yesoreyeram-infinity-datasource", got)
	})

	t.Run("TC-12-U-004: secrets via grafana-secrets, none inlined", func(t *testing.T) {
		needSecret := map[string]bool{
			"GF_SECURITY_ADMIN_USER":     false,
			"GF_SECURITY_ADMIN_PASSWORD": false,
			"SONAR_TOKEN":                false,
			"GITHUB_TOKEN":               false,
		}
		for _, e := range asSlice(grafanaC["env"]) {
			em := asMap(e)
			name, _ := em["name"].(string)
			if _, want := needSecret[name]; want {
				ref := asMap(asMap(em["valueFrom"])["secretKeyRef"])
				assert.Equalf(t, "grafana-secrets", ref["name"], "%s not from grafana-secrets", name)
				needSecret[name] = true
			}
		}
		for name, ok := range needSecret {
			assert.Truef(t, ok, "%s missing from deployment env", name)
		}
		raw := string(readFile(t, root, manifestPath))
		for _, marker := range secretLiteralMarkers {
			assert.NotContainsf(t, raw, marker, "manifest contains a committed token literal %q", marker)
		}
	})

	t.Run("TC-12-U-005: NodePort service maps 3000 -> 30300", func(t *testing.T) {
		svc := byKind["Service"]
		require.NotNil(t, svc)
		assert.Equal(t, "NodePort", asMap(svc["spec"])["type"])
		ports := asSlice(asMap(svc["spec"])["ports"])
		require.NotEmpty(t, ports)
		p := asMap(ports[0])
		port, _ := asInt(p["port"])
		nodePort, _ := asInt(p["nodePort"])
		assert.Equal(t, 3000, port)
		assert.Equal(t, 30300, nodePort)
	})

	t.Run("TC-12-U-006: grafana port matches the PORTS_README contract", func(t *testing.T) {
		ports := asSlice(grafanaC["ports"])
		require.NotEmpty(t, ports)
		cp, _ := asInt(asMap(ports[0])["containerPort"])
		assert.Equal(t, 3000, cp)
		env := string(readFile(t, root, envPath))
		assert.Contains(t, env, "GRAFANA_PORT=3000")
	})

	t.Run("TC-12-U-007: no blocked port is used", func(t *testing.T) {
		for _, d := range allDocs {
			for _, key := range []string{"port", "nodePort", "targetPort", "containerPort"} {
				walkInts(d, key, func(n int) {
					assert.Falsef(t, blockedPorts[n], "blocked port %d used in %s", n, d["kind"])
				})
			}
		}
	})

	t.Run("TC-12-U-008: PVC declared and mounted at /var/lib/grafana", func(t *testing.T) {
		pvc := byKind["PersistentVolumeClaim"]
		require.NotNil(t, pvc)
		assert.Equal(t, "grafana-data", asMap(pvc["metadata"])["name"])
		assert.Contains(t, asSlice(asMap(pvc["spec"])["accessModes"]), "ReadWriteOnce")
		var mounted bool
		for _, m := range asSlice(grafanaC["volumeMounts"]) {
			if asMap(m)["mountPath"] == "/var/lib/grafana" {
				mounted = true
			}
		}
		assert.True(t, mounted, "grafana data not mounted at /var/lib/grafana")
	})

	t.Run("TC-12-U-009: probes hit /api/health", func(t *testing.T) {
		for _, probe := range []string{"readinessProbe", "livenessProbe"} {
			path := asMap(asMap(asMap(grafanaC[probe])["httpGet"]))["path"]
			assert.Equalf(t, "/api/health", path, "%s wrong path", probe)
		}
	})

	// ── Datasource provisioning ────────────────────────────────────────────
	dsDoc := asMap(yamlDocs(t, readFile(t, root, datasourcesPath))[0])
	datasources := asSlice(dsDoc["datasources"])
	dsByName := map[string]map[string]any{}
	for _, d := range datasources {
		dm := asMap(d)
		dsByName[dm["name"].(string)] = dm
	}

	t.Run("TC-12-U-010: datasource provisioning defines SonarQube + GitHub", func(t *testing.T) {
		assert.Equal(t, 1, asInt1(dsDoc["apiVersion"]))
		assert.Contains(t, dsByName, "SonarQube")
		assert.Contains(t, dsByName, "GitHub")
	})

	t.Run("TC-12-U-011: both datasources are the Infinity type", func(t *testing.T) {
		for _, name := range []string{"SonarQube", "GitHub"} {
			assert.Equalf(t, "yesoreyeram-infinity-datasource", dsByName[name]["type"], "%s type", name)
		}
	})

	t.Run("TC-12-U-012: datasource UIDs are sonarqube_api / github_api", func(t *testing.T) {
		assert.Equal(t, "sonarqube_api", dsByName["SonarQube"]["uid"])
		assert.Equal(t, "github_api", dsByName["GitHub"]["uid"])
	})

	t.Run("TC-12-U-013: datasource secrets come from env, not literals", func(t *testing.T) {
		raw := string(readFile(t, root, datasourcesPath))
		assert.Contains(t, raw, "${SONAR_TOKEN}")
		assert.Contains(t, raw, "${GITHUB_TOKEN}")
		for _, marker := range secretLiteralMarkers {
			assert.NotContainsf(t, raw, marker, "datasources.yml contains a token literal %q", marker)
		}
	})

	t.Run("TC-12-U-014: dashboard provider points at the dashboards path", func(t *testing.T) {
		provDoc := asMap(yamlDocs(t, readFile(t, root, providerPath))[0])
		assert.Equal(t, 1, asInt1(provDoc["apiVersion"]))
		providers := asSlice(provDoc["providers"])
		require.NotEmpty(t, providers)
		assert.Equal(t, "/var/lib/grafana/dashboards", asMap(asMap(providers[0])["options"])["path"])
	})

	// ── Dashboard JSON ─────────────────────────────────────────────────────
	var dash map[string]any
	require.NoError(t, json.Unmarshal(readFile(t, root, dashboardPath), &dash), "dashboard JSON must parse")
	panels := asSlice(dash["panels"])
	declaredUIDs := map[string]bool{"sonarqube_api": true, "github_api": true}

	t.Run("TC-12-U-015: dashboard JSON parses with the expected identity", func(t *testing.T) {
		assert.Equal(t, "mergemarket-quality", dash["uid"])
		assert.NotEmpty(t, dash["title"])
		sv, _ := asInt(dash["schemaVersion"])
		assert.GreaterOrEqual(t, sv, 36)
	})

	t.Run("TC-12-U-016: 7 panels, each bound to a known datasource UID", func(t *testing.T) {
		require.Len(t, panels, 7)
		for _, p := range panels {
			pm := asMap(p)
			uid, _ := asMap(pm["datasource"])["uid"].(string)
			assert.Truef(t, declaredUIDs[uid], "panel %q has unknown datasource uid %q", pm["title"], uid)
			for _, tg := range asSlice(pm["targets"]) {
				tuid, _ := asMap(asMap(tg)["datasource"])["uid"].(string)
				assert.Truef(t, declaredUIDs[tuid], "panel %q target has unknown uid %q", pm["title"], tuid)
			}
		}
	})

	t.Run("TC-12-U-017: SonarQube panels query the correct API + metric keys", func(t *testing.T) {
		var sawComponent, sawHistory bool
		for _, url := range targetURLsForDS(panels, "sonarqube_api") {
			if strings.Contains(url, "/api/measures/component") {
				sawComponent = true
				assert.Truef(t,
					strings.Contains(url, "coverage") || strings.Contains(url, "bugs") || strings.Contains(url, "vulnerabilities"),
					"measures/component target missing a metric key: %s", url)
			}
			if strings.Contains(url, "/api/measures/search_history") {
				sawHistory = true
			}
		}
		assert.True(t, sawComponent, "no /api/measures/component target")
		assert.True(t, sawHistory, "no /api/measures/search_history target")
	})

	t.Run("TC-12-U-018: GitHub panels query Actions runs; dashboard vars present", func(t *testing.T) {
		var sawRuns bool
		for _, url := range targetURLsForDS(panels, "github_api") {
			if strings.Contains(url, "/actions/runs") {
				sawRuns = true
				assert.Contains(t, url, "/repos/", "Actions runs URL should target /repos/{repo}")
			}
		}
		assert.True(t, sawRuns, "no GitHub Actions runs target")

		vars := map[string]string{}
		for _, v := range asSlice(asMap(dash["templating"])["list"]) {
			vm := asMap(v)
			name, _ := vm["name"].(string)
			vars[name] = asMap(vm["current"])["value"].(string)
		}
		assert.Equal(t, "mergemarket", vars["sonar_project"])
		assert.Equal(t, "Vislaren/MergeMarket", vars["gh_repo"])
	})
}

// targetURLsForDS returns every target URL among panels bound to the given
// datasource uid (panel-level or target-level binding).
func targetURLsForDS(panels []any, uid string) []string {
	var urls []string
	for _, p := range panels {
		pm := asMap(p)
		panelUID, _ := asMap(pm["datasource"])["uid"].(string)
		for _, tg := range asSlice(pm["targets"]) {
			tm := asMap(tg)
			tUID, _ := asMap(tm["datasource"])["uid"].(string)
			effective := tUID
			if effective == "" {
				effective = panelUID
			}
			if effective == uid {
				if u, ok := tm["url"].(string); ok {
					urls = append(urls, u)
				}
			}
		}
	}
	return urls
}

// asInt1 coerces apiVersion (may decode as int or float64) to int.
func asInt1(v any) int {
	n, _ := asInt(v)
	return n
}

// walkInts recurses through a decoded structure invoking fn for every int value
// found under the given key name.
func walkInts(v any, key string, fn func(int)) {
	switch t := v.(type) {
	case map[string]any:
		for k, val := range t {
			if k == key {
				if n, ok := asInt(val); ok {
					fn(n)
				}
			}
			walkInts(val, key, fn)
		}
	case []any:
		for _, item := range t {
			walkInts(item, key, fn)
		}
	}
}
