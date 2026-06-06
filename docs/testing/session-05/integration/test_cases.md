# Integration Test Cases — Session 05 — B-02 Mock Server

All cases executed against the live binary this session. Legend: [PASS]/[FAIL]/[PENDING].

Common preconditions: the mock-server binary is built and started on an
OS-assigned free port via `MOCK_SERVER_PORT`; `/health` returned 200 before the
case ran.

---

### TC-05-I-001: GET /health returns the contract shape
| Field | Value |
|-------|-------|
| Task reference | B-02 |
| Type | Integration |
| Input | `GET /health` |
| Steps | 1. GET /health 2. Assert 200 3. Assert body `status=ok`, `service=mock-server`, `version` present |
| Expected result | 200 with `{status, service, version}` |
| Actual result | [PASS] |
| Notes | Also used as the readiness probe |

---

### TC-05-I-002: GET /search returns multi-store results; total_cost = price + shipping
| Field | Value |
|-------|-------|
| Task reference | B-02 |
| Type | Integration |
| Input | `GET /api/v1/search?q=phone&location=CM` |
| Steps | 1. GET search 2. Assert 200 3. Assert `query=="phone"` 4. Assert ≥2 results 5. For each: `total_cost==price+shipping`, `0≤deal_score≤100` |
| Expected result | 200, multi-store results, invariant holds |
| Actual result | [PASS] |
| Notes | 3 stores: Jumia/Kilimall/AfricShop |

---

### TC-05-I-003: Search error paths (400 missing_query, 504 timeout)
| Field | Value |
|-------|-------|
| Task reference | B-02 |
| Type | Integration |
| Input | `GET /api/v1/search` (no q); `GET /api/v1/search?q=timeout` |
| Steps | 1. GET without q → assert 400 2. GET with q=timeout → assert 504 |
| Expected result | 400 then 504 |
| Actual result | [PASS] |
| Notes | Sentinel-driven, stateless |

---

### TC-05-I-004: Product history — 6 points + aggregates; unknown → 404
| Field | Value |
|-------|-------|
| Task reference | B-02 |
| Type | Integration |
| Input | `GET /api/v1/products/prod-001/history`; `.../unknown/history` |
| Steps | 1. GET history → assert 6 points, `average_6m>0`, `lowest_30d>0` 2. GET unknown → assert 404 |
| Expected result | 200 with aggregates; 404 for sentinel |
| Actual result | [PASS] |
| Notes | — |

---

### TC-05-I-005: Auth flows (201 register, 401 login, 401 refresh)
| Field | Value |
|-------|-------|
| Task reference | B-02 |
| Type | Integration |
| Input | register new email; login wrong password; refresh expired token |
| Steps | 1. POST register → 201 2. POST login `wrongpassword` → 401 3. POST refresh `expired` → 401 |
| Expected result | 201, 401, 401 |
| Actual result | [PASS] |
| Notes | Token bundle returned on success path |

---

### TC-05-I-006: Wishlist + alerts CRUD status codes
| Field | Value |
|-------|-------|
| Task reference | B-02 |
| Type | Integration |
| Input | wishlist list/add-duplicate/delete; alert create/delete-unknown |
| Steps | 1. GET wishlist → 200 2. POST wishlist `prod-001` → 409 3. DELETE wishlist/wl-001 → 204 4. POST alerts → 201 5. DELETE alerts/unknown → 404 |
| Expected result | 200, 409, 204, 201, 404 |
| Actual result | [PASS] |
| Notes | 204 verified to carry no body in handler unit tests |

---

### TC-05-I-007: Savings total equals sum of transactions
| Field | Value |
|-------|-------|
| Task reference | B-02 |
| Type | Integration |
| Input | `GET /api/v1/savings` |
| Steps | 1. GET savings 2. Sum `transactions[].saved` 3. Assert `total_saved == sum` |
| Expected result | Consistent total |
| Actual result | [PASS] |
| Notes | Gamified dashboard data integrity |

---

### TC-05-I-008: CORS preflight returns 204 with allow-origin
| Field | Value |
|-------|-------|
| Task reference | B-02 |
| Type | Integration |
| Input | `OPTIONS /api/v1/search` |
| Steps | 1. OPTIONS request 2. Assert 204 3. Assert `Access-Control-Allow-Origin: *` |
| Expected result | Preflight handled for Flutter web/dev clients |
| Actual result | [PASS] |
| Notes | — |
