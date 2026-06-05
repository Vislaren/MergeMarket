// Unit tests for the wishlist data models (B-05).
//
// Test artefacts: docs/testing/session-08/unit/.

import 'package:flutter_test/flutter_test.dart';
import 'package:mergemarket/models/wishlist.dart';

import '../mocks/wishlist_mock_data.dart';

void main() {
  group('B-05 Wishlist model', () {
    test('TC-08-U-001: decodes the wishlist contract', () {
      final wl = Wishlist.fromJson(kWishlistJson);
      expect(wl.items, hasLength(2));
      expect(wl.items.first.wishlistId, 'wl-001');
      expect(wl.items.first.title, 'Samsung Galaxy A54 128GB');
      expect(wl.items.first.stores, hasLength(2));
    });

    test('TC-08-U-002: storeCount and bestTotalCost are derived', () {
      final item = Wishlist.fromJson(kWishlistJson).items.first;
      expect(item.storeCount, 2);
      // Cheapest of 250000 / 247900.
      expect(item.bestTotalCost, 247900);
    });

    test('TC-08-U-003: empty wishlist decodes to no items', () {
      final wl = Wishlist.fromJson(kWishlistEmptyJson);
      expect(wl.items, isEmpty);
    });

    test('TC-08-U-004: an item with no stores is safe', () {
      final item = WishlistItem.fromJson({
        'wishlist_id': 'wl-x',
        'product_id': 'p',
        'title': 't',
      });
      expect(item.stores, isEmpty);
      expect(item.storeCount, 0);
      expect(item.bestTotalCost, isNull);
    });
  });
}
