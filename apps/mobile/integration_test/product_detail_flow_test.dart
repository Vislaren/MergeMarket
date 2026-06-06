// Integration test for the B-04 Product Detail flow (USER_FLOWS Flow 2).
//
// Drives the REAL app end-to-end against a running B-02 mock server: search →
// tap a result → Product Detail renders the price history, deal meter, store
// comparison, and truth score from the live HTTP repositories. No mocks.
//
// Requires:
//   1. A connected device/emulator (or `-d chrome`).
//   2. The mock server running and reachable at the configured base URL:
//        cd services/mock-server && MOCK_SERVER_PORT=8081 go run ./cmd/mock-server
//   3. The base URL passed in (Android emulator reaches the host at 10.0.2.2):
//        flutter test integration_test/product_detail_flow_test.dart \
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
import 'package:mergemarket/widgets/mm_deal_meter.dart';
import 'package:mergemarket/widgets/mm_price_chart.dart';
import 'package:mergemarket/widgets/mm_product_card.dart';
import 'package:mergemarket/widgets/mm_search_bar.dart';
import 'package:mergemarket/widgets/mm_store_comparison_table.dart';
import 'package:mergemarket/widgets/mm_truth_score.dart';

void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  group('B-04 Product Detail flow integration tests', () {
    testWidgets('TC-07-I-001: opening a product shows all detail sections',
        (tester) async {
      await tester.pumpWidget(const ProviderScope(child: MergeMarketApp()));
      await tester.pumpAndSettle();

      await tester.enterText(find.byType(MMSearchBar), 'galaxy');
      await tester.testTextInput.receiveAction(TextInputAction.search);
      await tester.pumpAndSettle(const Duration(seconds: 5));

      await tester.tap(find.byType(MMProductCard).first);
      await tester.pumpAndSettle(const Duration(seconds: 5));

      expect(find.byType(MMDealMeter), findsOneWidget);
      expect(find.byType(MMPriceChart), findsOneWidget);
      expect(find.byType(MMStoreComparisonTable), findsOneWidget);
      expect(find.byType(MMTruthScore), findsOneWidget);
      expect(find.textContaining('Go to Best Store'), findsOneWidget);
    });

    testWidgets('TC-07-I-002: Add to Wishlist confirms via SnackBar',
        (tester) async {
      await tester.pumpWidget(const ProviderScope(child: MergeMarketApp()));
      await tester.pumpAndSettle();

      await tester.enterText(find.byType(MMSearchBar), 'galaxy');
      await tester.testTextInput.receiveAction(TextInputAction.search);
      await tester.pumpAndSettle(const Duration(seconds: 5));
      await tester.tap(find.byType(MMProductCard).first);
      await tester.pumpAndSettle(const Duration(seconds: 5));

      await tester.tap(find.byIcon(Icons.favorite_border_rounded));
      await tester.pumpAndSettle();

      // prod-001 is the mock's "already in wishlist" sentinel → 409 message.
      expect(find.byType(SnackBar), findsOneWidget);
    });
  });
}
