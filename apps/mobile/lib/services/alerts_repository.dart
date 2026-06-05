import 'dart:convert';

import 'package:http/http.dart' as http;

import '../config/app_config.dart';
import '../models/alert.dart';
import 'api_exception.dart';

/// Data-layer access to the alerts contract (`/api/v1/alerts`).
///
/// The only component that knows the alerts API shape: it lists alerts, creates
/// a price-threshold alert, and deletes one, mapping the contract's status codes
/// to a typed [ApiException]. The injected [http.Client] makes it unit testable
/// with `package:http`'s `MockClient`.
class AlertsRepository {
  AlertsRepository({
    required http.Client client,
    AppConfig config = AppConfig.env,
  })  : _client = client,
        _config = config;

  final http.Client _client;
  final AppConfig _config;

  Uri _uri(String path) => Uri.parse('${_config.apiBaseUrl}$path');

  /// Lists the user's alerts. Returns a decoded [AlertList] on 200; throws
  /// [ApiException] for any other status or a transport error.
  Future<AlertList> list() async {
    final http.Response response;
    try {
      response = await _client.get(
        _uri('/api/v1/alerts'),
        headers: const {'Accept': 'application/json'},
      );
    } on Exception {
      throw _networkError;
    }
    if (response.statusCode == 200) {
      return AlertList.fromJson(
          jsonDecode(response.body) as Map<String, dynamic>);
    }
    throw ApiException(
      ApiErrorKind.server,
      _messageOr(response, 'Could not load your alerts. Please try again.'),
    );
  }

  /// Creates a price alert for [productId] at [thresholdPrice] in [currency].
  /// Returns normally on 201; throws [ApiException] for 400, any other status,
  /// or a transport error.
  Future<void> create({
    required String productId,
    required double thresholdPrice,
    required String currency,
  }) async {
    final http.Response response;
    try {
      response = await _client.post(
        _uri('/api/v1/alerts'),
        headers: const {
          'Accept': 'application/json',
          'Content-Type': 'application/json',
        },
        body: jsonEncode({
          'product_id': productId,
          'threshold_price': thresholdPrice,
          'currency': currency,
        }),
      );
    } on Exception {
      throw _networkError;
    }
    switch (response.statusCode) {
      case 201:
        return;
      case 400:
        throw ApiException(
          ApiErrorKind.badRequest,
          _messageOr(response, 'Please enter a valid alert price.'),
        );
      default:
        throw ApiException(
          ApiErrorKind.server,
          _messageOr(response, 'Could not set the alert. Please try again.'),
        );
    }
  }

  /// Deletes the alert [alertId]. Returns normally on 204; throws [ApiException]
  /// for 404, any other status, or a transport error.
  Future<void> remove(String alertId) async {
    final http.Response response;
    try {
      response = await _client.delete(
        _uri('/api/v1/alerts/$alertId'),
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
          _messageOr(response, 'This alert no longer exists.'),
        );
      default:
        throw ApiException(
          ApiErrorKind.server,
          _messageOr(response, 'Could not remove the alert. Please try again.'),
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
