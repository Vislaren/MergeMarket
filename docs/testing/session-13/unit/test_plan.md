## Test Plan — Unit — Session 13 — B-11 Integration (Swap Mocks for Real Backend)

**Scope:** The B-11 client/BFF integration plumbing, isolated with mocks:
- `AuthenticatedClient` — Bearer-token injection and refresh-on-401 + replay.
- `AuthController.refreshSession()` — the refresh the interceptor drives.
- BFF aggregate — forwarding the caller's `Authorization` header upstream.

**Out of scope:** Live calls to Kong or any real Agent A service (no Docker, no
deployed backend); the search/wishlist/alerts/savings/truth-score *real services*
(they do not exist yet — see `CONTRACT_AUDIT.md`); native integrations
(url_launcher, receive_sharing_intent, firebase_messaging). Re-verifying B-01…B-10
screens (covered by their own sessions) beyond confirming the suite stays green.

**Approach:** Unit tests isolate each unit with `package:http`'s `MockClient`
and an in-memory `FakeSessionStore` (Flutter) / an `httptest` upstream (Go). No
network, no platform channels — the whole suite runs offline.

**Entry criteria:** App + BFF build; baseline suite green (was 141/141 Flutter).

**Exit criteria:** All B-11 unit cases pass; full Flutter suite and BFF
`go test ./...` stay green; `flutter analyze` clean.

**Tools:** `flutter_test` + `package:http/testing.dart`; Go `testing` +
`net/http/httptest`.

**Assumptions:** Where a real service exists (auth, history) its JSON shape
equals the mock's, so swapping the transport does not change decoding (confirmed
in the contract audit). Signed-out requests pass through the authenticated
client unchanged, preserving the existing screen/repository tests.

**Risk:** The interceptor's refresh-replay must buffer and replay request bodies
correctly (covered by TC-13-U-008) and must not loop on a persistent 401
(covered by TC-13-U-004). Concurrent 401 bursts must coalesce to one refresh
(TC-13-U-007). Rotating refresh tokens under high concurrency are mitigated by a
pre-refresh token re-check but not exhaustively load-tested — noted as a known
limitation.
