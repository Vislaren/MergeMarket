# Test Cases — Integration — Session 12 — A-12 Grafana Dashboard on VPS

_(U = Unit, I = Integration)_
All cases reference Agent A task **A-12**. Type: Integration (live Grafana).
**All PENDING this session** — no reachable K3s cluster / Grafana from local dev
(VPS unreachable). The Go suite (`-tags=integration`) is written and self-skips
unless `GRAFANA_URL` is set.

---

### TC-12-I-001: Manifest applies and the deployment rolls out
| Field | Value |
|-------|-------|
| Task reference | A-12 |
| Type | Integration |
| Preconditions | `grafana-secrets` + `grafana-dashboards` ConfigMap exist; kubectl context set |
| Input | `infra/k3s/grafana.yml` |
| Steps | 1. `kubectl apply -f infra/k3s/grafana.yml` 2. `kubectl -n mergemarket-observability rollout status deploy/grafana` |
| Expected result | Apply succeeds; rollout reports "successfully rolled out"; PVC `grafana-data` Bound |
| Actual result | [PENDING] |
| Notes | Performed by the operator; gates the HTTP cases below |

### TC-12-I-002: Grafana is healthy
| Field | Value |
|-------|-------|
| Task reference | A-12 |
| Type | Integration |
| Preconditions | TC-12-I-001 done; `GRAFANA_URL` reachable |
| Input | `GET {GRAFANA_URL}/api/health` |
| Steps | 1. HTTP GET 2. Parse JSON |
| Expected result | 200; body `database == "ok"` |
| Actual result | [PENDING] |
| Notes | No auth required for `/api/health` |

### TC-12-I-003: Both datasources are provisioned
| Field | Value |
|-------|-------|
| Task reference | A-12 |
| Type | Integration |
| Preconditions | Grafana healthy; `GRAFANA_TOKEN` set (admin/service-account) |
| Input | `GET {GRAFANA_URL}/api/datasources` (Bearer token) |
| Steps | 1. HTTP GET 2. Collect datasource names + types |
| Expected result | 200; contains SonarQube + GitHub, both type `yesoreyeram-infinity-datasource`; Infinity plugin installed |
| Actual result | [PENDING] |
| Notes | Confirms provisioning + `GF_INSTALL_PLUGINS` took effect |

### TC-12-I-004: Dashboard provisioned and loads by uid
| Field | Value |
|-------|-------|
| Task reference | A-12 |
| Type | Integration |
| Preconditions | Grafana healthy; token set |
| Input | `GET {GRAFANA_URL}/api/dashboards/uid/mergemarket-quality` |
| Steps | 1. HTTP GET 2. Parse dashboard 3. Count panels |
| Expected result | 200; `dashboard.uid == mergemarket-quality`; 7 panels; folder "MergeMarket" |
| Actual result | [PENDING] |
| Notes | Confirms the file provider loaded the ConfigMap dashboard |

### TC-12-I-005: SonarQube coverage panel returns live data
| Field | Value |
|-------|-------|
| Task reference | A-12 |
| Type | Integration |
| Preconditions | SonarQube scanned the `mergemarket` project; SonarQube datasource reachable from Grafana |
| Input | Query the SonarQube datasource (`/api/measures/component?...coverage`) via Grafana's datasource proxy/query API |
| Steps | 1. Issue the query through Grafana 2. Read the returned measure |
| Expected result | A numeric coverage value 0–100 |
| Actual result | [PENDING] |
| Notes | Requires a prior SonarQube scan (A-10 pipeline). Validates the end-to-end SonarQube → Infinity → panel path |

### TC-12-I-006: GitHub pipeline panel query path is reachable
| Field | Value |
|-------|-------|
| Task reference | A-12 |
| Type | Integration |
| Preconditions | `GITHUB_TOKEN` valid; GitHub datasource reachable |
| Input | Query the GitHub datasource (`/repos/{repo}/actions/runs`) via Grafana |
| Steps | 1. Issue the query 2. Inspect the HTTP status of the upstream call |
| Expected result | Upstream returns 200 (an **empty** `workflow_runs` array is acceptable — CI is on Jenkins, no Actions workflow yet) |
| Actual result | [PENDING] |
| Notes | Asserts reachability/auth, NOT data presence — documented limitation in `docs/grafana-setup.md` |

### TC-12-I-007: Data survives a pod restart (persistence)
| Field | Value |
|-------|-------|
| Task reference | A-12 |
| Type | Integration |
| Preconditions | Grafana healthy with provisioned content |
| Input | `kubectl rollout restart deploy/grafana` |
| Steps | 1. Restart 2. Wait for rollout 3. Re-fetch `/api/dashboards/uid/mergemarket-quality` and `/api/datasources` |
| Expected result | Dashboard + datasources still present after restart (PVC-backed `/var/lib/grafana`) |
| Actual result | [PENDING] |
| Notes | Mirrors the A-11 SonarQube persistence check |
