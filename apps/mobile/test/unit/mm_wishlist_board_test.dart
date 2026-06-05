// Widget tests for MMWishlistBoard (B-05).
//
// Test artefacts: docs/testing/session-08/unit/.

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mergemarket/models/wishlist.dart';
import 'package:mergemarket/widgets/mm_wishlist_board.dart';

import '../mocks/wishlist_mock_data.dart';

List<WishlistItem> _items() => Wishlist.fromJson(kWishlistJson).items;

Widget _wrap({
  void Function(String)? onTap,
  void Function(String)? onRemove,
  void Function(String)? onSetAlert,
}) {
  return MaterialApp(
    home: Scaffold(
      body: MMWishlistBoard(
        items: _items(),
        onTap: onTap ?? (_) {},
        onRemove: onRemove ?? (_) {},
        onSetAlert: onSetAlert ?? (_) {},
      ),
    ),
  );
}

void main() {
  group('B-05 MMWishlistBoard', () {
    testWidgets('TC-08-U-011: renders a tile per item with store count',
        (tester) async {
      await tester.pumpWidget(_wrap());
      await tester.pump(const Duration(milliseconds: 100));

      expect(find.text('Samsung Galaxy A54 128GB'), findsOneWidget);
      expect(find.text('Anker PowerCore 20000mAh'), findsOneWidget);
      expect(find.text('2 stores tracking'), findsOneWidget);
      expect(find.text('1 store tracking'), findsOneWidget);
    });

    testWidgets('TC-08-U-012: bell triggers onSetAlert with the product id',
        (tester) async {
      String? alerted;
      await tester.pumpWidget(_wrap(onSetAlert: (id) => alerted = id));
      await tester.pump(const Duration(milliseconds: 100));

      await tester.tap(find.byIcon(Icons.notifications_none_rounded).first);
      expect(alerted, 'prod-001');
    });

    testWidgets('TC-08-U-013: tapping a tile triggers onTap with product id',
        (tester) async {
      String? tapped;
      await tester.pumpWidget(_wrap(onTap: (id) => tapped = id));
      await tester.pump(const Duration(milliseconds: 100));

      await tester.tap(find.text('Anker PowerCore 20000mAh'));
      expect(tapped, 'prod-014');
    });

    testWidgets('TC-08-U-014: swipe-to-remove triggers onRemove',
        (tester) async {
      String? removed;
      await tester.pumpWidget(_wrap(onRemove: (id) => removed = id));
      await tester.pump(const Duration(milliseconds: 100));

      await tester.drag(
          find.text('Samsung Galaxy A54 128GB'), const Offset(-500, 0));
      // Bounded pumps (not pumpAndSettle: the shimmer placeholder animates
      // forever) to let the dismiss + resize animations run to completion.
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 400));
      await tester.pump(const Duration(milliseconds: 400));
      expect(removed, 'wl-001');
    });
  });
}
