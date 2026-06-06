# Test Cases — Unit — Session 06 — B-03 Flutter Search Screen

_(U = Unit. All executed with `cd apps/mobile && flutter test test/unit`.)_

Source of truth: `project_docs/api/API_CONTRACTS.md` (Search),
`project_docs/ui/COMPONENT_LIBRARY.md` (widgets),
`project_docs/flows/USER_FLOWS.md` (Flow 2).

---

### TC-06-U-001: SearchResponse.fromJson decodes all fields
| Field | Value |
|-------|-------|
| Task reference | B-03 |
| Type | Unit |
| Preconditions | `kSearchSuccessJson` fixture |
| Input | Decode the success fixture |
| Steps | 1. `SearchResponse.fromJson(json)` 2. Inspect fields |
| Expected result | query=`galaxy`, 3 results, cached=true, latencyMs=42 |
| Actual result | [PASS] |
| Notes | Mirrors mock server `Search` fixture |

### TC-06-U-002: total_cost equals price + shipping for every offer
| Field | Value |
|-------|-------|
| Task reference | B-03 |
| Type | Unit |
| Preconditions | Decoded success response |
| Input | Each offer |
| Steps | 1. For each offer assert `totalCost == price + shipping` |
| Expected result | Holds for all 3 offers |
| Actual result | [PASS] |
| Notes | Core contract invariant |

### TC-06-U-003: storeCount counts distinct stores
| Field | Value |
|-------|-------|
| Task reference | B-03 |
| Type | Unit |
| Input | Decoded success response |
| Steps | 1. Read `storeCount` |
| Expected result | 3 (Jumia, Kilimall, AfricShop) |
| Actual result | [PASS] |

### TC-06-U-004: empty results decode to an empty list (not an error)
| Field | Value |
|-------|-------|
| Task reference | B-03 |
| Type | Unit |
| Input | `kSearchEmptyJson` |
| Steps | 1. Decode 2. Inspect results |
| Expected result | results empty, cached=false |
| Actual result | [PASS] |
| Notes | Empty ≠ transport error |

### TC-06-U-005: Product.fromJson tolerates missing fields
| Field | Value |
|-------|-------|
| Task reference | B-03 |
| Type | Unit |
| Input | `{ product_id: 'p1' }` |
| Steps | 1. Decode 2. Inspect defaults |
| Expected result | title='', price=0, dealScore=0 |
| Actual result | [PASS] |
| Notes | One malformed offer never breaks the list |

### TC-06-U-006: 200 decodes into a SearchResponse
| Field | Value |
|-------|-------|
| Task reference | B-03 |
| Type | Unit |
| Preconditions | `MockClient` returns 200 + fixture |
| Input | `search('galaxy')` |
| Steps | 1. Await search |
| Expected result | 3 results, query='galaxy' |
| Actual result | [PASS] |

### TC-06-U-007: request targets /api/v1/search with q and location
| Field | Value |
|-------|-------|
| Task reference | B-03 |
| Type | Unit |
| Input | `search('galaxy', location: 'NG')` |
| Steps | 1. Capture request URL |
| Expected result | path `/api/v1/search`, q=galaxy, location=NG |
| Actual result | [PASS] |

### TC-06-U-008: 400 throws badRequest with contract message
| Field | Value |
|-------|-------|
| Task reference | B-03 |
| Type | Unit |
| Input | `MockClient` 400 + `missing_query` body |
| Steps | 1. Expect throw |
| Expected result | `ApiException(badRequest)`, message contains "required" |
| Actual result | [PASS] |

### TC-06-U-009: 504 throws a timeout ApiException
| Field | Value |
|-------|-------|
| Task reference | B-03 |
| Type | Unit |
| Input | `MockClient` 504 |
| Steps | 1. Expect throw |
| Expected result | `ApiException(timeout)` |
| Actual result | [PASS] |

### TC-06-U-010: transport failure maps to a network ApiException
| Field | Value |
|-------|-------|
| Task reference | B-03 |
| Type | Unit |
| Input | `MockClient` throws `ClientException` |
| Steps | 1. Expect throw |
| Expected result | `ApiException(network)` |
| Actual result | [PASS] |

### TC-06-U-011: unexpected 500 maps to a server ApiException
| Field | Value |
|-------|-------|
| Task reference | B-03 |
| Type | Unit |
| Input | `MockClient` 500 |
| Steps | 1. Expect throw |
| Expected result | `ApiException(server)` |
| Actual result | [PASS] |

### TC-06-U-012: bestPrice orders by ascending total cost
| Field | Value |
|-------|-------|
| Task reference | B-03 |
| Type | Unit |
| Input | The 3 sample offers |
| Steps | 1. `sortResults(.., bestPrice)` |
| Expected result | Kilimall (247,900) → Jumia (250,000) → AfricShop (252,000) |
| Actual result | [PASS] |
| Notes | Default Results ordering |

### TC-06-U-013: topRated orders by descending deal score
| Field | Value |
|-------|-------|
| Task reference | B-03 |
| Type | Unit |
| Steps | 1. `sortResults(.., topRated)` |
| Expected result | Jumia (88) → Kilimall (81) → AfricShop (74) |
| Actual result | [PASS] |

### TC-06-U-014: fastestShip orders by ascending shipping
| Field | Value |
|-------|-------|
| Task reference | B-03 |
| Type | Unit |
| Steps | 1. `sortResults(.., fastestShip)` |
| Expected result | AfricShop (0) → Jumia (5,000) → Kilimall (8,000) |
| Actual result | [PASS] |

### TC-06-U-015: sortResults does not mutate the input list
| Field | Value |
|-------|-------|
| Task reference | B-03 |
| Type | Unit |
| Steps | 1. Snapshot order 2. Sort 3. Re-check input |
| Expected result | Input order unchanged |
| Actual result | [PASS] |

### TC-06-U-016: MMProductCard renders store, title, formatted total, badge
| Field | Value |
|-------|-------|
| Task reference | B-03 |
| Type | Unit (widget) |
| Input | Jumia offer, score 88 |
| Steps | 1. Pump card 2. Find texts 3. Tap |
| Expected result | "Jumia", title, "XAF 250,000", "Hot Deal"; onTap fires |
| Actual result | [PASS] |

### TC-06-U-017: free shipping is labelled, not shown as 0
| Field | Value |
|-------|-------|
| Task reference | B-03 |
| Type | Unit (widget) |
| Input | AfricShop offer, shipping 0, score 74 |
| Steps | 1. Pump card 2. Find texts |
| Expected result | "Free shipping" shown; badge "Good Value" |
| Actual result | [PASS] |

### TC-06-U-018: tapping the action icon fires onSearch
| Field | Value |
|-------|-------|
| Task reference | B-03 |
| Type | Unit (widget) |
| Steps | 1. Enter text 2. Tap arrow icon |
| Expected result | onSearch called once; controller text retained |
| Actual result | [PASS] |

### TC-06-U-019: submitting from the keyboard fires onSearch
| Field | Value |
|-------|-------|
| Task reference | B-03 |
| Type | Unit (widget) |
| Steps | 1. Enter text 2. Send `TextInputAction.search` |
| Expected result | onSearch called |
| Actual result | [PASS] |

### TC-06-U-020: loading shows a spinner instead of the icon
| Field | Value |
|-------|-------|
| Task reference | B-03 |
| Type | Unit (widget) |
| Input | `isLoading: true` |
| Steps | 1. Pump bar |
| Expected result | `CircularProgressIndicator` present; arrow icon absent |
| Actual result | [PASS] |

### TC-06-U-021: Results success shows a card per offer, cheapest first
| Field | Value |
|-------|-------|
| Task reference | B-03 |
| Type | Unit (widget) |
| Preconditions | `httpClientProvider` overridden, 200 + fixture |
| Steps | 1. Pump Results(query='galaxy') 2. Bounded pumps |
| Expected result | 3 `MMProductCard`s; first store = Kilimall; header "3 offers from 3 stores" |
| Actual result | [PASS] |

### TC-06-U-022: loading shows skeleton placeholders
| Field | Value |
|-------|-------|
| Task reference | B-03 |
| Type | Unit (widget) |
| Preconditions | `MockClient` future held open via `Completer` |
| Steps | 1. Pump first frame 2. Assert skeletons 3. Complete 4. Pump |
| Expected result | `MMSkeletonLoader`s while pending; 3 cards after completion |
| Actual result | [PASS] |

### TC-06-U-023: empty results show a no-results message
| Field | Value |
|-------|-------|
| Task reference | B-03 |
| Type | Unit (widget) |
| Input | 200 + empty results, query='zxqw' |
| Steps | 1. Pump + settle |
| Expected result | No cards; "No results found" shown |
| Actual result | [PASS] |

### TC-06-U-024: transport error shows MMErrorState with retry
| Field | Value |
|-------|-------|
| Task reference | B-03 |
| Type | Unit (widget) |
| Input | `MockClient` throws `ClientException` |
| Steps | 1. Pump + settle |
| Expected result | `MMErrorState` + "Try Again" button |
| Actual result | [PASS] |

### TC-06-U-025: empty query prompts for a search
| Field | Value |
|-------|-------|
| Task reference | B-03 |
| Type | Unit (widget) |
| Input | `ResultsScreen(query: '')` |
| Steps | 1. Pump + settle |
| Expected result | "Search for a product" prompt; no cards; no request made |
| Actual result | [PASS] |
