## Test Plan — Integration — Session 13 — B-11 (Swap Mocks for Real Backend)

**Scope:** The full user journey over the **real backend** through Kong:
register/login → search → results → product detail → add to wishlist → set
alert → receive price-drop notification. Verifies cross-service contracts and
that the authenticated client / BFF JWT forwarding work end-to-end against live
services.

**Out of scope:** Unit-level logic (covered by the unit suite); load/perf;
native push delivery via real FCM (no Firebase project in the $0 bootstrap).

**Approach:** Integration tests make **real HTTP calls** (no mocks): bring up the
stack (`docker compose up -d` + the Go services, or point at the VPS Kong), set
`API_BASE_URL` to Kong (`:8088`), and drive the flow. The Go-side cross-service
checks hit Kong directly; the Flutter integration test drives the app on a
device/emulator against the live backend.

**Entry criteria:** All Agent A services deployed and healthy behind Kong
(auth A-08, history A-07, BFF B-09, **and** the not-yet-built search / wishlist /
alerts / savings / truth-score services — see `CONTRACT_AUDIT.md`); a running
Docker daemon or reachable VPS; a connected device/emulator for the Flutter leg.

**Exit criteria:** Each step returns the contracted status/shape; a price drop
below an alert threshold produces a notification payload that deep-links to the
product.

**Tools:** `flutter_test` + `integration_test`; `go test -tags=integration`;
`docker compose`.

**Assumptions:** When the real services exist, their JSON equals the mock
fixtures (verified statically for the services that do exist). Kong enforces JWT
on all non-auth routes, so the authenticated client's Bearer header is required.

**Risk / status:** **All integration cases are PENDING this session.** Docker is
not running locally and Agent A's services are not deployed, so no live stack is
reachable. Additionally, the search / wishlist / alerts / savings / truth-score
real services **do not exist yet**, so even with Docker the full flow cannot run
until they are built. The suite is written and self-skips when `API_BASE_URL` is
unset or the backend is unreachable.
