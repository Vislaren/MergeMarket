// Unit tests for WishlistRepository (B-05) using package:http MockClient.
//
// Covers list (200), add (201 / 409), remove (204 / 404), correct verbs+URLs,
// and the transport-error → network mapping.
//
// Test artefacts: docs/testing/session-08/unit/.

import 'dart:convert';

import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:mergemarket/services/api_exception.dart';
import 'package:mergemarket/services/wishlist_repository.dart';

import '../mocks/wishlist_mock_data.dart';

void main() {
  group('B-05 WishlistRepository.list', () {
    test('TC-08-U-005: 200 decodes the wishlist', () async {
      final repo = WishlistRepository(
          client: MockClient((_) async => http.Response(
              jsonEncode(kWishlistJson), 200)));
      final wl = await repo.list();
      expect(wl.items, hasLength(2));
    });

    test('TC-08-U-006: transport failure maps to ApiException(network)',
        () async {
      final repo = WishlistRepository(
          client: MockClient((_) async => throw http.ClientException('x')));
      await expectLater(
        repo.list(),
        throwsA(isA<ApiException>()
            .having((e) => e.kind, 'kind', ApiErrorKind.network)),
      );
    });
  });

  group('B-05 WishlistRepository.add', () {
    test('TC-08-U-007: 201 POSTs product_id to /api/v1/wishlist', () async {
      late http.Request captured;
      final repo = WishlistRepository(client: MockClient((req) async {
        captured = req;
        return http.Response(jsonEncode(kWishlistCreatedJson), 201);
      }));
      await repo.add('prod-014');
      expect(captured.method, 'POST');
      expect(captured.url.path, '/api/v1/wishlist');
      expect(jsonDecode(captured.body), {'product_id': 'prod-014'});
    });

    test('TC-08-U-008: 409 maps to ApiException(badRequest)', () async {
      final repo = WishlistRepository(
          client: MockClient((_) async => http.Response(
              jsonEncode(kAlreadyInWishlistJson), 409)));
      await expectLater(
        repo.add('prod-001'),
        throwsA(isA<ApiException>()
            .having((e) => e.kind, 'kind', ApiErrorKind.badRequest)
            .having((e) => e.message, 'message', contains('already'))),
      );
    });
  });

  group('B-05 WishlistRepository.remove', () {
    test('TC-08-U-009: 204 DELETEs the entry by id', () async {
      late http.Request captured;
      final repo = WishlistRepository(client: MockClient((req) async {
        captured = req;
        return http.Response('', 204);
      }));
      await repo.remove('wl-002');
      expect(captured.method, 'DELETE');
      expect(captured.url.path, '/api/v1/wishlist/wl-002');
    });

    test('TC-08-U-010: 404 maps to ApiException(notFound)', () async {
      final repo = WishlistRepository(
          client: MockClient((_) async => http.Response('{}', 404)));
      await expectLater(
        repo.remove('unknown'),
        throwsA(isA<ApiException>()
            .having((e) => e.kind, 'kind', ApiErrorKind.notFound)),
      );
    });
  });
}
