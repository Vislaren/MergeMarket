// Widget tests for the Set-Alert flow (B-06): the MMSetAlertSheet body and the
// showSetAlertSheet helper that creates the alert via the repository.
//
// Test artefacts: docs/testing/session-09/unit/.

import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:mergemarket/providers/search_providers.dart';
import 'package:mergemarket/widgets/mm_set_alert_sheet.dart';

import '../mocks/alerts_mock_data.dart';

void main() {
  group('B-06 MMSetAlertSheet body', () {
    testWidgets('TC-09-U-013: renders the product, current price, and bounds',
        (tester) async {
      await tester.pumpWidget(const MaterialApp(
        home: Scaffold(
          body: MMSetAlertSheet(
            productTitle: 'Nike Air Max 270',
            currentPrice: 55000,
            averagePrice: 48000,
            currency: 'XAF',
          ),
        ),
      ));
      await tester.pump(const Duration(milliseconds: 100));

      expect(find.text('Set Price Alert'), findsOneWidget);
      expect(find.text('Nike Air Max 270'), findsOneWidget);
      expect(find.textContaining('Current: XAF 55,000'), findsOneWidget);
      expect(find.textContaining('(Min)'), findsOneWidget);
      expect(find.textContaining('(Current)'), findsOneWidget);
      // Default suggestion clamps to the 6-month average (48,000).
      expect(find.text('48000'), findsOneWidget);
    });
  });

  group('B-06 showSetAlertSheet flow', () {
    testWidgets('TC-09-U-014: confirming creates the alert and returns the price',
        (tester) async {
      // Tall surface so the whole sheet (incl. the bottom actions) is visible.
      tester.view.physicalSize = const Size(1000, 2200);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(tester.view.reset);

      late http.Request captured;
      final client = MockClient((req) async {
        captured = req;
        return http.Response(jsonEncode(kAlertCreatedJson), 201);
      });

      double? result;
      await tester.pumpWidget(ProviderScope(
        overrides: [httpClientProvider.overrideWithValue(client)],
        child: MaterialApp(
          home: Consumer(builder: (context, ref, _) {
            return Scaffold(
              body: Center(
                child: ElevatedButton(
                  onPressed: () async {
                    result = await showSetAlertSheet(
                      context,
                      ref,
                      productId: 'prod-001',
                      productTitle: 'Galaxy A54',
                      currentPrice: 250000,
                      averagePrice: 256000,
                      currency: 'XAF',
                    );
                  },
                  child: const Text('open'),
                ),
              ),
            );
          }),
        ),
      ));

      await tester.tap(find.text('open'));
      // Bounded pumps (not pumpAndSettle: the product image shimmer animates
      // forever) to let the modal route animate in.
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 400));
      await tester.tap(find.text('Set Alert'));
      await tester.pump(); // sheet pops
      await tester.pump(const Duration(milliseconds: 400)); // create resolves

      expect(captured.method, 'POST');
      expect(captured.url.path, '/api/v1/alerts');
      expect(result, isNotNull);
    });

    testWidgets('TC-09-U-015: cancelling returns null and creates nothing',
        (tester) async {
      tester.view.physicalSize = const Size(1000, 2200);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(tester.view.reset);

      var posted = false;
      final client = MockClient((req) async {
        posted = true;
        return http.Response('{}', 201);
      });

      double? result = 1; // sentinel to prove it is set to null
      await tester.pumpWidget(ProviderScope(
        overrides: [httpClientProvider.overrideWithValue(client)],
        child: MaterialApp(
          home: Consumer(builder: (context, ref, _) {
            return Scaffold(
              body: Center(
                child: ElevatedButton(
                  onPressed: () async {
                    result = await showSetAlertSheet(
                      context,
                      ref,
                      productId: 'prod-001',
                      productTitle: 'Galaxy A54',
                      currentPrice: 250000,
                      currency: 'XAF',
                    );
                  },
                  child: const Text('open'),
                ),
              ),
            );
          }),
        ),
      ));

      await tester.tap(find.text('open'));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 400));
      await tester.tap(find.text('Cancel'));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 400));

      expect(result, isNull);
      expect(posted, isFalse);
    });
  });
}
