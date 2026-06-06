## Test Plan — Integration — Session 11 — B-08 Auth, B-09 BFF, B-10 Push

**Scope:** End-to-end paths that cross process boundaries:
- **B-08:** real app → running backend login from the Home account button →
  authenticated Home (USER_FLOWS Flow 1).
- **B-09:** the BFF against a real upstream (mock server or Kong) — the
  aggregated `/detail` view and pass-through forwarding. _(Exercised in the unit
  suite via an `httptest` upstream, which runs the real reverse proxy + HTTP
  client; a deployed end-to-end run is the remaining step.)_
- **B-10:** a real FCM/APNs message tap deep-linking to Product Detail
  (USER_FLOWS Flow 6).

**Out of scope:** load/perf; multi-device token management.

**Approach:** Integration tests use real HTTP / real navigation, no mocks. The
Flutter cases run on a device/emulator via `integration_test`; the BFF case runs
the binary against a live upstream.

**Entry criteria:** a device/emulator is connected; the B-02 mock server (or real
services) is reachable at `API_BASE_URL`; for B-10, a Firebase project +
`FirebasePushBackend` override are configured.

**Exit criteria:** all integration cases pass on at least one platform.

**Tools:** `integration_test` + `flutter_test`; Go binary + curl.

**Assumptions:** backend honours the API contracts; push payloads follow the
B-10 data shape.

**Risk:** **all integration cases are PENDING** locally — no device/emulator this
session and no Firebase project in the $0 bootstrap. The handling logic is fully
covered offline by the unit suites; only the transport edges are unverified.
