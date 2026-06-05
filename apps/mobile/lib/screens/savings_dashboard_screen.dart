import 'package:flutter/material.dart';

import '../widgets/placeholder_screen.dart';

/// Savings Dashboard screen (USER_FLOWS Flow 7). Real UI lands in B-07.
class SavingsDashboardScreen extends StatelessWidget {
  const SavingsDashboardScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Savings')),
      body: const PlaceholderScreen(
        title: 'Savings Dashboard',
        implementedIn: 'B-07',
        icon: Icons.savings_outlined,
      ),
    );
  }
}
