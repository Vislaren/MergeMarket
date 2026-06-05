import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/wishlist.dart';
import '../services/wishlist_repository.dart';
import 'search_providers.dart';

/// The wishlist data-layer repository, wired to the shared HTTP client from
/// [httpClientProvider].
final wishlistRepositoryProvider = Provider<WishlistRepository>((ref) {
  return WishlistRepository(client: ref.watch(httpClientProvider));
});

/// The user's wishlist as an [AsyncValue] so the Wishlist screen can render
/// loading / data / error from one source of truth.
///
/// `ref.invalidate(wishlistProvider)` re-fetches after an add/remove (the B-02
/// mock is stateless, so mutations don't change subsequent GETs — the refetch
/// keeps the data layer honest for the real backend at B-11).
final wishlistProvider = FutureProvider<Wishlist>((ref) async {
  return ref.watch(wishlistRepositoryProvider).list();
});

/// Wishlist add/remove actions, exposed as a single object the UI calls.
///
/// Kept thin: it forwards to the repository and invalidates [wishlistProvider]
/// so the list reflects the change. Errors propagate as [ApiException] for the
/// caller to surface.
class WishlistActions {
  WishlistActions(this._ref);

  final Ref _ref;

  /// Adds [productId] to the wishlist, then refreshes the list.
  Future<void> add(String productId) async {
    await _ref.read(wishlistRepositoryProvider).add(productId);
    _ref.invalidate(wishlistProvider);
  }

  /// Removes the entry [wishlistId], then refreshes the list.
  Future<void> remove(String wishlistId) async {
    await _ref.read(wishlistRepositoryProvider).remove(wishlistId);
    _ref.invalidate(wishlistProvider);
  }
}

/// Provides the [WishlistActions] for the current container.
final wishlistActionsProvider = Provider<WishlistActions>(WishlistActions.new);
