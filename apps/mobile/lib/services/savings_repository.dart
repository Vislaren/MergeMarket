import 'dart:convert';

import 'package:http/http.dart' as http;

import '../config/app_config.dart';
import '../models/savings.dart';
import 'api_exception.dart';

/// Data-layer access to the savings contract (`/api/v1/savings`).
///
/// This repository is the only component that knows the endpoint path and
/// status-code mapping. It returns a decoded [SavingsSummary] on 200 and maps
/// transport/server failures to [ApiException] for the UI layer.
class SavingsRepository {
  SavingsRepository({
    required http.Client client,
    AppConfig config = AppConfig.env,
  }) : _client = client,
       _config = config;

  final http.Client _client;
  final AppConfig _config;

  Uri _uri(String path) => Uri.parse('${_config.apiBaseUrl}$path');

  /// Loads the user's cumulative savings dashboard data.
  Future<SavingsSummary> getSavings() async {
    final http.Response response;
    try {
      response = await _client.get(
        _uri('/api/v1/savings'),
        headers: const {'Accept': 'application/json'},
      );
    } on Exception {
      throw _networkError;
    }

    if (response.statusCode == 200) {
      return SavingsSummary.fromJson(
        jsonDecode(response.body) as Map<String, dynamic>,
      );
    }

    throw ApiException(
      ApiErrorKind.server,
      _messageOr(response, 'Could not load your savings. Please try again.'),
    );
  }

  static const ApiException _networkError = ApiException(
    ApiErrorKind.network,
    'Could not reach the server. Check your connection and try again.',
  );

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
