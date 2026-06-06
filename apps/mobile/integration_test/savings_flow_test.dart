// Integration test for the B-07 Savings Dashboard flow (USER_FLOWS Flow 7).
//
// Drives the REAL app end-to-end against a running B-02 mock server: open the
// Savings tab, load the live savings endpoint, and tap a recent saving.
//
// Requires:
//   1. A connected device/emulator (or `-d chrome`).
//   2. The mock server running and reachable at the configured base URL:
//        cd services/mock-server && MOCK_SERVER_PORT=8081 go run ./cmd/mock-server
//   3. The base URL passed in (Android emulator reaches the host at 10.0.2.2):
//        flutter test integration_test/savings_flow_test.dart \
//          --dart-define=API_BASE_URL=http://10.0.2.2:8081
//
// Status: PENDING - no device/emulator available in the dev environment this
// session. The suite compiles and is ready to run once a device + mock server
// are present.

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';
import 'package:mergemarket/main.dart';
import 'package:mergemarket/widgets/mm_savings_card.dart';

void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  group('B-07 Savings Dashboard integration tests', () {
    testWidgets('TC-10-I-001: the Savings tab loads the savings dashboard', (
      tester,
    ) async {
      await tester.pumpWidget(const ProviderScope(child: MergeMarketApp()));
      await tester.pumpAndSettle();

      await tester.tap(find.byIcon(Icons.savings_outlined).first);
      await tester.pumpAndSettle(const Duration(seconds: 5));

      expect(find.byType(MMSavingsCard), findsOneWidget);
      expect(find.text('Recent savings'), findsOneWidget);
      expect(find.textContaining('savings recorded'), findsOneWidget);
    });

    testWidgets('TC-10-I-002: tapping a savings event opens Product Detail', (
      tester,
    ) async {
      await tester.pumpWidget(const ProviderScope(child: MergeMarketApp()));
      await tester.pumpAndSettle();

      await tester.tap(find.byIcon(Icons.savings_outlined).first);
      await tester.pumpAndSettle(const Duration(seconds: 5));

      await tester.tap(find.text('Samsung Galaxy A54 128GB').first);
      await tester.pumpAndSettle(const Duration(seconds: 5));

      expect(find.textContaining('Galaxy'), findsWidgets);
    });
  });
}
