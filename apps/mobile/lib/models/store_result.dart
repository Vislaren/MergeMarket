import 'product.dart';

/// One store's offer for a single product, as shown in the Product Detail
/// store-comparison table (`COMPONENT_LIBRARY.md → MMStoreComparisonTable`).
///
/// A trimmed projection of [Product] carrying only the columns the table needs,
/// so the widget is decoupled from the full search model.
class StoreResult {
  const StoreResult({
    required this.store,
    required this.price,
    required this.shipping,
    required this.totalCost,
    this.affiliateUrl = '',
  });

  final String store;
  final double price;
  final double shipping;
  final double totalCost;

  /// Affiliate link followed by the row's "Go to Store" action.
  final String affiliateUrl;

  /// Projects a search [Product] offer into a [StoreResult].
  factory StoreResult.fromProduct(Product p) => StoreResult(
        store: p.store,
        price: p.price,
        shipping: p.shipping,
        totalCost: p.totalCost,
        affiliateUrl: p.affiliateUrl,
      );
}
