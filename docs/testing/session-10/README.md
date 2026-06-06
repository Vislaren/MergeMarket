# Session 10 - Test Artefacts

**Task under test:** B-07 - Flutter Savings Dashboard Screen (Agent B, client-developer role)
**Tester:** Agent B (Quality Engineer role)
**Date:** 2026-06-06

---

## What was tested

B-07 built the Savings Dashboard flow (USER_FLOWS Flow 7) on the B-01
foundation, pointed at the B-02 mock server:

- **Data layer** - `SavingsSummary` / `SavingsTransaction` models decoding
  `GET /api/v1/savings`; `SavingsRepository.getSavings` for 200, non-200, and
  transport-error mapping to `ApiException`.
- **Business logic** - `savingsProvider`, backed by the shared
  `httpClientProvider`.
- **Widget** - `MMSavingsCard` with a large total saved amount, savings level,
  progress-to-next-level bar, and share action.
- **Screen** - Savings Dashboard with the four states: loading skeletons,
  success dashboard + recent savings list, empty prompt, and error with retry.

Out of scope: real backend persistence (B-11) and native share-sheet
integration. The share action currently reports through a SnackBar.

---

## Layout

```
docs/testing/session-10/
|-- README.md
|-- unit/{test_plan,test_cases,test_oracle}.md
`-- integration/{test_plan,test_cases,test_oracle}.md
```

**Runnable test code lives in the Flutter package:**

| Suite | File |
|-------|------|
| Unit - savings models | `apps/mobile/test/unit/savings_models_test.dart` |
| Unit - repository (MockClient) | `apps/mobile/test/unit/savings_repository_test.dart` |
| Unit - MMSavingsCard | `apps/mobile/test/unit/mm_savings_card_test.dart` |
| Unit - savings screen states | `apps/mobile/test/unit/savings_dashboard_screen_test.dart` |
| Shared fixtures | `apps/mobile/test/mocks/savings_mock_data.dart` |
| Integration - savings flow | `apps/mobile/integration_test/savings_flow_test.dart` |

Run unit tests: `cd apps/mobile && flutter test test/unit`
Run integration: `cd apps/mobile && flutter test integration_test/savings_flow_test.dart --dart-define=API_BASE_URL=http://10.0.2.2:8081`

---

## Results summary

| Suite | Cases | Status |
|-------|-------|--------|
| Unit | TC-10-U-001 ... 014 | **14/14 PASS** (executed; `flutter test`) |
| Integration | TC-10-I-001 ... 002 | **PENDING** - no device/emulator this session |

Full Flutter suite: **104/104 PASS**.
`flutter analyze`: **No issues found.**

**Quality gate:** PENDING PIPELINE RUN - no SonarQube scan has run against this
branch. Analyzer clean; all executable cases pass.

**Known gaps:** Integration cases need a device/emulator + running mock server.
