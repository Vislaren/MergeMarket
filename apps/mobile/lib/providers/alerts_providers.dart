import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/alert.dart';
import '../services/alerts_repository.dart';
import 'auth_providers.dart';

/// The alerts data-layer repository, wired to the authenticated HTTP client so
/// alert calls carry the Bearer token Kong requires (B-11).
final alertsRepositoryProvider = Provider<AlertsRepository>((ref) {
  return AlertsRepository(client: ref.watch(authedHttpClientProvider));
});

/// The user's alerts as an [AsyncValue] so the Alerts screen can render
/// loading / data / error from one source of truth.
///
/// `ref.invalidate(alertsProvider)` re-fetches after a create/remove.
final alertsProvider = FutureProvider<AlertList>((ref) async {
  return ref.watch(alertsRepositoryProvider).list();
});

/// Alert create/remove actions, exposed as a single object the UI calls.
///
/// Kept thin: it forwards to the repository and invalidates [alertsProvider] so
/// the list reflects the change. Errors propagate as [ApiException].
class AlertsActions {
  AlertsActions(this._ref);

  final Ref _ref;

  /// Creates a price alert for [productId] at [thresholdPrice], then refreshes.
  Future<void> create({
    required String productId,
    required double thresholdPrice,
    required String currency,
  }) async {
    await _ref.read(alertsRepositoryProvider).create(
          productId: productId,
          thresholdPrice: thresholdPrice,
          currency: currency,
        );
    _ref.invalidate(alertsProvider);
  }

  /// Removes the alert [alertId], then refreshes the list.
  Future<void> remove(String alertId) async {
    await _ref.read(alertsRepositoryProvider).remove(alertId);
    _ref.invalidate(alertsProvider);
  }
}

/// Provides the [AlertsActions] for the current container.
final alertsActionsProvider = Provider<AlertsActions>(AlertsActions.new);
