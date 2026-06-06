import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/savings.dart';
import '../services/savings_repository.dart';
import 'search_providers.dart';

/// The savings data-layer repository, wired to the shared HTTP client.
final savingsRepositoryProvider = Provider<SavingsRepository>((ref) {
  return SavingsRepository(client: ref.watch(httpClientProvider));
});

/// The current user's savings summary for the Savings Dashboard.
final savingsProvider = FutureProvider<SavingsSummary>((ref) async {
  return ref.watch(savingsRepositoryProvider).getSavings();
});
