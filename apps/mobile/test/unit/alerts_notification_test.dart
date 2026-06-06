// Widget test for the Alerts screen's in-app push banner (B-10).
//
// Overrides pushBackendProvider with a FakePushBackend and emits a foreground
// price-drop; the Alerts screen should surface it as a SnackBar with a "View"
// action. With the default NoopPushBackend nothing emits (existing B-06 tests
// stay green).
//
// Test artefacts: docs/testing/session-11/unit/.

import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:mergemarket/providers/notification_providers.dart';
import 'package:mergemarket/providers/search_providers.dart';
import 'package:mergemarket/screens/alerts_screen.dart';

import '../mocks/alerts_mock_data.dart';
import '../mocks/push_mock_data.dart';

void main() {
  testWidgets('TC-11-U-037: a foreground push shows an Alerts banner with View',
      (tester) async {
    final backend = FakePushBackend();
    addTearDown(backend.close);

    final client = MockClient((_) async => http.Response(
        jsonEncode(kAlertsEmptyJson), 200,
        headers: {'content-type': 'application/json'}));

    // A minimal router so the SnackBar's "View" action has somewhere to go.
    final router = GoRouter(routes: [
      GoRoute(path: '/', builder: (_, _) => const AlertsScreen()),
      GoRoute(
          path: '/product/:id',
          builder: (_, _) => const Scaffold(body: Text('PRODUCT-STUB'))),
    ]);

    await tester.pumpWidget(ProviderScope(
      overrides: [
        httpClientProvider.overrideWithValue(client),
        pushBackendProvider.overrideWithValue(backend),
      ],
      child: MaterialApp.router(routerConfig: router),
    ));
    await tester.pumpAndSettle();

    backend.emitForeground(kPriceDropData);
    await tester.pump(); // deliver stream event
    await tester.pump(); // build the SnackBar

    expect(
        find.textContaining('dropped to XAF 799,000'), findsOneWidget);
    expect(find.text('View'), findsOneWidget);
  });
}
