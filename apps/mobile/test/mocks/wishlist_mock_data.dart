/// Canned wishlist fixtures for B-05 tests.
///
/// Mirrors the B-02 mock server's wishlist bodies
/// (`services/mock-server/internal/fixtures/fixtures.go → Wishlist`) so the
/// Flutter data layer is exercised against the exact shapes it sees at runtime.
library;

/// `GET /api/v1/wishlist` — two tracked products, multi-store on the first.
const Map<String, dynamic> kWishlistJson = {
  'items': [
    {
      'wishlist_id': 'wl-001',
      'product_id': 'prod-001',
      'title': 'Samsung Galaxy A54 128GB',
      'image_url': 'https://img.example/galaxy-a54.jpg',
      'stores': [
        {'store': 'Jumia', 'price': 245000, 'total_cost': 250000},
        {'store': 'Kilimall', 'price': 239900, 'total_cost': 247900},
      ],
      'added_at': '2026-05-31T09:00:00Z',
    },
    {
      'wishlist_id': 'wl-002',
      'product_id': 'prod-014',
      'title': 'Anker PowerCore 20000mAh',
      'image_url': 'https://img.example/anker-20k.jpg',
      'stores': [
        {'store': 'Jumia', 'price': 28500, 'total_cost': 31500},
      ],
      'added_at': '2026-05-24T09:00:00Z',
    },
  ],
};

/// An empty-but-valid wishlist response.
const Map<String, dynamic> kWishlistEmptyJson = {'items': <dynamic>[]};

/// The 201 body for POST /api/v1/wishlist.
const Map<String, dynamic> kWishlistCreatedJson = {
  'wishlist_id': 'wl-003',
  'added_at': '2026-06-05T09:00:00Z',
};

/// The 409 body for adding a product already in the wishlist.
const Map<String, dynamic> kAlreadyInWishlistJson = {
  'error': 'already_in_wishlist',
  'message': 'product is already in the wishlist',
};
