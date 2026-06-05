/// Data models for the alerts contract (`/api/v1/alerts`).
///
/// Field names and JSON keys mirror `project_docs/api/API_CONTRACTS.md`
/// exactly, so these decode identically against the B-02 mock server and the
/// real services later. The data layer is the only place that parses JSON —
/// the Alerts screen and its widgets consume typed models only.
library;

/// One price-threshold alert on a tracked product.
class Alert {
  const Alert({
    required this.alertId,
    required this.productId,
    required this.title,
    required this.thresholdPrice,
    required this.currency,
    required this.isActive,
    required this.createdAt,
  });

  final String alertId;
  final String productId;
  final String title;

  /// The price at or below which the user is notified.
  final double thresholdPrice;
  final String currency;

  /// Whether the alert is currently watching for drops.
  final bool isActive;

  /// ISO 8601 timestamp of when the alert was created.
  final String createdAt;

  /// Decodes one alert. Missing or mistyped fields fall back to safe values.
  factory Alert.fromJson(Map<String, dynamic> json) {
    return Alert(
      alertId: json['alert_id'] as String? ?? '',
      productId: json['product_id'] as String? ?? '',
      title: json['title'] as String? ?? '',
      thresholdPrice: (json['threshold_price'] as num?)?.toDouble() ?? 0,
      currency: json['currency'] as String? ?? '',
      isActive: json['is_active'] as bool? ?? false,
      createdAt: json['created_at'] as String? ?? '',
    );
  }
}

/// The full `GET /api/v1/alerts` response body.
class AlertList {
  const AlertList({required this.alerts});

  final List<Alert> alerts;

  /// Decodes the alerts response; a missing `alerts` array decodes to empty.
  factory AlertList.fromJson(Map<String, dynamic> json) {
    final raw = json['alerts'] as List<dynamic>? ?? const [];
    return AlertList(
      alerts: raw
          .map((e) => Alert.fromJson(e as Map<String, dynamic>))
          .toList(growable: false),
    );
  }
}
