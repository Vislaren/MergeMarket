import 'package:flutter/material.dart';

import '../theme/colours.dart';
import '../theme/spacing.dart';
import '../theme/typography.dart';

/// Full-width rounded search input.
///
/// `COMPONENT_LIBRARY.md → MMSearchBar`. White surface, 1px border, 8px
/// radius, navy border on focus, trailing search icon (or a spinner while
/// [isLoading]). Submitting via the keyboard or tapping the icon calls
/// [onSearch].
class MMSearchBar extends StatelessWidget {
  const MMSearchBar({
    super.key,
    required this.controller,
    required this.onSearch,
    this.hint = 'Search 50+ stores...',
    this.isLoading = false,
  });

  final TextEditingController controller;
  final VoidCallback onSearch;
  final String hint;
  final bool isLoading;

  @override
  Widget build(BuildContext context) {
    return TextField(
      controller: controller,
      textInputAction: TextInputAction.search,
      onSubmitted: (_) => onSearch(),
      style: bodyRegular.copyWith(color: textPrimary),
      decoration: InputDecoration(
        hintText: hint,
        hintStyle: bodyRegular.copyWith(color: textSecondary),
        filled: true,
        fillColor: surfaceWhite,
        contentPadding: const EdgeInsets.symmetric(horizontal: md, vertical: md),
        prefixIcon: const Icon(Icons.search_rounded, color: textSecondary),
        suffixIcon: isLoading
            ? const Padding(
                padding: EdgeInsets.all(md),
                child: SizedBox(
                  width: 18,
                  height: 18,
                  child: CircularProgressIndicator(strokeWidth: 2),
                ),
              )
            : IconButton(
                icon: const Icon(Icons.arrow_forward_rounded, color: primaryNavy),
                onPressed: onSearch,
                tooltip: 'Search',
              ),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(sm),
          borderSide: const BorderSide(color: borderGrey),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(sm),
          borderSide: const BorderSide(color: borderGrey),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(sm),
          borderSide: const BorderSide(color: primaryNavy, width: 1.5),
        ),
      ),
    );
  }
}
