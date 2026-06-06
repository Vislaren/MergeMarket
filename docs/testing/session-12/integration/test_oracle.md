## Oracle — A-12 Grafana Dashboard on VPS (live integration)

Source of truth: the Grafana HTTP API, the A-12 deliverable
(`infra/k3s/grafana.yml` + provisioning + `mergemarket-quality.json`), the A-12
runbook (`docs/grafana-setup.md`), and the SonarQube / GitHub Actions REST APIs.

| Input | Condition | Expected Output | Source of Truth |
|-------|-----------|-----------------|-----------------|
| `kubectl apply` + rollout | secrets + dashboard ConfigMap exist | rollout succeeds; `grafana-data` PVC Bound | `grafana.yml`; runbook |
| GET /api/health | Grafana up | 200, `database: "ok"` | Grafana API |
| GET /api/datasources | provisioned + token | SonarQube + GitHub present, type `yesoreyeram-infinity-datasource` | datasources.yml; `GF_INSTALL_PLUGINS` |
| GET /api/dashboards/uid/mergemarket-quality | dashboard ConfigMap loaded | 200; uid matches; 7 panels; folder "MergeMarket" | dashboard JSON; provider.yml |
| SonarQube coverage query via Grafana | `mergemarket` scanned | numeric coverage 0–100 | SonarQube measures API |
| GitHub Actions runs query via Grafana | valid PAT | upstream 200; `workflow_runs` array (possibly **empty**) | GitHub Actions API; runbook limitation |
| rollout restart → re-fetch | PVC-backed state | dashboard + datasources persist | PVC `/var/lib/grafana`; A-11 persistence pattern |

### Pass/PENDING policy

- A case is **PASS** only when executed against a live Grafana and the expected
  output is observed.
- All cases are **PENDING** this session: no reachable cluster/Grafana from
  local dev (the VPS is unreachable — consistent with Sessions 02/04 and every
  prior session). The suite self-skips when `GRAFANA_URL` is unset.
- **TC-12-I-006 explicitly does not require data** — an empty `workflow_runs`
  array passes, because CI runs on Jenkins and the GitHub Actions workflow was
  removed; the panel is wired correctly but unfed. This is the one place where
  "expected output" is reachability, not content.

### Re-run command

```bash
GRAFANA_URL=http://95.111.228.35:3000 GRAFANA_TOKEN=<token> \
  go test -tags=integration ./docs/testing/session-12/integration/test_suite/... -v
```
