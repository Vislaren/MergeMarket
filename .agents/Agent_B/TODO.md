# Agent B — TODO

> Tasks are listed in the order they must be completed.
> Pick the first task that is not blocked.
> When a task is completed, remove it from this file and add it to DONE.md.
> When a task is blocked, move it to BLOCKED.md and pick the next one.

---

## Task Queue

---

### B-01 — Flutter Project Setup
**Description:** Initialise the Flutter project inside `apps/mobile/`
with Riverpod for state management and `go_router` for navigation.
Set up the folder structure: `lib/screens/`, `lib/widgets/`,
`lib/providers/`, `lib/services/`, `lib/models/`, `test/unit/`,
`test/mocks/`, `integration_test/`.

**Depends on:** A-01

**Output:** `apps/mobile/` initialised Flutter project

---

### B-02 — Mock Server for All API Contracts
**Description:** Build a lightweight Go HTTP server inside
`services/mock-server/` that returns hardcoded responses matching every
endpoint in `Agent_A/INSTRUCTIONS.md §4`. This is what all Flutter
development runs against until Agent A's real services are ready.

**Depends on:** A-01

**Output:** `services/mock-server/` (runnable Go server on port 8080)

---

### B-03 — Flutter Search Screen
**Description:** Build the Search screen with query input, concurrent
results list showing all stores, total cost (price + shipping) display,
and a loading/error state. Point to mock server.

**Depends on:** B-01, B-02

**Output:** `lib/screens/search_screen.dart` and related widgets/providers

---

### B-04 — Flutter Product Detail Screen
**Description:** Build the Product Detail screen showing price history
chart (last 6 months), store comparison table, AI Deal Meter rating,
and the "True Cost" breakdown. Include the Share to Scrape integration.

**Depends on:** B-03

**Output:** `lib/screens/product_detail_screen.dart`

---

### B-05 — Flutter Wishlist Screen
**Description:** Build the Wishlist screen with dynamic visual boards,
multi-store tracking per product, and add/remove functionality.

**Depends on:** B-03

**Output:** `lib/screens/wishlist_screen.dart`

---

### B-06 — Flutter Alerts Screen
**Description:** Build the Alerts screen where users set a price threshold
per wishlist product and manage their notification preferences.

**Depends on:** B-05

**Output:** `lib/screens/alerts_screen.dart`

---

### B-07 — Flutter Savings Dashboard Screen
**Description:** Build the Savings Dashboard with a gamified cumulative
savings tracker. Pull savings data from the BFF mock endpoint.

**Depends on:** B-05

**Output:** `lib/screens/savings_dashboard_screen.dart`

---

### B-08 — Flutter Auth Screens
**Description:** Build login and register screens with form validation
and session persistence using secure storage. This task requires the
real Auth service from Agent A (A-08) — use mock until it is ready.

**Depends on:** A-08 (real service), B-01

**Output:** `lib/screens/login_screen.dart`, `lib/screens/register_screen.dart`

---

### B-09 — BFF Service
**Description:** Build the Go Backend-for-Frontend service that sits
between Kong and the Flutter app. It aggregates and shapes data from
the scraper, normalization, history, and auth services. No business
logic — only data shaping and forwarding.

**Depends on:** A-09

**Output:** `services/bff/`

---

### B-10 — Push Notifications (Client Side)
**Description:** Integrate Firebase (Android) and APNs (iOS) into the
Flutter app. Handle incoming price-drop and restock notifications.
Display them correctly from the Alerts screen context.

**Depends on:** B-06

**Output:** `lib/services/notification_service.dart`

---

### B-11 — Integration (Swap Mocks for Real Backend)
**Description:** Replace all mock server references in the Flutter app
and BFF with the real backend services. Run the full end-to-end flow:
search → results → wishlist → alert → notification. Fix any contract
mismatches discovered during integration.

**Depends on:** A-09, B-09

**Output:** Updated service layer, passing integration tests

---

### B-12 — Final Documentation Pass
**Description:** Review all documentation files and ensure they
accurately reflect the final implementation. Update any sections that
drifted during development. Produce a final test coverage report.

**Depends on:** B-11

**Output:** Updated `Documentation.md`, final coverage report
