import 'dart:convert';

import 'package:http/http.dart' as http;

import '../config/app_config.dart';
import '../models/price_history.dart';
import '../models/truth_score.dart';
import 'api_exception.dart';

/// Data-layer access to the product-detail contracts: price history
/// (`GET /api/v1/products/{id}/history`) and the truth score
/// (`GET /api/v1/products/{id}/truth-score`).
///
/// This is the only component that knows these API shapes: it builds requests,
/// maps status codes to a typed [ApiException], and decodes bodies into typed
/// models. The injected [http.Client] makes it unit testable with `package:http`'s
/// `MockClient`.
class ProductRepository {
  ProductRepository({
    required http.Client client,
    AppConfig config = AppConfig.env,
  })  : _client = client,
        _config = config;

  final http.Client _client;
  final AppConfig _config;

  /// Fetches the six-month price history for [productId]. Returns a decoded
  /// [PriceHistory] on 200; throws [ApiException] for 404, any other status, or
  /// a transport error.
  Future<PriceHistory> history(String productId) async {
    final body = await _get('/api/v1/products/$productId/history');
    return PriceHistory.fromJson(body);
  }

  /// Fetches the review truth score for [productId]. Returns a decoded
  /// [TruthScore] on 200; throws [ApiException] for 404, any other status, or a
  /// transport error.
  Future<TruthScore> truthScore(String productId) async {
    final body = await _get('/api/v1/products/$productId/truth-score');
    return TruthScore.fromJson(body);
  }

  /// Performs a GET against [path] and returns the decoded JSON object, mapping
  /// the contract's failure paths to a typed [ApiException].
  Future<Map<String, dynamic>> _get(String path) async {
    final uri = Uri.parse('${_config.apiBaseUrl}$path');

    final http.Response response;
    try {
      response = await _client.get(
        uri,
        headers: const {'Accept': 'application/json'},
      );
    } on Exception {
      throw const ApiException(
        ApiErrorKind.network,
        'Could not reach the server. Check your connection and try again.',
      );
    }

    switch (response.statusCode) {
      case 200:
        return jsonDecode(response.body) as Map<String, dynamic>;
      case 404:
        throw ApiException(
          ApiErrorKind.notFound,
          _messageOr(response, "We couldn't find this product."),
        );
      default:
        throw ApiException(
          ApiErrorKind.server,
          _messageOr(response, 'Something went wrong. Please try again.'),
        );
    }
  }

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
