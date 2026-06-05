## Test Plan — Integration — Session 09 — B-06 Flutter Alerts Screen

**Scope:** The Alerts flow end-to-end against the **real** B-02 mock server over
HTTP (no `MockClient`): open the Alerts tab → the active-alerts list loads from
the live repository; and set an alert from a wishlist bell → confirm → land on
the Alerts tab.

**Out of scope:** The real backend services (swapped in at B-11); push
notification delivery (B-10); pixel fidelity.

**Approach:** `integration_test` drives the real `MergeMarketApp` on a
device/emulator, hitting the mock server at the `--dart-define` base URL.

**Entry criteria:** A device/emulator is connected; the mock server is running
and reachable.

**Exit criteria:** TC-09-I-001 … 002 pass on a device against the mock server.

**Tools:** `integration_test`, `flutter_test`, `flutter_riverpod`.

**Assumptions:** Android emulator reaches the host at `10.0.2.2`; mock server on
port 8081.

**Risk:** PENDING this session — no device/emulator available. Suite compiles and
is ready; the run command is in the test header.
