# Session 11 — Test Artefacts (Agent B dev tasks B-08, B-09, B-10)

Session 11 delivered three of Agent B's own development tasks. This folder holds
the QE artefacts (plan, cases, oracle) for all three; the executable suites live
beside the code they test:

| Task | Code | Unit suite | Integration suite |
|------|------|-----------|-------------------|
| B-08 Flutter Auth | `apps/mobile/lib/{models/auth_session,services/{auth_repository,session_store},providers/auth_providers,screens/{login,register}_screen,utils/validators}.dart` | `apps/mobile/test/unit/{auth_session,validators,auth_repository,auth_controller,login_screen,register_screen}_test.dart` | `apps/mobile/integration_test/auth_flow_test.dart` |
| B-09 Go BFF | `services/bff/` | `services/bff/internal/{config,server}/*_test.go` | (covered by `server_test.go` via httptest upstream) |
| B-10 Push (client) | `apps/mobile/lib/{models/push_notification,services/notification_service,providers/notification_providers}.dart` | `apps/mobile/test/unit/{push_notification,notification_service,alerts_notification}_test.dart` | `apps/mobile/integration_test/notification_flow_test.dart` |

## Run

```bash
# Flutter (B-08 + B-10) — 141/141 unit/widget cases
cd apps/mobile && flutter analyze && flutter test

# Go BFF (B-09) — config 100%, server 91%
cd services/bff && go test ./... -cover
```

## Result summary

- **Flutter:** `flutter analyze` clean; **141/141** unit/widget tests pass
  (104 prior + 37 new: B-08 TC-11-U-001…027, B-10 TC-11-U-028…037).
- **Go BFF:** `go build` / `go vet` / `gofmt -l` clean; all package tests pass
  (config 100%, server 91% — also exercises the `upstream` client).
- **Integration:** 3 cases authored, PENDING (need a device/emulator + running
  backend, and for B-10 a Firebase/APNs backend).
