# Integration Test Cases — Session 14

Type = Integration. Tools = Go testing + testify (`-tags=integration`).
Executed result: **3/8 PASS (executed), 5/8 PENDING** (gated on a live stack).

## Search service (A-14)

### TC-14-I-001: fails fast when Postgres is unreachable
| Field | Value |
|-------|-------|
| Task reference | A-14 |
| Type | Integration |
| Preconditions | search binary builds |
| Input | run binary with `DB_HOST=127.0.0.1 DB_PORT=1` (refused) |
| Steps | 1. build binary 2. run with bad DB 3. observe exit |
| Expected result | process exits non-zero, logs `search-service exited with error` |
| Actual result | [PASS] |
| Notes | Postgres is a hard dependency (ARCHITECTURE §2) |

### TC-14-I-002: live search returns the contract shape
| Input | `GET /api/v1/search?q=phone&location=CM` against a running binary + DB |
| Expected result | 200 with `results`, `cached`, `latency_ms` keys |
| Actual result | [PENDING] — set `DB_TEST_DSN` |

### TC-14-I-003: missing query parameter → 400 missing_query
| Input | `GET /api/v1/search?location=CM` (no `q`) |
| Expected result | 400 `{error:"missing_query"}` |
| Actual result | [PENDING] — set `DB_TEST_DSN` (binary needs DB to boot) |

### TC-14-I-004: identical query served from cache on the second call
| Input | same query twice |
| Expected result | 1st `cached:false`, 2nd `cached:true` |
| Actual result | [PENDING] — set `DB_TEST_DSN` + `REDIS_TEST_ADDR` |

## User-data service (A-16..A-18)

### TC-14-I-005: refuses to start without JWT_SECRET
| Input | run binary with `JWT_SECRET` unset |
| Expected result | exits non-zero, output mentions `JWT_SECRET` |
| Actual result | [PASS] |
| Notes | guards the shared-secret contract with auth (A-08) |

### TC-14-I-006: fails fast when Postgres is unreachable
| Input | run with `JWT_SECRET=test-secret`, `DB_HOST=127.0.0.1 DB_PORT=1` |
| Expected result | exits non-zero, logs `userdata-service exited with error` |
| Actual result | [PASS] |

### TC-14-I-007: protected route returns 401 without a bearer token
| Input | `GET /api/v1/wishlist` with no `Authorization` header |
| Expected result | 401 `{error:"unauthorized"}` |
| Actual result | [PENDING] — set `DB_TEST_DSN` (binary needs DB to boot) |

### TC-14-I-008: full live E2E (Kong→BFF→search/userdata; search→wishlist→alert)
| Input | the session-13 TC-13-I E2E flow, end to end |
| Expected result | search → add to wishlist → set alert → (drop) notification |
| Actual result | [PENDING] — needs the whole stack (Kong+BFF+auth+search+userdata+Postgres+Redis) |
| Notes | This is the gap session-13 flagged; the services now exist, so it is unblocked **in principle** — it just needs a running stack to execute. |

---

**Summary:** Executed 3/8 PASS. PENDING 5/8 (need a live backend / full stack).
