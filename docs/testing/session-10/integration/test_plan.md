## Test Plan - Integration - Session 10 - B-07 Flutter Savings Dashboard Screen

**Scope:** End-to-end Savings Dashboard flow in the real Flutter app against
the B-02 mock server: navigate to the Savings tab, load `GET /api/v1/savings`,
render the dashboard, and tap a recent savings event.

**Out of scope:** Real backend persistence, real auth, native share-sheet
integration, and notification delivery.

**Approach:** Flutter integration tests launch `MergeMarketApp` with the real
router and repository stack. The only external dependency is the running mock
server configured by `API_BASE_URL`.

**Entry criteria:** A connected device/emulator or Chrome target is available;
`services/mock-server` is running; `API_BASE_URL` points to it.

**Exit criteria:** TC-10-I-001 and TC-10-I-002 pass on a device/emulator.

**Tools:** `integration_test`, `flutter_test`, B-02 mock server.

**Assumptions:** The mock server exposes `/api/v1/savings` and the product ids
from the savings fixture are valid enough for existing Product Detail routing.

**Risk:** The current environment has no device/emulator, so integration cases
are written but not executed this session.
