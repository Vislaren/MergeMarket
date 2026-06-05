# Unit Test Cases — Session 05 — B-02 Mock Server

All cases were executed this session. Status legend: [PASS] / [FAIL] / [PENDING].

---

### TC-05-U-001: Service compiles (`go build ./...`)
| Field | Value |
|-------|-------|
| Task reference | B-02 |
| Type | Unit |
| Preconditions | Go 1.22+ on PATH; `services/mock-server/` present |
| Input | `go build ./...` run in the service dir |
| Steps | 1. Exec `go build ./...` 2. Assert exit 0 |
| Expected result | Build succeeds, no output |
| Actual result | [PASS] |
| Notes | — |

---

### TC-05-U-002: Developer in-package tests pass (`go test ./...`)
| Field | Value |
|-------|-------|
| Task reference | B-02 |
| Type | Unit |
| Preconditions | As above |
| Input | `go test ./...` |
| Steps | 1. Exec `go test ./...` 2. Assert exit 0 |
| Expected result | All package tests pass (config 100%, server 88.9% cov) |
| Actual result | [PASS] |
| Notes | Re-runs the service's own black-box handler tests |

---

### TC-05-U-003: `go vet` clean
| Field | Value |
|-------|-------|
| Task reference | B-02 |
| Type | Unit |
| Input | `go vet ./...` |
| Steps | 1. Exec `go vet ./...` 2. Assert exit 0 |
| Expected result | No vet diagnostics |
| Actual result | [PASS] |
| Notes | — |

---

### TC-05-U-004: `gofmt` clean
| Field | Value |
|-------|-------|
| Task reference | B-02 |
| Type | Unit |
| Input | `gofmt -l .` |
| Steps | 1. Exec `gofmt -l .` 2. Assert empty output |
| Expected result | No unformatted files |
| Actual result | [PASS] |
| Notes | — |

---

### TC-05-U-005: Dependency-free (`go.mod` has no `require` block)
| Field | Value |
|-------|-------|
| Task reference | B-02 |
| Type | Unit |
| Input | `services/mock-server/go.mod` |
| Steps | 1. Read go.mod 2. Assert it contains no `require` |
| Expected result | stdlib-only; always builds/runs offline |
| Actual result | [PASS] |
| Notes | Deliberate design choice for an offline dev mock |

---

### TC-05-U-006: Config is env-driven with default port 8080
| Field | Value |
|-------|-------|
| Task reference | B-02 |
| Type | Unit |
| Input | `internal/config/config.go` |
| Steps | 1. Read source 2. Assert `MOCK_SERVER_PORT` + `os.Getenv` present 3. Assert `defaultPort = 8080` |
| Expected result | No hardcoded listen addr; default 8080 per task spec |
| Actual result | [PASS] |
| Notes | `PORT` is honoured as a fallback |

---

### TC-05-U-007: `/health` returns `{status, service, version}`
| Field | Value |
|-------|-------|
| Task reference | B-02 |
| Type | Unit |
| Input | `internal/server/server.go`, `internal/fixtures/fixtures.go` |
| Steps | 1. Assert `GET /health` route registered 2. Assert HealthResponse has status/service/version JSON tags |
| Expected result | Matches API_CONTRACTS Health shape |
| Actual result | [PASS] |
| Notes | — |

---

### TC-05-U-008: Every API_CONTRACTS endpoint is routed
| Field | Value |
|-------|-------|
| Task reference | B-02 |
| Type | Unit |
| Input | `internal/server/server.go` |
| Steps | 1. Read source 2. Assert each of the 13 `/api/v1/*` method+path patterns is registered |
| Expected result | auth (3), search, products history/truth-score, wishlist (3), alerts (3), savings — all present |
| Actual result | [PASS] |
| Notes | Health asserted separately in U-007 |

---

### TC-05-U-009: Search result keeps `total_cost = price + shipping`
| Field | Value |
|-------|-------|
| Task reference | B-02 |
| Type | Unit |
| Input | `internal/fixtures/fixtures.go` |
| Steps | 1. Assert the offer builder sets `TotalCost: price + shipping` |
| Expected result | Invariant guaranteed in fixture construction |
| Actual result | [PASS] |
| Notes | Also verified at runtime in TC-05-I-002 |

---

### TC-05-U-010: Error shape is `{error, message}`
| Field | Value |
|-------|-------|
| Task reference | B-02 |
| Type | Unit |
| Input | `internal/fixtures/fixtures.go` |
| Steps | 1. Assert ErrorResponse has `error` + `message` JSON tags |
| Expected result | Matches API_CONTRACTS canonical error shape |
| Actual result | [PASS] |
| Notes | — |

---

### TC-05-U-011: Structured logging via `log/slog`
| Field | Value |
|-------|-------|
| Task reference | B-02 |
| Type | Unit |
| Input | `cmd/mock-server/main.go` |
| Steps | 1. Assert `log/slog` imported |
| Expected result | Standards §3 structured logging satisfied |
| Actual result | [PASS] |
| Notes | Middleware logs method/path/status/duration |

---

### TC-05-U-012: Every package has a GoDoc comment
| Field | Value |
|-------|-------|
| Task reference | B-02 |
| Type | Unit |
| Input | main.go, config.go, fixtures.go, server.go |
| Steps | 1. For each file assert a `// Package ` or `// Command ` comment exists |
| Expected result | All four documented |
| Actual result | [PASS] |
| Notes | — |
