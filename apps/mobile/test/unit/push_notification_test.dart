// Unit tests for the PushNotification model (B-10).
//
// Test artefacts: docs/testing/session-11/unit/.

import 'package:flutter_test/flutter_test.dart';
import 'package:mergemarket/models/push_notification.dart';

import '../mocks/push_mock_data.dart';

void main() {
  group('B-10 PushNotification.fromData', () {
    test('TC-11-U-028: parses a price-drop payload, coercing string numbers',
        () {
      final n = PushNotification.fromData(kPriceDropData);
      expect(n.type, PushType.priceDrop);
      expect(n.productId, 'prod-001');
      expect(n.price, 799000);
      expect(n.threshold, 850000);
      expect(n.store, 'Amazon');
      expect(n.isRoutable, isTrue);
    });

    test('TC-11-U-029: parses a restock payload with optional fields absent',
        () {
      final n = PushNotification.fromData(kRestockData);
      expect(n.type, PushType.restock);
      expect(n.productId, 'prod-002');
      expect(n.price, isNull);
      expect(n.isRoutable, isTrue);
    });

    test('TC-11-U-030: an unknown type with no product id is not routable', () {
      final n = PushNotification.fromData(kUnknownData);
      expect(n.type, PushType.unknown);
      expect(n.isRoutable, isFalse);
    });

    test('TC-11-U-031: missing/empty data decodes to safe values', () {
      final n = PushNotification.fromData(const {});
      expect(n.type, PushType.unknown);
      expect(n.productId, '');
      expect(n.title, '');
      expect(n.isRoutable, isFalse);
    });
  });
}
