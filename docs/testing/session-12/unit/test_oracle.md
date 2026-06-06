## Oracle — A-12 Grafana Dashboard on VPS (static validation)

The source of truth for each expected output: the project port contract
(`.agents/Agent_*/PORTS_README.md`), the system architecture
(`project_docs/architecture/ARCHITECTURE.md §9` monitoring), the A-12 task spec
(`.agents/Agent_A/TODO.md` → A-12) and its DONE entry, and the Grafana / Infinity
provisioning schemas.

| Input | Condition | Expected Output | Source of Truth |
|-------|-----------|-----------------|-----------------|
| `infra/k3s/grafana.yml` | YAML decode | 6 docs: Namespace, PVC, ConfigMap×2, Deployment, Service | A-12 DONE entry |
| manifest resources | namespace field | all in `mergemarket-observability` | A-11 manifest (shared ns) |
| Deployment env | `GF_INSTALL_PLUGINS` | `yesoreyeram-infinity-datasource` | A-12 decision (no native datasource) |
| Deployment env | admin/sonar/github secrets | all via `secretKeyRef → grafana-secrets`; no literal token | A-12 decision; NFR-4 |
| Service | type/ports | `NodePort`, 3000 → nodePort 30300 | PORTS_README (Grafana 3000) |
| `.env.example` | Grafana block | `GRAFANA_PORT=3000` present | PORTS_README |
| all ports | blocked-port check | none ∈ {80,443,8080,8000,8081,9090} | PORTS_README blocked table |
| Deployment volumes | grafana data | `grafana-data` PVC mounted at `/var/lib/grafana` | A-12 deliverable |
| Deployment probes | readiness+liveness | both `httpGet /api/health` | Grafana health endpoint; A-11 probe convention |
| datasources.yml | parse + count | 2 datasources: SonarQube, GitHub | A-12 task ("two data sources") |
| datasources.yml | each `type` | `yesoreyeram-infinity-datasource` | A-12 decision |
| datasources.yml | each `uid` | `sonarqube_api`, `github_api` | A-12 deliverable |
| datasources.yml | secret fields | `${SONAR_TOKEN}` / `${GITHUB_TOKEN}`; no literal | NFR-4; A-12 decision |
| provider.yml | options.path | `/var/lib/grafana/dashboards` | matches Deployment mount |
| dashboard JSON | parse + identity | `uid=mergemarket-quality`, title set, schemaVersion ≥ 36 | Grafana dashboard schema |
| dashboard JSON | panels | exactly 7; each datasource uid ∈ declared uids | A-12 task (5 metrics → 7 panels) |
| SonarQube panels | target URL | `/api/measures/component` (coverage,bugs,vulnerabilities) + `/api/measures/search_history` | SonarQube measures API |
| GitHub panels | target URL | `/repos/${gh_repo}/actions/runs` | GitHub Actions API |
| dashboard JSON | templating | vars `sonar_project`(=mergemarket), `gh_repo`(=Vislaren/MergeMarket) | A-12 deliverable |

### Known truth-gaps (covered by the integration suite, not unit)

| Input | Condition | Expected Output | Source of Truth |
|-------|-----------|-----------------|-----------------|
| GET /api/health on live Grafana | after `kubectl apply` | `{"database":"ok",...}` 200 | integration TC-12-I-002 |
| GET /api/datasources | provisioned | SonarQube + GitHub present, reachable | integration TC-12-I-003 |
| Coverage panel query | SonarQube has scanned `mergemarket` | numeric coverage 0–100 | integration TC-12-I-004 |
| Pipeline panels | GitHub Actions workflow has run | non-empty conclusions/durations | integration TC-12-I-005 (expected EMPTY today — CI is on Jenkins, Actions workflow deleted; documented in `docs/grafana-setup.md`) |
