import 'dart:convert';

import 'package:http/http.dart' as http;

import '../config/app_config.dart';
import '../models/product.dart';
import 'api_exception.dart';

/// Data-layer access to the Search contract (`GET /api/v1/search`).
///
/// This is the only component that knows the API shape for search: it builds
/// the request, maps status codes to a typed [ApiException], and decodes the
/// body into a [SearchResponse]. The injected [http.Client] makes it unit
/// testable with `package:http`'s `MockClient`.
class SearchRepository {
  SearchRepository({required http.Client client, AppConfig config = AppConfig.env})
      : _client = client,
        _config = config;

  final http.Client _client;
  final AppConfig _config;

  /// Searches all stores for [query] in [location] (defaults to the configured
  /// locale). Returns a decoded [SearchResponse] on 200; throws [ApiException]
  /// for the contract's 400/504 paths, any other status, or a transport error.
  Future<SearchResponse> search(String query, {String? location}) async {
    final uri = Uri.parse('${_config.apiBaseUrl}/api/v1/search').replace(
      queryParameters: {
        'q': query,
        'location': location ?? _config.defaultLocation,
      },
    );

    final http.Response response;
    try {
      response = await _client.get(uri, headers: const {'Accept': 'application/json'});
    } on Exception {
      throw const ApiException(
        ApiErrorKind.network,
        'Could not reach the server. Check your connection and try again.',
      );
    }

    switch (response.statusCode) {
      case 200:
        return SearchResponse.fromJson(
          jsonDecode(response.body) as Map<String, dynamic>,
        );
      case 400:
        throw ApiException(
          ApiErrorKind.badRequest,
          _messageOr(response, 'Please enter a product to search for.'),
        );
      case 504:
        throw ApiException(
          ApiErrorKind.timeout,
          _messageOr(response, 'Results took too long. Try again.'),
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
