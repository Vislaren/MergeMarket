# Session 14 — QE Test Artefacts

End-of-session testing protocol (Agent B, Quality Engineer role) for **Agent A's
session-8 work**:
- **A-14** — Search read service (`services/search`)
- **A-16..A-18** — Consolidated user-data service (`services/userdata`:
  wishlist + alerts + savings)

These two services close most of the live-E2E gap that
`docs/testing/session-13/CONTRACT_AUDIT.md` flagged: search, wishlist, alerts,
and savings now have real backends (only truth-score, A-15, remains mock-only).

## Layout
```
session-14/
├── go.mod                       # QE test module (testify)
├── unit/
│   ├── test_plan.md  test_cases.md  test_oracle.md
│   └── test_suite/{helpers,search,userdata}_test.go
└── integration/
    ├── test_plan.md  test_cases.md  test_oracle.md
    └── test_suite/{helpers,search,userdata}_integration_test.go
```

## Results

| Suite | Executed | Result |
|-------|----------|--------|
| Unit | 25/25 | **PASS** |
| Integration | 3/8 | **PASS** (fail-fast / config-guard) |
| Integration | 5/8 | **PENDING** (need a live backend / full stack) |

Both services are `go build` / `go vet` / `gofmt` clean and their own in-package
suites pass (run by TC-14-U-002 / -015).

## Run it
```bash
# Unit (fully offline, executes now):
go test ./docs/testing/session-14/unit/...

# Integration — executes the no-DB fail-fast cases now; gated cases run when a
# backend is provided:
DB_TEST_DSN=postgres://postgres:pass@localhost:5432/mergemarket?sslmode=disable \
REDIS_TEST_ADDR=localhost:6379 \
go test -tags=integration ./docs/testing/session-14/integration/test_suite/...
```

## QE findings

- **No defects found** in either service. Contract shapes, error codes, env-driven
  config, JWT verification, and per-user data scoping all hold.
- **Observation (non-blocking, carried from Agent A's DONE notes):** neither new
  service exposes `/metrics`, though `PORTS_README` mandates it for every Go
  service. No A-04..A-08 service does either — this is a real cross-service gap,
  not specific to A-14/A-16-18. Recommend a dedicated follow-up to add
  `prometheus/client_golang` (or the BFF's hand-rolled counter pattern) uniformly.
- **PENDING integration** is environmental only (no Docker; host Postgres is of
  unknown ownership so not used; full stack not running) — the suite is written
  and ready; TC-14-I-008 is the session-13 E2E, now unblocked in principle.

## Quality gate (§5)

SonarQube at `http://95.111.228.35:9000` is **UP**, but project-status APIs
require a token (no `SONAR_TOKEN` available locally), so the gate could not be
read this session. No failing gate to fix; both services are build/vet/gofmt
clean. Re-verify once CI (A-10/Jenkins) runs against the branch.

## Branch note

Committed on **`co-courage`** (beside Agent A's service code) — consistent with
the session-04 and session-12 A-service testing sessions, because the suite must
sit beside the services to build and run them.
