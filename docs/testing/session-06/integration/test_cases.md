# Test Cases — Integration — Session 06 — B-03 Flutter Search Screen

_(I = Integration. Real app + running B-02 mock server, over HTTP.)_

File: `apps/mobile/integration_test/search_flow_test.dart`

---

### TC-06-I-001: search from Home renders store results
| Field | Value |
|-------|-------|
| Task reference | B-03 |
| Type | Integration |
| Preconditions | Mock server running; app launched with `API_BASE_URL` set; device/emulator connected |
| Input | Query "galaxy" typed into `MMSearchBar`, search action submitted |
| Steps | 1. Launch app 2. Enter "galaxy" 3. Submit 4. Wait for results |
| Expected result | Navigates to Results; ≥1 `MMProductCard` rendered; total cost text (`XAF …`) shown |
| Actual result | [PENDING] |
| Notes | Needs device + running mock server (none this session) |

### TC-06-I-002: tapping a result opens Product Detail
| Field | Value |
|-------|-------|
| Task reference | B-03 |
| Type | Integration |
| Preconditions | As TC-06-I-001 |
| Input | Tap on the first result card |
| Steps | 1. Run search 2. Tap first `MMProductCard` 3. Wait for navigation |
| Expected result | Product Detail route (`/product/{product_id}`) reached (placeholder until B-04) |
| Actual result | [PENDING] |
| Notes | Verifies the Results → Product Detail navigation wiring |
