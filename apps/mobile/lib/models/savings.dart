/// One purchase event that contributed to the user's savings total.
class SavingsTransaction {
  const SavingsTransaction({
    required this.productId,
    required this.title,
    required this.saved,
    required this.boughtAt,
  });

  /// Decodes a transaction from the `/api/v1/savings` contract.
  factory SavingsTransaction.fromJson(Map<String, dynamic> json) {
    final saved = json['saved'];
    return SavingsTransaction(
      productId: json['product_id'] as String? ?? '',
      title: json['title'] as String? ?? '',
      saved: saved is num ? saved.toDouble() : 0,
      boughtAt:
          DateTime.tryParse(json['bought_at'] as String? ?? '') ??
          DateTime.fromMillisecondsSinceEpoch(0),
    );
  }

  final String productId;
  final String title;
  final double saved;
  final DateTime boughtAt;
}

/// The full Savings Dashboard payload.
class SavingsSummary {
  const SavingsSummary({
    required this.totalSaved,
    required this.currency,
    required this.transactions,
  });

  /// Decodes the `/api/v1/savings` response with safe zero-value fallbacks.
  factory SavingsSummary.fromJson(Map<String, dynamic> json) {
    final rawTransactions = json['transactions'];
    final totalSaved = json['total_saved'];
    return SavingsSummary(
      totalSaved: totalSaved is num ? totalSaved.toDouble() : 0,
      currency: json['currency'] as String? ?? '',
      transactions: rawTransactions is List
          ? rawTransactions
                .whereType<Map<String, dynamic>>()
                .map(SavingsTransaction.fromJson)
                .toList()
          : const [],
    );
  }

  final double totalSaved;
  final String currency;
  final List<SavingsTransaction> transactions;

  /// Level 1-10, advancing once per 50,000 saved.
  int get savingsLevel => (totalSaved ~/ _levelSize + 1).clamp(1, 10);

  /// Progress within the current savings level, capped at 100%.
  double get progressToNextLevel {
    if (savingsLevel == 10) return 1;
    final currentLevelStart = (savingsLevel - 1) * _levelSize;
    return ((totalSaved - currentLevelStart) / _levelSize).clamp(0, 1);
  }

  /// Amount left before the next level. Returns zero at level 10.
  double get remainingToNextLevel {
    if (savingsLevel == 10) return 0;
    final nextLevelStart = savingsLevel * _levelSize;
    return (nextLevelStart - totalSaved).clamp(0, _levelSize).toDouble();
  }

  static const int _levelSize = 50000;
}
