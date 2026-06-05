// Widget tests for MMSearchBar (B-03).
//
// Source of truth: project_docs/ui/COMPONENT_LIBRARY.md → MMSearchBar.
// Test artefacts: docs/testing/session-06/unit/.

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mergemarket/widgets/mm_search_bar.dart';

void main() {
  group('B-03 MMSearchBar widget tests', () {
    testWidgets('TC-06-U-018: tapping the action icon fires onSearch',
        (tester) async {
      final controller = TextEditingController();
      addTearDown(controller.dispose);
      var searches = 0;

      await tester.pumpWidget(MaterialApp(
        home: Scaffold(
          body: MMSearchBar(
            controller: controller,
            onSearch: () => searches++,
          ),
        ),
      ));

      await tester.enterText(find.byType(TextField), 'galaxy');
      await tester.tap(find.byIcon(Icons.arrow_forward_rounded));
      expect(searches, 1);
      expect(controller.text, 'galaxy');
    });

    testWidgets('TC-06-U-019: submitting from the keyboard fires onSearch',
        (tester) async {
      final controller = TextEditingController();
      addTearDown(controller.dispose);
      var searched = false;

      await tester.pumpWidget(MaterialApp(
        home: Scaffold(
          body: MMSearchBar(
            controller: controller,
            onSearch: () => searched = true,
          ),
        ),
      ));

      await tester.enterText(find.byType(TextField), 'phone');
      await tester.testTextInput.receiveAction(TextInputAction.search);
      expect(searched, isTrue);
    });

    testWidgets('TC-06-U-020: loading shows a spinner instead of the icon',
        (tester) async {
      final controller = TextEditingController();
      addTearDown(controller.dispose);

      await tester.pumpWidget(MaterialApp(
        home: Scaffold(
          body: MMSearchBar(
            controller: controller,
            onSearch: () {},
            isLoading: true,
          ),
        ),
      ));

      expect(find.byType(CircularProgressIndicator), findsOneWidget);
      expect(find.byIcon(Icons.arrow_forward_rounded), findsNothing);
    });
  });
}
