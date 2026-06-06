## Test Plan — Unit — Session 12 — A-12 Grafana Dashboard on VPS

**Scope:** Static validation of the A-12 deliverables as committed to the repo:
- `infra/k3s/grafana.yml` — K3s manifest (namespace, PVC, provisioning
  ConfigMaps, Deployment, NodePort Service).
- `infra/grafana/provisioning/datasources/datasources.yml` — SonarQube + GitHub
  Infinity datasources.
- `infra/grafana/provisioning/dashboards/provider.yml` — file-based provider.
- `infra/grafana/dashboards/mergemarket-quality.json` — the 7-panel dashboard.
- `.env.example` — Grafana env block.

These checks confirm the artefacts are well-formed and internally consistent:
YAML/JSON parse, the manifest is a valid multi-document set, datasource UIDs
match the UIDs the dashboard panels target, panels hit the correct SonarQube /
GitHub endpoints with the correct metric keys, the deployment installs the
Infinity plugin and wires secrets via `grafana-secrets`, ports match the
`PORTS_README` contract (Grafana 3000, NodePort 30300), and no blocked port is
used.

**Out of scope:** Anything requiring a running Grafana, K3s cluster, SonarQube,
or GitHub API — i.e. that provisioning actually loads, that datasources connect,
that panels return data, and that the dashboard renders. All of that is the
**integration** suite and is PENDING this session (no reachable cluster).
Visual/UX correctness of the dashboard is also out of scope.

**Approach:** Unit tests isolate the artefacts and assert their structure by
parsing the real committed files (no mocks of the files themselves) — YAML via
`gopkg.in/yaml.v3`, JSON via stdlib `encoding/json` — and making structural
assertions with `testify`. No cluster, no network.

**Entry criteria:** A-12 artefacts committed (done — commit
`session(A-12): grafana quality & pipeline dashboard`). Go toolchain available
with `testify` + `yaml.v3` resolvable.

**Exit criteria:** All unit cases pass; every A-12 artefact parses and the
datasource↔panel↔endpoint wiring is internally consistent.

**Tools:** Go `testing` + `testify/assert` + `testify/require` + `yaml.v3`.

**Assumptions:** The dashboard's SonarQube project key defaults to `mergemarket`
and the GitHub repo to `Vislaren/MergeMarket` (template variables) — asserted as
the documented defaults, not as live-correct values.

**Risk:** Static checks cannot catch runtime issues — a structurally valid
Infinity datasource can still fail to authenticate, and a valid panel query can
still return no data (notably the GitHub Actions panels, which are empty until
an Actions workflow runs — CI currently runs on Jenkins). These are covered by
the integration suite and called out in the A-12 DONE entry.
