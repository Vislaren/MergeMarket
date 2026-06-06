import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/price_history.dart';
import '../models/product.dart';
import '../models/truth_score.dart';
import '../services/product_repository.dart';
import 'search_providers.dart';

/// The Product Detail data-layer repository (history + truth score), wired to
/// the shared HTTP client from [httpClientProvider].
final productRepositoryProvider = Provider<ProductRepository>((ref) {
  return ProductRepository(client: ref.watch(httpClientProvider));
});

/// Everything the Product Detail screen needs, aggregated into one view model.
///
/// The BFF will eventually shape this server-side (B-09); until then the
/// provider layer composes it from the existing contracts — no business logic
/// leaks into the widgets, which consume only this typed aggregate.
class ProductDetail {
  const ProductDetail({
    required this.history,
    required this.offers,
    required this.truthScore,
  });

  /// Six-month price history, plus the title / 6-month average / 30-day low.
  final PriceHistory history;

  /// Every store offer for this product, ascending by total cost (cheapest
  /// first), so [bestOffer] is `offers.first`.
  final List<Product> offers;

  /// Review-authenticity assessment for this product.
  final TruthScore truthScore;

  /// The cheapest landed-cost offer, or `null` when no stores returned one.
  Product? get bestOffer => offers.isEmpty ? null : offers.first;

  /// The AI Deal Meter score for the headline (best) offer, 0–100.
  int get dealScore => bestOffer?.dealScore ?? 0;
}

/// Loads the full [ProductDetail] aggregate for a product id.
///
/// Keyed by product id (`.family`) so each product is cached and retried
/// independently; `ref.invalidate(productDetailProvider(id))` re-runs it for the
/// retry button. History is fetched first because its `title` is what the store
/// comparison searches for; the search and truth-score then run concurrently.
final productDetailProvider =
    FutureProvider.family<ProductDetail, String>((ref, productId) async {
  final productRepo = ref.watch(productRepositoryProvider);
  final searchRepo = ref.watch(searchRepositoryProvider);

  final history = await productRepo.history(productId);

  // Title is only known after history loads; the offer search and truth score
  // then run concurrently.
  final searchFuture = searchRepo.search(history.title);
  final truthFuture = productRepo.truthScore(productId);
  final searchResponse = await searchFuture;
  final truthScore = await truthFuture;

  final offers = [...searchResponse.results]
    ..sort((a, b) => a.totalCost.compareTo(b.totalCost));

  return ProductDetail(
    history: history,
    offers: offers,
    truthScore: truthScore,
  );
});
