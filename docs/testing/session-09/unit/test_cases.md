# Test Cases — Unit — Session 09 — B-06 Alerts

_(U = Unit. All executed via `flutter test`; all PASS.)_

| ID | Task | Preconditions | Input | Expected | Actual |
|----|------|---------------|-------|----------|--------|
| TC-09-U-001 | B-06 | — | `kAlertsJson` | decodes 2 alerts; first active (230000), second inactive | [PASS] |
| TC-09-U-002 | B-06 | — | `kAlertsEmptyJson` | empty alerts list | [PASS] |
| TC-09-U-003 | B-06 | — | `{alert_id}` only | defaults threshold 0 / inactive / '' currency | [PASS] |
| TC-09-U-004 | B-06 | MockClient 200 | list() | decodes 2 alerts | [PASS] |
| TC-09-U-005 | B-06 | MockClient throws | list() | `ApiException(network)` | [PASS] |
| TC-09-U-006 | B-06 | MockClient 201 | create(prod-001,230000,XAF) | POST `/api/v1/alerts`, body has all 3 fields | [PASS] |
| TC-09-U-007 | B-06 | MockClient 400 | create('',0,XAF) | `ApiException(badRequest)` | [PASS] |
| TC-09-U-008 | B-06 | MockClient 204 | remove('al-002') | DELETE `/api/v1/alerts/al-002` | [PASS] |
| TC-09-U-009 | B-06 | MockClient 404 | remove('unknown') | `ApiException(notFound)` | [PASS] |
| TC-09-U-010 | B-06 | — | `MMAlertCard(active)` | title + threshold 230,000 + "Active" | [PASS] |
| TC-09-U-011 | B-06 | — | `MMAlertCard(inactive)` | "Inactive" chip | [PASS] |
| TC-09-U-012 | B-06 | — | swipe card left | `onDelete()` fires | [PASS] |
| TC-09-U-013 | B-06 | — | sheet(cur 55000, avg 48000) | title, "Current: XAF 55,000", Min/Current, default 48000 | [PASS] |
| TC-09-U-014 | B-06 | MockClient 201, tall surface | open sheet → Set Alert | POST `/api/v1/alerts`; helper returns the price | [PASS] |
| TC-09-U-015 | B-06 | tall surface | open sheet → Cancel | helper returns null; no POST | [PASS] |
| TC-09-U-016 | B-06 | MockClient 200 | open Alerts | 2 `MMAlertCard`, "Tracking 2 items", Active + Inactive | [PASS] |
| TC-09-U-017 | B-06 | MockClient 200 empty | open Alerts | "No price alerts yet" prompt | [PASS] |
| TC-09-U-018 | B-06 | MockClient throws | open Alerts | `MMErrorState` + "Try Again" | [PASS] |
