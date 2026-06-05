## Oracle — Product Detail (B-04)

| Input | Condition | Expected Output | Source of Truth |
|-------|-----------|-----------------|-----------------|
| `GET /products/{id}/history` 200 | valid product | `{product_id,title,history[],average_6m,lowest_30d}` decodes | API_CONTRACTS §Products |
| `GET /products/{id}/history` | product_id "unknown" | 404 `not_found` → `ApiException(notFound)` | API_CONTRACTS §Products |
| `GET /products/{id}/truth-score` 200 | valid product | `{score,sentiment,fake_review_risk,summary}` decodes | API_CONTRACTS §Products |
| history points | always | oldest-first; `latestPrice` = last point | API_CONTRACTS (sample order) |
| Deal meter score | 81–100 | band "Exceptional", gold | COMPONENT_LIBRARY §MMDealMeter |
| Deal meter score | 61–80 | band "Hot Deal", green | COMPONENT_LIBRARY §MMDealMeter |
| Deal meter score | 31–60 | band "Average", amber | COMPONENT_LIBRARY §MMDealMeter |
| Deal meter score | 0–30 | band "Poor Deal", red | COMPONENT_LIBRARY §MMDealMeter |
| current < average | comparison text | "X% below the 6-month average" | UI sample (product_detail) |
| Store comparison rows | always | sorted by total cost asc; row 0 = best deal (green) | COMPONENT_LIBRARY §MMStoreComparisonTable |
| Best price headline | offers present | = cheapest `total_cost` across stores | USER_FLOWS Flow 2 |
| Detail load | history 404 | `MMErrorState` with retry, no sections | UI_MIGRATION §Step 5 |
| Detail load | in flight | `MMSkeletonLoader` placeholders | UI_MIGRATION §Step 5 |
