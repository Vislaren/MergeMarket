# Test Artefacts — Session 12 (Agent A task A-12)

End-of-session testing protocol output for Agent A's **A-12 — Grafana Dashboard
on VPS** (see `.agents/Agent_A/DONE.md`).

A-12 delivered infrastructure artefacts (a K3s manifest, Grafana provisioning
YAML, a dashboard JSON, env vars, and a runbook) — **not application code** — so,
exactly like Session-02 (A-02), the verification is **static validation (unit)**
plus **live cluster smoke (integration)**, not application unit tests.

```
session-12/
├── unit/            ← static validation of each A-12 artefact (no cluster)
│   ├── test_plan.md
│   ├── test_cases.md
│   ├── test_oracle.md
│   └── test_suite/grafana_config_test.go
├── integration/     ← live Grafana smoke tests (kubectl apply + HTTP)
│   ├── test_plan.md
│   ├── test_cases.md
│   ├── test_oracle.md
│   └── test_suite/grafana_integration_test.go
└── go.mod           ← module for both test suites
```

## Execution summary (this session)

| Suite | How run | Result |
|-------|---------|--------|
| Unit (static) | `go test ./docs/testing/session-12/unit/test_suite/...` | **18/18 PASS** |
| Integration (live) | `go test -tags=integration ...` against a deployed Grafana | **PENDING** — no reachable K3s cluster / Grafana from local dev (the VPS is unreachable, same as every prior session); rerun once Grafana is deployed (or in CI) |

The Go suites under `test_suite/` are the committed, CI-runnable form of the
checks. The unit suite parses the real committed A-12 files (YAML + JSON) and
asserts their structure; it was executed this session and all cases passed. The
integration suite is written, compiles under `-tags=integration`, and self-skips
unless `GRAFANA_URL` (and, for the datasource panels, `GRAFANA_TOKEN`) is set.

## How to run

```bash
# Unit (static) — offline, no cluster:
go test ./docs/testing/session-12/unit/test_suite/...

# Integration — against a live Grafana:
GRAFANA_URL=http://95.111.228.35:3000 GRAFANA_TOKEN=<api-token> \
  go test -tags=integration ./docs/testing/session-12/integration/test_suite/...
```
