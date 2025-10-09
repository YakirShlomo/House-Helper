import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/app_providers.dart';
import '../models/models.dart';
import '../generated/l10n/app_localizations.dart';
import 'home_screen.dart';

class TasksScreen extends ConsumerWidget {
  const TasksScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final tasksAsync = ref.watch(tasksProvider);

    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.tasksTitle),
        actions: [
          IconButton(
            icon: const Icon(Icons.add),
            onPressed: () => _showAddTaskDialog(context, ref),
          ),
        ],
      ),
      body: tasksAsync.when(
        data: (tasks) {
          if (tasks.isEmpty) {
            return Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const Icon(Icons.task, size: 64, color: Colors.grey),
                  const SizedBox(height: 16),
                  Text(
                    'No tasks yet',
                    style: Theme.of(context).textTheme.titleLarge,
                  ),
                  const SizedBox(height: 8),
                  Text(
                    'Add your first task to get started',
                    style: Theme.of(context).textTheme.bodyMedium,
                  ),
                  const SizedBox(height: 24),
                  ElevatedButton.icon(
                    onPressed: () => _showAddTaskDialog(context, ref),
                    icon: const Icon(Icons.add),
                    label: Text(l10n.addTask),
                  ),
                ],
              ),
            );
          }

          return ListView.builder(
            padding: const EdgeInsets.all(16),
            itemCount: tasks.length,
            itemBuilder: (context, index) {
              final task = tasks[index];
              return Card(
                child: ListTile(
                  leading: IconButton(
                    icon: Icon(
                      task.status == TaskStatus.completed
                          ? Icons.check_circle
                          : Icons.radio_button_unchecked,
                      color: task.status == TaskStatus.completed
                          ? Colors.green
                          : null,
                    ),
                    onPressed: () {
                      final updatedTask = task.copyWith(
                        status: task.status == TaskStatus.completed
                            ? TaskStatus.pending
                            : TaskStatus.completed,
                        completedAt: task.status == TaskStatus.completed
                            ? null
                            : DateTime.now(),
                      );
                      ref.read(tasksProvider.notifier).updateTask(updatedTask);
                    },
                  ),
                  title: Text(
                    task.title,
                    style: TextStyle(
                      decoration: task.status == TaskStatus.completed
                          ? TextDecoration.lineThrough
                          : null,
                    ),
                  ),
                  subtitle: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      if (task.description != null)
                        Text(task.description!),
                      if (task.dueAt != null)
                        Text(
                          'Due: ${_formatDate(task.dueAt!)}',
                          style: TextStyle(
                            color: task.dueAt!.isBefore(DateTime.now())
                                ? Colors.red
                                : Colors.grey,
                          ),
                        ),
                    ],
                  ),
                  trailing: PopupMenuButton(
                    itemBuilder: (context) => [
                      PopupMenuItem(
                        child: const Text('Edit'),
                        onTap: () => _showEditTaskDialog(context, ref, task),
                      ),
                      PopupMenuItem(
                        child: const Text('Delete'),
                        onTap: () => ref.read(tasksProvider.notifier).deleteTask(task.id),
                      ),
                    ],
                  ),
                ),
              );
            },
          );
        },
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (error, stack) => Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const Icon(Icons.error, size: 64, color: Colors.red),
              const SizedBox(height: 16),
              Text('Error: $error'),
              const SizedBox(height: 16),
              ElevatedButton(
                onPressed: () => ref.read(tasksProvider.notifier).loadTasks(),
                child: const Text('Retry'),
              ),
            ],
          ),
        ),
      ),
      bottomNavigationBar: const AppBottomNavigationBar(currentIndex: 1),
    );
  }

  String _formatDate(DateTime date) {
    return '${date.day}/${date.month}/${date.year}';
  }

  void _showAddTaskDialog(BuildContext context, WidgetRef ref) {
    _showTaskDialog(context, ref, null);
  }

  void _showEditTaskDialog(BuildContext context, WidgetRef ref, Task task) {
    _showTaskDialog(context, ref, task);
  }

  void _showTaskDialog(BuildContext context, WidgetRef ref, Task? existingTask) {
    final titleController = TextEditingController(text: existingTask?.title ?? '');
    final descriptionController = TextEditingController(text: existingTask?.description ?? '');
    DateTime? selectedDueDate = existingTask?.dueAt;

    showDialog(
      context: context,
      builder: (context) => StatefulBuilder(
        builder: (context, setState) => AlertDialog(
          title: Text(existingTask == null ? 'Add Task' : 'Edit Task'),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              TextField(
                controller: titleController,
                decoration: const InputDecoration(
                  labelText: 'Task Title',
                  border: OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 16),
              TextField(
                controller: descriptionController,
                decoration: const InputDecoration(
                  labelText: 'Description (optional)',
                  border: OutlineInputBorder(),
                ),
                maxLines: 3,
              ),
              const SizedBox(height: 16),
              Row(
                children: [
                  Expanded(
                    child: Text(
                      selectedDueDate == null
                          ? 'No due date'
                          : 'Due: ${_formatDate(selectedDueDate!)}',
                    ),
                  ),
                  TextButton(
                    onPressed: () async {
                      final date = await showDatePicker(
                        context: context,
                        initialDate: selectedDueDate ?? DateTime.now(),
                        firstDate: DateTime.now(),
                        lastDate: DateTime.now().add(const Duration(days: 365)),
                      );
                      if (date != null) {
                        setState(() {
                          selectedDueDate = date;
                        });
                      }
                    },
                    child: const Text('Set Due Date'),
                  ),
                ],
              ),
            ],
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.of(context).pop(),
              child: const Text('Cancel'),
            ),
            ElevatedButton(
              onPressed: () {
                if (titleController.text.trim().isEmpty) return;

                final authState = ref.read(authStateProvider);
                final householdId = authState is AuthAuthenticated 
                    ? authState.user.currentHouseholdId ?? 'default'
                    : 'default';
                
                final task = Task(
                  id: existingTask?.id ?? DateTime.now().millisecondsSinceEpoch.toString(),
                  title: titleController.text.trim(),
                  description: descriptionController.text.trim().isEmpty
                      ? null
                      : descriptionController.text.trim(),
                  createdAt: existingTask?.createdAt ?? DateTime.now(),
                  dueAt: selectedDueDate,
                  householdId: householdId,
                  status: existingTask?.status ?? TaskStatus.pending,
                  completedAt: existingTask?.completedAt,
                );

                if (existingTask == null) {
                  ref.read(tasksProvider.notifier).addTask(task);
                } else {
                  ref.read(tasksProvider.notifier).updateTask(task);
                }

                Navigator.of(context).pop();
              },
              child: Text(existingTask == null ? 'Add' : 'Save'),
            ),
          ],
        ),
      ),
    );
  }
}
