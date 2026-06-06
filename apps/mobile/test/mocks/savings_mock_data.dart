/// Canned savings fixtures for B-07 tests.
///
/// Mirrors the B-02 mock server's savings body
/// (`services/mock-server/internal/fixtures/fixtures.go -> Savings`) so the
/// Flutter data layer is exercised against the exact shape it sees at runtime.
library;

/// `GET /api/v1/savings` with three realised savings transactions.
const Map<String, dynamic> kSavingsJson = {
  'total_saved': 33500,
  'currency': 'XAF',
  'transactions': [
    {
      'product_id': 'prod-001',
      'title': 'Samsung Galaxy A54 128GB',
      'saved': 23000,
      'bought_at': '2026-05-16T09:00:00Z',
    },
    {
      'product_id': 'prod-009',
      'title': 'JBL Tune 510BT Headphones',
      'saved': 6500,
      'bought_at': '2026-04-26T09:00:00Z',
    },
    {
      'product_id': 'prod-014',
      'title': 'Anker PowerCore 20000mAh',
      'saved': 4000,
      'bought_at': '2026-04-11T09:00:00Z',
    },
  ],
};

/// A top-level savings response, used to verify level clamping.
const Map<String, dynamic> kSavingsTopLevelJson = {
  'total_saved': 510000,
  'currency': 'XAF',
  'transactions': [
    {
      'product_id': 'prod-020',
      'title': 'MacBook Air M3',
      'saved': 510000,
      'bought_at': '2026-06-01T09:00:00Z',
    },
  ],
};

/// An empty-but-valid savings response.
const Map<String, dynamic> kSavingsEmptyJson = {
  'total_saved': 0,
  'currency': 'XAF',
  'transactions': <dynamic>[],
};

/// A standard API error body for savings failures.
const Map<String, dynamic> kSavingsErrorJson = {
  'error': 'server_error',
  'message': 'Savings are unavailable right now.',
};
