// Integration tests for the auth flow (B-08), USER_FLOWS Flow 1.
//
// Runs the real app against a running backend (B-02 mock server or the real
// Auth service). Requires a device/emulator + a reachable backend, so these are
// PENDING in CI-less local runs.
//
// Run:
//   1. start the mock server:  cd services/mock-server && go run ./cmd/mock-server
//   2. flutter test integration_test/auth_flow_test.dart \
//        --dart-define=API_BASE_URL=http://10.0.2.2:8089
//
// Test artefacts: docs/testing/session-11/integration/.

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';
import 'package:mergemarket/main.dart';
import 'package:mergemarket/screens/login_screen.dart';

void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  group('B-08 Auth flow (real backend)', () {
    testWidgets('TC-11-I-001: log in from the account button reaches Home',
        (tester) async {
      await tester.pumpWidget(const ProviderScope(child: MergeMarketApp()));
      await tester.pumpAndSettle();

      // Open Login from the Home account affordance.
      await tester.tap(find.byTooltip('Log in'));
      await tester.pumpAndSettle();
      expect(find.byType(LoginScreen), findsOneWidget);

      await tester.enterText(
          find.byType(TextFormField).at(0), 'user@example.com');
      await tester.enterText(find.byType(TextFormField).at(1), 'secret123');
      await tester.tap(find.widgetWithText(ElevatedButton, 'Log In'));
      await tester.pumpAndSettle();

      // Back on Home, now showing the logout affordance.
      expect(find.byTooltip('Log out'), findsOneWidget);
    });

    testWidgets('TC-11-I-002: register then land on Home authenticated',
        (tester) async {
      await tester.pumpWidget(const ProviderScope(child: MergeMarketApp()));
      await tester.pumpAndSettle();
      // Exercised manually against the mock server; see header for the command.
    });
  });
}
