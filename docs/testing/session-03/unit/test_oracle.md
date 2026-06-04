## Oracle — B-01 Flutter Project Setup (Unit)

The source of truth for each expectation is the project documentation that
B-01 implements. Tokens come from the Component Library; navigation from its
Navigation Structure; the framework choices from Agent B's INSTRUCTIONS §3.

| Input | Condition | Expected Output | Source of Truth |
|-------|-----------|----------------|----------------|
| `primaryNavy` | token defined | `Color(0xFF1A2B4A)` | COMPONENT_LIBRARY §Design Tokens → Colours |
| `accentRed` | token defined | `Color(0xFFC0392B)` | COMPONENT_LIBRARY §Colours |
| `backgroundLight` | token defined | `Color(0xFFF4F6F9)` | COMPONENT_LIBRARY §Colours |
| `headingLarge` | token defined | size 26, weight w700 | COMPONENT_LIBRARY §Typography |
| `headingSmall` | token defined | size 15, weight w600 | COMPONENT_LIBRARY §Typography |
| `labelBold` | token defined | size 11, w700, letterSpacing 0.5 | COMPONENT_LIBRARY §Typography |
| spacing tokens | grid | `[4,8,16,24,32,48]` | COMPONENT_LIBRARY §Spacing |
| `buildAppTheme()` | built | `useMaterial3 == true` | INSTRUCTIONS §3 (Material 3) |
| `buildAppTheme()` | built | `colorScheme.primary == primaryNavy`, `secondary == accentRed` | DESIGN.md §Colors / COMPONENT_LIBRARY |
| `buildAppTheme()` | built | `scaffoldBackgroundColor == backgroundLight` | DESIGN.md §Elevation (base background) |
| Theme font | runtime | Inter loaded from bundled asset, **no network fetch** | INSTRUCTIONS §3 + $0/offline ethos (ARCHITECTURE §1) |
| `routerProvider` | read from container | returns a `GoRouter` | INSTRUCTIONS §3 (go_router) |
| Router routes | configured | StatefulShellRoute(4 tabs) + results + product/:id + login + register | COMPONENT_LIBRARY §Navigation Structure |
| App launch | pumped in ProviderScope | Home/Search renders at `/`; "MergeMarket" AppBar | COMPONENT_LIBRARY §Navigation (`/` → Home/Search) |
| Bottom nav | on Home | 4 destinations: Search, Wishlist, Alerts, Savings | COMPONENT_LIBRARY §Navigation (bottom nav screens) |
