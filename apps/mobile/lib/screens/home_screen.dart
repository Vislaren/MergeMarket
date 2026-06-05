import 'package:flutter/material.dart';

import '../widgets/placeholder_screen.dart';

/// Home / Search screen (USER_FLOWS Flow 2). Real UI lands in B-03.
class HomeScreen extends StatelessWidget {
  const HomeScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('MergeMarket')),
      body: const PlaceholderScreen(
        title: 'Search',
        implementedIn: 'B-03',
        icon: Icons.search_rounded,
      ),
    );
  }
}
