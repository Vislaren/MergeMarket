// Widget tests for the Savings Dashboard screen states (B-07).
//
// Drives the real provider chain (savingsProvider -> SavingsRepository) with
// httpClientProvider overridden by a MockClient.
//
// Test artefacts: docs/testing/session-10/unit/.

import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:mergemarket/providers/search_providers.dart';
import 'package:mergemarket/screens/product_detail_screen.dart';
import 'package:mergemarket/screens/savings_dashboard_screen.dart';
import 'package:mergemarket/widgets/mm_error_state.dart';
import 'package:mergemarket/widgets/mm_savings_card.dart';

import '../mocks/savings_mock_data.dart';

Widget _app(http.Client client) {
  final router = GoRouter(
    routes: [
      GoRoute(
        path: '/',
        builder: (context, state) => const SavingsDashboardScreen(),
      ),
      GoRoute(
        path: '/product/:id',
        builder: (context, state) =>
            ProductDetailScreen(productId: state.pathParameters['id']!),
      ),
      GoRoute(
        path: '/wishlist',
        builder: (context, state) => const Scaffold(body: Text('Wishlist')),
      ),
    ],
  );

  return ProviderScope(
    overrides: [httpClientProvider.overrideWithValue(client)],
    child: MaterialApp.router(routerConfig: router),
  );
}

http.Client _json(int status, Object body) => MockClient(
  (_) async => http.Response(
    jsonEncode(body),
    status,
    headers: {'content-type': 'application/json'},
  ),
);

void main() {
  group('B-07 Savings Dashboard screen states', () {
    testWidgets('TC-10-U-011: success shows savings card and recent events', (
      tester,
    ) async {
      await tester.pumpWidget(_app(_json(200, kSavingsJson)));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 100));

      expect(find.byType(MMSavingsCard), findsOneWidget);
      expect(find.text('XAF 33,500'), findsOneWidget);
      expect(find.text('Recent savings'), findsOneWidget);
      expect(find.text('Samsung Galaxy A54 128GB'), findsOneWidget);
      expect(find.textContaining('3 savings recorded'), findsOneWidget);
    });

    testWidgets('TC-10-U-012: empty savings show the empty prompt', (
      tester,
    ) async {
      await tester.pumpWidget(_app(_json(200, kSavingsEmptyJson)));
      await tester.pumpAndSettle();

      expect(find.byType(MMSavingsCard), findsNothing);
      expect(find.text('No savings yet'), findsOneWidget);
      expect(find.text('Open Wishlist'), findsOneWidget);
    });

    testWidgets('TC-10-U-013: load failure shows MMErrorState with retry', (
      tester,
    ) async {
      await tester.pumpWidget(
        _app(MockClient((_) async => throw http.ClientException('x'))),
      );
      await tester.pumpAndSettle();

      expect(find.byType(MMErrorState), findsOneWidget);
      expect(find.text('Try Again'), findsOneWidget);
    });

    testWidgets('TC-10-U-014: share action reports the savings total', (
      tester,
    ) async {
      await tester.pumpWidget(_app(_json(200, kSavingsJson)));
      await tester.pumpAndSettle();

      await tester.tap(find.byIcon(Icons.ios_share_rounded));
      await tester.pumpAndSettle();

      expect(find.textContaining('Shared XAF 33,500'), findsOneWidget);
    });
  });
}
