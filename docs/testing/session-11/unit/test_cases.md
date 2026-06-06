## Test Cases — Unit — Session 11

_(U = Unit. Flutter cases run via `flutter test`; Go cases via `go test`.)_

### B-08 — Flutter Auth

| ID | Test | Expected | Result |
|----|------|----------|--------|
| TC-11-U-001 | `AuthSession.fromJson` decodes the token bundle | fields set, not expired | [PASS] |
| TC-11-U-002 | past `expires_at` | `isExpired == true` | [PASS] |
| TC-11-U-003 | missing/invalid fields | safe values, expired (epoch) | [PASS] |
| TC-11-U-004 | `validateEmail` empty/malformed | error string | [PASS] |
| TC-11-U-005 | `validateEmail` valid (trimmed) | null | [PASS] |
| TC-11-U-006 | `validateLoginPassword` presence only | null when non-empty | [PASS] |
| TC-11-U-007 | `validateNewPassword` min length | error < 8 chars | [PASS] |
| TC-11-U-008 | `validateConfirmPassword` match | error unless identical | [PASS] |
| TC-11-U-009 | register 201 POSTs creds, decodes session | body+url correct | [PASS] |
| TC-11-U-010 | register 400 | `ApiException(badRequest)` | [PASS] |
| TC-11-U-011 | register 409 | `ApiException(conflict)` | [PASS] |
| TC-11-U-012 | login 200 decodes session | refresh token set | [PASS] |
| TC-11-U-013 | login 401 | `ApiException(unauthorized)` | [PASS] |
| TC-11-U-014 | login transport failure | `ApiException(network)` | [PASS] |
| TC-11-U-015 | refresh 200 POSTs refresh token | body+url correct | [PASS] |
| TC-11-U-016 | refresh 401 | `ApiException(unauthorized)` | [PASS] |
| TC-11-U-017 | controller restores a valid persisted session | authenticated | [PASS] |
| TC-11-U-018 | controller with expired persisted session | signed out | [PASS] |
| TC-11-U-019 | controller login persists + authenticates | store written | [PASS] |
| TC-11-U-020 | controller logout clears storage | signed out, cleared | [PASS] |
| TC-11-U-021 | Login empty submit | field validation errors, no nav | [PASS] |
| TC-11-U-022 | Login valid creds | navigates Home, session stored | [PASS] |
| TC-11-U-023 | Login 401 | inline error banner, no nav | [PASS] |
| TC-11-U-024 | Login password visibility toggle | reveals password | [PASS] |
| TC-11-U-025 | Register mismatched confirm | validation error, no submit | [PASS] |
| TC-11-U-026 | Register valid | account created, navigates Home | [PASS] |
| TC-11-U-027 | Register 409 taken email | conflict error banner | [PASS] |

### B-10 — Push Notifications

| ID | Test | Expected | Result |
|----|------|----------|--------|
| TC-11-U-028 | parse price-drop (string→num coercion) | typed, routable | [PASS] |
| TC-11-U-029 | parse restock, optional fields absent | typed, routable | [PASS] |
| TC-11-U-030 | unknown type, no product id | not routable | [PASS] |
| TC-11-U-031 | empty data | safe defaults, not routable | [PASS] |
| TC-11-U-032 | foreground message → `inbound` parsed | emits PushNotification | [PASS] |
| TC-11-U-033 | tap → `taps` stream | emits PushNotification | [PASS] |
| TC-11-U-034 | launch message delivered as tap on init | emits + token set | [PASS] |
| TC-11-U-035 | unroutable tap dropped | no `taps` emission | [PASS] |
| TC-11-U-036 | `routeFor` deep link / null | `/product/{id}` or null | [PASS] |
| TC-11-U-037 | Alerts shows banner + View on foreground push | SnackBar + action | [PASS] |

### B-09 — Go BFF (`go test ./...`)

| ID | Test | Expected | Result |
|----|------|----------|--------|
| TC-11-U-038 | `TestLoadDefaults` | port 8082, default upstream | [PASS] |
| TC-11-U-039 | `TestLoadPortPrecedenceAndOverride` | BFF_PORT wins, upstream override | [PASS] |
| TC-11-U-040 | `TestLoadInvalidPort` | error (non-numeric / out of range) | [PASS] |
| TC-11-U-041 | `TestLoadInvalidUpstream` | error on bad URL | [PASS] |
| TC-11-U-042 | `TestHealth` | `{status:ok, service:bff}` | [PASS] |
| TC-11-U-043 | `TestMetricsCountsRequests` | `bff_requests_total` exposed | [PASS] |
| TC-11-U-044 | `TestProductDetailAggregates` | merged, sorted, best offer + deal score + store count | [PASS] |
| TC-11-U-045 | `TestProductDetailNotFound` | 404 `{error:not_found}` | [PASS] |
| TC-11-U-046 | `TestForwardsUnknownRoutesToUpstream` | `/api/v1/alerts` reverse-proxied | [PASS] |
