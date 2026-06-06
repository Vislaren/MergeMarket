# Test Cases — Unit — Session 13 — B-11

Executable sources:
- Flutter: `apps/mobile/test/unit/authenticated_client_test.dart`
- Go (BFF): `services/bff/internal/server/server_test.go`
  (`TestProductDetailForwardsAuthHeaderUpstream`)

---

### TC-13-U-001: Attaches Bearer token when signed in
| Field | Value |
|-------|-------|
| Task reference | B-11 |
| Type | Unit |
| Preconditions | `AuthenticatedClient` with `readToken → "tok-123"` |
| Input | `GET /api/v1/search?q=phone` |
| Steps | 1. Send request through the client. 2. Inspect the header seen by the inner client. |
| Expected result | Inner request carries `Authorization: Bearer tok-123`; 200 returned |
| Actual result | [PASS] |
| Notes | — |

### TC-13-U-002: No Authorization header when signed out
| Field | Value |
|-------|-------|
| Task reference | B-11 |
| Type | Unit |
| Preconditions | `readToken → null` |
| Input | `GET /api/v1/search` |
| Steps | 1. Send. 2. Check header presence. |
| Expected result | No `Authorization` header (pass-through preserves mock/test behaviour) |
| Actual result | [PASS] |
| Notes | This is why the 141 prior screen/repo tests still pass. |

### TC-13-U-003: Refreshes once on 401 and replays with the new token
| Field | Value |
|-------|-------|
| Task reference | B-11 |
| Type | Unit |
| Preconditions | Token "old"; inner returns 401 for "old", 200 otherwise; refresh sets token "new" |
| Input | `GET /api/v1/wishlist` |
| Steps | 1. Send. 2. Observe refresh + replay. |
| Expected result | refresh called exactly once; requests seen = `[Bearer old, Bearer new]`; final 200 |
| Actual result | [PASS] |
| Notes | — |

### TC-13-U-004: A failed refresh surfaces the original 401
| Field | Value |
|-------|-------|
| Task reference | B-11 |
| Type | Unit |
| Preconditions | Token "old"; inner always 401; refresh returns false |
| Input | `GET /api/v1/wishlist` |
| Expected result | Final status 401; refresh attempted once (no infinite loop) |
| Actual result | [PASS] |
| Notes | The router guard routes to login on the surfaced 401. |

### TC-13-U-005: Never refreshes a 401 on auth endpoints
| Field | Value |
|-------|-------|
| Task reference | B-11 |
| Type | Unit |
| Preconditions | Token present; inner returns 401 |
| Input | `POST /api/v1/auth/login` |
| Expected result | 401 returned; refresh NOT called (a 401 here is a real credential failure) |
| Actual result | [PASS] |
| Notes | — |

### TC-13-U-006: No refresh when no token was sent
| Field | Value |
|-------|-------|
| Task reference | B-11 |
| Type | Unit |
| Preconditions | `readToken → null`; inner returns 401 |
| Input | `GET /api/v1/search` |
| Expected result | 401 returned; refresh NOT called |
| Actual result | [PASS] |
| Notes | Guards the signed-out path from spurious refreshes. |

### TC-13-U-007: Concurrent 401s coalesce into a single refresh
| Field | Value |
|-------|-------|
| Task reference | B-11 |
| Type | Unit |
| Preconditions | Three parallel protected GETs; refresh has a 10ms delay |
| Input | wishlist + alerts + savings in `Future.wait` |
| Expected result | all 200; refresh called exactly once |
| Actual result | [PASS] |
| Notes | In-flight refresh future is shared; a pre-check also avoids re-spending a rotated token. |

### TC-13-U-008: Replays the POST body intact after a refresh
| Field | Value |
|-------|-------|
| Task reference | B-11 |
| Type | Unit |
| Preconditions | Token "old"→"new"; inner 401 for "old", 201 otherwise |
| Input | `POST /api/v1/alerts` with a JSON body |
| Expected result | 201; the body seen on both attempts is byte-identical |
| Actual result | [PASS] |
| Notes | Body buffered once so the finalized request can be replayed. |

### TC-13-U-009: refreshSession swaps in the new session
| Field | Value |
|-------|-------|
| Task reference | B-11 |
| Type | Unit |
| Preconditions | Valid persisted session; refresh endpoint returns a new bundle |
| Input | `AuthController.refreshSession()` |
| Expected result | returns true; state + store hold the new token; email preserved; still authenticated |
| Actual result | [PASS] |
| Notes | — |

### TC-13-U-010: refreshSession clears the session on failure
| Field | Value |
|-------|-------|
| Task reference | B-11 |
| Type | Unit |
| Preconditions | Valid persisted session; refresh endpoint returns 401 |
| Input | `AuthController.refreshSession()` |
| Expected result | returns false; session cleared (clearCount 1); signed out |
| Actual result | [PASS] |
| Notes | — |

### TC-13-U-011 (BFF): Aggregate forwards the Authorization header upstream
| Field | Value |
|-------|-------|
| Task reference | B-11 |
| Type | Unit (Go) |
| Preconditions | BFF wired to an httptest upstream that records `Authorization` |
| Input | `GET /api/v1/products/prod-001/detail` with `Authorization: Bearer caller-token` |
| Steps | 1. Serve the request. 2. Read the header on each of the 3 upstream calls. |
| Expected result | 200; all three upstream calls (history, search, truth-score) saw `Bearer caller-token` |
| Actual result | [PASS] |
| Notes | Lets the BFF aggregate work behind Kong's JWT gate. |
