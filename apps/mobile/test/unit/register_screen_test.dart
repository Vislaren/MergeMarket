// Widget tests for the Register screen (B-08).
//
// Drives the real auth provider chain with httpClientProvider +
// sessionStoreProvider overridden, and a lightweight GoRouter for navigation.
//
// Test artefacts: docs/testing/session-11/unit/.

import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:mergemarket/providers/auth_providers.dart';
import 'package:mergemarket/providers/search_providers.dart';
import 'package:mergemarket/screens/register_screen.dart';
import 'package:mergemarket/widgets/auth_form_widgets.dart';

import '../mocks/auth_mock_data.dart';

Widget _app(http.Client client, FakeSessionStore store) {
  final router = GoRouter(
    initialLocation: '/register',
    routes: [
      GoRoute(path: '/register', builder: (_, _) => const RegisterScreen()),
      GoRoute(
          path: '/',
          builder: (_, _) => const Scaffold(body: Text('HOME-STUB'))),
      GoRoute(
          path: '/login',
          builder: (_, _) => const Scaffold(body: Text('LOGIN-STUB'))),
    ],
  );
  return ProviderScope(
    overrides: [
      httpClientProvider.overrideWithValue(client),
      sessionStoreProvider.overrideWithValue(store),
    ],
    child: MaterialApp.router(routerConfig: router),
  );
}

http.Client _status(int status, Object body) => MockClient((_) async =>
    http.Response(jsonEncode(body), status,
        headers: {'content-type': 'application/json'}));

void main() {
  group('B-08 Register screen', () {
    testWidgets('TC-11-U-025: mismatched confirm password blocks submit',
        (tester) async {
      await tester.pumpWidget(
          _app(_status(201, authTokenJson()), FakeSessionStore()));
      await tester.pump();

      await tester.enterText(
          find.byType(TextFormField).at(0), 'new@example.com');
      await tester.enterText(find.byType(TextFormField).at(1), 'secret123');
      await tester.enterText(find.byType(TextFormField).at(2), 'different1');
      await tester.tap(find.widgetWithText(ElevatedButton, 'Create Account'));
      await tester.pump();

      expect(find.text('Passwords do not match.'), findsOneWidget);
      expect(find.text('HOME-STUB'), findsNothing);
    });

    testWidgets('TC-11-U-026: valid registration creates the account and navigates home',
        (tester) async {
      final store = FakeSessionStore();
      await tester.pumpWidget(_app(_status(201, authTokenJson()), store));
      await tester.pump();

      await tester.enterText(
          find.byType(TextFormField).at(0), 'new@example.com');
      await tester.enterText(find.byType(TextFormField).at(1), 'secret123');
      await tester.enterText(find.byType(TextFormField).at(2), 'secret123');
      await tester.tap(find.widgetWithText(ElevatedButton, 'Create Account'));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 200));

      expect(find.text('HOME-STUB'), findsOneWidget);
      expect(await store.read(), isNotNull);
    });

    testWidgets('TC-11-U-027: a taken email shows the conflict error banner',
        (tester) async {
      await tester.pumpWidget(
          _app(_status(409, kEmailExistsJson), FakeSessionStore()));
      await tester.pump();

      await tester.enterText(
          find.byType(TextFormField).at(0), 'taken@mergemarket.app');
      await tester.enterText(find.byType(TextFormField).at(1), 'secret123');
      await tester.enterText(find.byType(TextFormField).at(2), 'secret123');
      await tester.tap(find.widgetWithText(ElevatedButton, 'Create Account'));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 200));

      expect(find.byType(AuthErrorBanner), findsOneWidget);
      expect(find.text('HOME-STUB'), findsNothing);
    });
  });
}
