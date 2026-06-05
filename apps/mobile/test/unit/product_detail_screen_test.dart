// Widget tests for the Product Detail screen's states (B-04).
//
// Drives the real provider chain (productDetailProvider → ProductRepository +
// SearchRepository) with httpClientProvider overridden by a routing MockClient,
// so the screen is tested exactly as it runs against the mock server. The
// provider fetches history, then searches by title and fetches the truth score.
//
// Test artefacts: docs/testing/session-07/unit/.

import 'dart:async';
import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:mergemarket/providers/search_providers.dart';
import 'package:mergemarket/screens/product_detail_screen.dart';
import 'package:mergemarket/widgets/mm_deal_meter.dart';
import 'package:mergemarket/widgets/mm_error_state.dart';
import 'package:mergemarket/widgets/mm_price_chart.dart';
import 'package:mergemarket/widgets/mm_skeleton_loader.dart';
import 'package:mergemarket/widgets/mm_store_comparison_table.dart';
import 'package:mergemarket/widgets/mm_truth_score.dart';

import '../mocks/product_detail_mock_data.dart';

Widget _app(http.Client client, {String id = 'prod-001'}) {
  return ProviderScope(
    overrides: [httpClientProvider.overrideWithValue(client)],
    child: MaterialApp(home: ProductDetailScreen(productId: id)),
  );
}

/// Routes by path to the right fixture, mirroring the mock server.
http.Client _routingClient() {
  return MockClient((req) async {
    final path = req.url.path;
    Object body;
    if (path.endsWith('/history')) {
      body = kHistoryJson;
    } else if (path.endsWith('/truth-score')) {
      body = kTruthScoreJson;
    } else if (path == '/api/v1/search') {
      body = kProductSearchJson;
    } else {
      return http.Response('{}', 404);
    }
    return http.Response(jsonEncode(body), 200,
        headers: {'content-type': 'application/json'});
  });
}

void main() {
  group('B-04 Product Detail screen states', () {
    testWidgets('TC-07-U-020: success renders all detail sections',
        (tester) async {
      await tester.pumpWidget(_app(_routingClient()));
      await tester.pump(); // loading frame
      await tester.pump(const Duration(milliseconds: 200)); // futures resolve
      await tester.pump(const Duration(milliseconds: 200)); // chart settle

      expect(find.text('Samsung Galaxy A54 128GB'), findsWidgets);
      expect(find.byType(MMDealMeter), findsOneWidget);
      expect(find.byType(MMPriceChart), findsOneWidget);
      expect(find.byType(MMStoreComparisonTable), findsOneWidget);
      expect(find.byType(MMTruthScore), findsOneWidget);
      // Headline best price is the cheapest total (Kilimall, 247,900).
      expect(find.textContaining('247,900'), findsWidgets);
      expect(find.textContaining('Go to Best Store'), findsOneWidget);
    });

    testWidgets('TC-07-U-021: loading shows skeleton placeholders',
        (tester) async {
      final completer = Completer<http.Response>();
      await tester.pumpWidget(_app(MockClient((_) => completer.future)));
      await tester.pump(); // first frame: history still pending

      expect(find.byType(MMSkeletonLoader), findsWidgets);
      completer.complete(http.Response('{}', 404)); // let it finish
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 100));
    });

    testWidgets('TC-07-U-022: 404 on history shows MMErrorState with retry',
        (tester) async {
      final client = MockClient((_) async => http.Response(
          jsonEncode(kNotFoundJson), 404,
          headers: {'content-type': 'application/json'}));
      await tester.pumpWidget(_app(client, id: 'unknown'));
      await tester.pumpAndSettle();

      expect(find.byType(MMErrorState), findsOneWidget);
      expect(find.text('Try Again'), findsOneWidget);
    });
  });
}
