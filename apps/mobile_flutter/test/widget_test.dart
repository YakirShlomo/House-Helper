import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:house_helper/main.dart';

void main() {
  testWidgets('App starts and shows home screen', (WidgetTester tester) async {
    await tester.pumpWidget(const ProviderScope(child: HouseHelperApp()));
    
    // Wait for initial load
    await tester.pumpAndSettle();
    
    // Verify app title is shown
    expect(find.text('House Helper'), findsOneWidget);
    
    // Verify welcome message is shown
    expect(find.textContaining('Welcome'), findsOneWidget);
  });

  testWidgets('Bottom navigation works', (WidgetTester tester) async {
    await tester.pumpWidget(const ProviderScope(child: HouseHelperApp()));
    await tester.pumpAndSettle();
    
    // Tap on Tasks tab
    await tester.tap(find.byIcon(Icons.task));
    await tester.pumpAndSettle();
    
    // Should show tasks screen
    expect(find.text('Tasks'), findsOneWidget);
    
    // Tap on Shopping tab
    await tester.tap(find.byIcon(Icons.shopping_cart));
    await tester.pumpAndSettle();
    
    // Should show shopping screen
    expect(find.text('Shopping'), findsOneWidget);
  });

  testWidgets('Quick actions are displayed', (WidgetTester tester) async {
    await tester.pumpWidget(const ProviderScope(child: HouseHelperApp()));
    await tester.pumpAndSettle();
    
    // Verify quick action cards are shown
    expect(find.text('Add Task'), findsOneWidget);
    expect(find.text('Start Timer'), findsOneWidget);
    expect(find.text('Add Item'), findsOneWidget);
    expect(find.text('Add Bill'), findsOneWidget);
  });
}
