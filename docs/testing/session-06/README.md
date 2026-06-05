# Session 06 — Test Artefacts

**Task under test:** B-03 — Flutter Search Screen (Agent B, client-developer role)
**Tester:** Agent B (Quality Engineer role)
**Date:** 2026-06-05

---

## What was tested

B-03 built the real Search flow (USER_FLOWS Flow 2) on top of the B-01
foundation, pointed at the B-02 mock server:

- **Data layer** — `Product` / `SearchResponse` models decoding the
  `GET /api/v1/search` contract; `SearchRepository` mapping status codes
  (200/400/504/5xx) and transport failures to a typed `ApiException`.
- **Business logic** — Riverpod providers (`searchResultsProvider` family,
  `resultSortProvider`) and the pure `sortResults` ordering function
  (Best Price / Top Rated / Fastest Ship).
- **Widgets** — `MMSearchBar`, `MMProductCard` (total-cost = price + shipping,
  deal badge), `MMSkeletonLoader`, `MMErrorState`.
- **Screens** — Home/Search (query input + trending shortcuts) and Results
  (the four required states: loading, success, empty, error).

Out of scope: Product Detail, Wishlist, Alerts, Savings, Auth (their own
tasks B-04–B-08); pixel-level fidelity to the Stitch samples; the real
backend (still mocked until A-09/B-11).

---

## Layout

```
docs/testing/session-06/
├── README.md                 (this file)
├── unit/
│   ├── test_plan.md
│   ├── test_cases.md
│   └── test_oracle.md
└── integration/
    ├── test_plan.md
    ├── test_cases.md
    └── test_oracle.md
```

**Runnable test code lives in the Flutter package** (Flutter tests must reside
under the app to compile against its `pubspec.yaml`):

| Suite | File |
|-------|------|
| Unit — search models | `apps/mobile/test/unit/search_models_test.dart` |
| Unit — repository (MockClient) | `apps/mobile/test/unit/search_repository_test.dart` |
| Unit — sort logic | `apps/mobile/test/unit/search_providers_test.dart` |
| Unit — MMProductCard | `apps/mobile/test/unit/mm_product_card_test.dart` |
| Unit — MMSearchBar | `apps/mobile/test/unit/mm_search_bar_test.dart` |
| Unit — Results states | `apps/mobile/test/unit/results_screen_test.dart` |
| Shared fixtures | `apps/mobile/test/mocks/search_mock_data.dart` |
| Integration — search flow | `apps/mobile/integration_test/search_flow_test.dart` |

Run unit tests: `cd apps/mobile && flutter test test/unit`
Run integration: `cd apps/mobile && flutter test integration_test/search_flow_test.dart --dart-define=API_BASE_URL=http://10.0.2.2:8081`

---

## Results summary

| Suite | Cases | Status |
|-------|-------|--------|
| Unit | TC-06-U-001 … 025 | **25/25 PASS** (executed; `flutter test`) |
| Integration | TC-06-I-001 … 002 | **PENDING** — no device/emulator this session |

`flutter analyze`: **No issues found.**

**Quality gate:** PENDING PIPELINE RUN — no SonarQube scan has run against this
branch (A-10 CI pipeline / A-11 VPS SonarQube not built). No failing gate to
fix; analyzer clean and all executable cases pass.

**Known gaps:** The two integration cases need a connected device/emulator and a
running mock server; the suite compiles and is ready (`--dart-define` documented
in its header). The B-02 mock is stateless static data, so these tests assert on
the fixed sample offers, not on mutation.
