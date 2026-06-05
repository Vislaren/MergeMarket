// Integration test for the B-06 Alerts flow (USER_FLOWS Flow 5).
//
// Drives the REAL app end-to-end against a running B-02 mock server: open the
// Alerts tab → see the active alerts list from the live HTTP repository; and
// set an alert from the wishlist bell. No mocks.
//
// Requires:
//   1. A connected device/emulator (or `-d chrome`).
//   2. The mock server running and reachable at the configured base URL:
//        cd services/mock-server && MOCK_SERVER_PORT=8081 go run ./cmd/mock-server
//   3. The base URL passed in (Android emulator reaches the host at 10.0.2.2):
//        flutter test integration_test/alerts_flow_test.dart \
//          --dart-define=API_BASE_URL=http://10.0.2.2:8081
//
// Status: PENDING — no device/emulator available in the dev environment this
// session. The suite compiles and is ready to run once a device + mock server
// are present.

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';
import 'package:mergemarket/main.dart';
import 'package:mergemarket/widgets/mm_alert_card.dart';

void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  group('B-06 Alerts flow integration tests', () {
    testWidgets('TC-09-I-001: the Alerts tab lists the active alerts',
        (tester) async {
      await tester.pumpWidget(const ProviderScope(child: MergeMarketApp()));
      await tester.pumpAndSettle();

      // Navigate to the Alerts tab via the bottom navigation bar.
      await tester.tap(find.byIcon(Icons.notifications_none_rounded).first);
      await tester.pumpAndSettle(const Duration(seconds: 5));

      expect(find.byType(MMAlertCard), findsWidgets);
      expect(find.textContaining('Tracking'), findsOneWidget);
    });

    testWidgets('TC-09-I-002: setting an alert from the wishlist lands on Alerts',
        (tester) async {
      await tester.pumpWidget(const ProviderScope(child: MergeMarketApp()));
      await tester.pumpAndSettle();

      await tester.tap(find.byIcon(Icons.favorite_border_rounded).first);
      await tester.pumpAndSettle(const Duration(seconds: 5));

      await tester.tap(find.byIcon(Icons.notifications_none_rounded).first);
      await tester.pumpAndSettle();
      await tester.tap(find.text('Set Alert'));
      await tester.pumpAndSettle(const Duration(seconds: 5));

      // After confirming, the app navigates to the Alerts tab.
      expect(find.text('Price Alerts'), findsOneWidget);
    });
  });
}
