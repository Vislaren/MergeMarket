## Test Cases — Integration — Session 11

_(I = Integration. PENDING until a device/emulator + reachable backend, and for
B-10 a Firebase/APNs backend, are available.)_

### TC-11-I-001: Log in from the account button reaches Home (B-08)

| Field | Value |
|-------|-------|
| Task reference | B-08 |
| Type | Integration |
| Preconditions | Device/emulator; mock server reachable at `API_BASE_URL` |
| Input | `user@example.com` / `secret123` |
| Steps | 1. Launch app 2. Tap the Home "Log in" account icon 3. Enter creds 4. Tap Log In |
| Expected result | Returns to Home showing the "Log out" affordance (authenticated) |
| Actual result | [PENDING] |
| Notes | `flutter test integration_test/auth_flow_test.dart --dart-define=API_BASE_URL=http://10.0.2.2:8089` |

### TC-11-I-002: Register then land on Home authenticated (B-08)

| Field | Value |
|-------|-------|
| Task reference | B-08 |
| Type | Integration |
| Preconditions | Device/emulator; mock server reachable |
| Input | a fresh email + matching passwords |
| Steps | 1. Launch 2. Open Login → Register 3. Fill form 4. Create Account |
| Expected result | Account created, navigates to Home authenticated |
| Actual result | [PENDING] |
| Notes | Manual against the mock server |

### TC-11-I-003: Tapping a price-drop push opens Product Detail (B-10)

| Field | Value |
|-------|-------|
| Task reference | B-10 |
| Type | Integration |
| Preconditions | Device with notification permission; Firebase/APNs configured; `FirebasePushBackend` override |
| Input | FCM data `{type:price_drop, product_id:prod-001, ...}` |
| Steps | 1. App backgrounded 2. Send test FCM message 3. Tap the notification |
| Expected result | App opens directly to `/product/prod-001` |
| Actual result | [PENDING] |
| Notes | Handling logic verified offline by TC-11-U-032…037 |

### TC-11-I-004: BFF aggregate + forwarding against a live upstream (B-09)

| Field | Value |
|-------|-------|
| Task reference | B-09 |
| Type | Integration |
| Preconditions | `go run ./cmd/bff` + a live upstream (mock server / Kong) |
| Input | `GET /api/v1/products/prod-001/detail`, `GET /api/v1/alerts` |
| Steps | 1. Start upstream 2. Start BFF 3. curl both endpoints |
| Expected result | `/detail` returns the merged view; `/alerts` is forwarded verbatim |
| Actual result | [PENDING] — covered by `server_test.go` via httptest upstream |
| Notes | `BFF_UPSTREAM_URL=http://localhost:8089 go run ./cmd/bff` |
