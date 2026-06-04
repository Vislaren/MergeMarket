import 'package:flutter/material.dart';

import '../widgets/placeholder_screen.dart';

/// Product Detail screen (USER_FLOWS Flows 2, 3, 4, 6). Real UI lands in B-04.
///
/// Standalone route (no bottom navigation bar). Reached via `/product/:id`,
/// including as the deep-link target of a price-drop notification (Flow 6).
class ProductDetailScreen extends StatelessWidget {
  const ProductDetailScreen({super.key, required this.productId});

  final String productId;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Product Detail')),
      body: PlaceholderScreen(
        title: 'Product $productId',
        implementedIn: 'B-04',
        icon: Icons.shopping_bag_outlined,
      ),
    );
  }
}
