// Unit tests for NotificationService (B-10): foreground republish, tap routing,
// launch-message handling, and route derivation — driven by a FakePushBackend.
//
// Test artefacts: docs/testing/session-11/unit/.

import 'package:flutter_test/flutter_test.dart';
import 'package:mergemarket/models/push_notification.dart';
import 'package:mergemarket/services/notification_service.dart';

import '../mocks/push_mock_data.dart';

void main() {
  group('B-10 NotificationService', () {
    test('TC-11-U-032: foreground messages are parsed onto the inbound stream',
        () async {
      final backend = FakePushBackend();
      final service = NotificationService(backend: backend);
      addTearDown(() {
        service.dispose();
        backend.close();
      });
      await service.init();

      final first = service.inbound.first;
      backend.emitForeground(kPriceDropData);
      final n = await first;
      expect(n.type, PushType.priceDrop);
      expect(n.productId, 'prod-001');
    });

    test('TC-11-U-033: tapping a notification emits it on the taps stream',
        () async {
      final backend = FakePushBackend();
      final service = NotificationService(backend: backend);
      addTearDown(() {
        service.dispose();
        backend.close();
      });
      await service.init();

      final first = service.taps.first;
      backend.emitOpened(kRestockData);
      final n = await first;
      expect(n.productId, 'prod-002');
    });

    test('TC-11-U-034: a launch message is delivered as a tap on init',
        () async {
      final backend = FakePushBackend(initial: kPriceDropData);
      final service = NotificationService(backend: backend);
      addTearDown(() {
        service.dispose();
        backend.close();
      });

      final first = service.taps.first;
      await service.init();
      final n = await first;
      expect(n.productId, 'prod-001');
      expect(service.token, 'fake-token');
    });

    test('TC-11-U-035: an unroutable tap is dropped (no product to open)',
        () async {
      final backend = FakePushBackend();
      final service = NotificationService(backend: backend);
      addTearDown(() {
        service.dispose();
        backend.close();
      });
      await service.init();

      var tapped = false;
      service.taps.listen((_) => tapped = true);
      backend.emitOpened(kUnknownData);
      await Future<void>.delayed(const Duration(milliseconds: 10));
      expect(tapped, isFalse);
    });

    test('TC-11-U-036: routeFor builds a product deep link, null when unroutable',
        () {
      expect(
        NotificationService.routeFor(
            PushNotification.fromData(kPriceDropData)),
        '/product/prod-001',
      );
      expect(
        NotificationService.routeFor(PushNotification.fromData(kUnknownData)),
        isNull,
      );
    });
  });
}
