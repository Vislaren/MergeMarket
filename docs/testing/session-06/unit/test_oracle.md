## Oracle — Search Feature (B-03)

| Input | Condition | Expected Output | Source of Truth |
|-------|-----------|-----------------|-----------------|
| `GET /api/v1/search?q=galaxy&location=CM` | 200 | `SearchResponse` with results, `cached`, `latency_ms` | API_CONTRACTS.md → Search |
| Any search offer | — | `total_cost == price + shipping` | API_CONTRACTS.md → Search / mock fixtures |
| `GET /api/v1/search` (no `q`) | 400 | `ApiException(badRequest)`, message from `{error,message}` | API_CONTRACTS.md (400 missing_query) |
| `q=timeout` | 504 | `ApiException(timeout)` | API_CONTRACTS.md (504 timeout) |
| Server unreachable | transport error | `ApiException(network)` | UI_MIGRATION_PROMPT §5 (error state required) |
| 5xx / unexpected status | — | `ApiException(server)` | API_CONTRACTS.md error shape |
| Results list, default sort | bestPrice | ordered by ascending `total_cost` | USER_FLOWS Flow 2 ("Sorted by Total Cost") |
| Results list, Top Rated | topRated | ordered by descending `deal_score` | COMPONENT_LIBRARY (deal score) |
| Results list, Fastest Ship | fastestShip | ordered by ascending `shipping` | Results design (Fastest Ship filter) |
| deal_score 81–100 | card badge | "Hot Deal" (gold) | COMPONENT_LIBRARY → MMDealMeter ranges |
| deal_score 61–80 | card badge | "Good Value" (green) | COMPONENT_LIBRARY → MMDealMeter ranges |
| shipping == 0 | card line | "Free shipping" | Results design ("Free shipping") |
| Search in flight | loading | `MMSkeletonLoader` placeholders | UI_MIGRATION_PROMPT §5 (loading state) |
| Results empty | empty | "No results found for …" | UI_MIGRATION_PROMPT §5 (empty state) |
| Search failed | error | `MMErrorState` + retry | UI_MIGRATION_PROMPT §5 (error state) |
| Home: submit query | navigation | route `/results?q=…` | COMPONENT_LIBRARY §Navigation |
| Results card tapped | navigation | route `/product/{product_id}` | USER_FLOWS Flow 2 |
