# Session 03 — Test Artefacts

**Task under test:** B-01 — Flutter Project Setup (Agent B, client-developer role)
**Tester:** Agent B (Quality Engineer role)
**Date:** 2026-06-04

---

## What was tested

B-01 initialised the Flutter project in `apps/mobile/` with Riverpod state
management and `go_router` navigation, the design-token theme from
`project_docs/ui/COMPONENT_LIBRARY.md`, and placeholder screens wired to every
route in `COMPONENT_LIBRARY.md §Navigation`.

Because B-01 is *project scaffolding*, the suite verifies the **foundation**
contracts rather than feature behaviour:

- Design tokens (colours, typography, spacing) match the documented values.
- The global `ThemeData` is Material 3 with the navy/red scheme.
- `go_router` exposes every documented route and boots at `/`.
- The app launches inside a `ProviderScope` and renders the bottom-nav shell
  with the four primary destinations.
- (Integration) Bottom-nav tab switching and `/product/:id` deep-linking work
  end-to-end through the real router.

Feature UI (search, product detail, wishlist, alerts, savings, auth) is **out
of scope** — those land in their own tasks (B-03–B-08) and will be tested then.

---

## Layout

```
docs/testing/session-03/
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
| Unit — theme tokens | `apps/mobile/test/unit/theme_tokens_test.dart` |
| Unit — router/boot | `apps/mobile/test/unit/router_test.dart` |
| Integration — navigation | `apps/mobile/integration_test/navigation_flow_test.dart` |

Run unit tests: `cd apps/mobile && flutter test`
Run integration tests: `cd apps/mobile && flutter test integration_test/navigation_flow_test.dart`

---

## Results summary

| Suite | Cases | Status |
|-------|-------|--------|
| Unit | TC-03-U-001 … 008 | see `unit/test_cases.md` |
| Integration | TC-03-I-001 … 002 | see `integration/test_cases.md` |

**Quality gate:** PENDING PIPELINE RUN — SonarQube instance at
`http://95.111.228.35:9000` is UP (v26.5.0) but no project has been scanned yet
(A-10 CI pipeline / A-11 VPS SonarQube not built). No failing gate to fix.
