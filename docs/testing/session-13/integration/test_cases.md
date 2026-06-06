# Test Cases — Integration — Session 13 — B-11

All cases are **PENDING**: no live backend is reachable (no Docker / services not
deployed), and several upstream services do not exist yet (`CONTRACT_AUDIT.md`).
Scaffold: `integration/test_suite/b11_e2e_integration_test.dart` (self-skips
without `API_BASE_URL`).

---

### TC-13-I-001: Register/login obtains a session over real auth (A-08)
| Field | Value |
|-------|-------|
| Task reference | B-11 |
| Type | Integration |
| Preconditions | Kong + auth service up; `API_BASE_URL` = Kong `:8088` |
| Input | `POST /api/v1/auth/login` valid credentials |
| Expected result | 200 with `token`/`refresh_token`/`expires_at`; session persisted |
| Actual result | [PENDING] |
| Notes | The one leg runnable today once auth is deployed (real service exists). |

### TC-13-I-002: Protected request without a token is rejected by Kong
| Field | Value |
|-------|-------|
| Task reference | B-11 |
| Type | Integration |
| Input | `GET /api/v1/wishlist` with no Authorization header |
| Expected result | 401 from Kong's JWT plugin |
| Actual result | [PENDING] |
| Notes | Confirms the authenticated client is actually required (vs. the auth-free mock). |

### TC-13-I-003: Search → results over the real backend
| Field | Value |
|-------|-------|
| Task reference | B-11 |
| Input | `GET /api/v1/search?q=phone&location=CM` with Bearer token |
| Expected result | 200 with `results[]`, `cached`, `latency_ms` per contract |
| Actual result | [PENDING] |
| Notes | **Blocked:** no real search service exists yet. |

### TC-13-I-004: Add to wishlist, then list reflects it
| Field | Value |
|-------|-------|
| Task reference | B-11 |
| Input | `POST /api/v1/wishlist {product_id}` then `GET /api/v1/wishlist` |
| Expected result | 201 then the item present (stateful real backend, unlike the mock) |
| Actual result | [PENDING] |
| Notes | **Blocked:** no real wishlist service exists yet. |

### TC-13-I-005: Set an alert below current price
| Field | Value |
|-------|-------|
| Task reference | B-11 |
| Input | `POST /api/v1/alerts {product_id, threshold_price, currency}` |
| Expected result | 201 with `alert_id`; alert listed as active |
| Actual result | [PENDING] |
| Notes | **Blocked:** no real alerts CRUD service exists (history emits events only). |

### TC-13-I-006: Token-refresh-on-401 against real auth
| Field | Value |
|-------|-------|
| Task reference | B-11 |
| Input | A protected call after the access token has expired |
| Expected result | client refreshes via `/auth/refresh`, replays, succeeds transparently |
| Actual result | [PENDING] |
| Notes | Runnable once auth + one protected real service are deployed. |

### TC-13-I-007: Price drop produces a deep-linking notification
| Field | Value |
|-------|-------|
| Task reference | B-11 |
| Input | History heartbeat records a price ≤ threshold (downward crossing) |
| Expected result | a notification payload routes to `/product/{id}` |
| Actual result | [PENDING] |
| Notes | **Blocked:** needs alerts service + FCM; client handling unit-tested in B-10. |

### TC-13-I-008: BFF `/detail` aggregate over real services behind Kong
| Field | Value |
|-------|-------|
| Task reference | B-11 |
| Input | `GET /api/v1/products/{id}/detail` with Bearer token |
| Expected result | 200 aggregate; JWT forwarded to upstream history/search/truth-score |
| Actual result | [PENDING] |
| Notes | History leg works today; search/truth-score legs blocked on missing services. |
