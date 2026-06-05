import 'package:flutter/material.dart';

import '../widgets/placeholder_screen.dart';

/// Results screen (USER_FLOWS Flows 2 & 3). Real UI lands in B-03.
///
/// Standalone route (no bottom navigation bar) per
/// `COMPONENT_LIBRARY.md §Navigation`.
class ResultsScreen extends StatelessWidget {
  const ResultsScreen({super.key, this.query});

  /// The search query that produced these results (passed via router extra).
  final String? query;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text(query == null ? 'Results' : 'Results: $query')),
      body: const PlaceholderScreen(
        title: 'Results',
        implementedIn: 'B-03',
        icon: Icons.list_alt_rounded,
      ),
    );
  }
}
