import 'dart:developer' as developer;

import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import '../models/models.dart';

class ApiClient {
  late final Dio _dio;
  final _storage = const FlutterSecureStorage();
  static const _tokenKey = 'auth_token';
  static const _refreshTokenKey = 'refresh_token';
  
  ApiClient({String? baseUrl}) {
    _dio = Dio(BaseOptions(
      baseUrl: baseUrl ?? 'http://localhost:8080',
      connectTimeout: const Duration(seconds: 10),
      receiveTimeout: const Duration(seconds: 10),
      headers: {
        'Content-Type': 'application/json',
      },
    ));

    _dio.interceptors.add(LogInterceptor(
      requestBody: true,
      responseBody: true,
      logPrint: (log) => developer.log('[API] $log', name: 'ApiClient'),
    ));

    _dio.interceptors.add(InterceptorsWrapper(
      onRequest: (options, handler) async {
        // Add JWT token if available
        final token = await _getToken();
        if (token != null) {
          options.headers['Authorization'] = 'Bearer $token';
        }
        handler.next(options);
      },
      onError: (error, handler) async {
        if (error.response?.statusCode == 401) {
          // Try to refresh token
          final refreshed = await _refreshToken();
          if (refreshed) {
            // Retry the request
            final request = error.requestOptions;
            final token = await _getToken();
            if (token != null) {
              request.headers['Authorization'] = 'Bearer $token';
            }
            try {
              final response = await _dio.fetch(request);
              handler.resolve(response);
              return;
            } catch (e) {
              // If retry fails, continue with original error
            }
          }
        }
        handler.next(error);
      },
    ));
  }

  // Auth endpoints
  Future<Map<String, dynamic>> login(String email, String password) async {
    final response = await _dio.post('/v1/auth/login', data: {
      'email': email,
      'password': password,
    });
    
    // Store tokens securely
    if (response.data['token'] != null) {
      await _storage.write(key: _tokenKey, value: response.data['token']);
    }
    if (response.data['refresh_token'] != null) {
      await _storage.write(key: _refreshTokenKey, value: response.data['refresh_token']);
    }
    
    return response.data;
  }

  Future<Map<String, dynamic>> signup(String email, String password, String name) async {
    final response = await _dio.post('/v1/auth/signup', data: {
      'email': email,
      'password': password,
      'name': name,
    });
    
    // Store tokens securely
    if (response.data['token'] != null) {
      await _storage.write(key: _tokenKey, value: response.data['token']);
    }
    if (response.data['refresh_token'] != null) {
      await _storage.write(key: _refreshTokenKey, value: response.data['refresh_token']);
    }
    
    return response.data;
  }

  Future<User> getCurrentUser() async {
    final response = await _dio.get('/v1/me');
    return User.fromJson(response.data);
  }

  // Tasks endpoints
  Future<List<Task>> getTasks({String? householdId}) async {
    final response = await _dio.get('/v1/tasks', queryParameters: {
      if (householdId != null) 'household_id': householdId,
    });
    return (response.data as List)
        .map((json) => Task.fromJson(json))
        .toList();
  }

  Future<Task> createTask(Task task) async {
    final response = await _dio.post('/v1/tasks', data: task.toJson());
    return Task.fromJson(response.data);
  }

  Future<Task> updateTask(String id, Task task) async {
    final response = await _dio.put('/v1/tasks/$id', data: task.toJson());
    return Task.fromJson(response.data);
  }

  Future<void> deleteTask(String id) async {
    await _dio.delete('/v1/tasks/$id');
  }

  // Shopping endpoints
  Future<List<ShoppingList>> getShoppingLists({String? householdId}) async {
    final response = await _dio.get('/v1/shopping/lists', queryParameters: {
      if (householdId != null) 'household_id': householdId,
    });
    return (response.data as List)
        .map((json) => ShoppingList.fromJson(json))
        .toList();
  }

  Future<ShoppingList> createShoppingList(ShoppingList list) async {
    final response = await _dio.post('/v1/shopping/lists', data: list.toJson());
    return ShoppingList.fromJson(response.data);
  }

  Future<ShoppingItem> addShoppingItem(String listId, ShoppingItem item) async {
    final response = await _dio.post('/v1/shopping/lists/$listId/items', data: item.toJson());
    return ShoppingItem.fromJson(response.data);
  }

  Future<ShoppingItem> updateShoppingItem(String listId, String itemId, ShoppingItem item) async {
    final response = await _dio.put('/v1/shopping/lists/$listId/items/$itemId', data: item.toJson());
    return ShoppingItem.fromJson(response.data);
  }

  // Bills endpoints
  Future<List<Bill>> getBills({String? householdId}) async {
    final response = await _dio.get('/v1/bills', queryParameters: {
      if (householdId != null) 'household_id': householdId,
    });
    return (response.data as List)
        .map((json) => Bill.fromJson(json))
        .toList();
  }

  Future<Bill> createBill(Bill bill) async {
    final response = await _dio.post('/v1/bills', data: bill.toJson());
    return Bill.fromJson(response.data);
  }

  Future<Bill> payBill(String id) async {
    final response = await _dio.post('/v1/bills/$id/pay');
    return Bill.fromJson(response.data);
  }

  // Timers endpoints
  Future<HouseTimer> startTimer(String type, Duration duration, {String? taskId}) async {
    final response = await _dio.post('/v1/timers/start', data: {
      'type': type,
      'duration_seconds': duration.inSeconds,
      if (taskId != null) 'task_id': taskId,
    });
    return HouseTimer.fromJson(response.data);
  }

  Future<void> cancelTimer(String id) async {
    await _dio.post('/v1/timers/$id/cancel');
  }

  Future<List<HouseTimer>> getActiveTimers() async {
    final response = await _dio.get('/v1/timers/active');
    return (response.data as List)
        .map((json) => HouseTimer.fromJson(json))
        .toList();
  }

  // Private methods for token management
  Future<String?> _getToken() async {
    return await _storage.read(key: _tokenKey);
  }

  Future<bool> _refreshToken() async {
    try {
      final refreshToken = await _storage.read(key: _refreshTokenKey);
      if (refreshToken == null) return false;

      final response = await _dio.post('/v1/auth/refresh', data: {
        'refresh_token': refreshToken,
      });

      if (response.data['token'] != null) {
        await _storage.write(key: _tokenKey, value: response.data['token']);
        if (response.data['refresh_token'] != null) {
          await _storage.write(key: _refreshTokenKey, value: response.data['refresh_token']);
        }
        return true;
      }
      return false;
    } catch (e) {
      developer.log('Token refresh failed: $e', name: 'ApiClient', error: e);
      return false;
    }
  }

  // Logout method to clear tokens
  Future<void> logout() async {
    await _storage.delete(key: _tokenKey);
    await _storage.delete(key: _refreshTokenKey);
  }
}

final apiClientProvider = Provider<ApiClient>((ref) {
  return ApiClient();
});
