import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/models.dart';
import '../services/api_client.dart';

// Auth state provider
final authStateProvider = StateNotifierProvider<AuthStateNotifier, AuthState>((ref) {
  return AuthStateNotifier(ref.read(apiClientProvider));
});

class AuthStateNotifier extends StateNotifier<AuthState> {
  final ApiClient _apiClient;

  AuthStateNotifier(this._apiClient) : super(const AuthState.unauthenticated());

  Future<void> login(String email, String password) async {
    state = const AuthState.loading();
    try {
      await _apiClient.login(email, password);
      final user = await _apiClient.getCurrentUser();
      state = AuthState.authenticated(user);
    } catch (e) {
      state = AuthState.error(e.toString());
    }
  }

  Future<void> signup(String email, String password, String name) async {
    state = const AuthState.loading();
    try {
      await _apiClient.signup(email, password, name);
      final user = await _apiClient.getCurrentUser();
      state = AuthState.authenticated(user);
    } catch (e) {
      state = AuthState.error(e.toString());
    }
  }

  Future<void> logout() async {
    state = const AuthState.unauthenticated();
  }

  Future<void> checkAuthStatus() async {
    try {
      final user = await _apiClient.getCurrentUser();
      state = AuthState.authenticated(user);
    } catch (e) {
      state = const AuthState.unauthenticated();
    }
  }
}

// Auth state sealed class
sealed class AuthState {
  const AuthState();

  const factory AuthState.loading() = AuthLoading;
  const factory AuthState.authenticated(User user) = AuthAuthenticated;
  const factory AuthState.unauthenticated() = AuthUnauthenticated;
  const factory AuthState.error(String message) = AuthError;
}

class AuthLoading extends AuthState {
  const AuthLoading();
}

class AuthAuthenticated extends AuthState {
  final User user;
  const AuthAuthenticated(this.user);
}

class AuthUnauthenticated extends AuthState {
  const AuthUnauthenticated();
}

class AuthError extends AuthState {
  final String message;
  const AuthError(this.message);
}

// Tasks provider
final tasksProvider = StateNotifierProvider<TasksNotifier, AsyncValue<List<Task>>>((ref) {
  return TasksNotifier(ref.read(apiClientProvider));
});

class TasksNotifier extends StateNotifier<AsyncValue<List<Task>>> {
  final ApiClient _apiClient;

  TasksNotifier(this._apiClient) : super(const AsyncValue.loading()) {
    loadTasks();
  }

  Future<void> loadTasks() async {
    state = const AsyncValue.loading();
    try {
      final tasks = await _apiClient.getTasks();
      state = AsyncValue.data(tasks);
    } catch (e, stackTrace) {
      state = AsyncValue.error(e, stackTrace);
    }
  }

  Future<void> addTask(Task task) async {
    try {
      final newTask = await _apiClient.createTask(task);
      state = state.whenData((tasks) => [...tasks, newTask]);
    } catch (e) {
      // Handle error
    }
  }

  Future<void> updateTask(Task task) async {
    try {
      final updatedTask = await _apiClient.updateTask(task.id, task);
      state = state.whenData((tasks) => 
        tasks.map((t) => t.id == task.id ? updatedTask : t).toList());
    } catch (e) {
      // Handle error
    }
  }

  Future<void> deleteTask(String id) async {
    try {
      await _apiClient.deleteTask(id);
      state = state.whenData((tasks) => tasks.where((t) => t.id != id).toList());
    } catch (e) {
      // Handle error
    }
  }
}

// Shopping provider
final shoppingListsProvider = StateNotifierProvider<ShoppingListsNotifier, AsyncValue<List<ShoppingList>>>((ref) {
  return ShoppingListsNotifier(ref.read(apiClientProvider));
});

class ShoppingListsNotifier extends StateNotifier<AsyncValue<List<ShoppingList>>> {
  final ApiClient _apiClient;

  ShoppingListsNotifier(this._apiClient) : super(const AsyncValue.loading()) {
    loadShoppingLists();
  }

  Future<void> loadShoppingLists() async {
    state = const AsyncValue.loading();
    try {
      final lists = await _apiClient.getShoppingLists();
      state = AsyncValue.data(lists);
    } catch (e, stackTrace) {
      state = AsyncValue.error(e, stackTrace);
    }
  }

  Future<void> addShoppingList(ShoppingList list) async {
    try {
      final newList = await _apiClient.createShoppingList(list);
      state = state.whenData((lists) => [...lists, newList]);
    } catch (e) {
      // Handle error
    }
  }

  Future<void> addItemToList(String listId, ShoppingItem item) async {
    try {
      final newItem = await _apiClient.addShoppingItem(listId, item);
      state = state.whenData((lists) => 
        lists.map((list) {
          if (list.id == listId) {
            return list.copyWith(items: [...list.items, newItem]);
          }
          return list;
        }).toList());
    } catch (e) {
      // Handle error
    }
  }

  Future<void> updateShoppingItem(String listId, ShoppingItem item) async {
    try {
      final updatedItem = await _apiClient.updateShoppingItem(listId, item.id, item);
      state = state.whenData((lists) => 
        lists.map((list) {
          if (list.id == listId) {
            return list.copyWith(
              items: list.items.map((i) => i.id == item.id ? updatedItem : i).toList()
            );
          }
          return list;
        }).toList());
    } catch (e) {
      // Handle error
    }
  }
}

// Bills provider
final billsProvider = StateNotifierProvider<BillsNotifier, AsyncValue<List<Bill>>>((ref) {
  return BillsNotifier(ref.read(apiClientProvider));
});

class BillsNotifier extends StateNotifier<AsyncValue<List<Bill>>> {
  final ApiClient _apiClient;

  BillsNotifier(this._apiClient) : super(const AsyncValue.loading()) {
    loadBills();
  }

  Future<void> loadBills() async {
    state = const AsyncValue.loading();
    try {
      final bills = await _apiClient.getBills();
      state = AsyncValue.data(bills);
    } catch (e, stackTrace) {
      state = AsyncValue.error(e, stackTrace);
    }
  }

  Future<void> addBill(Bill bill) async {
    try {
      final newBill = await _apiClient.createBill(bill);
      state = state.whenData((bills) => [...bills, newBill]);
    } catch (e) {
      // Handle error
    }
  }

  Future<void> payBill(String id) async {
    try {
      final paidBill = await _apiClient.payBill(id);
      state = state.whenData((bills) => 
        bills.map((b) => b.id == id ? paidBill : b).toList());
    } catch (e) {
      // Handle error
    }
  }
}

// Timers provider
final timersProvider = StateNotifierProvider<TimersNotifier, AsyncValue<List<HouseTimer>>>((ref) {
  return TimersNotifier(ref.read(apiClientProvider));
});

class TimersNotifier extends StateNotifier<AsyncValue<List<HouseTimer>>> {
  final ApiClient _apiClient;

  TimersNotifier(this._apiClient) : super(const AsyncValue.loading()) {
    loadActiveTimers();
  }

  Future<void> loadActiveTimers() async {
    state = const AsyncValue.loading();
    try {
      final timers = await _apiClient.getActiveTimers();
      state = AsyncValue.data(timers);
    } catch (e, stackTrace) {
      state = AsyncValue.error(e, stackTrace);
    }
  }

  Future<void> startTimer(String type, Duration duration, {String? taskId}) async {
    try {
      final timer = await _apiClient.startTimer(type, duration, taskId: taskId);
      state = state.whenData((timers) => [...timers, timer]);
    } catch (e) {
      // Handle error
    }
  }

  Future<void> cancelTimer(String id) async {
    try {
      await _apiClient.cancelTimer(id);
      state = state.whenData((timers) => timers.where((t) => t.id != id).toList());
    } catch (e) {
      // Handle error
    }
  }
}

// Theme mode provider
final themeModeProvider = StateProvider<ThemeMode>((ref) => ThemeMode.system);

// Locale provider
final localeProvider = StateProvider<Locale>((ref) => const Locale('en'));

// Notifications enabled provider
final notificationsEnabledProvider = StateProvider<bool>((ref) => true);
