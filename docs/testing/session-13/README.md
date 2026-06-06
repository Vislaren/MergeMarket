# Session 13 — B-11 Integration (Swap Mocks for Real Backend)

**Agent:** B (Client Developer + Quality Engineer)
**Task:** B-11 — replace mock-server references in the Flutter app and BFF with
the real backend, run the search → results → wishlist → alert → notification
flow, and fix any contract mismatches found.
**Branch:** `co-luiza` (Agent B's permanent branch — the only tree with the full
B-01…B-10 app + BFF).

---

## What B-11 delivered (code)

The substance of "swap mocks for the real backend" is the plumbing the
auth-free B-02 mock never needed but Kong (A-09) requires:

1. **Authenticated HTTP client** — `lib/services/authenticated_client.dart`
   wraps the shared `http.Client` and attaches `Authorization: Bearer <token>`
   to every protected request. Wired in via `authedHttpClientProvider`; the
   five protected repositories (search, product, wishlist, alerts, savings) now
   use it, while the auth repository keeps the bare client.
2. **Refresh-on-401 interceptor** (deferred from B-08) — a 401 on a protected
   route triggers a single coalesced session refresh and replays the original
   request with the new token; a failed refresh lets the 401 surface so the
   router guard routes to login. `AuthController.refreshSession()` performs the
   refresh + persist.
3. **BFF JWT forwarding** — the BFF aggregate (`/products/{id}/detail`) now
   forwards the caller's `Authorization` header to its upstream history /
   search / truth-score calls, so the BFF works behind Kong's JWT gate.
4. **Backend target config** — `AppConfig.apiBaseUrl` default corrected from the
   blocked `:8080` to the mock's assigned `:8089` (PORTS_README); Kong (`:8088`)
   is selected via `--dart-define=API_BASE_URL=...`. The BFF already targets the
   upstream via `BFF_UPSTREAM_URL` (default mock `:8089`, Kong/real in prod).

## What B-11 could **not** complete (and why)

- **Live end-to-end run is PENDING.** Docker is not running locally and Agent
  A's services are not deployed, so the real stack cannot be brought up. The
  integration suite is written and self-skips without a reachable backend.
- **Most real services do not exist yet.** See `CONTRACT_AUDIT.md`: only Auth
  and Price History are real HTTP endpoints. Search, Truth-Score, Wishlist,
  Alerts and Savings are mock-only, so those routes cannot be swapped to a real
  backend regardless of the client work. This is an Agent A gap, not a client
  defect.
- **Native integrations still stubbed** (carried from B-04/B-10): affiliate-link
  opening (`url_launcher`), Share-to-Scrape (`receive_sharing_intent`), and the
  real `FirebasePushBackend` (`firebase_messaging`) need native plugins + a
  device and remain release-time follow-ups.

---

## Results

| Suite | Result |
|-------|--------|
| Flutter unit (full app) | **151/151 PASS** (141 prior + 10 new B-11) |
| `flutter analyze` | clean |
| BFF Go (`go test ./...`) | PASS (incl. new auth-forward test) |
| Integration (live E2E) | **PENDING** — no Docker / no deployed backend |

Executable tests (canonical, run in CI):
- `apps/mobile/test/unit/authenticated_client_test.dart` (TC-13-U-001…010)
- `services/bff/internal/server/server_test.go` →
  `TestProductDetailForwardsAuthHeaderUpstream` (TC-13-U-011)

Artefacts in this folder mirror those and add the test plan / cases / oracle and
the contract audit.

**Quality gate:** SonarQube (`http://95.111.228.35:9000`) is reachable this
session and the token authenticates, but the instance has **zero projects** — no
scan has ever run, so there is no gate to pass/fail. Status: PENDING PIPELINE RUN.
