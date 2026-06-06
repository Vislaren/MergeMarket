import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/savings.dart';
import '../services/savings_repository.dart';
import 'auth_providers.dart';

/// The savings data-layer repository, wired to the authenticated HTTP client so
/// savings calls carry the Bearer token Kong requires (B-11).
final savingsRepositoryProvider = Provider<SavingsRepository>((ref) {
  return SavingsRepository(client: ref.watch(authedHttpClientProvider));
});

/// The current user's savings summary for the Savings Dashboard.
final savingsProvider = FutureProvider<SavingsSummary>((ref) async {
  return ref.watch(savingsRepositoryProvider).getSavings();
});
