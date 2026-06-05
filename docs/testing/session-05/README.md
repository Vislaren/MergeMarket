# Session 05 — Test Artefacts

**Task under test:** B-02 — Mock Server for All API Contracts (Agent B, Client Developer role)
**Tester:** Agent B (Quality Engineer role)
**Date:** 2026-06-05

---

## What was tested

B-02 delivered `services/mock-server/` — a dependency-free (stdlib-only) Go HTTP
server that returns hardcoded fixtures matching every endpoint in
`project_docs/api/API_CONTRACTS.md`. It is what all Flutter (B-03+) and BFF (B-09)
development runs against until Agent A's real services are ready. Default port
`8080`, env-overridable via `MOCK_SERVER_PORT` / `PORT`.

Because the server's logic lives in `internal/` packages — which Go forbids a
**separate** test module from importing — this QE suite uses the same approach as
Session-02 (A-02) and Session-04 (A-04):

- **Unit** — a subprocess toolchain gate (`go build`/`test`/`vet`/`gofmt`) plus
  structural source assertions of the contract-level invariants: env-driven config
  with default port 8080, `/health` shape, **every** API_CONTRACTS route
  registered, `total_cost = price + shipping`, the `{error, message}` shape,
  `log/slog` logging, dependency-free `go.mod`, and GoDoc on every package.
- **Integration** — builds and runs the **real binary** on an OS-assigned free
  port, waits for `/health`, then exercises every contract over HTTP: success
  shapes, the `total_cost` invariant, all error sentinels (400/401/404/409/504),
  CRUD status codes, and CORS preflight.

Unlike Sessions 02/04, **the whole suite runs offline** — the mock server needs no
Redis, Docker, or network — so every integration case **executed and passed** this
session (no PENDING cases).

Out of scope: the Flutter app's consumption of the mock (B-03+), the BFF (B-09),
and the swap to real services (B-11). Fixtures are static sample data, not a
stateful store — sentinel inputs drive the error paths deterministically.

---

## Layout

```
docs/testing/session-05/
├── README.md                (this file)
├── go.mod                   (stdlib-only module for both Go suites)
├── unit/
│   ├── test_plan.md
│   ├── test_cases.md
│   ├── test_oracle.md
│   └── test_suite/mock_server_test.go
└── integration/
    ├── test_plan.md
    ├── test_cases.md
    ├── test_oracle.md
    └── test_suite/mock_server_integration_test.go   (//go:build integration)
```

Run unit: `go test ./docs/testing/session-05/unit/test_suite/...`
Run integration: `go test -tags=integration ./docs/testing/session-05/integration/test_suite/...`

The service's own in-package developer tests also pass:
`cd services/mock-server && go test ./... -cover` → config 100%, server 88.9%.

---

## Results summary

| Suite | Cases | Status |
|-------|-------|--------|
| Unit | TC-05-U-001 … 012 | **12/12 PASS** (executed this session) |
| Integration | TC-05-I-001 … 008 | **8/8 PASS** (executed — real binary over HTTP, offline) |

**QE findings:** No defects in B-02. The service behaves to every contract in
API_CONTRACTS.md, error sentinels return the documented codes, and the
`total_cost = price + shipping` invariant holds across all sample data. One
environmental note surfaced during the live smoke test: **port 8080 collides with
a local Jenkins** (Jetty) instance — the env-driven config (`MOCK_SERVER_PORT`)
already handles this, and it is documented in the service README.

**Quality gate:** PENDING PIPELINE RUN — SonarQube at `http://95.111.228.35:9000`
was not reachable / no scan has run against this branch yet (consistent with prior
sessions; the CI pipeline is A-10's deliverable and the VPS SonarQube is A-11's).
No failing gate to fix; the service is build/vet/gofmt clean and all 20 cases pass.
Re-verify once CI runs against the pushed branch.
