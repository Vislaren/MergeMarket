# Test Cases — Unit — Session 07 — B-04 Product Detail

_(U = Unit. All executed via `flutter test`; all PASS.)_

| ID | Task | Preconditions | Input | Expected | Actual |
|----|------|---------------|-------|----------|--------|
| TC-07-U-001 | B-04 | — | `kHistoryJson` | `PriceHistory` decodes: 6 points, avg 256000, low 245000 | [PASS] |
| TC-07-U-002 | B-04 | — | `kHistoryJson` | `latestPrice`=245000, `currency`=XAF, oldest-first order | [PASS] |
| TC-07-U-003 | B-04 | — | `{}`-ish history | empty series, `latestPrice`=null, safe defaults | [PASS] |
| TC-07-U-004 | B-04 | — | good + bad `recorded_at` | parses date; bad date → epoch | [PASS] |
| TC-07-U-005 | B-04 | — | `kTruthScoreJson` | score 82, positive, low risk, summary set | [PASS] |
| TC-07-U-006 | B-04 | — | `{product_id}` only | defaults score 0 / mixed / medium / '' | [PASS] |
| TC-07-U-007 | B-04 | MockClient 200 | history call | decodes `PriceHistory` | [PASS] |
| TC-07-U-008 | B-04 | MockClient | history('prod-099') | URL `/api/v1/products/prod-099/history` | [PASS] |
| TC-07-U-009 | B-04 | MockClient 404 | history('unknown') | throws `ApiException(notFound)` | [PASS] |
| TC-07-U-010 | B-04 | MockClient throws | history call | throws `ApiException(network)` | [PASS] |
| TC-07-U-011 | B-04 | MockClient 200 | truthScore call | decodes `TruthScore` | [PASS] |
| TC-07-U-012 | B-04 | MockClient | truthScore('prod-007') | URL `.../prod-007/truth-score` | [PASS] |
| TC-07-U-013 | B-04 | — | `MMDealMeter(score:88, avg>cur)` | shows `88/100` + `Exceptional` | [PASS] |
| TC-07-U-014 | B-04 | — | `MMDealMeter(cur<avg)` | "below the 6-month average" text | [PASS] |
| TC-07-U-015 | B-04 | — | `MMDealMeter(avg:0,cur:0)` | "Not enough price history" text | [PASS] |
| TC-07-U-016 | B-04 | — | 3 unsorted stores | rows sorted asc; cheapest = "Best deal" + first | [PASS] |
| TC-07-U-017 | B-04 | — | tap Go-to-Store (row 1) | callback fires for cheapest store | [PASS] |
| TC-07-U-018 | B-04 | — | `MMTruthScore(82,positive,low)` | renders 82, "Positive sentiment", "Low fake-review risk" | [PASS] |
| TC-07-U-019 | B-04 | — | tap "Read more" | summary expands → "Show less" | [PASS] |
| TC-07-U-020 | B-04 | routing MockClient | open prod-001 | all 4 sections + best price 247,900 + Go-to-Best-Store | [PASS] |
| TC-07-U-021 | B-04 | pending Completer | first frame | skeleton placeholders shown | [PASS] |
| TC-07-U-022 | B-04 | MockClient 404 | open 'unknown' | `MMErrorState` + "Try Again" | [PASS] |
