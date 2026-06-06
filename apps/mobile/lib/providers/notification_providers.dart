import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/push_notification.dart';
import '../services/notification_service.dart';

/// The push transport backend. Defaults to [NoopPushBackend] so the app runs
/// without any native push configuration; override it with a real
/// `FirebasePushBackend` in `main()` once Firebase/APNs are configured, or with
/// a fake in tests.
final pushBackendProvider = Provider<PushMessagingBackend>((ref) {
  return const NoopPushBackend();
});

/// The app-wide [NotificationService], wired to [pushBackendProvider] and
/// initialised eagerly so it begins consuming messages as soon as it is read.
final notificationServiceProvider = Provider<NotificationService>((ref) {
  final service = NotificationService(backend: ref.watch(pushBackendProvider));
  // Fire-and-forget: subscribes to the backend and drains any launch message.
  unawaited(service.init());
  ref.onDispose(service.dispose);
  return service;
});

/// Foreground notifications, for in-app display (e.g. the Alerts banner).
final incomingNotificationsProvider = StreamProvider<PushNotification>((ref) {
  return ref.watch(notificationServiceProvider).inbound;
});

/// Notification taps, for deep-linking to the relevant product (Flow 6).
final notificationTapsProvider = StreamProvider<PushNotification>((ref) {
  return ref.watch(notificationServiceProvider).taps;
});
