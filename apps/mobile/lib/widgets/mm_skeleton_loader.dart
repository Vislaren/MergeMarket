import 'package:flutter/material.dart';

import '../theme/colours.dart';

/// Animated shimmer placeholder shown while content loads.
///
/// `COMPONENT_LIBRARY.md → MMSkeletonLoader`. Renders a sweeping highlight
/// between [backgroundLight] and [borderGrey]. Use it in the same layout as
/// the real content it stands in for (e.g. card-shaped blocks on Results).
class MMSkeletonLoader extends StatefulWidget {
  const MMSkeletonLoader({
    super.key,
    required this.width,
    required this.height,
    this.borderRadius = 8.0,
  });

  final double width;
  final double height;
  final double borderRadius;

  @override
  State<MMSkeletonLoader> createState() => _MMSkeletonLoaderState();
}

class _MMSkeletonLoaderState extends State<MMSkeletonLoader>
    with SingleTickerProviderStateMixin {
  late final AnimationController _controller = AnimationController(
    vsync: this,
    duration: const Duration(milliseconds: 1200),
  )..repeat();

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: _controller,
      builder: (context, _) {
        // Slide a highlight band from left (-1) to right (2) across the block.
        final offset = (_controller.value * 3) - 1;
        return Container(
          width: widget.width,
          height: widget.height,
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(widget.borderRadius),
            gradient: LinearGradient(
              begin: Alignment(offset - 1, 0),
              end: Alignment(offset + 1, 0),
              colors: const [backgroundLight, borderGrey, backgroundLight],
              stops: const [0.1, 0.5, 0.9],
            ),
          ),
        );
      },
    );
  }
}
