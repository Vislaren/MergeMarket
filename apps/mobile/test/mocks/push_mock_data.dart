// Controllable fake PushMessagingBackend + sample payloads for the B-10 tests.

import 'dart:async';

import 'package:mergemarket/services/notification_service.dart';

/// A [PushMessagingBackend] whose streams and launch message are driven by the
/// test, so notification handling can be exercised without Firebase or a device.
class FakePushBackend implements PushMessagingBackend {
  FakePushBackend({this.tokenValue = 'fake-token', this.initial});

  final String? tokenValue;

  /// The notification that "launched" the app, returned once by [initialMessage].
  Map<String, dynamic>? initial;

  final _onMessage = StreamController<Map<String, dynamic>>.broadcast();
  final _onOpened = StreamController<Map<String, dynamic>>.broadcast();

  /// Simulates a foreground notification arriving.
  void emitForeground(Map<String, dynamic> data) => _onMessage.add(data);

  /// Simulates the user tapping a notification that opens the app.
  void emitOpened(Map<String, dynamic> data) => _onOpened.add(data);

  @override
  Future<String?> requestPermissionAndToken() async => tokenValue;

  @override
  Stream<Map<String, dynamic>> get onMessage => _onMessage.stream;

  @override
  Stream<Map<String, dynamic>> get onMessageOpenedApp => _onOpened.stream;

  @override
  Future<Map<String, dynamic>?> initialMessage() async => initial;

  void close() {
    _onMessage.close();
    _onOpened.close();
  }
}

/// A price-drop payload as delivered by FCM/APNs (data values are strings).
const Map<String, dynamic> kPriceDropData = {
  'type': 'price_drop',
  'product_id': 'prod-001',
  'title': 'iPhone 15',
  'body': 'iPhone 15 dropped to XAF 799,000 on Amazon — below your alert',
  'price': '799000',
  'store': 'Amazon',
  'threshold': '850000',
};

/// A restock payload.
const Map<String, dynamic> kRestockData = {
  'type': 'restock',
  'product_id': 'prod-002',
  'title': 'PS5',
  'body': 'PS5 is back in stock at Jumia',
};

/// An unrecognised payload that must not be routed.
const Map<String, dynamic> kUnknownData = {
  'type': 'newsletter',
  'product_id': '',
  'title': 'Weekly deals',
  'body': 'Check out this week’s deals',
};
