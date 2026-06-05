## Test Plan — Integration — Session 06 — B-03 Flutter Search Screen

**Scope:** The end-to-end Search flow (USER_FLOWS Flow 2) running the **real**
app against a **running B-02 mock server** over HTTP — no mocks, no provider
overrides:
- Type a query on Home → submit → navigate to Results.
- Results screen issues a real `GET /api/v1/search` and renders the multi-store
  list sorted by total cost.
- Tapping a result navigates to the Product Detail route.

**Out of scope:** Product Detail UI (placeholder until B-04); the real backend
(A-09 / B-11); offline/error injection at the network layer (covered by the
unit suite via `MockClient`).

**Approach:** `integration_test` driving `MergeMarketApp` on a device/emulator.
The mock server is started out-of-band; the app is pointed at it with
`--dart-define=API_BASE_URL=…`. Assertions verify real widgets appear after a
real round-trip.

**Entry criteria:**
- A connected device/emulator (or `-d chrome`).
- Mock server running and reachable (`go run ./cmd/mock-server`).
- `API_BASE_URL` define set (Android emulator → `http://10.0.2.2:<port>`).

**Exit criteria:** TC-06-I-001 and TC-06-I-002 pass.

**Tools:** `integration_test`, `flutter_test`, the B-02 Go mock server.

**Assumptions:** The mock returns the static `Search` fixture (3 Galaxy
offers); tests assert on those fixed shapes, not on mutation (the mock is
stateless).

**Risk / status:** **PENDING this session** — no device/emulator or running
mock server was available in the dev environment. The suite
(`apps/mobile/integration_test/search_flow_test.dart`) compiles and is ready to
run; the exact command is documented in its file header and this folder's
README.
