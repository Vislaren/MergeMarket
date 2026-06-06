/// Data models for the wishlist contract (`/api/v1/wishlist`).
///
/// Field names and JSON keys mirror `project_docs/api/API_CONTRACTS.md`
/// exactly, so these decode identically against the B-02 mock server and the
/// real services later. The data layer is the only place that parses JSON —
/// the Wishlist screen and its widgets consume typed models only.
library;

/// One store's offer for a wishlisted product.
class WishlistStore {
  const WishlistStore({
    required this.store,
    required this.price,
    required this.totalCost,
  });

  final String store;
  final double price;
  final double totalCost;

  /// Decodes one store entry, with safe zero-value fallbacks.
  factory WishlistStore.fromJson(Map<String, dynamic> json) {
    return WishlistStore(
      store: json['store'] as String? ?? '',
      price: (json['price'] as num?)?.toDouble() ?? 0,
      totalCost: (json['total_cost'] as num?)?.toDouble() ?? 0,
    );
  }
}

/// One entry in the wishlist, tracked across one or more stores.
class WishlistItem {
  const WishlistItem({
    required this.wishlistId,
    required this.productId,
    required this.title,
    required this.imageUrl,
    required this.stores,
    required this.addedAt,
  });

  final String wishlistId;
  final String productId;
  final String title;
  final String imageUrl;

  /// Per-store offers for this product, in the order returned by the API.
  final List<WishlistStore> stores;

  /// ISO 8601 timestamp of when the product was added.
  final String addedAt;

  /// Number of distinct stores tracking this product.
  int get storeCount => stores.map((s) => s.store).toSet().length;

  /// The lowest total cost across [stores], or `null` when none are tracked.
  double? get bestTotalCost => stores.isEmpty
      ? null
      : stores.map((s) => s.totalCost).reduce((a, b) => a < b ? a : b);

  /// Decodes one wishlist item. A missing `stores` array decodes to an empty
  /// list rather than throwing.
  factory WishlistItem.fromJson(Map<String, dynamic> json) {
    final rawStores = json['stores'] as List<dynamic>? ?? const [];
    return WishlistItem(
      wishlistId: json['wishlist_id'] as String? ?? '',
      productId: json['product_id'] as String? ?? '',
      title: json['title'] as String? ?? '',
      imageUrl: json['image_url'] as String? ?? '',
      stores: rawStores
          .map((e) => WishlistStore.fromJson(e as Map<String, dynamic>))
          .toList(growable: false),
      addedAt: json['added_at'] as String? ?? '',
    );
  }
}

/// The full `GET /api/v1/wishlist` response body.
class Wishlist {
  const Wishlist({required this.items});

  final List<WishlistItem> items;

  /// Decodes the wishlist response; a missing `items` array decodes to empty.
  factory Wishlist.fromJson(Map<String, dynamic> json) {
    final rawItems = json['items'] as List<dynamic>? ?? const [];
    return Wishlist(
      items: rawItems
          .map((e) => WishlistItem.fromJson(e as Map<String, dynamic>))
          .toList(growable: false),
    );
  }
}
