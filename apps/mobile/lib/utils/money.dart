import 'package:intl/intl.dart';

/// Formats [amount] as `"CURRENCY 1,234"` (e.g. `XAF 48,500`).
///
/// Thousands are grouped and decimals are shown only when non-zero, matching
/// the results-screen design (whole-number XAF prices). Centralised here so
/// every screen renders money identically.
String formatMoney(double amount, String currency) {
  final number = NumberFormat('#,##0.##').format(amount);
  return currency.isEmpty ? number : '$currency $number';
}
