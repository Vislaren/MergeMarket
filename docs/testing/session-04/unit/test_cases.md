# Test Cases — Unit — Session 04 — A-04

All cases below were executed this session via
`go test ./docs/testing/session-04/unit/test_suite/...` against the service at
commit `session(A-04)`. **Result: 12/12 PASS.**

---

### TC-04-U-001: Service compiles
| Field | Value |
|-------|-------|
| Task reference | A-04 |
| Type | Unit |
| Preconditions | Go toolchain; service `go.sum` committed; module cache populated |
| Input | `services/proxy-validator` |
| Steps | 1. `go build ./...` in the service dir |
| Expected result | Exit 0, no compile errors |
| Actual result | [PASS] |
| Notes | First Go service in the repo; sets the build pattern for A-05..A-08. |

---

### TC-04-U-002: Developer unit tests pass
| Field | Value |
|-------|-------|
| Task reference | A-04 |
| Type | Unit |
| Preconditions | Service compiles |
| Input | The service's in-package `_test.go` suites |
| Steps | 1. `go test ./...` in the service dir |
| Expected result | All packages `ok`; exit 0 |
| Actual result | [PASS] |
| Notes | QE re-runs Agent A's own suite as evidence. Per-package coverage 78–100%; store's live-Redis tests self-skip without `REDIS_TEST_ADDR`. |

---

### TC-04-U-003: `go vet` is clean
| Field | Value |
|-------|-------|
| Task reference | A-04 |
| Type | Unit |
| Preconditions | Service compiles |
| Input | `services/proxy-validator` |
| Steps | 1. `go vet ./...` |
| Expected result | No diagnostics; exit 0 |
| Actual result | [PASS] |
| Notes | — |

---

### TC-04-U-004: `gofmt` is clean
| Field | Value |
|-------|-------|
| Task reference | A-04 |
| Type | Unit |
| Preconditions | — |
| Input | `services/proxy-validator` |
| Steps | 1. `gofmt -l .` 2. Assert output empty |
| Expected result | No files listed |
| Actual result | [PASS] |
| Notes | Format-of-record for SonarQube/CI. |

---

### TC-04-U-005: Configuration is environment-driven
| Field | Value |
|-------|-------|
| Task reference | A-04 |
| Type | Unit |
| Preconditions | — |
| Input | `internal/config/config.go`, `cmd/proxy-validator/main.go` |
| Steps | 1. Assert `os.LookupEnv` used 2. Assert `PROXY_VALIDATOR_PORT`, `REDIS_HOST`, `REDIS_PORT`, `PROXY_POOL_TTL` honoured 3. Assert main binds `":"+cfg.Port` |
| Expected result | All config flows from env; no hardcoded listen literal |
| Actual result | [PASS] |
| Notes | INSTRUCTIONS §3 "all config from env, never hardcoded". |

---

### TC-04-U-006: `/health` matches the API contract shape
| Field | Value |
|-------|-------|
| Task reference | A-04 |
| Type | Unit |
| Preconditions | — |
| Input | `internal/server/server.go` |
| Steps | 1. Assert `/health` route registered 2. Assert `Status`/`Service`/`Version` fields 3. Assert service name `proxy-validator` |
| Expected result | Handler returns `{status, service, version}` |
| Actual result | [PASS] |
| Notes | API_CONTRACTS.md "Health (all services)". |

---

### TC-04-U-007: Writes `proxy_pool` Set with a 5m TTL
| Field | Value |
|-------|-------|
| Task reference | A-04 |
| Type | Unit |
| Preconditions | — |
| Input | `internal/config/config.go`, `internal/store/store.go` |
| Steps | 1. Assert default key `proxy_pool` 2. Assert default TTL `5*time.Minute` 3. Assert store uses `SAdd` + `Expire` |
| Expected result | Pool is a Redis Set with a 5-minute expiry |
| Actual result | [PASS] |
| Notes | DATABASE_SCHEMA §3 (`proxy_pool`, Set, 5 min). **QE finding:** initial assertion used `5 * time.Minute` (spaced); corrected to gofmt form `5*time.Minute` — the source was correct, the test was. |

---

### TC-04-U-008: Pool is replaced atomically
| Field | Value |
|-------|-------|
| Task reference | A-04 |
| Type | Unit |
| Preconditions | — |
| Input | `internal/store/store.go` |
| Steps | 1. Assert `Rename` used 2. Assert a `staging` temp key |
| Expected result | New set staged then RENAMEd over the live key |
| Actual result | [PASS] |
| Notes | Guarantees readers never see a half-written/empty pool. |

---

### TC-04-U-009: Politeness is an adaptive random delay
| Field | Value |
|-------|-------|
| Task reference | A-04 |
| Type | Unit |
| Preconditions | — |
| Input | `internal/politeness/politeness.go`, `internal/runner/runner.go` |
| Steps | 1. Assert randomised delay (`rand`) 2. Assert `Failure()`/`Success()` adapt 3. Assert runner calls `limiter.Wait` and reports `Failure`/`Success` |
| Expected result | Adaptive random delay, applied between dispatches |
| Actual result | [PASS] |
| Notes | A-04 task: "politeness protocol (adaptive random delays)". |

---

### TC-04-U-010: Resilient to partial source failure
| Field | Value |
|-------|-------|
| Task reference | A-04 |
| Type | Unit |
| Preconditions | — |
| Input | `internal/fetcher/fetcher.go` |
| Steps | 1. Assert it only errors when `failures == len(f.sources)` 2. Assert a failing source is `continue`d |
| Expected result | One bad source does not abort the others |
| Actual result | [PASS] |
| Notes | NFR-2 (single failure must not halt others). |

---

### TC-04-U-011: Default port is 8086
| Field | Value |
|-------|-------|
| Task reference | A-04 |
| Type | Unit |
| Preconditions | — |
| Input | `internal/config/config.go` |
| Steps | 1. Assert `getEnv("PROXY_VALIDATOR_PORT", "8086")` |
| Expected result | Default port 8086 |
| Actual result | [PASS] |
| Notes | ARCHITECTURE §2 + `.env.example`. |

---

### TC-04-U-012: Every package carries a GoDoc comment
| Field | Value |
|-------|-------|
| Task reference | A-04 |
| Type | Unit |
| Preconditions | — |
| Input | all `internal/*` package files + `cmd/proxy-validator/main.go` |
| Steps | 1. Assert each package file has `// Package ...` 2. Assert main has `// Command proxy-validator` |
| Expected result | Documented packages throughout |
| Actual result | [PASS] |
| Notes | INSTRUCTIONS §3 "every exported function has a GoDoc comment"; checked at package granularity here. |
