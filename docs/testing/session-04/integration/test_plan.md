## Test Plan — Integration — Session 04 — A-04 Proxy-Validator Service

**Scope:** Verify the **compiled A-04 binary** behaves correctly end-to-end:
- It serves `GET /health` with the API-contract shape while running.
- Given a (fake) proxy-list source and a (fake) working proxy, it runs a full
  scrape → validate → write cycle and persists the working proxy to a **live
  Redis** `proxy_pool` Set with a TTL ≤ 5m, and reports the cycle via `/stats`.

The "fake working proxy" is a local httptest server that answers `204` to any
relayed request, so the validator accepts it; the "fake source" is a local
httptest server returning that proxy's `host:port`. This isolates A-04's pipeline
from the open internet while exercising the real validate-and-write path.

**Out of scope:**
- Real public-proxy reachability and yield (flaky; ops concern).
- TLS-proxy (`CONNECT`) tunnelling — the fake proxy short-circuits plain HTTP.
- A-05's consumption of the pool; Kong/auth (A-08/A-09).

**Approach:** Integration tests use **no mocks of A-04 itself** — they build the
real binary, launch it as a subprocess with env pointing at the local fakes and a
live Redis, then make real HTTP calls (`/health`, `/stats`) and real Redis reads
(`SMEMBERS`, `TTL`). The fakes stand in only for the *external* world (proxy list +
proxy endpoint).

**Entry criteria:**
- A-04 committed and buildable.
- Go toolchain available (to compile the binary).
- A live Redis reachable at `REDIS_TEST_ADDR` (default `localhost:6379`) for the
  pipeline cases.

**Exit criteria:** `/health` case passes; with Redis available, the pipeline writes
the expected member to `proxy_pool` with a valid TTL.

**Tools:** Go `testing` + `testify`; `redis/go-redis/v9`; `net/http/httptest`;
`go build` + subprocess launch. Suite is tagged `//go:build integration`.

**Assumptions:**
- For plain-HTTP targets, Go's `http.Transport{Proxy: …}` sends the absolute-URI
  request to the proxy, so the fake proxy can answer `204` directly and be deemed
  valid — no real upstream is contacted.
- A dedicated test key `proxy_pool:it-session04` is used and cleaned up, so the
  test never touches a real pool.

**Risk:**
- The pipeline cases (TC-04-I-002…004) require Redis; where none is available they
  **skip** (reported PENDING), leaving the live write path unverified until CI
  (A-10) or a local `docker compose up` provides Redis.
- Subprocess port binding uses fixed high ports (18186/18187); a conflicting
  listener on the test host would cause a flake.
