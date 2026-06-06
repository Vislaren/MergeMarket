# Session 09 — Test Artefacts

**Task under test:** B-06 — Flutter Alerts Screen (Agent B, client-developer role)
**Tester:** Agent B (Quality Engineer role)
**Date:** 2026-06-05

---

## What was tested

B-06 built the Alerts flow (USER_FLOWS Flow 5) on the B-01 foundation, pointed
at the B-02 mock server, and wired the Set-Alert sheet behind the wishlist bell:

- **Data layer** — `Alert` / `AlertList` models decoding `GET /api/v1/alerts`;
  `AlertsRepository` for list (200), create (201/400), and remove (204/404),
  mapping status codes + transport failures to a typed `ApiException`.
- **Business logic** — `alertsProvider` (the list) and `AlertsActions`
  (create/remove → invalidate the list).
- **Widgets** — `MMAlertCard` (title, bold threshold, active/inactive chip,
  swipe-to-delete) and `MMSetAlertSheet` + `showSetAlertSheet` (current price,
  amount field synced to a slider, average hint, Set Alert / Cancel; on confirm
  it creates the alert).
- **Screen** — Alerts with the four states (loading skeletons, success list +
  "Tracking N items" summary, empty prompt, error with retry).
- **Cross-task wiring** — the wishlist bell now opens the Set-Alert sheet and,
  on success, navigates to the Alerts tab (B-05 had a SnackBar stub).

Out of scope: push notifications / delivery (B-10); real backend (B-11). The
B-02 mock is stateless, so create/remove don't change subsequent GETs.

---

## Layout

```
docs/testing/session-09/
├── README.md                 (this file)
├── unit/{test_plan,test_cases,test_oracle}.md
└── integration/{test_plan,test_cases,test_oracle}.md
```

**Runnable test code lives in the Flutter package:**

| Suite | File |
|-------|------|
| Unit — alert models | `apps/mobile/test/unit/alerts_models_test.dart` |
| Unit — repository (MockClient) | `apps/mobile/test/unit/alerts_repository_test.dart` |
| Unit — MMAlertCard | `apps/mobile/test/unit/mm_alert_card_test.dart` |
| Unit — Set-Alert sheet + flow | `apps/mobile/test/unit/set_alert_sheet_test.dart` |
| Unit — alerts screen states | `apps/mobile/test/unit/alerts_screen_test.dart` |
| Shared fixtures | `apps/mobile/test/mocks/alerts_mock_data.dart` |
| Integration — alerts flow | `apps/mobile/integration_test/alerts_flow_test.dart` |

Run unit tests: `cd apps/mobile && flutter test test/unit`
Run integration: `cd apps/mobile && flutter test integration_test/alerts_flow_test.dart --dart-define=API_BASE_URL=http://10.0.2.2:8081`

---

## Results summary

| Suite | Cases | Status |
|-------|-------|--------|
| Unit | TC-09-U-001 … 018 | **18/18 PASS** (executed; `flutter test`) |
| Integration | TC-09-I-001 … 002 | **PENDING** — no device/emulator this session |

`flutter analyze`: **No issues found.**

**Quality gate:** PENDING PIPELINE RUN — no SonarQube scan has run against this
branch. Analyzer clean; all executable cases pass.

**Known gaps:** Integration cases need a device/emulator + running mock server.
Mutations are not persisted by the stateless mock.
