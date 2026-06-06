# Test Cases — Integration — Session 03 — B-01 Flutter Project Setup

Suite file: `apps/mobile/integration_test/navigation_flow_test.dart`

Status this session: **PENDING** — no device/emulator connected. The host
exposes only Windows/web targets, which this project (Android/iOS) does not
build. Code is written and `flutter analyze` is clean.

Rerun: `cd apps/mobile && flutter test integration_test/navigation_flow_test.dart`

---

### TC-03-I-001: Bottom nav switches between primary destinations

| Field | Value |
|-------|-------|
| Task reference | B-01 |
| Type | Integration |
| Preconditions | App built; device/emulator connected |
| Input | Taps on the "Wishlist" then "Alerts" bottom-nav labels |
| Steps | 1. Pump `ProviderScope(MergeMarketApp)`. 2. Settle (starts on Search). 3. Tap "Wishlist" → settle. 4. Tap "Alerts" → settle. |
| Expected result | AppBar title shows "Wishlist", then "Alerts"; indexed-stack branch switches |
| Actual result | [PENDING] |
| Notes | Verifies StatefulShellRoute branch navigation |

---

### TC-03-I-002: Deep link /product/:id renders Product Detail

| Field | Value |
|-------|-------|
| Task reference | B-01 |
| Type | Integration |
| Preconditions | App built; device/emulator connected |
| Input | `router.go('/product/abc123')` |
| Steps | 1. Pump app, capturing the GoRouter via a Consumer. 2. Settle. 3. `go('/product/abc123')` → settle. |
| Expected result | "Product Detail" AppBar shown; the id `abc123` appears (path param passed through) |
| Actual result | [PENDING] |
| Notes | Models the Flow 6 notification deep-link target; confirms `/product/:id` path-param wiring |
