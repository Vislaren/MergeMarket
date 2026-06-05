# Test Cases — Unit — Session 03 — B-01 Flutter Project Setup

Suite files:
- `apps/mobile/test/unit/theme_tokens_test.dart`
- `apps/mobile/test/unit/router_test.dart`

Executed: `cd apps/mobile && flutter test` → **All tests passed (8/8)**.
`flutter analyze` → **No issues found.**

---

### TC-03-U-001: Colour tokens match COMPONENT_LIBRARY hex values

| Field | Value |
|-------|-------|
| Task reference | B-01 |
| Type | Unit |
| Preconditions | `lib/theme/colours.dart` compiled |
| Input | The eight documented colour constants |
| Steps | 1. Import tokens. 2. Compare each to the hex in COMPONENT_LIBRARY. |
| Expected result | Every constant equals its documented `Color(0x…)` |
| Actual result | [PASS] |
| Notes | primaryNavy, accentRed, backgroundLight, surfaceWhite, borderGrey, successGreen, warningAmber, dealGold |

---

### TC-03-U-002: Typography scale matches documented sizes/weights

| Field | Value |
|-------|-------|
| Task reference | B-01 |
| Type | Unit |
| Preconditions | `lib/theme/typography.dart` compiled |
| Input | headingLarge/Medium/Small, bodyRegular/Small, labelBold |
| Steps | 1. Assert each `fontSize` and `fontWeight`; assert `labelBold.letterSpacing`. |
| Expected result | 26/w700, 18/w700, 15/w600, 14/w400, 12/w400, 11/w700 (ls 0.5) |
| Actual result | [PASS] |
| Notes | — |

---

### TC-03-U-003: Spacing tokens follow the 8pt grid

| Field | Value |
|-------|-------|
| Task reference | B-01 |
| Type | Unit |
| Preconditions | `lib/theme/spacing.dart` compiled |
| Input | xs, sm, md, lg, xl, xxl |
| Steps | 1. Compare list to `[4, 8, 16, 24, 32, 48]`. |
| Expected result | Exact match |
| Actual result | [PASS] |
| Notes | — |

---

### TC-03-U-004: buildAppTheme is Material 3 with the navy/red scheme

| Field | Value |
|-------|-------|
| Task reference | B-01 |
| Type | Unit |
| Preconditions | `lib/theme/app_theme.dart` compiled |
| Input | `buildAppTheme()` result |
| Steps | 1. Build theme. 2. Assert useMaterial3, scaffold bg, primary, secondary, appBar bg. |
| Expected result | useMaterial3=true; scaffoldBg=backgroundLight; primary=navy; secondary=red; appBar bg=navy |
| Actual result | [PASS] |
| Notes | Previously failed when the theme fetched Inter via google_fonts at runtime; fixed by bundling Inter as an asset (no network/binding dependency). |

---

### TC-03-U-005: Route path constants match COMPONENT_LIBRARY

| Field | Value |
|-------|-------|
| Task reference | B-01 |
| Type | Unit |
| Preconditions | `lib/router/app_router.dart` compiled |
| Input | `Routes` constants |
| Steps | 1. Assert each constant equals the documented path. |
| Expected result | `/`, `/results`, `/product`, `/wishlist`, `/alerts`, `/savings`, `/login`, `/register` |
| Actual result | [PASS] |
| Notes | `/product` is the base; full route is `/product/:id` |

---

### TC-03-U-006: routerProvider builds a GoRouter with all routes

| Field | Value |
|-------|-------|
| Task reference | B-01 |
| Type | Unit |
| Preconditions | A `ProviderContainer` |
| Input | `container.read(routerProvider)` |
| Steps | 1. Read provider. 2. Assert it is a GoRouter. 3. Assert 5 top-level routes. |
| Expected result | GoRouter with 1 StatefulShellRoute + 4 top-level routes = 5 |
| Actual result | [PASS] |
| Notes | Initial-location verification is covered behaviourally by TC-03-U-007 (Home renders on boot), since `currentConfiguration` is only populated once the router is mounted in a widget. |

---

### TC-03-U-007: App boots in a ProviderScope and shows Search

| Field | Value |
|-------|-------|
| Task reference | B-01 |
| Type | Unit (widget) |
| Preconditions | Full app compiled |
| Input | `ProviderScope(child: MergeMarketApp())` |
| Steps | 1. pumpWidget. 2. pumpAndSettle. 3. Find MaterialApp, "Search", "MergeMarket". |
| Expected result | MaterialApp present; Home/Search placeholder + "MergeMarket" AppBar title shown |
| Actual result | [PASS] |
| Notes | Confirms ProviderScope + MaterialApp.router + initial route `/` |

---

### TC-03-U-008: Bottom nav shows the four primary destinations

| Field | Value |
|-------|-------|
| Task reference | B-01 |
| Type | Unit (widget) |
| Preconditions | Full app compiled |
| Input | Booted app |
| Steps | 1. pumpAndSettle. 2. Find NavigationBar. 3. Assert 4 destinations. |
| Expected result | One NavigationBar with Search/Wishlist/Alerts/Savings (4) |
| Actual result | [PASS] |
| Notes | Matches COMPONENT_LIBRARY: bottom nav on Home, Wishlist, Alerts, Savings |
