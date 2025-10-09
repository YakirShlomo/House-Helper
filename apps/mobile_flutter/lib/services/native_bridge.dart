import 'dart:developer' as developer;

import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/models.dart';

class NativeBridgeService {
  static const MethodChannel _channel = MethodChannel('app.househelper.mobile/native');

  // Update widgets with latest data
  Future<void> updateWidgets() async {
    try {
      await _channel.invokeMethod('updateWidgets');
    } catch (e) {
      developer.log('Error updating widgets: $e', name: 'NativeBridge', error: e);
    }
  }

  // Create app shortcuts (Android 7.1+)
  Future<void> createShortcuts() async {
    try {
      await _channel.invokeMethod('createShortcuts');
    } catch (e) {
      developer.log('Error creating shortcuts: $e', name: 'NativeBridge', error: e);
    }
  }

  // Update dynamic shortcuts with current tasks
  Future<void> updateDynamicShortcuts(List<Task> tasks) async {
    try {
      final tasksData = tasks.map((task) => {
        'id': task.id,
        'title': task.title,
        'isCompleted': task.status == TaskStatus.completed,
      }).toList();

      await _channel.invokeMethod('updateDynamicShortcuts', {
        'tasks': tasksData,
      });
    } catch (e) {
      developer.log('Error updating dynamic shortcuts: $e', name: 'NativeBridge', error: e);
    }
  }

  // Schedule widget data sync
  Future<void> scheduleWidgetSync() async {
    try {
      await _channel.invokeMethod('scheduleWidgetSync');
    } catch (e) {
      developer.log('Error scheduling widget sync: $e', name: 'NativeBridge', error: e);
    }
  }

  // Get pending actions from widgets/shortcuts
  Future<List<String>> getPendingActions() async {
    try {
      final result = await _channel.invokeMethod('getPendingActions');
      return List<String>.from(result ?? []);
    } catch (e) {
      developer.log('Error getting pending actions: $e', name: 'NativeBridge', error: e);
      return [];
    }
  }

  // Clear pending actions after processing
  Future<void> clearPendingActions() async {
    try {
      await _channel.invokeMethod('clearPendingActions');
    } catch (e) {
      developer.log('Error clearing pending actions: $e', name: 'NativeBridge', error: e);
    }
  }

  // Set up intent handling
  void setupIntentHandling(Function(Map<String, dynamic>) onIntent) {
    _channel.setMethodCallHandler((call) async {
      if (call.method == 'handleIntent') {
        final args = Map<String, dynamic>.from(call.arguments);
        onIntent(args);
      }
    });
  }
}

// Provider for native bridge service
final nativeBridgeProvider = Provider<NativeBridgeService>((ref) {
  return NativeBridgeService();
});

// Service to handle native integration
class NativeIntegrationService {
  final NativeBridgeService _bridge;
  Function(Map<String, dynamic>)? _onIntentCallback;
  
  NativeIntegrationService(this._bridge);
  
  void setIntentCallback(Function(Map<String, dynamic>) callback) {
    _onIntentCallback = callback;
  }

  // Initialize native integrations
  Future<void> initialize() async {
    await _bridge.createShortcuts();
    await _bridge.scheduleWidgetSync();
    
    // Set up intent handling
    _bridge.setupIntentHandling((intent) {
      handleNativeIntent(intent);
    });
  }

  // Handle intents from widgets/shortcuts
  void handleNativeIntent(Map<String, dynamic> intent) {
    final action = intent['action'] as String?;
    
    developer.log('Handling native intent: $action', name: 'NativeIntegration');
    
    // Notify callback if set (for navigation/state updates in app)
    if (_onIntentCallback != null) {
      _onIntentCallback!(intent);
    }
    
    // Log for debugging purposes
    switch (action) {
      case 'add_task':
        final title = intent['title'] as String?;
        if (title != null) {
          developer.log('Add task intent: $title', name: 'NativeIntegration');
        }
        break;
        
      case 'start_timer':
        final timerType = intent['timer_type'] as String?;
        final duration = intent['duration'] as int?;
        if (timerType != null && duration != null) {
          developer.log('Start timer intent: $timerType for ${duration}s', name: 'NativeIntegration');
        }
        break;
        
      case 'complete_task':
        final taskId = intent['task_id'] as String?;
        if (taskId != null) {
          developer.log('Complete task intent: $taskId', name: 'NativeIntegration');
        }
        break;
        
      case 'view_tasks':
        developer.log('View tasks intent', name: 'NativeIntegration');
        break;
        
      case 'add_shopping_item':
        final itemName = intent['item_name'] as String?;
        if (itemName != null) {
          developer.log('Add shopping item intent: $itemName', name: 'NativeIntegration');
        }
        break;
    }
  }

  // Process pending actions from widgets
  Future<void> processPendingActions() async {
    final actions = await _bridge.getPendingActions();
    
    for (final action in actions) {
      final parts = action.split(':');
      if (parts.length < 3) continue;
      
      final actionType = parts[0];
      final data = parts[1];
      final timestamp = int.tryParse(parts[2]) ?? 0;
      
      // Skip old actions (older than 1 hour)
      if (DateTime.now().millisecondsSinceEpoch - timestamp > 3600000) {
        continue;
      }
      
      // Create intent map and handle via callback
      final intent = <String, dynamic>{
        'action': actionType,
        'data': data,
        'timestamp': timestamp,
      };
      
      if (_onIntentCallback != null) {
        _onIntentCallback!(intent);
      }
      
      developer.log('Processing action: $actionType with data: $data', name: 'NativeIntegration');
    }
    
    // Clear processed actions
    await _bridge.clearPendingActions();
  }

  // Update native components with app data
  Future<void> syncWithNative(List<Task> tasks) async {
    await _bridge.updateWidgets();
    await _bridge.updateDynamicShortcuts(tasks);
  }
}

// Provider for native integration service
final nativeIntegrationProvider = Provider<NativeIntegrationService>((ref) {
  final bridge = ref.read(nativeBridgeProvider);
  return NativeIntegrationService(bridge);
});
