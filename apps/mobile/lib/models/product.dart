/// Data models for the Search contract (`GET /api/v1/search`).
///
/// Field names and JSON keys mirror `project_docs/api/API_CONTRACTS.md`
/// exactly, so these decode identically against the B-02 mock server and the
/// real services later. The data layer is the only place that parses JSON —
/// screens and widgets consume typed models only.
library;

/// A single store offer returned by a search.
///
/// Multiple [Product] entries can share the same [productId] when the same
/// item is found across several stores — the Results screen lists every offer
/// and sorts by [totalCost].
class Product {
  const Product({
    required this.productId,
    required this.title,
    required this.price,
    required this.currency,
    required this.shipping,
    required this.totalCost,
    required this.imageUrl,
    required this.store,
    required this.affiliateUrl,
    required this.dealScore,
    required this.scrapedAt,
  });

  final String productId;
  final String title;
  final double price;
  final String currency;
  final double shipping;
  final double totalCost;
  final String imageUrl;
  final String store;
  final String affiliateUrl;

  /// AI Deal Meter score, 0–100. Drives the deal badge on the product card.
  final int dealScore;

  /// ISO 8601 timestamp of when this offer was scraped.
  final String scrapedAt;

  /// Decodes one offer from a search-results JSON object. Missing or
  /// mistyped fields fall back to safe zero values rather than throwing, so a
  /// single malformed offer never breaks the whole results list.
  factory Product.fromJson(Map<String, dynamic> json) {
    return Product(
      productId: json['product_id'] as String? ?? '',
      title: json['title'] as String? ?? '',
      price: (json['price'] as num?)?.toDouble() ?? 0,
      currency: json['currency'] as String? ?? '',
      shipping: (json['shipping'] as num?)?.toDouble() ?? 0,
      totalCost: (json['total_cost'] as num?)?.toDouble() ?? 0,
      imageUrl: json['image_url'] as String? ?? '',
      store: json['store'] as String? ?? '',
      affiliateUrl: json['affiliate_url'] as String? ?? '',
      dealScore: (json['deal_score'] as num?)?.toInt() ?? 0,
      scrapedAt: json['scraped_at'] as String? ?? '',
    );
  }
}

/// The full `GET /api/v1/search` response body.
class SearchResponse {
  const SearchResponse({
    required this.query,
    required this.results,
    required this.cached,
    required this.latencyMs,
  });

  final String query;
  final List<Product> results;
  final bool cached;
  final int latencyMs;

  /// Number of distinct stores represented across [results].
  int get storeCount => results.map((r) => r.store).toSet().length;

  /// Decodes a search response. A missing `results` array decodes to an empty
  /// list (an empty-but-valid response), distinct from a transport error.
  factory SearchResponse.fromJson(Map<String, dynamic> json) {
    final rawResults = json['results'] as List<dynamic>? ?? const [];
    return SearchResponse(
      query: json['query'] as String? ?? '',
      results: rawResults
          .map((e) => Product.fromJson(e as Map<String, dynamic>))
          .toList(growable: false),
      cached: json['cached'] as bool? ?? false,
      latencyMs: (json['latency_ms'] as num?)?.toInt() ?? 0,
    );
  }
}
