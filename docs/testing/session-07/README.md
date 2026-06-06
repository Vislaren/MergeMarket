# Session 07 — Test Artefacts

**Task under test:** B-04 — Flutter Product Detail Screen (Agent B, client-developer role)
**Tester:** Agent B (Quality Engineer role)
**Date:** 2026-06-05

---

## What was tested

B-04 built the Product Detail flow (USER_FLOWS Flow 2/3/4/6) on the B-01
foundation, pointed at the B-02 mock server:

- **Data layer** — `PriceHistory` / `PricePoint` / `TruthScore` models decoding
  the history (`GET /api/v1/products/{id}/history`) and truth-score
  (`.../truth-score`) contracts; `ProductRepository` mapping status codes
  (200/404/5xx) and transport failures to a typed `ApiException`.
- **Business logic** — the `productDetailProvider` family aggregates history +
  store offers (via the existing `SearchRepository`, keyed by the product
  title) + truth score into one `ProductDetail` view model, offers sorted
  cheapest-first.
- **Widgets** — `MMDealMeter` (gauge + band label + comparison), `MMPriceChart`
  (fl_chart line chart), `MMStoreComparisonTable` (sorted, best-deal highlight,
  Go-to-Store), `MMTruthScore` (circular badge + risk chip + expandable summary).
- **Screen** — Product Detail with the four states (loading skeletons, success
  with all sections, 404 not-found error, retry).

Out of scope: real backend (still mocked until A-09/B-11); opening affiliate
links / outbound share / inbound Share-to-Scrape (deferred to B-11 — the buttons
confirm intent via SnackBar this session); pixel-level fidelity to the sample.

---

## Layout

```
docs/testing/session-07/
├── README.md                 (this file)
├── unit/{test_plan,test_cases,test_oracle}.md
└── integration/{test_plan,test_cases,test_oracle}.md
```

**Runnable test code lives in the Flutter package** (Flutter tests must reside
under the app to compile against its `pubspec.yaml`):

| Suite | File |
|-------|------|
| Unit — detail models | `apps/mobile/test/unit/product_models_test.dart` |
| Unit — repository (MockClient) | `apps/mobile/test/unit/product_repository_test.dart` |
| Unit — detail widgets | `apps/mobile/test/unit/product_detail_widgets_test.dart` |
| Unit — detail screen states | `apps/mobile/test/unit/product_detail_screen_test.dart` |
| Shared fixtures | `apps/mobile/test/mocks/product_detail_mock_data.dart` |
| Integration — detail flow | `apps/mobile/integration_test/product_detail_flow_test.dart` |

Run unit tests: `cd apps/mobile && flutter test test/unit`
Run integration: `cd apps/mobile && flutter test integration_test/product_detail_flow_test.dart --dart-define=API_BASE_URL=http://10.0.2.2:8081`

---

## Results summary

| Suite | Cases | Status |
|-------|-------|--------|
| Unit | TC-07-U-001 … 022 | **22/22 PASS** (executed; `flutter test`) |
| Integration | TC-07-I-001 … 002 | **PENDING** — no device/emulator this session |

`flutter analyze`: **No issues found.**

**Quality gate:** PENDING PIPELINE RUN — no SonarQube scan has run against this
branch (A-10 CI / A-11 VPS SonarQube not built). No failing gate to fix;
analyzer clean and all executable cases pass.

**Known gaps:** The two integration cases need a connected device/emulator and a
running mock server; the suite compiles and is ready (`--dart-define`
documented in its header). The B-02 mock is stateless, so detail data reflects
fixed sample offers.
