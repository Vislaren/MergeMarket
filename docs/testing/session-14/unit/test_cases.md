# Unit Test Cases — Session 14

All cases were **executed this session** and **PASS** (`go test ./unit/...`).
Type = Unit. Tools = Go testing + testify.

## Search service (A-14)

### TC-14-U-001: service compiles
| Field | Value |
|-------|-------|
| Task reference | A-14 |
| Preconditions | `services/search` present |
| Input | `go build ./...` in the service dir |
| Steps | 1. Run build as a subprocess |
| Expected result | Exit 0 |
| Actual result | [PASS] |

### TC-14-U-002: developer in-package tests pass
| Field | Value |
|-------|-------|
| Input | `go test ./...` in `services/search` |
| Expected result | Exit 0 (config, cache-key, orchestrator, Deal Meter, HTTP) |
| Actual result | [PASS] |

### TC-14-U-003: `go vet` clean — [PASS]
### TC-14-U-004: `gofmt -l` reports nothing — [PASS]

### TC-14-U-005: port is env-driven (SEARCH_PORT, default 8087), not hardcoded
| Input | source of `internal/config` + `cmd/search/main.go` |
| Expected result | uses `os.LookupEnv`, `SEARCH_PORT`/`8087`; main binds `":"+cfg.Port` |
| Actual result | [PASS] |

### TC-14-U-006: `/health` returns `{status, service, version}` — [PASS]
### TC-14-U-007: `GET /api/v1/search` route registered — [PASS]
### TC-14-U-008: missing `q`/`location` → 400 `missing_query` — [PASS]
### TC-14-U-009: cache key is `search:{sha256(q|location)}` — [PASS]
### TC-14-U-010: Deal Meter present (cheapest scores 100) — [PASS]
### TC-14-U-011: stale-while-revalidate (stale hit refreshes in background) — [PASS]
### TC-14-U-012: response carries `cached` + `latency_ms` — [PASS]
### TC-14-U-013: every internal package has a GoDoc comment — [PASS]

## User-data service (A-16..A-18)

### TC-14-U-014: service compiles — [PASS]
### TC-14-U-015: developer in-package tests pass — [PASS]
_(config, JWT verify valid/expired/wrong-secret/wrong-issuer/tampered/malformed,
and every HTTP route with fakes)_
### TC-14-U-016: `go vet` clean — [PASS]
### TC-14-U-017: `gofmt -l` clean — [PASS]

### TC-14-U-018: env-driven (USERDATA_PORT 8090) and requires JWT_SECRET
| Expected result | config rejects an empty `JWT_SECRET`; default port 8090 |
| Actual result | [PASS] |

### TC-14-U-019: all 7 API routes + `/health` registered
| Expected result | wishlist GET/POST/DELETE, alerts GET/POST/DELETE, savings GET, health GET |
| Actual result | [PASS] |

### TC-14-U-020: every API route is JWT-protected; `/health` is not — [PASS]
### TC-14-U-021: token verifier checks HS256 signature, expiry, issuer — [PASS]
### TC-14-U-022: persistence scoped by `user_id` (deletes require id AND user_id) — [PASS]
### TC-14-U-023: contract error codes emitted (already_in_wishlist, not_found, invalid_input, token_expired, unauthorized) — [PASS]
### TC-14-U-024: `purchases` table defined (EnsureSchema + canonical `01-schema.sql` + index) — [PASS]
### TC-14-U-025: every internal package has a GoDoc comment — [PASS]

---

**Summary:** 25/25 PASS.
