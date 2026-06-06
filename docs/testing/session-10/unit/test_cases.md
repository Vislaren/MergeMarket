# Test Cases - Unit - Session 10 - B-07 Savings Dashboard

_(U = Unit. All executed via `flutter test`; all PASS.)_

| ID | Task | Preconditions | Input | Expected | Actual |
|----|------|---------------|-------|----------|--------|
| TC-10-U-001 | B-07 | - | `kSavingsJson` | decodes total 33500, XAF, and 3 transactions | [PASS] |
| TC-10-U-002 | B-07 | - | `kSavingsJson` | derives level 1, progress 0.67, remaining 16500 | [PASS] |
| TC-10-U-003 | B-07 | - | `kSavingsTopLevelJson` | clamps to level 10, progress 1, remaining 0 | [PASS] |
| TC-10-U-004 | B-07 | - | malformed fields | safe zero-value fallbacks; bad date becomes epoch | [PASS] |
| TC-10-U-005 | B-07 | MockClient 200 | getSavings() | decodes savings summary with 3 transactions | [PASS] |
| TC-10-U-006 | B-07 | MockClient 200 | getSavings() | GET `/api/v1/savings`, `Accept: application/json` | [PASS] |
| TC-10-U-007 | B-07 | MockClient 500 | getSavings() | `ApiException(server)` with contract message | [PASS] |
| TC-10-U-008 | B-07 | MockClient throws | getSavings() | `ApiException(network)` | [PASS] |
| TC-10-U-009 | B-07 | - | `MMSavingsCard(level 1)` | total, level, remaining hint, share callback | [PASS] |
| TC-10-U-010 | B-07 | - | `MMSavingsCard(level 10)` | "Top level reached" completion copy | [PASS] |
| TC-10-U-011 | B-07 | MockClient 200 | open Savings | card, total, recent savings, momentum panel | [PASS] |
| TC-10-U-012 | B-07 | MockClient 200 empty | open Savings | "No savings yet" prompt + Wishlist CTA | [PASS] |
| TC-10-U-013 | B-07 | MockClient throws | open Savings | `MMErrorState` + "Try Again" | [PASS] |
| TC-10-U-014 | B-07 | MockClient 200 | tap share | SnackBar includes shared savings total | [PASS] |
