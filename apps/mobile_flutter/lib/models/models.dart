import 'package:freezed_annotation/freezed_annotation.dart';

part 'models.freezed.dart';
part 'models.g.dart';

@freezed
class Task with _$Task {
  const factory Task({
    required String id,
    required String title,
    String? description,
    required DateTime createdAt,
    DateTime? completedAt,
    DateTime? dueAt,
    required String householdId,
    String? assignedToUserId,
    required TaskStatus status,
    TaskPriority? priority,
    List<String>? tags,
  }) = _Task;

  factory Task.fromJson(Map<String, dynamic> json) => _$TaskFromJson(json);
}

@freezed
class ShoppingList with _$ShoppingList {
  const factory ShoppingList({
    required String id,
    required String name,
    required String householdId,
    required DateTime createdAt,
    DateTime? updatedAt,
    required List<ShoppingItem> items,
  }) = _ShoppingList;

  factory ShoppingList.fromJson(Map<String, dynamic> json) => _$ShoppingListFromJson(json);
}

@freezed
class ShoppingItem with _$ShoppingItem {
  const factory ShoppingItem({
    required String id,
    required String name,
    String? note,
    int? quantity,
    double? estimatedPrice,
    required bool completed,
    required DateTime createdAt,
    String? completedByUserId,
  }) = _ShoppingItem;

  factory ShoppingItem.fromJson(Map<String, dynamic> json) => _$ShoppingItemFromJson(json);
}

@freezed
class Bill with _$Bill {
  const factory Bill({
    required String id,
    required String name,
    String? description,
    required double amount,
    required String currency,
    required DateTime dueDate,
    DateTime? paidAt,
    required String householdId,
    required BillStatus status,
    BillRecurrence? recurrence,
    String? category,
  }) = _Bill;

  factory Bill.fromJson(Map<String, dynamic> json) => _$BillFromJson(json);
}

@freezed
class HouseTimer with _$HouseTimer {
  const factory HouseTimer({
    required String id,
    required String type,
    required String title,
    required Duration duration,
    required DateTime startedAt,
    DateTime? completedAt,
    String? taskId,
    required String householdId,
    required TimerStatus status,
  }) = _HouseTimer;

  factory HouseTimer.fromJson(Map<String, dynamic> json) => _$HouseTimerFromJson(json);
}

@freezed
class User with _$User {
  const factory User({
    required String id,
    required String email,
    required String name,
    String? avatarUrl,
    required DateTime createdAt,
    String? currentHouseholdId,
    UserPreferences? preferences,
  }) = _User;

  factory User.fromJson(Map<String, dynamic> json) => _$UserFromJson(json);
}

@freezed
class UserPreferences with _$UserPreferences {
  const factory UserPreferences({
    @Default('en') String language,
    @Default(false) bool darkMode,
    @Default(true) bool notificationsEnabled,
    @Default('22:00') String quietHoursStart,
    @Default('07:00') String quietHoursEnd,
  }) = _UserPreferences;

  factory UserPreferences.fromJson(Map<String, dynamic> json) => _$UserPreferencesFromJson(json);
}

enum TaskStatus {
  @JsonValue('pending')
  pending,
  @JsonValue('in_progress')
  inProgress,
  @JsonValue('completed')
  completed,
  @JsonValue('cancelled')
  cancelled,
}

enum TaskPriority {
  @JsonValue('low')
  low,
  @JsonValue('medium')
  medium,
  @JsonValue('high')
  high,
  @JsonValue('urgent')
  urgent,
}

enum BillStatus {
  @JsonValue('pending')
  pending,
  @JsonValue('paid')
  paid,
  @JsonValue('overdue')
  overdue,
  @JsonValue('cancelled')
  cancelled,
}

enum BillRecurrence {
  @JsonValue('none')
  none,
  @JsonValue('weekly')
  weekly,
  @JsonValue('monthly')
  monthly,
  @JsonValue('quarterly')
  quarterly,
  @JsonValue('yearly')
  yearly,
}

enum TimerStatus {
  @JsonValue('running')
  running,
  @JsonValue('completed')
  completed,
  @JsonValue('cancelled')
  cancelled,
}
