// Widget tests for the Login screen (B-08).
//
// Drives the real auth provider chain (AuthController → AuthRepository) with
// httpClientProvider + sessionStoreProvider overridden. A lightweight GoRouter
// provides the /, /login and /register destinations so navigation can be
// asserted without the app's redirect guard.
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
import 'package:mergemarket/screens/login_screen.dart';
import 'package:mergemarket/widgets/auth_form_widgets.dart';

import '../mocks/auth_mock_data.dart';

Widget _app(http.Client client, FakeSessionStore store) {
  final router = GoRouter(
    initialLocation: '/login',
    routes: [
      GoRoute(path: '/login', builder: (_, _) => const LoginScreen()),
      GoRoute(
          path: '/',
          builder: (_, _) => const Scaffold(body: Text('HOME-STUB'))),
      GoRoute(
          path: '/register',
          builder: (_, _) => const Scaffold(body: Text('REGISTER-STUB'))),
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
  group('B-08 Login screen', () {
    testWidgets('TC-11-U-021: empty submit shows field validation errors',
        (tester) async {
      await tester.pumpWidget(_app(_status(200, authTokenJson()), FakeSessionStore()));
      await tester.pump();

      await tester.tap(find.widgetWithText(ElevatedButton, 'Log In'));
      await tester.pump();

      expect(find.text('Email is required.'), findsOneWidget);
      expect(find.text('Password is required.'), findsOneWidget);
      expect(find.text('HOME-STUB'), findsNothing);
    });

    testWidgets('TC-11-U-022: valid credentials log in and navigate home',
        (tester) async {
      final store = FakeSessionStore();
      await tester.pumpWidget(_app(_status(200, authTokenJson()), store));
      await tester.pump();

      await tester.enterText(
          find.byType(TextFormField).at(0), 'user@example.com');
      await tester.enterText(find.byType(TextFormField).at(1), 'secret123');
      await tester.tap(find.widgetWithText(ElevatedButton, 'Log In'));
      await tester.pump(); // start submit
      await tester.pump(const Duration(milliseconds: 200)); // resolve + navigate

      expect(find.text('HOME-STUB'), findsOneWidget);
      expect(await store.read(), isNotNull);
    });

    testWidgets('TC-11-U-023: invalid credentials show an inline error banner',
        (tester) async {
      await tester.pumpWidget(
          _app(_status(401, kInvalidCredentialsJson), FakeSessionStore()));
      await tester.pump();

      await tester.enterText(
          find.byType(TextFormField).at(0), 'user@example.com');
      await tester.enterText(
          find.byType(TextFormField).at(1), 'wrongpassword');
      await tester.tap(find.widgetWithText(ElevatedButton, 'Log In'));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 200));

      expect(find.byType(AuthErrorBanner), findsOneWidget);
      expect(find.text('HOME-STUB'), findsNothing);
    });

    testWidgets('TC-11-U-024: password visibility toggle reveals the password',
        (tester) async {
      await tester.pumpWidget(
          _app(_status(200, authTokenJson()), FakeSessionStore()));
      await tester.pump();

      bool passwordObscured() => tester
          .widget<EditableText>(find.descendant(
            of: find.byType(TextFormField).at(1),
            matching: find.byType(EditableText),
          ))
          .obscureText;

      expect(passwordObscured(), isTrue);
      await tester.tap(find.byTooltip('Show password'));
      await tester.pump();
      expect(passwordObscured(), isFalse);
    });
  });
}
