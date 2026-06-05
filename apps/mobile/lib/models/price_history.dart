/// Data models for the price-history contract
/// (`GET /api/v1/products/{product_id}/history`).
///
/// Field names and JSON keys mirror `project_docs/api/API_CONTRACTS.md`
/// exactly, so these decode identically against the B-02 mock server and the
/// real History service later. The data layer is the only place that parses
/// JSON — the Product Detail screen and its widgets consume typed models only.
library;

/// A single historical price observation for a product.
class PricePoint {
  const PricePoint({
    required this.price,
    required this.currency,
    required this.recordedAt,
  });

  final double price;
  final String currency;

  /// ISO 8601 timestamp of when this price was recorded.
  final String recordedAt;

  /// [recordedAt] parsed to a [DateTime]; falls back to the Unix epoch when the
  /// string is missing or unparseable, so the chart never throws on bad data.
  DateTime get recordedAtDate =>
      DateTime.tryParse(recordedAt)?.toLocal() ??
      DateTime.fromMillisecondsSinceEpoch(0);

  /// Decodes one history point. Missing or mistyped fields fall back to safe
  /// zero values rather than throwing, so one bad point never breaks the chart.
  factory PricePoint.fromJson(Map<String, dynamic> json) {
    return PricePoint(
      price: (json['price'] as num?)?.toDouble() ?? 0,
      currency: json['currency'] as String? ?? '',
      recordedAt: json['recorded_at'] as String? ?? '',
    );
  }
}

/// The full `GET /api/v1/products/{id}/history` response body.
class PriceHistory {
  const PriceHistory({
    required this.productId,
    required this.title,
    required this.history,
    required this.average6m,
    required this.lowest30d,
  });

  final String productId;
  final String title;

  /// Price points, oldest first (as returned by the contract).
  final List<PricePoint> history;

  /// Mean price over the last six months — drives the AI Deal Meter baseline.
  final double average6m;

  /// Lowest observed price in the last 30 days.
  final double lowest30d;

  /// The most recent recorded price, or `null` when there is no history.
  double? get latestPrice => history.isEmpty ? null : history.last.price;

  /// Currency of the series (taken from the latest point), empty when unknown.
  String get currency => history.isEmpty ? '' : history.last.currency;

  /// Decodes a history response. A missing `history` array decodes to an empty
  /// list (a valid response with no recorded prices).
  factory PriceHistory.fromJson(Map<String, dynamic> json) {
    final rawHistory = json['history'] as List<dynamic>? ?? const [];
    return PriceHistory(
      productId: json['product_id'] as String? ?? '',
      title: json['title'] as String? ?? '',
      history: rawHistory
          .map((e) => PricePoint.fromJson(e as Map<String, dynamic>))
          .toList(growable: false),
      average6m: (json['average_6m'] as num?)?.toDouble() ?? 0,
      lowest30d: (json['lowest_30d'] as num?)?.toDouble() ?? 0,
    );
  }
}
