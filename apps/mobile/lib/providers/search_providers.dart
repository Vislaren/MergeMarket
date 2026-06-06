import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:http/http.dart' as http;

import '../models/product.dart';
import '../services/search_repository.dart';
import 'auth_providers.dart';

/// Shared HTTP client for the data layer. Closed automatically when the
/// provider container is disposed. Override this in tests with a `MockClient`.
final httpClientProvider = Provider<http.Client>((ref) {
  final client = http.Client();
  ref.onDispose(client.close);
  return client;
});

/// The Search data-layer repository, wired to the authenticated HTTP client so
/// search calls carry the Bearer token Kong requires (B-11).
final searchRepositoryProvider = Provider<SearchRepository>((ref) {
  return SearchRepository(client: ref.watch(authedHttpClientProvider));
});

/// Runs a search for [query] and exposes it as an [AsyncValue] so the Results
/// screen can render loading / data / error from one source of truth.
///
/// Keyed by query string (`.family`) so each distinct search is cached and
/// retried independently; `ref.invalidate(searchResultsProvider(query))`
/// re-runs it for the retry button.
final searchResultsProvider =
    FutureProvider.family<SearchResponse, String>((ref, query) async {
  return ref.watch(searchRepositoryProvider).search(query);
});

/// How the Results list is ordered. The three modes map to the design's
/// "Best Price / Top Rated / Fastest Ship" segmented control.
enum ResultSort {
  /// Lowest total cost (price + shipping) first — the default.
  bestPrice,

  /// Highest AI Deal Meter score first.
  topRated,

  /// Lowest shipping cost first.
  fastestShip,
}

/// Holds the currently selected ordering for the Results list.
class ResultSortNotifier extends Notifier<ResultSort> {
  @override
  ResultSort build() => ResultSort.bestPrice;

  /// Switches the active ordering.
  void select(ResultSort sort) => state = sort;
}

/// Currently selected ordering for the Results list.
final resultSortProvider =
    NotifierProvider<ResultSortNotifier, ResultSort>(ResultSortNotifier.new);

/// Returns a new list of [results] ordered according to [sort].
List<Product> sortResults(List<Product> results, ResultSort sort) {
  final sorted = [...results];
  switch (sort) {
    case ResultSort.bestPrice:
      sorted.sort((a, b) => a.totalCost.compareTo(b.totalCost));
    case ResultSort.topRated:
      sorted.sort((a, b) => b.dealScore.compareTo(a.dealScore));
    case ResultSort.fastestShip:
      sorted.sort((a, b) => a.shipping.compareTo(b.shipping));
  }
  return sorted;
}
