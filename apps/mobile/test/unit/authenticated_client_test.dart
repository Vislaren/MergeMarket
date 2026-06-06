// Unit tests for B-11 "swap mocks for the real backend": the AuthenticatedClient
// interceptor (Bearer auth + refresh-on-401 + replay) and
// AuthController.refreshSession (the refresh the interceptor drives).
//
// Fully offline via package:http MockClient + an in-memory FakeSessionStore —
// no live backend required. Test artefacts: docs/testing/session-13/unit/.

import 'dart:convert';

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:mergemarket/models/auth_session.dart';
import 'package:mergemarket/providers/auth_providers.dart';
import 'package:mergemarket/providers/search_providers.dart';
import 'package:mergemarket/services/authenticated_client.dart';

import '../mocks/auth_mock_data.dart';

void main() {
  group('B-11 AuthenticatedClient — Bearer auth', () {
    test('TC-13-U-001: attaches "Bearer <token>" when signed in', () async {
      String? seen;
      final inner = MockClient((req) async {
        seen = req.headers['Authorization'];
        return http.Response('{}', 200);
      });
      final client = AuthenticatedClient(
        inner: inner,
        readToken: () => 'tok-123',
        refreshToken: () async => false,
      );

      final res = await client.get(Uri.parse('http://x/api/v1/search?q=phone'));

      expect(res.statusCode, 200);
      expect(seen, 'Bearer tok-123');
    });

    test('TC-13-U-002: sends no Authorization header when signed out', () async {
      var hadAuth = true;
      final inner = MockClient((req) async {
        hadAuth = req.headers.containsKey('Authorization');
        return http.Response('{}', 200);
      });
      final client = AuthenticatedClient(
        inner: inner,
        readToken: () => null,
        refreshToken: () async => false,
      );

      await client.get(Uri.parse('http://x/api/v1/search'));

      expect(hadAuth, isFalse);
    });
  });

  group('B-11 AuthenticatedClient — refresh on 401', () {
    test('TC-13-U-003: refreshes once on 401 and replays with the new token',
        () async {
      var token = 'old';
      var refreshCalls = 0;
      final seen = <String?>[];
      final inner = MockClient((req) async {
        seen.add(req.headers['Authorization']);
        if (req.headers['Authorization'] == 'Bearer old') {
          return http.Response('{"error":"unauthorized"}', 401);
        }
        return http.Response('{"ok":true}', 200);
      });
      final client = AuthenticatedClient(
        inner: inner,
        readToken: () => token,
        refreshToken: () async {
          refreshCalls++;
          token = 'new';
          return true;
        },
      );

      final res = await client.get(Uri.parse('http://x/api/v1/wishlist'));

      expect(res.statusCode, 200);
      expect(refreshCalls, 1);
      expect(seen, ['Bearer old', 'Bearer new']);
    });

    test('TC-13-U-004: a failed refresh surfaces the original 401', () async {
      var refreshCalls = 0;
      final inner =
          MockClient((req) async => http.Response('{"error":"unauthorized"}', 401));
      final client = AuthenticatedClient(
        inner: inner,
        readToken: () => 'old',
        refreshToken: () async {
          refreshCalls++;
          return false;
        },
      );

      final res = await client.get(Uri.parse('http://x/api/v1/wishlist'));

      expect(res.statusCode, 401);
      expect(refreshCalls, 1);
    });

    test('TC-13-U-005: never refreshes a 401 on auth endpoints', () async {
      var refreshCalls = 0;
      final inner = MockClient(
          (req) async => http.Response('{"error":"invalid_credentials"}', 401));
      final client = AuthenticatedClient(
        inner: inner,
        readToken: () => 'old',
        refreshToken: () async {
          refreshCalls++;
          return true;
        },
      );

      final res = await client.post(Uri.parse('http://x/api/v1/auth/login'),
          body: '{}');

      expect(res.statusCode, 401);
      expect(refreshCalls, 0);
    });

    test('TC-13-U-006: does not refresh a 401 when no token was sent', () async {
      var refreshCalls = 0;
      final inner = MockClient((req) async => http.Response('{}', 401));
      final client = AuthenticatedClient(
        inner: inner,
        readToken: () => null,
        refreshToken: () async {
          refreshCalls++;
          return true;
        },
      );

      final res = await client.get(Uri.parse('http://x/api/v1/search'));

      expect(res.statusCode, 401);
      expect(refreshCalls, 0);
    });

    test('TC-13-U-007: concurrent 401s coalesce into a single refresh',
        () async {
      var token = 'old';
      var refreshCalls = 0;
      final inner = MockClient((req) async {
        if (req.headers['Authorization'] == 'Bearer old') {
          return http.Response('{}', 401);
        }
        return http.Response('{}', 200);
      });
      final client = AuthenticatedClient(
        inner: inner,
        readToken: () => token,
        refreshToken: () async {
          refreshCalls++;
          await Future<void>.delayed(const Duration(milliseconds: 10));
          token = 'new';
          return true;
        },
      );

      final results = await Future.wait([
        client.get(Uri.parse('http://x/api/v1/wishlist')),
        client.get(Uri.parse('http://x/api/v1/alerts')),
        client.get(Uri.parse('http://x/api/v1/savings')),
      ]);

      expect(results.map((r) => r.statusCode), everyElement(200));
      expect(refreshCalls, 1);
    });

    test('TC-13-U-008: replays the POST body intact after a refresh', () async {
      var token = 'old';
      final bodies = <String>[];
      final inner = MockClient((req) async {
        bodies.add(req.body);
        if (req.headers['Authorization'] == 'Bearer old') {
          return http.Response('{}', 401);
        }
        return http.Response('{"alert_id":"a1"}', 201);
      });
      final client = AuthenticatedClient(
        inner: inner,
        readToken: () => token,
        refreshToken: () async {
          token = 'new';
          return true;
        },
      );

      final res = await client.post(
        Uri.parse('http://x/api/v1/alerts'),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode(
            {'product_id': 'p1', 'threshold_price': 100, 'currency': 'XAF'}),
      );

      expect(res.statusCode, 201);
      expect(bodies, hasLength(2));
      expect(bodies[0], bodies[1]); // body survived the replay
      expect(jsonDecode(bodies[1])['product_id'], 'p1');
    });
  });

  group('B-11 AuthController.refreshSession', () {
    ProviderContainer container({
      required http.Client client,
      required FakeSessionStore store,
    }) {
      final c = ProviderContainer(overrides: [
        httpClientProvider.overrideWithValue(client),
        sessionStoreProvider.overrideWithValue(store),
      ]);
      addTearDown(c.dispose);
      return c;
    }

    test('TC-13-U-009: a successful refresh swaps in the new session', () async {
      final store = FakeSessionStore(
        initial: AuthSession.fromJson(authTokenJson()),
        initialEmail: 'user@example.com',
      );
      final refreshed = {
        'token': 'refreshed.access-token',
        'refresh_token': 'refreshed.refresh-token',
        'expires_at':
            DateTime.now().add(const Duration(hours: 1)).toIso8601String(),
      };
      final c = container(
        client: MockClient((_) async => http.Response(jsonEncode(refreshed), 200)),
        store: store,
      );
      await c.read(authControllerProvider.future);

      final ok = await c.read(authControllerProvider.notifier).refreshSession();

      expect(ok, isTrue);
      expect(c.read(isAuthenticatedProvider), isTrue);
      expect(c.read(authControllerProvider).value?.session?.token,
          'refreshed.access-token');
      expect((await store.read())?.token, 'refreshed.access-token');
      expect(await store.readEmail(), 'user@example.com');
    });

    test('TC-13-U-010: a failed refresh clears the session and returns false',
        () async {
      final store = FakeSessionStore(
        initial: AuthSession.fromJson(authTokenJson()),
        initialEmail: 'user@example.com',
      );
      final c = container(
        client: MockClient((_) async =>
            http.Response(jsonEncode(kInvalidCredentialsJson), 401)),
        store: store,
      );
      await c.read(authControllerProvider.future);
      expect(c.read(isAuthenticatedProvider), isTrue);

      final ok = await c.read(authControllerProvider.notifier).refreshSession();

      expect(ok, isFalse);
      expect(c.read(isAuthenticatedProvider), isFalse);
      expect(store.clearCount, 1);
      expect(await store.read(), isNull);
    });
  });
}
