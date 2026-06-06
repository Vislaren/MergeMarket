/// Data model for the product truth-score contract
/// (`GET /api/v1/products/{product_id}/truth-score`).
///
/// Field names and JSON keys mirror `project_docs/api/API_CONTRACTS.md`
/// exactly. The data layer is the only place that parses JSON — the
/// `MMTruthScore` widget consumes this typed model only.
library;

/// Review-authenticity assessment for a product (the "Product Truth Score").
class TruthScore {
  const TruthScore({
    required this.productId,
    required this.score,
    required this.sentiment,
    required this.fakeReviewRisk,
    required this.summary,
  });

  final String productId;

  /// Overall authenticity/quality score, 0–100.
  final int score;

  /// Aggregate review sentiment: `positive`, `mixed`, or `negative`.
  final String sentiment;

  /// Likelihood that reviews are inauthentic: `low`, `medium`, or `high`.
  final String fakeReviewRisk;

  /// One-line human summary of the review analysis.
  final String summary;

  /// Decodes a truth-score response. Missing or mistyped fields fall back to
  /// safe defaults rather than throwing.
  factory TruthScore.fromJson(Map<String, dynamic> json) {
    return TruthScore(
      productId: json['product_id'] as String? ?? '',
      score: (json['score'] as num?)?.toInt() ?? 0,
      sentiment: json['sentiment'] as String? ?? 'mixed',
      fakeReviewRisk: json['fake_review_risk'] as String? ?? 'medium',
      summary: json['summary'] as String? ?? '',
    );
  }
}
