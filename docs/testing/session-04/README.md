# Session 04 — Test Artefacts

**Task under test:** A-04 — Proxy-Validator Service (Agent A, infrastructure/backend role)
**Tester:** Agent B (Quality Engineer role)
**Date:** 2026-06-05

---

## What was tested

A-04 delivered the first real Go service in the repo (`services/proxy-validator/`):
it scrapes public proxy lists, validates each proxy against a real endpoint with
an adaptive politeness delay and bounded concurrency, and writes the working
`ip:port` set to Redis `proxy_pool` (atomic swap, 5-minute TTL). It serves
`GET /health` and `GET /stats` on port 8086.

Because A-04's behaviour lives in `internal/` packages — which Go will not let a
**separate** test module import — this suite verifies A-04 the same way Session-02
verified A-02's infra: **static + subprocess** for unit, **live binary** for
integration. Both are independent QE checks designed from the oracle
(API_CONTRACTS, DATABASE_SCHEMA §3, ARCHITECTURE §2/§8, the A-04 task spec), not a
copy of the developer's in-package tests.

- **Unit** — runs `go build`/`go test`/`go vet`/`gofmt` against the service, then
  asserts the contract-level facts structurally: env-only config, `/health` shape,
  `proxy_pool` Set + 5m TTL, atomic RENAME swap, adaptive politeness, resilience to
  partial source failure, default port 8086, GoDoc on every package.
- **Integration** — builds and runs the **real binary**, drives a full
  fetch→validate→write cycle through local httptest fakes (a fake proxy list + a
  fake working proxy), and reads `proxy_pool` back from a live Redis to assert Set
  membership and TTL.

Out of scope: real public-proxy reachability/yield (inherently flaky, ops concern),
the scraper's consumption of the pool (A-05), and Kong/auth routing (A-08/A-09).

---

## Layout

```
docs/testing/session-04/
├── README.md                (this file)
├── go.mod / go.sum          (module for both Go suites)
├── unit/
│   ├── test_plan.md
│   ├── test_cases.md
│   ├── test_oracle.md
│   └── test_suite/proxy_validator_test.go
└── integration/
    ├── test_plan.md
    ├── test_cases.md
    ├── test_oracle.md
    └── test_suite/proxy_validator_integration_test.go   (//go:build integration)
```

Run unit: `go test ./docs/testing/session-04/unit/test_suite/...`
Run integration: `REDIS_TEST_ADDR=localhost:6379 go test -tags=integration ./docs/testing/session-04/integration/test_suite/...`

---

## Results summary

| Suite | Cases | Status |
|-------|-------|--------|
| Unit | TC-04-U-001 … 012 | **12/12 PASS** (executed this session) |
| Integration | TC-04-I-001 | **PASS** (executed — real binary, `/health`) |
| Integration | TC-04-I-002 … 004 | **PENDING** — require a live Redis (none running this session) |

The service's own developer tests (Agent A's in-package suite) were re-run as part
of TC-04-U-002 and pass (per-package coverage 78–100%).

**Quality gate:** PENDING PIPELINE RUN — SonarQube at `http://95.111.228.35:9000`
is unreachable from the local dev environment this session, so the last push's gate
could not be read. No failing gate to fix; the service is build/vet/gofmt clean and
all executed cases pass. Re-verify once CI (A-10) runs against the pushed branch.
