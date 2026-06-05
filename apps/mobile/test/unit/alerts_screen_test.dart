// Widget tests for the Alerts screen's states (B-06).
//
// Drives the real provider chain (alertsProvider → AlertsRepository) with
// httpClientProvider overridden by a MockClient.
//
// Test artefacts: docs/testing/session-09/unit/.

import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:mergemarket/providers/search_providers.dart';
import 'package:mergemarket/screens/alerts_screen.dart';
import 'package:mergemarket/widgets/mm_alert_card.dart';
import 'package:mergemarket/widgets/mm_error_state.dart';

import '../mocks/alerts_mock_data.dart';

Widget _app(http.Client client) {
  return ProviderScope(
    overrides: [httpClientProvider.overrideWithValue(client)],
    child: const MaterialApp(home: AlertsScreen()),
  );
}

http.Client _json(int status, Object body) => MockClient(
    (_) async => http.Response(jsonEncode(body), status,
        headers: {'content-type': 'application/json'}));

void main() {
  group('B-06 Alerts screen states', () {
    testWidgets('TC-09-U-016: success lists alert cards with a summary',
        (tester) async {
      await tester.pumpWidget(_app(_json(200, kAlertsJson)));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 100));

      expect(find.byType(MMAlertCard), findsNWidgets(2));
      expect(find.textContaining('Tracking 2 items'), findsOneWidget);
      expect(find.text('Active'), findsOneWidget);
      expect(find.text('Inactive'), findsOneWidget);
    });

    testWidgets('TC-09-U-017: empty alerts show the empty prompt',
        (tester) async {
      await tester.pumpWidget(_app(_json(200, kAlertsEmptyJson)));
      await tester.pumpAndSettle();

      expect(find.byType(MMAlertCard), findsNothing);
      expect(find.text('No price alerts yet'), findsOneWidget);
    });

    testWidgets('TC-09-U-018: load failure shows MMErrorState with retry',
        (tester) async {
      await tester.pumpWidget(
          _app(MockClient((_) async => throw http.ClientException('x'))));
      await tester.pumpAndSettle();

      expect(find.byType(MMErrorState), findsOneWidget);
      expect(find.text('Try Again'), findsOneWidget);
    });
  });
}
