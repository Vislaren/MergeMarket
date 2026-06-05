## Test Plan — Unit — Session 04 — A-04 Proxy-Validator Service

**Scope:** Verify the A-04 service in isolation, without external infrastructure
(no Redis, no live proxies), via two complementary techniques:
- **Subprocess toolchain checks** — the service compiles (`go build`), the
  developer's own tests pass (`go test`), and it is `go vet` / `gofmt` clean.
- **Structural source assertions** — the committed source satisfies the
  contract-level invariants the oracle requires:
  - configuration is read only from the environment (no hardcoded config);
  - `GET /health` returns the API-contract shape `{status, service, version}`;
  - the working pool is a Redis **Set** named `proxy_pool` with a 5-minute TTL
    (DATABASE_SCHEMA §3) and is replaced **atomically** (staging key + RENAME);
  - the politeness protocol is an **adaptive random** delay (back off on failure,
    relax on success) and is actually applied by the runner between dispatches;
  - a single failing proxy source does not abort the others (NFR-2);
  - the default port is 8086 (ARCHITECTURE §2);
  - every package carries a GoDoc package comment.

**Out of scope:**
- Starting Redis, writing to a real pool, or contacting live proxies — that is the
  integration suite for this session.
- Real-world proxy yield/latency — inherently flaky, an ops/tuning concern.
- A-05's consumption of `proxy_pool`, and Kong/auth (A-08/A-09).

**Approach:** Unit tests isolate the service. Toolchain checks shell out to the Go
tools against `services/proxy-validator`. Structural checks read the committed
source and assert required tokens/shapes. No mocks are needed — the developer's
in-package suite already covers logic with stubs/httptest; this suite confirms that
suite passes and that the contract-level facts hold. (Go forbids importing the
service's `internal/` packages from this separate module, so behaviour is verified
via the binary in the integration suite.)

**Entry criteria:**
- A-04 committed (`session(A-04): proxy-validator service`) and present in the tree.
- Go toolchain available; the service's `go.sum` is committed and the module cache
  is populated (so `go build`/`go test` run offline).

**Exit criteria:** All unit cases PASS. The developer suite (re-run via TC-04-U-002)
is green. `go vet` and `gofmt -l` report nothing.

**Tools:** Go `testing` + `testify`; `go build`/`go test`/`go vet`/`gofmt` as
subprocesses.

**Assumptions:**
- `gofmt` canonical form is the format-of-record (e.g. `5*time.Minute` without
  spaces) — the structural assertions match gofmt output, not hand-spacing.
- The module cache already holds the service's dependencies; in a cold CI a
  `go mod download` precedes this suite.

**Risk:**
- Structural assertions confirm the source *expresses* each invariant; they cannot
  prove runtime behaviour end-to-end — that gap is closed by the integration suite
  (PASS for `/health`; PENDING for the Redis pipeline until a Redis is available).
- String-based structural checks are sensitive to large refactors of the service
  (e.g. renaming `staging`); they are intentionally tied to the current design and
  should be updated alongside any such refactor.
