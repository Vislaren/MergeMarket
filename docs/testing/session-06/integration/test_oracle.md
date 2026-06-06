## Oracle — Search Flow Integration (B-03)

| Input | Condition | Expected Output | Source of Truth |
|-------|-----------|-----------------|-----------------|
| Home: type "galaxy" + submit | mock server reachable | Navigate to `/results?q=galaxy` | COMPONENT_LIBRARY §Navigation / USER_FLOWS Flow 2 |
| Results loads | real `GET /api/v1/search` returns 200 | ≥1 `MMProductCard` rendered, sorted by total cost | USER_FLOWS Flow 2 ("Sorted by Total Cost") |
| Result card visible | success | Total cost shown with currency (`XAF …`) | API_CONTRACTS.md → Search / Results design |
| Tap first result card | — | Navigate to `/product/{product_id}` | USER_FLOWS Flow 2 |
| Mock server unreachable | transport error | `MMErrorState` with "Try Again" | UI_MIGRATION_PROMPT §5 (error state) |
