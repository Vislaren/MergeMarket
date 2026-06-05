/// Canned product-detail fixtures for B-04 tests.
///
/// Mirrors the B-02 mock server's history, truth-score, and search bodies
/// (`services/mock-server/internal/fixtures/fixtures.go`) so the Flutter data
/// layer is exercised against the exact shapes it sees at runtime.
library;

/// `GET /api/v1/products/{id}/history` — six monthly points, oldest first.
const Map<String, dynamic> kHistoryJson = {
  'product_id': 'prod-001',
  'title': 'Samsung Galaxy A54 128GB',
  'history': [
    {'price': 268000, 'currency': 'XAF', 'recorded_at': '2026-01-05T09:00:00Z'},
    {'price': 261000, 'currency': 'XAF', 'recorded_at': '2026-02-05T09:00:00Z'},
    {'price': 255000, 'currency': 'XAF', 'recorded_at': '2026-03-05T09:00:00Z'},
    {'price': 258000, 'currency': 'XAF', 'recorded_at': '2026-04-05T09:00:00Z'},
    {'price': 249000, 'currency': 'XAF', 'recorded_at': '2026-05-05T09:00:00Z'},
    {'price': 245000, 'currency': 'XAF', 'recorded_at': '2026-06-05T09:00:00Z'},
  ],
  'average_6m': 256000,
  'lowest_30d': 245000,
};

/// `GET /api/v1/products/{id}/truth-score`.
const Map<String, dynamic> kTruthScoreJson = {
  'product_id': 'prod-001',
  'score': 82,
  'sentiment': 'positive',
  'fake_review_risk': 'low',
  'summary':
      'Reviews are consistent across stores with low duplication; buyers '
          'praise battery life and value.',
};

/// `GET /api/v1/search?q=...` for the product's title — three store offers.
const Map<String, dynamic> kProductSearchJson = {
  'query': 'Samsung Galaxy A54 128GB',
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

/// The contract's 404 body for an unknown product.
const Map<String, dynamic> kNotFoundJson = {
  'error': 'not_found',
  'message': 'product not found',
};
