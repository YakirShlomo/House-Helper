import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../providers/app_providers.dart';
import '../generated/l10n/app_localizations.dart';
import '../models/models.dart';

class HomeScreen extends ConsumerWidget {
  const HomeScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final tasksAsync = ref.watch(tasksProvider);
    final timersAsync = ref.watch(timersProvider);

    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.appTitle),
        actions: [
          IconButton(
            icon: const Icon(Icons.settings),
            onPressed: () => context.go('/settings'),
          ),
        ],
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Welcome section
            Card(
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      l10n.welcome,
                      style: Theme.of(context).textTheme.headlineSmall,
                    ),
                    const SizedBox(height: 8),
                    Text(
                      l10n.manageHousehold,
                      style: Theme.of(context).textTheme.bodyMedium,
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 24),

            // Quick actions
            Text(
              'Quick Actions',
              style: Theme.of(context).textTheme.titleLarge,
            ),
            const SizedBox(height: 16),
            Row(
              children: [
                Expanded(
                  child: _QuickActionCard(
                    icon: Icons.add_task,
                    title: l10n.addTask,
                    onTap: () => context.go('/tasks/add'),
                  ),
                ),
                const SizedBox(width: 16),
                Expanded(
                  child: _QuickActionCard(
                    icon: Icons.timer,
                    title: l10n.startTimer,
                    onTap: () => _showTimerDialog(context, ref),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 16),
            Row(
              children: [
                Expanded(
                  child: _QuickActionCard(
                    icon: Icons.shopping_cart,
                    title: l10n.addShoppingItem,
                    onTap: () => context.go('/shopping'),
                  ),
                ),
                const SizedBox(width: 16),
                Expanded(
                  child: _QuickActionCard(
                    icon: Icons.receipt,
                    title: l10n.addBill,
                    onTap: () => context.go('/bills/add'),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 24),

            // Active timers
            if (timersAsync.hasValue && timersAsync.value!.isNotEmpty) ...[
              Text(
                'Active Timers',
                style: Theme.of(context).textTheme.titleLarge,
              ),
              const SizedBox(height: 16),
              ...timersAsync.value!.map((timer) => Card(
                child: ListTile(
                  leading: const Icon(Icons.timer),
                  title: Text(timer.title),
                  subtitle: Text(timer.type),
                  trailing: IconButton(
                    icon: const Icon(Icons.stop),
                    onPressed: () => ref.read(timersProvider.notifier).cancelTimer(timer.id),
                  ),
                ),
              )),
              const SizedBox(height: 24),
            ],

            // Recent tasks
            Text(
              'Recent Tasks',
              style: Theme.of(context).textTheme.titleLarge,
            ),
            const SizedBox(height: 16),
            tasksAsync.when(
              data: (tasks) {
                final recentTasks = tasks.take(3).toList();
                if (recentTasks.isEmpty) {
                  return const Card(
                    child: Padding(
                      padding: EdgeInsets.all(16),
                      child: Text('No tasks yet. Add your first task!'),
                    ),
                  );
                }
                return Column(
                  children: recentTasks.map((task) => Card(
                    child: ListTile(
                      leading: Icon(
                        task.status == TaskStatus.completed 
                          ? Icons.check_circle 
                          : Icons.radio_button_unchecked,
                        color: task.status == TaskStatus.completed 
                          ? Colors.green 
                          : null,
                      ),
                      title: Text(task.title),
                      subtitle: task.description != null ? Text(task.description!) : null,
                      onTap: () => context.go('/tasks/${task.id}'),
                    ),
                  )).toList(),
                );
              },
              loading: () => const CircularProgressIndicator(),
              error: (error, stack) => Text('Error: $error'),
            ),
          ],
        ),
      ),
      bottomNavigationBar: const AppBottomNavigationBar(currentIndex: 0),
    );
  }

  void _showTimerDialog(BuildContext context, WidgetRef ref) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Start Timer'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              leading: const Icon(Icons.local_laundry_service),
              title: const Text('Laundry Timer'),
              subtitle: const Text('45 minutes'),
              onTap: () {
                ref.read(timersProvider.notifier).startTimer(
                  'laundry',
                  const Duration(minutes: 45),
                );
                Navigator.of(context).pop();
              },
            ),
            ListTile(
              leading: const Icon(Icons.kitchen),
              title: const Text('Cooking Timer'),
              subtitle: const Text('30 minutes'),
              onTap: () {
                ref.read(timersProvider.notifier).startTimer(
                  'cooking',
                  const Duration(minutes: 30),
                );
                Navigator.of(context).pop();
              },
            ),
          ],
        ),
      ),
    );
  }
}

class _QuickActionCard extends StatelessWidget {
  final IconData icon;
  final String title;
  final VoidCallback onTap;

  const _QuickActionCard({
    required this.icon,
    required this.title,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            children: [
              Icon(icon, size: 32),
              const SizedBox(height: 8),
              Text(
                title,
                textAlign: TextAlign.center,
                style: Theme.of(context).textTheme.bodyMedium,
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class AppBottomNavigationBar extends StatelessWidget {
  final int currentIndex;

  const AppBottomNavigationBar({
    super.key,
    required this.currentIndex,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);

    return BottomNavigationBar(
      currentIndex: currentIndex,
      type: BottomNavigationBarType.fixed,
      items: [
        BottomNavigationBarItem(
          icon: const Icon(Icons.home),
          label: l10n.homeTitle,
        ),
        BottomNavigationBarItem(
          icon: const Icon(Icons.task),
          label: l10n.tasksTitle,
        ),
        BottomNavigationBarItem(
          icon: const Icon(Icons.shopping_cart),
          label: l10n.shoppingTitle,
        ),
        BottomNavigationBarItem(
          icon: const Icon(Icons.receipt),
          label: l10n.billsTitle,
        ),
        BottomNavigationBarItem(
          icon: const Icon(Icons.history),
          label: l10n.activityTitle,
        ),
      ],
      onTap: (index) {
        switch (index) {
          case 0:
            context.go('/');
            break;
          case 1:
            context.go('/tasks');
            break;
          case 2:
            context.go('/shopping');
            break;
          case 3:
            context.go('/bills');
            break;
          case 4:
            context.go('/activity');
            break;
        }
      },
    );
  }
}
