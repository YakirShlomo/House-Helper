// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'models.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_$TaskImpl _$$TaskImplFromJson(Map<String, dynamic> json) => _$TaskImpl(
      id: json['id'] as String,
      title: json['title'] as String,
      description: json['description'] as String?,
      createdAt: DateTime.parse(json['createdAt'] as String),
      completedAt: json['completedAt'] == null
          ? null
          : DateTime.parse(json['completedAt'] as String),
      dueAt: json['dueAt'] == null
          ? null
          : DateTime.parse(json['dueAt'] as String),
      householdId: json['householdId'] as String,
      assignedToUserId: json['assignedToUserId'] as String?,
      status: $enumDecode(_$TaskStatusEnumMap, json['status']),
      priority: $enumDecodeNullable(_$TaskPriorityEnumMap, json['priority']),
      tags: (json['tags'] as List<dynamic>?)?.map((e) => e as String).toList(),
    );

Map<String, dynamic> _$$TaskImplToJson(_$TaskImpl instance) =>
    <String, dynamic>{
      'id': instance.id,
      'title': instance.title,
      'description': instance.description,
      'createdAt': instance.createdAt.toIso8601String(),
      'completedAt': instance.completedAt?.toIso8601String(),
      'dueAt': instance.dueAt?.toIso8601String(),
      'householdId': instance.householdId,
      'assignedToUserId': instance.assignedToUserId,
      'status': _$TaskStatusEnumMap[instance.status]!,
      'priority': _$TaskPriorityEnumMap[instance.priority],
      'tags': instance.tags,
    };

const _$TaskStatusEnumMap = {
  TaskStatus.pending: 'pending',
  TaskStatus.inProgress: 'in_progress',
  TaskStatus.completed: 'completed',
  TaskStatus.cancelled: 'cancelled',
};

const _$TaskPriorityEnumMap = {
  TaskPriority.low: 'low',
  TaskPriority.medium: 'medium',
  TaskPriority.high: 'high',
  TaskPriority.urgent: 'urgent',
};

_$ShoppingListImpl _$$ShoppingListImplFromJson(Map<String, dynamic> json) =>
    _$ShoppingListImpl(
      id: json['id'] as String,
      name: json['name'] as String,
      householdId: json['householdId'] as String,
      createdAt: DateTime.parse(json['createdAt'] as String),
      updatedAt: json['updatedAt'] == null
          ? null
          : DateTime.parse(json['updatedAt'] as String),
      items: (json['items'] as List<dynamic>)
          .map((e) => ShoppingItem.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$$ShoppingListImplToJson(_$ShoppingListImpl instance) =>
    <String, dynamic>{
      'id': instance.id,
      'name': instance.name,
      'householdId': instance.householdId,
      'createdAt': instance.createdAt.toIso8601String(),
      'updatedAt': instance.updatedAt?.toIso8601String(),
      'items': instance.items,
    };

_$ShoppingItemImpl _$$ShoppingItemImplFromJson(Map<String, dynamic> json) =>
    _$ShoppingItemImpl(
      id: json['id'] as String,
      name: json['name'] as String,
      note: json['note'] as String?,
      quantity: (json['quantity'] as num?)?.toInt(),
      estimatedPrice: (json['estimatedPrice'] as num?)?.toDouble(),
      completed: json['completed'] as bool,
      createdAt: DateTime.parse(json['createdAt'] as String),
      completedByUserId: json['completedByUserId'] as String?,
    );

Map<String, dynamic> _$$ShoppingItemImplToJson(_$ShoppingItemImpl instance) =>
    <String, dynamic>{
      'id': instance.id,
      'name': instance.name,
      'note': instance.note,
      'quantity': instance.quantity,
      'estimatedPrice': instance.estimatedPrice,
      'completed': instance.completed,
      'createdAt': instance.createdAt.toIso8601String(),
      'completedByUserId': instance.completedByUserId,
    };

_$BillImpl _$$BillImplFromJson(Map<String, dynamic> json) => _$BillImpl(
      id: json['id'] as String,
      name: json['name'] as String,
      description: json['description'] as String?,
      amount: (json['amount'] as num).toDouble(),
      currency: json['currency'] as String,
      dueDate: DateTime.parse(json['dueDate'] as String),
      paidAt: json['paidAt'] == null
          ? null
          : DateTime.parse(json['paidAt'] as String),
      householdId: json['householdId'] as String,
      status: $enumDecode(_$BillStatusEnumMap, json['status']),
      recurrence:
          $enumDecodeNullable(_$BillRecurrenceEnumMap, json['recurrence']),
      category: json['category'] as String?,
    );

Map<String, dynamic> _$$BillImplToJson(_$BillImpl instance) =>
    <String, dynamic>{
      'id': instance.id,
      'name': instance.name,
      'description': instance.description,
      'amount': instance.amount,
      'currency': instance.currency,
      'dueDate': instance.dueDate.toIso8601String(),
      'paidAt': instance.paidAt?.toIso8601String(),
      'householdId': instance.householdId,
      'status': _$BillStatusEnumMap[instance.status]!,
      'recurrence': _$BillRecurrenceEnumMap[instance.recurrence],
      'category': instance.category,
    };

const _$BillStatusEnumMap = {
  BillStatus.pending: 'pending',
  BillStatus.paid: 'paid',
  BillStatus.overdue: 'overdue',
  BillStatus.cancelled: 'cancelled',
};

const _$BillRecurrenceEnumMap = {
  BillRecurrence.none: 'none',
  BillRecurrence.weekly: 'weekly',
  BillRecurrence.monthly: 'monthly',
  BillRecurrence.quarterly: 'quarterly',
  BillRecurrence.yearly: 'yearly',
};

_$HouseTimerImpl _$$HouseTimerImplFromJson(Map<String, dynamic> json) =>
    _$HouseTimerImpl(
      id: json['id'] as String,
      type: json['type'] as String,
      title: json['title'] as String,
      duration: Duration(microseconds: (json['duration'] as num).toInt()),
      startedAt: DateTime.parse(json['startedAt'] as String),
      completedAt: json['completedAt'] == null
          ? null
          : DateTime.parse(json['completedAt'] as String),
      taskId: json['taskId'] as String?,
      householdId: json['householdId'] as String,
      status: $enumDecode(_$TimerStatusEnumMap, json['status']),
    );

Map<String, dynamic> _$$HouseTimerImplToJson(_$HouseTimerImpl instance) =>
    <String, dynamic>{
      'id': instance.id,
      'type': instance.type,
      'title': instance.title,
      'duration': instance.duration.inMicroseconds,
      'startedAt': instance.startedAt.toIso8601String(),
      'completedAt': instance.completedAt?.toIso8601String(),
      'taskId': instance.taskId,
      'householdId': instance.householdId,
      'status': _$TimerStatusEnumMap[instance.status]!,
    };

const _$TimerStatusEnumMap = {
  TimerStatus.running: 'running',
  TimerStatus.completed: 'completed',
  TimerStatus.cancelled: 'cancelled',
};

_$UserImpl _$$UserImplFromJson(Map<String, dynamic> json) => _$UserImpl(
      id: json['id'] as String,
      email: json['email'] as String,
      name: json['name'] as String,
      avatarUrl: json['avatarUrl'] as String?,
      createdAt: DateTime.parse(json['createdAt'] as String),
      currentHouseholdId: json['currentHouseholdId'] as String?,
      preferences: json['preferences'] == null
          ? null
          : UserPreferences.fromJson(
              json['preferences'] as Map<String, dynamic>),
    );

Map<String, dynamic> _$$UserImplToJson(_$UserImpl instance) =>
    <String, dynamic>{
      'id': instance.id,
      'email': instance.email,
      'name': instance.name,
      'avatarUrl': instance.avatarUrl,
      'createdAt': instance.createdAt.toIso8601String(),
      'currentHouseholdId': instance.currentHouseholdId,
      'preferences': instance.preferences,
    };

_$UserPreferencesImpl _$$UserPreferencesImplFromJson(
        Map<String, dynamic> json) =>
    _$UserPreferencesImpl(
      language: json['language'] as String? ?? 'en',
      darkMode: json['darkMode'] as bool? ?? false,
      notificationsEnabled: json['notificationsEnabled'] as bool? ?? true,
      quietHoursStart: json['quietHoursStart'] as String? ?? '22:00',
      quietHoursEnd: json['quietHoursEnd'] as String? ?? '07:00',
    );

Map<String, dynamic> _$$UserPreferencesImplToJson(
        _$UserPreferencesImpl instance) =>
    <String, dynamic>{
      'language': instance.language,
      'darkMode': instance.darkMode,
      'notificationsEnabled': instance.notificationsEnabled,
      'quietHoursStart': instance.quietHoursStart,
      'quietHoursEnd': instance.quietHoursEnd,
    };
