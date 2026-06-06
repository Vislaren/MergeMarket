# Test Cases — Unit — Session 12 — A-12 Grafana Dashboard on VPS

_(U = Unit, I = Integration)_
All cases reference Agent A task **A-12**. Type: Unit (static validation).
Executed this session via `go test ./docs/testing/session-12/unit/test_suite/...`.

---

### TC-12-U-001: K3s manifest parses as a multi-document set
| Field | Value |
|-------|-------|
| Preconditions | `infra/k3s/grafana.yml` committed |
| Input | The manifest file |
| Steps | 1. `yaml.Decoder` over the file 2. Count non-nil documents 3. Collect `kind`s |
| Expected result | 6 documents: Namespace, PersistentVolumeClaim, ConfigMap, ConfigMap, Deployment, Service |
| Actual result | [PASS] |
| Notes | Mirrors the A-12 DONE verification |

### TC-12-U-002: Namespace reuses `mergemarket-observability`
| Field | Value |
|-------|-------|
| Preconditions | manifest parsed |
| Input | Namespace doc + every resource's `metadata.namespace` |
| Steps | 1. Find the Namespace 2. Assert name 3. Assert every namespaced resource uses it |
| Expected result | All resources in `mergemarket-observability` (shared with A-11 SonarQube) |
| Actual result | [PASS] |
| Notes | Avoids a second observability namespace |

### TC-12-U-003: Deployment installs the Infinity plugin
| Field | Value |
|-------|-------|
| Preconditions | manifest parsed |
| Input | Deployment container env |
| Steps | 1. Locate the grafana container 2. Read `GF_INSTALL_PLUGINS` |
| Expected result | `GF_INSTALL_PLUGINS=yesoreyeram-infinity-datasource` |
| Actual result | [PASS] |
| Notes | Neither SonarQube nor GitHub has a native datasource |

### TC-12-U-004: Secrets sourced from `grafana-secrets`, never inlined
| Field | Value |
|-------|-------|
| Preconditions | manifest parsed |
| Input | Deployment env entries |
| Steps | 1. Assert admin user/password + sonar-token + github-token use `secretKeyRef.name: grafana-secrets` 2. Assert no plaintext token literal (`squ_`, `ghp_`, `github_pat_`) anywhere in the manifest |
| Expected result | All four secrets via `grafana-secrets`; zero committed credentials |
| Actual result | [PASS] |
| Notes | Security check — secrets are created on the VPS per the runbook |

### TC-12-U-005: NodePort service maps 3000 → 30300
| Field | Value |
|-------|-------|
| Preconditions | manifest parsed |
| Input | Service doc |
| Steps | 1. Assert `type: NodePort` 2. Assert `port: 3000`, `nodePort: 30300` |
| Expected result | Grafana reachable on container 3000 / node 30300 |
| Actual result | [PASS] |
| Notes | 3000 matches PORTS_README; 30300 is in the K8s NodePort range |

### TC-12-U-006: Grafana port matches the PORTS_README contract
| Field | Value |
|-------|-------|
| Preconditions | manifest parsed |
| Input | container port + `.env.example` |
| Steps | 1. Assert containerPort 3000 2. Assert `.env.example` `GRAFANA_PORT=3000` |
| Expected result | 3000 everywhere (Grafana row of the port table) |
| Actual result | [PASS] |
| Notes | — |

### TC-12-U-007: No blocked port is used
| Field | Value |
|-------|-------|
| Preconditions | manifest parsed |
| Input | All `port`/`containerPort`/`nodePort` numbers |
| Steps | 1. Collect ports 2. Assert none is a blocked port (80, 443, 8080, 8000, 8081, 9090) |
| Expected result | No blocked port used |
| Actual result | [PASS] |
| Notes | Per PORTS_README blocked-ports table |

### TC-12-U-008: PVC declared for Grafana data
| Field | Value |
|-------|-------|
| Preconditions | manifest parsed |
| Input | PVC doc + Deployment volumes |
| Steps | 1. Assert a `grafana-data` PVC (RWO) 2. Assert it is mounted at `/var/lib/grafana` |
| Expected result | Persistent storage for Grafana DB/state |
| Actual result | [PASS] |
| Notes | — |

### TC-12-U-009: Liveness/readiness probes hit `/api/health`
| Field | Value |
|-------|-------|
| Preconditions | manifest parsed |
| Input | Deployment probes |
| Steps | 1. Assert readiness + liveness `httpGet.path == /api/health` |
| Expected result | Both probes target the Grafana health endpoint |
| Actual result | [PASS] |
| Notes | Same health-probe convention as the A-11 SonarQube manifest |

### TC-12-U-010: Datasource provisioning parses; SonarQube + GitHub defined
| Field | Value |
|-------|-------|
| Preconditions | `infra/grafana/provisioning/datasources/datasources.yml` committed |
| Input | The datasources file |
| Steps | 1. Parse YAML 2. Assert `apiVersion: 1` 3. Assert two datasources named SonarQube + GitHub |
| Expected result | Both datasources present |
| Actual result | [PASS] |
| Notes | — |

### TC-12-U-011: Both datasources are the Infinity type
| Field | Value |
|-------|-------|
| Preconditions | datasources parsed |
| Input | each datasource `type` |
| Steps | 1. Assert both `type == yesoreyeram-infinity-datasource` |
| Expected result | Matches the installed plugin |
| Actual result | [PASS] |
| Notes | — |

### TC-12-U-012: Datasource UIDs are `sonarqube_api` / `github_api`
| Field | Value |
|-------|-------|
| Preconditions | datasources parsed |
| Input | each datasource `uid` |
| Steps | 1. Assert SonarQube `uid == sonarqube_api` 2. GitHub `uid == github_api` |
| Expected result | Stable UIDs the dashboard panels can bind to |
| Actual result | [PASS] |
| Notes | Cross-checked against panels in TC-12-U-016 |

### TC-12-U-013: Datasource secrets come from env, not literals
| Field | Value |
|-------|-------|
| Preconditions | datasources parsed |
| Input | datasources file text |
| Steps | 1. Assert `${SONAR_TOKEN}` and `${GITHUB_TOKEN}` placeholders used 2. Assert no `squ_`/`ghp_`/`github_pat_` literal |
| Expected result | Tokens injected at runtime; none committed |
| Actual result | [PASS] |
| Notes | Security check |

### TC-12-U-014: Dashboard provider points at the dashboards path
| Field | Value |
|-------|-------|
| Preconditions | `infra/grafana/provisioning/dashboards/provider.yml` committed |
| Input | provider file |
| Steps | 1. Parse YAML 2. Assert `apiVersion: 1` 3. Assert provider `options.path == /var/lib/grafana/dashboards` |
| Expected result | File provider loads dashboards from the mounted path |
| Actual result | [PASS] |
| Notes | Matches the Deployment dashboards volume mount |

### TC-12-U-015: Dashboard JSON parses with the expected identity
| Field | Value |
|-------|-------|
| Preconditions | `infra/grafana/dashboards/mergemarket-quality.json` committed |
| Input | the dashboard file |
| Steps | 1. `json.Unmarshal` 2. Assert `uid == mergemarket-quality` 3. Assert non-empty title + `schemaVersion >= 36` |
| Expected result | Valid Grafana dashboard document |
| Actual result | [PASS] |
| Notes | — |

### TC-12-U-016: All 7 panels present, each bound to a known datasource UID
| Field | Value |
|-------|-------|
| Preconditions | dashboard parsed |
| Input | `panels[]` |
| Steps | 1. Assert exactly 7 panels 2. Assert each panel datasource UID ∈ {`sonarqube_api`,`github_api`} 3. Assert each target's datasource UID matches a declared datasource |
| Expected result | Coverage %, Open Bugs, Vulnerabilities, Pass/Fail rate, Coverage-over-time, Bugs/Vulns-over-time, Build-duration |
| Actual result | [PASS] |
| Notes | Closes the datasource↔panel wiring loop with TC-12-U-012 |

### TC-12-U-017: SonarQube panels query the correct API + metric keys
| Field | Value |
|-------|-------|
| Preconditions | dashboard parsed |
| Input | targets of the SonarQube panels |
| Steps | 1. Assert measures panels call `/api/measures/component` with metricKeys `coverage`/`bugs`/`vulnerabilities` 2. Assert trend panels call `/api/measures/search_history` |
| Expected result | Coverage/bugs/vulnerabilities sourced from the documented SonarQube endpoints |
| Actual result | [PASS] |
| Notes | Metric keys match SonarQube's measures API |

### TC-12-U-018: GitHub panels query the Actions runs API; dashboard vars present
| Field | Value |
|-------|-------|
| Preconditions | dashboard parsed |
| Input | GitHub panel targets + `templating.list` |
| Steps | 1. Assert GitHub panels call `/repos/${gh_repo}/actions/runs` 2. Assert template vars `sonar_project` (default `mergemarket`) and `gh_repo` (default `Vislaren/MergeMarket`) exist |
| Expected result | Pipeline panels source from GitHub Actions; both retarget vars present |
| Actual result | [PASS] |
| Notes | Pipeline panels are empty until an Actions workflow runs (CI on Jenkins) — see oracle |
