/// Canned Search-contract fixtures for B-03 tests.
///
/// Mirrors the B-02 mock server's `GET /api/v1/search` body
/// (`services/mock-server/internal/fixtures/fixtures.go → Search`) so the
/// Flutter data layer is exercised against the exact shape it sees at runtime.
/// `total_cost` always equals `price + shipping`, per the API contract.
library;

/// A successful three-store search response for "galaxy".
const Map<String, dynamic> kSearchSuccessJson = {
  'query': 'galaxy',
  'results': [
    {
      'product_id': 'prod-001',
      'title': 'Samsung Galaxy A54 128GB',
      'price': 245000,
      'currency': 'XAF',
      'shipping': 5000,
      'total_cost': 250000,
      'image_url': 'https://img.example/galaxy-a54.jpg',
      'store': 'Jumia',
      'affiliate_url': 'https://aff.example/jumia/prod-001',
      'deal_score': 88,
      'scraped_at': '2026-06-05T09:00:00Z',
    },
    {
      'product_id': 'prod-001',
      'title': 'Samsung Galaxy A54 128GB',
      'price': 239900,
      'currency': 'XAF',
      'shipping': 8000,
      'total_cost': 247900,
      'image_url': 'https://img.example/galaxy-a54.jpg',
      'store': 'Kilimall',
      'affiliate_url': 'https://aff.example/kilimall/prod-001',
      'deal_score': 81,
      'scraped_at': '2026-06-05T09:00:00Z',
    },
    {
      'product_id': 'prod-001',
      'title': 'Samsung Galaxy A54 (128GB)',
      'price': 252000,
      'currency': 'XAF',
      'shipping': 0,
      'total_cost': 252000,
      'image_url': 'https://img.example/galaxy-a54.jpg',
      'store': 'AfricShop',
      'affiliate_url': 'https://aff.example/africshop/prod-001',
      'deal_score': 74,
      'scraped_at': '2026-06-05T09:00:00Z',
    },
  ],
  'cached': true,
  'latency_ms': 42,
};

/// An empty-but-valid search response (no offers found).
const Map<String, dynamic> kSearchEmptyJson = {
  'query': 'zxqw',
  'results': <dynamic>[],
  'cached': false,
  'latency_ms': 18,
};

/// The contract's 400 body for a missing query.
const Map<String, dynamic> kMissingQueryJson = {
  'error': 'missing_query',
  'message': "query parameter 'q' is required",
};

/// The contract's 504 body for a search timeout.
const Map<String, dynamic> kTimeoutJson = {
  'error': 'timeout',
  'message': 'Results took too long. Try again.',
};
