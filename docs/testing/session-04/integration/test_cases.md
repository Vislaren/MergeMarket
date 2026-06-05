# Test Cases — Integration — Session 04 — A-04

TC-04-I-001 was **executed and PASSED** this session (real binary, real HTTP).
TC-04-I-002…004 are **PENDING** — they require a live Redis, which was not running
this session; they self-skip until one is available (CI / `docker compose up`).

---

### TC-04-I-001: Running binary serves the `/health` contract
| Field | Value |
|-------|-------|
| Task reference | A-04 |
| Type | Integration |
| Preconditions | Go toolchain; service builds |
| Input | Binary started with `PROXY_VALIDATOR_PORT=18186` and an unreachable source/Redis |
| Steps | 1. `go build` the command 2. launch it 3. poll `GET /health` |
| Expected result | `200`, body `{status:"ok", service:"proxy-validator", version:non-empty}`; health serves even while Redis and the source are down |
| Expected result | (cont.) |
| Actual result | [PASS] |
| Notes | Confirms the health server is independent of the validation loop and degrades gracefully (logs warnings, keeps serving). |

---

### TC-04-I-002: Full pipeline validates a working proxy and reports it
| Field | Value |
|-------|-------|
| Task reference | A-04 |
| Type | Integration |
| Preconditions | Live Redis; Go toolchain |
| Input | `PROXY_SOURCES`=local httptest list returning the fake proxy `host:port`; `PROXY_TEST_URL`=dummy; fake proxy answers `204` |
| Steps | 1. start fake proxy + fake source 2. launch binary pointed at them + Redis 3. poll `GET /stats` until `has_run` |
| Expected result | `stats.has_run == true` and `stats.working >= 1` |
| Actual result | [PENDING] — no live Redis this session |
| Notes | `RunOnce` only records stats after a successful Redis write, so this also proves the write path ran. |

---

### TC-04-I-003: Working proxy is written to `proxy_pool` Set
| Field | Value |
|-------|-------|
| Task reference | A-04 |
| Type | Integration |
| Preconditions | TC-04-I-002 ran |
| Input | Redis key `proxy_pool:it-session04` |
| Steps | 1. `SMEMBERS proxy_pool:it-session04` |
| Expected result | Set contains the fake proxy's `host:port` |
| Actual result | [PENDING] — no live Redis this session |
| Notes | DATABASE_SCHEMA §3 — pool is a Redis Set of `ip:port`. |

---

### TC-04-I-004: Pool carries a TTL no greater than 5m
| Field | Value |
|-------|-------|
| Task reference | A-04 |
| Type | Integration |
| Preconditions | TC-04-I-003 ran |
| Input | Redis key `proxy_pool:it-session04` |
| Steps | 1. `TTL proxy_pool:it-session04` |
| Expected result | TTL `> 0` and `<= 5m` |
| Actual result | [PENDING] — no live Redis this session |
| Notes | DATABASE_SCHEMA §3 — proxy pool TTL is 5 minutes. |

---

## How to run the PENDING cases

```bash
# bring up Redis (or use the A-02 compose stack), then:
REDIS_TEST_ADDR=localhost:6379 \
  go test -tags=integration -v \
  ./docs/testing/session-04/integration/test_suite/...
```
