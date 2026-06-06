## Test Plan — Integration — Session 07 — B-04 Flutter Product Detail Screen

**Scope:** The Product Detail flow end-to-end against the **real** B-02 mock
server over HTTP (no `MockClient`): search → tap a result → Product Detail loads
its history, deal meter, store comparison, and truth score from the live
repositories; the Add-to-Wishlist action confirms via SnackBar.

**Out of scope:** The real backend services (A-05…A-09, swapped in at B-11);
affiliate-link navigation; visual fidelity.

**Approach:** `integration_test` drives the real `MergeMarketApp` widget on a
device/emulator. HTTP goes to the mock server at the `--dart-define` base URL.
Assertions check that the detail widgets mount and the headline CTA renders.

**Entry criteria:** A device/emulator is connected; the mock server is running
and reachable at the configured base URL.

**Exit criteria:** TC-07-I-001 … 002 pass on a device against the mock server.

**Tools:** `integration_test`, `flutter_test`, `flutter_riverpod`.

**Assumptions:** The Android emulator reaches the host at `10.0.2.2`; the mock
server is started on the documented port (8081 to avoid the Jenkins/Coolify
8080 collision noted in PORTS_README).

**Risk:** PENDING this session — no device/emulator in the dev environment. The
suite compiles and is ready to run; the run command is in the test header.
