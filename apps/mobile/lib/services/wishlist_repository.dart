import 'dart:convert';

import 'package:http/http.dart' as http;

import '../config/app_config.dart';
import '../models/wishlist.dart';
import 'api_exception.dart';

/// Data-layer access to the wishlist contract (`/api/v1/wishlist`).
///
/// The only component that knows the wishlist API shape: it lists items, adds a
/// product, and removes an entry, mapping the contract's status codes to a typed
/// [ApiException]. The injected [http.Client] makes it unit testable with
/// `package:http`'s `MockClient`.
class WishlistRepository {
  WishlistRepository({
    required http.Client client,
    AppConfig config = AppConfig.env,
  })  : _client = client,
        _config = config;

  final http.Client _client;
  final AppConfig _config;

  Uri _uri(String path) => Uri.parse('${_config.apiBaseUrl}$path');

  /// Lists the user's wishlist. Returns a decoded [Wishlist] on 200; throws
  /// [ApiException] for any other status or a transport error.
  Future<Wishlist> list() async {
    final http.Response response;
    try {
      response = await _client.get(
        _uri('/api/v1/wishlist'),
        headers: const {'Accept': 'application/json'},
      );
    } on Exception {
      throw _networkError;
    }
    if (response.statusCode == 200) {
      return Wishlist.fromJson(jsonDecode(response.body) as Map<String, dynamic>);
    }
    throw ApiException(
      ApiErrorKind.server,
      _messageOr(response, 'Could not load your wishlist. Please try again.'),
    );
  }

  /// Adds [productId] to the wishlist. Returns normally on 201; throws
  /// [ApiException] for 409 (already added), 400, any other status, or a
  /// transport error.
  Future<void> add(String productId) async {
    final http.Response response;
    try {
      response = await _client.post(
        _uri('/api/v1/wishlist'),
        headers: const {
          'Accept': 'application/json',
          'Content-Type': 'application/json',
        },
        body: jsonEncode({'product_id': productId}),
      );
    } on Exception {
      throw _networkError;
    }
    switch (response.statusCode) {
      case 201:
        return;
      case 409:
        throw ApiException(
          ApiErrorKind.badRequest,
          _messageOr(response, 'This product is already in your wishlist.'),
        );
      case 400:
        throw ApiException(
          ApiErrorKind.badRequest,
          _messageOr(response, 'Could not add this product.'),
        );
      default:
        throw ApiException(
          ApiErrorKind.server,
          _messageOr(response, 'Could not add to wishlist. Please try again.'),
        );
    }
  }

  /// Removes the entry [wishlistId]. Returns normally on 204; throws
  /// [ApiException] for 404, any other status, or a transport error.
  Future<void> remove(String wishlistId) async {
    final http.Response response;
    try {
      response = await _client.delete(
        _uri('/api/v1/wishlist/$wishlistId'),
        headers: const {'Accept': 'application/json'},
      );
    } on Exception {
      throw _networkError;
    }
    switch (response.statusCode) {
      case 204:
        return;
      case 404:
        throw ApiException(
          ApiErrorKind.notFound,
          _messageOr(response, 'This item is no longer in your wishlist.'),
        );
      default:
        throw ApiException(
          ApiErrorKind.server,
          _messageOr(response, 'Could not remove the item. Please try again.'),
        );
    }
  }

  static const ApiException _networkError = ApiException(
    ApiErrorKind.network,
    'Could not reach the server. Check your connection and try again.',
  );

  /// Extracts the contract's `{error, message}` human message, falling back to
  /// [fallback] when the body is absent or not the expected error shape.
  String _messageOr(http.Response response, String fallback) {
    try {
      final body = jsonDecode(response.body) as Map<String, dynamic>;
      final message = body['message'] as String?;
      if (message != null && message.isNotEmpty) return message;
    } catch (_) {
      // Fall through to the fallback message.
    }
    return fallback;
  }
}
