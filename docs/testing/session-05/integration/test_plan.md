## Test Plan — Integration — Session 05 — B-02 Mock Server for All API Contracts

**Scope:** End-to-end verification of the **running** mock-server binary over
HTTP. The suite builds the service, starts it on an OS-assigned free port, waits
for `/health`, then drives every API_CONTRACTS endpoint and asserts: success
response shapes, the `total_cost = price + shipping` invariant, every error
sentinel (400 missing_query, 504 timeout, 401 invalid_credentials / token_expired,
404 not_found, 409 already_in_wishlist), CRUD status codes (201/204), and CORS
preflight (204 + `Access-Control-Allow-Origin`).

**Out of scope:** Flutter/BFF consumers (B-03+/B-09); persistence/statefulness
(the mock is intentionally stateless); the swap to real backends (B-11).

**Approach:** Integration tests make real HTTP calls against the compiled binary —
no mocks, no in-process handler. Because the mock server has zero external
dependencies (no Redis/DB/network), the entire suite runs offline and
deterministically; unlike Sessions 02/04 there are **no PENDING cases**.

**Entry criteria:** Go 1.22+ toolchain on PATH; a free TCP port available
(obtained programmatically via `net.Listen` on `:0`).

**Exit criteria:** All 8 integration cases pass against the live binary.

**Tools:** Go `testing` + `net/http` + `os/exec` (stdlib). Build tag
`//go:build integration`.

**Assumptions:** Sample fixture data is stable; error paths are reached via the
documented sentinel inputs (e.g. `q=timeout`, `product_id=unknown`,
`password=wrongpassword`).

**Risk:** Port binding can race on a busy host; mitigated by requesting an
OS-assigned ephemeral port immediately before launch and polling `/health` with a
10s readiness deadline.
