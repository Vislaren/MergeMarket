// Integration tests for B-01 — full app + router wired together, no mocks of
// the navigation stack. Verifies the bottom-nav shell and deep-link routes
// from COMPONENT_LIBRARY §Navigation behave end-to-end.
//
// Run on a device/emulator:  flutter test integration_test/navigation_flow_test.dart
// Test artefacts: docs/testing/session-03/integration/.

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:integration_test/integration_test.dart';

import 'package:mergemarket/main.dart';
import 'package:mergemarket/router/app_router.dart';

void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  group('B-01 Navigation Integration Tests', () {
    testWidgets('TC-03-I-001: bottom nav switches between primary destinations',
        (tester) async {
      await tester.pumpWidget(const ProviderScope(child: MergeMarketApp()));
      await tester.pumpAndSettle();

      // Starts on Search (Home).
      expect(find.text('MergeMarket'), findsOneWidget);

      // Tap the Wishlist tab.
      await tester.tap(find.text('Wishlist'));
      await tester.pumpAndSettle();
      expect(find.widgetWithText(AppBar, 'Wishlist'), findsOneWidget);

      // Tap the Alerts tab.
      await tester.tap(find.text('Alerts'));
      await tester.pumpAndSettle();
      expect(find.widgetWithText(AppBar, 'Alerts'), findsOneWidget);
    });

    testWidgets('TC-03-I-002: deep link /product/:id renders Product Detail',
        (tester) async {
      late final GoRouter router;
      await tester.pumpWidget(
        ProviderScope(
          child: Consumer(
            builder: (context, ref, _) {
              router = ref.watch(routerProvider);
              return const MergeMarketApp();
            },
          ),
        ),
      );
      await tester.pumpAndSettle();

      router.go('/product/abc123');
      await tester.pumpAndSettle();

      expect(find.widgetWithText(AppBar, 'Product Detail'), findsOneWidget);
      expect(find.textContaining('abc123'), findsOneWidget);
    });
  });
}
