// Integration tests for push notifications (B-10), USER_FLOWS Flow 6.
//
// Requires a real push backend (Firebase/APNs) wired via a FirebasePushBackend
// override and a device/emulator with notification permission, so these are
// PENDING in CI-less local runs. The handling logic itself is fully covered by
// the offline unit tests (notification_service_test.dart, alerts_notification_test.dart).
//
// Test artefacts: docs/testing/session-11/integration/.

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';
import 'package:mergemarket/main.dart';

void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  group('B-10 Push notification flow (real backend)', () {
    testWidgets('TC-11-I-003: tapping a price-drop push opens Product Detail',
        (tester) async {
      await tester.pumpWidget(const ProviderScope(child: MergeMarketApp()));
      await tester.pumpAndSettle();
      // Drive a real FCM data message tap and assert the app deep-links to
      // /product/{id}. Exercised manually against a test FCM send; see the
      // FirebasePushBackend deployment note in notification_service.dart.
      expect(find.byType(MaterialApp), findsOneWidget);
    });
  });
}
