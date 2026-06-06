/// Canned alerts fixtures for B-06 tests.
///
/// Mirrors the B-02 mock server's alerts bodies
/// (`services/mock-server/internal/fixtures/fixtures.go → Alerts`) so the
/// Flutter data layer is exercised against the exact shapes it sees at runtime.
library;

/// `GET /api/v1/alerts` — one active, one inactive.
const Map<String, dynamic> kAlertsJson = {
  'alerts': [
    {
      'alert_id': 'al-001',
      'product_id': 'prod-001',
      'title': 'Samsung Galaxy A54 128GB',
      'threshold_price': 230000,
      'currency': 'XAF',
      'is_active': true,
      'created_at': '2026-06-01T09:00:00Z',
    },
    {
      'alert_id': 'al-002',
      'product_id': 'prod-014',
      'title': 'Anker PowerCore 20000mAh',
      'threshold_price': 25000,
      'currency': 'XAF',
      'is_active': false,
      'created_at': '2026-05-27T09:00:00Z',
    },
  ],
};

/// An empty-but-valid alerts response.
const Map<String, dynamic> kAlertsEmptyJson = {'alerts': <dynamic>[]};

/// The 201 body for POST /api/v1/alerts.
const Map<String, dynamic> kAlertCreatedJson = {
  'alert_id': 'al-003',
  'created_at': '2026-06-05T09:00:00Z',
};

/// The 400 body for an invalid alert request.
const Map<String, dynamic> kAlertInvalidJson = {
  'error': 'invalid_input',
  'message': 'product_id and a positive threshold_price are required',
};
