# Test Cases — Integration — Session 08 — B-05 Wishlist

_(I = Integration. Require a device/emulator + running mock server.)_

| ID | Task | Preconditions | Input | Expected | Actual |
|----|------|---------------|-------|----------|--------|
| TC-08-I-001 | B-05 | device + mock server up | open Wishlist tab | `MMWishlistBoard` shows tracked products with "tracking" counts | [PENDING] |
| TC-08-I-002 | B-05 | device + mock server up | Wishlist → tap a bell | "Set Price Alert" sheet opens | [PENDING] |

**Why PENDING:** No connected device/emulator in the dev environment this
session. Run with:

```
cd apps/mobile && flutter test integration_test/wishlist_flow_test.dart \
  --dart-define=API_BASE_URL=http://10.0.2.2:8081
```
