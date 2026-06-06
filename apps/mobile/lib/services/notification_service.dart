import 'dart:async';

import '../models/push_notification.dart';

/// Transport abstraction over the platform push SDK (B-10).
///
/// The app depends on this interface, not on `firebase_messaging` directly, so
/// the notification handling, parsing, and routing are fully unit-testable
/// without a Firebase project (none exists in the $0 bootstrap) or a device.
///
/// Two implementations exist:
/// * [NoopPushBackend] — the default; never emits. Lets the app run and its
///   tests pass with no native push configuration.
/// * `FirebasePushBackend` (added at deployment, see the class doc below) —
///   wires Firebase Cloud Messaging (Android) and APNs (iOS).
abstract interface class PushMessagingBackend {
  /// Requests notification permission and returns the device token, or null if
  /// denied / unavailable.
  Future<String?> requestPermissionAndToken();

  /// Notifications received while the app is in the foreground. Each event is
  /// the raw FCM/APNs `data` map.
  Stream<Map<String, dynamic>> get onMessage;

  /// Fired when the user taps a notification that opened (or resumed) the app.
  Stream<Map<String, dynamic>> get onMessageOpenedApp;

  /// The notification that launched the app from a terminated state, if any —
  /// consumed once on startup to deep-link (Flow 6).
  Future<Map<String, dynamic>?> initialMessage();
}

/// A backend that does nothing — the safe default when push is not configured.
///
/// Streams are empty broadcast streams, so wiring it up changes no behaviour;
/// the app simply never surfaces a push until a real backend is provided.
class NoopPushBackend implements PushMessagingBackend {
  const NoopPushBackend();

  @override
  Future<String?> requestPermissionAndToken() async => null;

  @override
  Stream<Map<String, dynamic>> get onMessage => const Stream.empty();

  @override
  Stream<Map<String, dynamic>> get onMessageOpenedApp => const Stream.empty();

  @override
  Future<Map<String, dynamic>?> initialMessage() async => null;
}

/// Handles incoming price-drop / restock notifications (USER_FLOWS Flow 6).
///
/// It owns no transport — it sits on top of a [PushMessagingBackend] and:
/// * parses every raw payload into a typed [PushNotification];
/// * republishes foreground messages on [inbound] (the Alerts screen shows an
///   in-app banner for these);
/// * republishes taps (background-opened + launch) on [taps] (the app deep-links
///   these to the product detail screen).
class NotificationService {
  NotificationService({required PushMessagingBackend backend})
      : _backend = backend;

  final PushMessagingBackend _backend;

  final _inbound = StreamController<PushNotification>.broadcast();
  final _taps = StreamController<PushNotification>.broadcast();
  final _subscriptions = <StreamSubscription<dynamic>>[];

  bool _initialised = false;

  /// Foreground notifications, for in-app display (e.g. an Alerts banner).
  Stream<PushNotification> get inbound => _inbound.stream;

  /// Notification taps, for deep-linking to the relevant product.
  Stream<PushNotification> get taps => _taps.stream;

  /// The most recent device token obtained from the backend, if any.
  String? token;

  /// Subscribes to the backend and consumes any launch notification. Safe to
  /// call once; subsequent calls are no-ops.
  Future<void> init() async {
    if (_initialised) return;
    _initialised = true;

    token = await _backend.requestPermissionAndToken();

    _subscriptions.add(_backend.onMessage.listen((data) {
      _inbound.add(PushNotification.fromData(data));
    }));
    _subscriptions.add(_backend.onMessageOpenedApp.listen((data) {
      _emitTap(PushNotification.fromData(data));
    }));

    final launch = await _backend.initialMessage();
    if (launch != null) {
      _emitTap(PushNotification.fromData(launch));
    }
  }

  /// Builds the in-app deep-link route for a tapped [notification], or null when
  /// the payload is not routable (no product id / unknown type).
  static String? routeFor(PushNotification notification) {
    if (!notification.isRoutable) return null;
    return '/product/${notification.productId}';
  }

  void _emitTap(PushNotification notification) {
    if (notification.isRoutable) _taps.add(notification);
  }

  /// Cancels subscriptions and closes the streams.
  void dispose() {
    for (final sub in _subscriptions) {
      sub.cancel();
    }
    _inbound.close();
    _taps.close();
  }
}

// ---------------------------------------------------------------------------
// Deployment note — wiring the real backend (B-11 / release):
//
// Add `firebase_messaging` + `firebase_core`, drop in the Android
// `google-services.json` / iOS `GoogleService-Info.plist`, then implement:
//
//   class FirebasePushBackend implements PushMessagingBackend {
//     final _m = FirebaseMessaging.instance;
//     Future<String?> requestPermissionAndToken() async {
//       final settings = await _m.requestPermission();          // iOS/APNs prompt
//       if (settings.authorizationStatus == AuthorizationStatus.denied) return null;
//       return _m.getToken();                                   // FCM token
//     }
//     Stream<Map<String,dynamic>> get onMessage =>
//         FirebaseMessaging.onMessage.map((m) => m.data);
//     Stream<Map<String,dynamic>> get onMessageOpenedApp =>
//         FirebaseMessaging.onMessageOpenedApp.map((m) => m.data);
//     Future<Map<String,dynamic>?> initialMessage() async =>
//         (await _m.getInitialMessage())?.data;
//   }
//
// Then override `pushBackendProvider` with `FirebasePushBackend()` in main()
// after `Firebase.initializeApp()`. Nothing else in this file changes.
// ---------------------------------------------------------------------------
