## Test Plan — Unit — Session 11 — B-08 Auth, B-09 BFF, B-10 Push

**Scope:**
- **B-08 (Flutter Auth):** `AuthSession` decode; form validators; `AuthRepository`
  status→`ApiException` mapping (register 201/400/409, login 200/401, refresh
  200/401, transport→network); `AuthController` session restore / login /
  register / logout with secure-storage persistence; Login & Register screen
  validation, success navigation, inline error banners, password visibility.
- **B-09 (Go BFF):** `config` env loading/validation; `server` health, `/metrics`
  counter, aggregated product-detail (history + truth-score + offers, sorted,
  best offer, deal score, store count), product-detail 404, and reverse-proxy
  forwarding of unknown routes to the upstream.
- **B-10 (Push):** `PushNotification.fromData` parsing (price_drop / restock /
  unknown / missing, string-number coercion); `NotificationService` foreground
  republish, tap routing, launch-message handling, unroutable-tap drop, and
  `routeFor`; Alerts in-app banner on a foreground push.

**Out of scope:** real Firebase/APNs transport (no project in the $0 bootstrap);
real secure-storage platform channel (faked); affiliate-link/Google OAuth/forgot-
password flows (stubbed); wiring the Flutter app to consume the BFF `/detail`
(that is B-11).

**Approach:** Unit tests isolate each unit with test doubles — `package:http`
`MockClient` for repositories, an in-memory `FakeSessionStore`, a controllable
`FakePushBackend`, and an `httptest` fake upstream for the Go BFF. Widget tests
drive the real provider chains with only the transport/storage overridden.

**Entry criteria:** code compiles; `flutter analyze` clean; `go build`/`vet`/
`gofmt` clean.

**Exit criteria:** all unit/widget cases pass; Go coverage config 100% / server
91%; no analyzer warnings.

**Tools:** `flutter_test` + `package:http/testing`; Go `testing` + `httptest`.

**Assumptions:** mock-server auth sentinels (`wrongpassword`→401,
`taken@mergemarket.app`→409) model the real Auth service (A-08); FCM/APNs deliver
a flat string `data` map.

**Risk:** integration paths (device/Firebase) remain unverified locally; the BFF
`upstream` package shows 0% in its own package report because it is exercised
cross-package from `server` tests.
