// Unit tests for AuthController (B-08): session restore, login, register,
// logout — wired to a MockClient and an in-memory FakeSessionStore.
//
// Test artefacts: docs/testing/session-11/unit/.

import 'dart:convert';

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:mergemarket/models/auth_session.dart';
import 'package:mergemarket/providers/auth_providers.dart';
import 'package:mergemarket/providers/search_providers.dart';

import '../mocks/auth_mock_data.dart';

ProviderContainer _container({
  required http.Client client,
  required FakeSessionStore store,
}) {
  final container = ProviderContainer(overrides: [
    httpClientProvider.overrideWithValue(client),
    sessionStoreProvider.overrideWithValue(store),
  ]);
  addTearDown(container.dispose);
  return container;
}

http.Client _ok() => MockClient(
    (_) async => http.Response(jsonEncode(authTokenJson()), 200));

void main() {
  group('B-08 AuthController.build (restore)', () {
    test('TC-11-U-017: restores a valid persisted session as authenticated',
        () async {
      final store = FakeSessionStore(
        initial: AuthSession.fromJson(authTokenJson()),
        initialEmail: 'user@example.com',
      );
      final container = _container(client: _ok(), store: store);

      final state = await container.read(authControllerProvider.future);
      expect(state.isAuthenticated, isTrue);
      expect(state.email, 'user@example.com');
    });

    test('TC-11-U-018: an expired persisted session restores as signed out',
        () async {
      final store = FakeSessionStore(
          initial: AuthSession.fromJson(expiredTokenJson()));
      final container = _container(client: _ok(), store: store);

      final state = await container.read(authControllerProvider.future);
      expect(state.isAuthenticated, isFalse);
    });
  });

  group('B-08 AuthController mutations', () {
    test('TC-11-U-019: login persists the session and becomes authenticated',
        () async {
      final store = FakeSessionStore();
      final container = _container(client: _ok(), store: store);
      await container.read(authControllerProvider.future);

      await container
          .read(authControllerProvider.notifier)
          .login('user@example.com', 'secret123');

      expect(container.read(isAuthenticatedProvider), isTrue);
      expect(await store.read(), isNotNull);
      expect(await store.readEmail(), 'user@example.com');
    });

    test('TC-11-U-020: logout clears storage and becomes signed out', () async {
      final store = FakeSessionStore(
          initial: AuthSession.fromJson(authTokenJson()),
          initialEmail: 'user@example.com');
      final container = _container(client: _ok(), store: store);
      await container.read(authControllerProvider.future);
      expect(container.read(isAuthenticatedProvider), isTrue);

      await container.read(authControllerProvider.notifier).logout();

      expect(container.read(isAuthenticatedProvider), isFalse);
      expect(store.clearCount, 1);
      expect(await store.read(), isNull);
    });
  });
}
