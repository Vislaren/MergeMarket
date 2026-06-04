## Test Plan — Integration — Session 03 — B-01 Flutter Project Setup

**Scope:** End-to-end navigation through the **real** `go_router` + app widget
tree (no stubbing of the navigation stack):
- Bottom-nav tab switching between the four primary destinations.
- Deep-linking to `/product/:id` (the price-drop notification target, Flow 6).

**Out of scope:** Backend/API integration — B-01 has no service layer or mock
server wiring yet (mock server is B-02; swap-to-real is B-11). Feature
behaviour of individual screens.

**Approach:** Integration tests in `apps/mobile/integration_test/` drive the
fully-composed `MergeMarketApp` inside a `ProviderScope` using
`IntegrationTestWidgetsFlutterBinding`, exercising the router exactly as a user
or a deep link would.

**Entry criteria:** Unit suite green; `flutter analyze` clean; a target
device/emulator (Android/iOS) connected, or a CI runner with one (A-10).

**Exit criteria:** TC-03-I-001 and TC-03-I-002 pass on a device/emulator.

**Tools:** `integration_test` + `flutter_test`.

**Assumptions:** The integration_test harness requires a connected device; it
cannot run on the bare host (the project targets Android/iOS only).

**Risk / limitation:** **No device/emulator was connected this session**, so
both integration cases are **PENDING** (same situation as session-02's live
integration tests). The suite is written and analyzer-clean; the navigation it
asserts is partially covered headless by unit widget tests TC-03-U-007/008.
Rerun with: `cd apps/mobile && flutter test integration_test/navigation_flow_test.dart`
once a device/emulator is available (or in the A-10 CI pipeline).
