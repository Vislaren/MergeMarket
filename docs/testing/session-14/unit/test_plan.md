## Test Plan — Unit — Session 14 — A-14 (Search) + A-16..A-18 (User-data)

**Scope:** Agent A's session-8 services:
- **A-14 search** (`services/search`) — `GET /api/v1/search` over normalized
  Postgres products with a stale-while-revalidate Redis cache + Deal Meter.
- **A-16..A-18 userdata** (`services/userdata`) — consolidated JWT-protected
  per-user CRUD for wishlist, alerts, and savings.

Verified at the unit level: toolchain health (build/test/vet/gofmt), the
contract invariants from `API_CONTRACTS.md`, the schema invariant from
`DATABASE_SCHEMA.md` (the new `purchases` table), and the MergeMarket Go
standards (env-driven config, `/health` shape, GoDoc, JWT verification,
per-user data scoping).

**Out of scope:** Live HTTP behaviour against a running Postgres/Redis (that is
the integration plan); the Flutter client; the truth-score service (A-15, not
built); the cross-service Kong/BFF wiring.

**Approach:** Go forbids importing a service's `internal/` packages from another
module, so this independent QE suite verifies three ways, all **executable**:
1. **Toolchain subprocess** — `go build`, `go test` (the services' own in-package
   suites), `go vet`, `gofmt -l` run against each service.
2. **Structural source assertions** — the suite reads the service source and
   asserts the contract/standard invariants (routes, error codes, env keys,
   cache key, Deal Meter, JWT checks, user_id scoping, DDL).
3. The services' own `_test.go` suites (run via step 1) cover behaviour with
   table tests + fakes.

**Entry criteria:** Agent A's A-14 + A-16..A-18 code present on the checkout;
Go 1.22+ and `gofmt` on PATH.

**Exit criteria:** All unit cases PASS; both services build/vet/gofmt clean and
their in-package suites pass.

**Tools:** Go `testing` + `testify`.

**Assumptions:** The services' own in-package unit tests are authoritative for
fine-grained behaviour (deal-score maths, JWT verify edge cases, HTTP handler
status codes); this QE suite confirms they pass and that the contract-level
invariants hold structurally.

**Risk:** Structural assertions match source patterns, so a future refactor that
changes a string (e.g. an error code constant) without changing behaviour could
flag here — intentional, it forces the contract to be re-confirmed.
