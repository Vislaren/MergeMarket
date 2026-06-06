# Test Cases — Integration — Session 07 — B-04 Product Detail

_(I = Integration. Require a device/emulator + running mock server.)_

| ID | Task | Preconditions | Input | Expected | Actual |
|----|------|---------------|-------|----------|--------|
| TC-07-I-001 | B-04 | device + mock server up | search "galaxy" → tap first card | Product Detail shows MMDealMeter, MMPriceChart, MMStoreComparisonTable, MMTruthScore, "Go to Best Store" | [PENDING] |
| TC-07-I-002 | B-04 | device + mock server up | on detail, tap favourite (Add to Wishlist) | a SnackBar confirms the outcome (prod-001 → "already in wishlist") | [PENDING] |

**Why PENDING:** No connected device/emulator in the dev environment this
session. Suite compiles under the integration harness and is ready; run with:

```
cd apps/mobile && flutter test integration_test/product_detail_flow_test.dart \
  --dart-define=API_BASE_URL=http://10.0.2.2:8081
```
