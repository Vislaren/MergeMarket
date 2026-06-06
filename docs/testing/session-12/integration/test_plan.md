## Test Plan — Integration — Session 12 — A-12 Grafana Dashboard on VPS

**Scope:** End-to-end smoke of a **deployed** Grafana built from the A-12
artefacts:
- Grafana comes up after `kubectl apply -f infra/k3s/grafana.yml` and is
  reachable (`/api/health`).
- The Infinity plugin is installed.
- Both datasources (SonarQube, GitHub) are provisioned.
- The `mergemarket-quality` dashboard is provisioned and loads.
- The SonarQube panels return live measures for the scanned project.
- The GitHub pipeline panels query the Actions API (data presence is
  environment-dependent — see Risk).

**Out of scope:** Static structure of the files (that is the **unit** suite,
all PASS). Visual rendering / panel layout. Provisioning of the underlying
SonarQube (A-11) and the GitHub PAT lifecycle. Long-term dashboard data trends.

**Approach:** Integration tests make **real HTTP calls** to a running Grafana
using its REST API (`/api/health`, `/api/datasources`,
`/api/dashboards/uid/{uid}`, and the datasource query/proxy endpoints). No
mocks. The suite is Go, gated by the `integration` build tag, and self-skips
unless `GRAFANA_URL` is set (and `GRAFANA_TOKEN` for the authenticated calls).

**Entry criteria:**
- `grafana-secrets` created on the cluster (admin creds + SonarQube + GitHub
  tokens) — see `docs/grafana-setup.md`.
- `grafana-dashboards` ConfigMap created from `infra/grafana/dashboards/`.
- `kubectl apply -f infra/k3s/grafana.yml` succeeded and the deployment rolled
  out.
- `GRAFANA_URL` reachable from the test runner.

**Exit criteria:** Grafana healthy; both datasources present; dashboard loads by
uid; SonarQube coverage query returns a numeric value. (GitHub panels assert the
query path is reachable; non-empty data is not required — see Risk.)

**Tools:** Go `testing` + `testify` + stdlib `net/http` (build tag
`integration`).

**Assumptions:** SonarQube has already scanned a project whose key matches the
dashboard's `sonar_project` variable (`mergemarket`). A Grafana service account
or admin API token is available as `GRAFANA_TOKEN`.

**Risk:**
- **No reachable cluster from local dev this session** — the VPS is unreachable
  (same as every prior session), so all integration cases are **PENDING**. The
  suite is written and compiles; rerun once Grafana is deployed or in CI.
- **GitHub pipeline panels may legitimately return empty.** CI was migrated to
  Jenkins and the GitHub Actions workflow directory was deleted, so the Actions
  runs API returns no runs until a workflow exists again. TC-12-I-005 therefore
  asserts only that the query path is reachable, not that data is present
  (documented in `docs/grafana-setup.md`).
