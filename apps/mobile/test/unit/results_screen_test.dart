// Widget tests for the Results screen's four states (B-03).
//
// Drives the real provider chain (searchResultsProvider → searchRepository)
// with httpClientProvider overridden by a package:http MockClient, so the
// screen is tested exactly as it runs against the mock server.
//
// Test artefacts: docs/testing/session-06/unit/.

import 'dart:async';
import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:mergemarket/providers/search_providers.dart';
import 'package:mergemarket/screens/results_screen.dart';
import 'package:mergemarket/widgets/mm_error_state.dart';
import 'package:mergemarket/widgets/mm_product_card.dart';
import 'package:mergemarket/widgets/mm_skeleton_loader.dart';

import '../mocks/search_mock_data.dart';

Widget _app(http.Client client, {String query = 'galaxy'}) {
  return ProviderScope(
    overrides: [httpClientProvider.overrideWithValue(client)],
    child: MaterialApp(home: ResultsScreen(query: query)),
  );
}

http.Client _jsonClient(int status, Object body) {
  return MockClient((_) async => http.Response(jsonEncode(body), status,
      headers: {'content-type': 'application/json'}));
}

void main() {
  group('B-03 Results screen state tests', () {
    testWidgets('TC-06-U-021: success shows a card per offer, cheapest first',
        (tester) async {
      await tester.pumpWidget(_app(_jsonClient(200, kSearchSuccessJson)));
      // Bounded pumps, not pumpAndSettle: the skeleton/image shimmer animates
      // forever and would never "settle".
      await tester.pump(); // loading frame
      await tester.pump(const Duration(milliseconds: 100)); // future resolves

      expect(find.byType(MMProductCard), findsNWidgets(3));
      // Best Price default → cheapest total (Kilimall, 247,900) is first.
      final firstCard = tester.widget<MMProductCard>(
          find.byType(MMProductCard).first);
      expect(firstCard.store, 'Kilimall');
      // Header summarises the multi-store result set.
      expect(find.textContaining('3 offers from 3 stores'), findsOneWidget);
    });

    testWidgets('TC-06-U-022: loading shows skeleton placeholders',
        (tester) async {
      final completer = Completer<http.Response>();
      await tester.pumpWidget(_app(MockClient((_) => completer.future)));
      await tester.pump(); // first frame: future still pending

      expect(find.byType(MMSkeletonLoader), findsWidgets);

      completer.complete(http.Response(jsonEncode(kSearchSuccessJson), 200));
      await tester.pump(); // process completion
      await tester.pump(const Duration(milliseconds: 100)); // rebuild with data
      expect(find.byType(MMProductCard), findsNWidgets(3));
    });

    testWidgets('TC-06-U-023: empty results show a no-results message',
        (tester) async {
      await tester.pumpWidget(
          _app(_jsonClient(200, kSearchEmptyJson), query: 'zxqw'));
      await tester.pumpAndSettle();

      expect(find.byType(MMProductCard), findsNothing);
      expect(find.textContaining('No results found'), findsOneWidget);
    });

    testWidgets('TC-06-U-024: transport error shows MMErrorState with retry',
        (tester) async {
      final client = MockClient((_) async => throw http.ClientException('x'));
      await tester.pumpWidget(_app(client));
      await tester.pumpAndSettle();

      expect(find.byType(MMErrorState), findsOneWidget);
      expect(find.text('Try Again'), findsOneWidget);
    });

    testWidgets('TC-06-U-025: empty query prompts for a search', (tester) async {
      await tester.pumpWidget(_app(_jsonClient(200, kSearchSuccessJson), query: ''));
      await tester.pumpAndSettle();

      expect(find.textContaining('Search for a product'), findsOneWidget);
      expect(find.byType(MMProductCard), findsNothing);
    });
  });
}
