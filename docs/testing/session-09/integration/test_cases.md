# Test Cases — Integration — Session 09 — B-06 Alerts

_(I = Integration. Require a device/emulator + running mock server.)_

| ID | Task | Preconditions | Input | Expected | Actual |
|----|------|---------------|-------|----------|--------|
| TC-09-I-001 | B-06 | device + mock server up | open Alerts tab | `MMAlertCard`s listed + "Tracking N items" summary | [PENDING] |
| TC-09-I-002 | B-06 | device + mock server up | Wishlist → bell → Set Alert | lands on the Alerts tab ("Price Alerts") | [PENDING] |

**Why PENDING:** No connected device/emulator in the dev environment this
session. Run with:

```
cd apps/mobile && flutter test integration_test/alerts_flow_test.dart \
  --dart-define=API_BASE_URL=http://10.0.2.2:8081
```
