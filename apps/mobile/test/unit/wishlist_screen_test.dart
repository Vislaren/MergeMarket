// Widget tests for the Wishlist screen's states (B-05).
//
// Drives the real provider chain (wishlistProvider → WishlistRepository) with
// httpClientProvider overridden by a MockClient.
//
// Test artefacts: docs/testing/session-08/unit/.

import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:mergemarket/providers/search_providers.dart';
import 'package:mergemarket/screens/wishlist_screen.dart';
import 'package:mergemarket/widgets/mm_error_state.dart';
import 'package:mergemarket/widgets/mm_wishlist_board.dart';

import '../mocks/wishlist_mock_data.dart';

Widget _app(http.Client client) {
  return ProviderScope(
    overrides: [httpClientProvider.overrideWithValue(client)],
    child: const MaterialApp(home: WishlistScreen()),
  );
}

http.Client _json(int status, Object body) => MockClient(
    (_) async => http.Response(jsonEncode(body), status,
        headers: {'content-type': 'application/json'}));

void main() {
  group('B-05 Wishlist screen states', () {
    testWidgets('TC-08-U-015: success shows the board with tracked items',
        (tester) async {
      await tester.pumpWidget(_app(_json(200, kWishlistJson)));
      await tester.pump(); // loading
      await tester.pump(const Duration(milliseconds: 100)); // data

      expect(find.byType(MMWishlistBoard), findsOneWidget);
      expect(find.text('Samsung Galaxy A54 128GB'), findsOneWidget);
    });

    testWidgets('TC-08-U-016: empty wishlist shows the empty prompt',
        (tester) async {
      await tester.pumpWidget(_app(_json(200, kWishlistEmptyJson)));
      await tester.pumpAndSettle();

      expect(find.byType(MMWishlistBoard), findsNothing);
      expect(find.text('Your wishlist is empty'), findsOneWidget);
      expect(find.text('Start Searching'), findsOneWidget);
    });

    testWidgets('TC-08-U-017: load failure shows MMErrorState with retry',
        (tester) async {
      await tester.pumpWidget(
          _app(MockClient((_) async => throw http.ClientException('x'))));
      await tester.pumpAndSettle();

      expect(find.byType(MMErrorState), findsOneWidget);
      expect(find.text('Try Again'), findsOneWidget);
    });
  });
}
