import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../router/app_router.dart';
import '../theme/colours.dart';
import '../theme/spacing.dart';
import '../theme/typography.dart';
import '../widgets/mm_search_bar.dart';

/// Home / Search screen — the entry point of the real-time search flow
/// (USER_FLOWS Flow 2). Hosts the query input and quick "trending" shortcuts;
/// submitting a query navigates to the Results screen, which runs the search
/// and renders the loading / results / error states.
class HomeScreen extends ConsumerStatefulWidget {
  const HomeScreen({super.key});

  @override
  ConsumerState<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends ConsumerState<HomeScreen> {
  final _controller = TextEditingController();

  /// Static discovery shortcuts. These are entry-point hints, not API data —
  /// a real "trending" feed is out of scope for B-03.
  static const _trending = [
    'iPhone 15 Pro',
    'Air Jordan 1',
    'Dyson V15',
    'Samsung Galaxy A54',
    'PS5',
  ];

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  void _search([String? term]) {
    final query = (term ?? _controller.text).trim();
    if (query.isEmpty) return;
    context.go('${Routes.results}?q=${Uri.encodeQueryComponent(query)}');
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: backgroundLight,
      body: SafeArea(
        child: SingleChildScrollView(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              _Header(controller: _controller, onSearch: _search),
              Padding(
                padding: const EdgeInsets.all(md),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text('Trending Searches',
                        style: headingMedium.copyWith(color: textPrimary)),
                    const SizedBox(height: md),
                    Wrap(
                      spacing: sm,
                      runSpacing: sm,
                      children: [
                        for (final term in _trending)
                          ActionChip(
                            avatar: const Icon(Icons.trending_up_rounded,
                                size: 16, color: primaryNavy),
                            label: Text(term),
                            onPressed: () => _search(term),
                          ),
                      ],
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

/// Navy header band with a greeting and the search bar.
class _Header extends StatelessWidget {
  const _Header({required this.controller, required this.onSearch});

  final TextEditingController controller;
  final VoidCallback onSearch;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      color: primaryNavy,
      padding: const EdgeInsets.fromLTRB(md, lg, md, lg),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('MergeMarket',
              style: headingLarge.copyWith(color: surfaceWhite)),
          const SizedBox(height: xs),
          Text('Find the best deals today',
              style: bodyRegular.copyWith(color: borderGrey)),
          const SizedBox(height: md),
          MMSearchBar(controller: controller, onSearch: onSearch),
        ],
      ),
    );
  }
}
