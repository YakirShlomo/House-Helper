# House Helper API Documentation

**Version**: 1.0  
**Base URL**: `https://api.house-helper.com/api/v1`  
**Authentication**: Bearer Token (JWT)

## Overview

The House Helper API is a RESTful API that provides endpoints for managing household tasks, families, users, and notifications.

### API Characteristics

- **Protocol**: HTTPS only
- **Data Format**: JSON
- **Authentication**: JWT tokens
- **Rate Limiting**: 1000 requests/hour per user
- **Versioning**: URI versioning (`/api/v1`)

### Response Format

All API responses follow this structure:

```json
{
  "data": {...},      // Response data
  "meta": {           // Metadata (optional)
    "total": 100,
    "page": 1,
    "per_page": 20
  },
  "error": {          // Error details (if error)
    "code": "INVALID_INPUT",
    "message": "Validation error",
    "details": []
  }
}
```

### HTTP Status Codes

| Code | Description |
|------|-------------|
| 200 | OK - Request successful |
| 201 | Created - Resource created |
| 204 | No Content - Successful with no response body |
| 400 | Bad Request - Invalid input |
| 401 | Unauthorized - Missing or invalid auth token |
| 403 | Forbidden - Insufficient permissions |
| 404 | Not Found - Resource not found |
| 409 | Conflict - Resource conflict |
| 422 | Unprocessable Entity - Validation error |
| 429 | Too Many Requests - Rate limit exceeded |
| 500 | Internal Server Error - Server error |
| 503 | Service Unavailable - Service temporarily unavailable |

## Authentication

### Register

Create a new user account.

**Endpoint**: `POST /auth/register`

**Request**:
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123!",
  "name": "John Doe"
}
```

**Response**: `201 Created`
```json
{
  "data": {
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "name": "John Doe",
      "created_at": "2024-01-15T10:00:00Z"
    },
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

### Login

Authenticate and receive access token.

**Endpoint**: `POST /auth/login`

**Request**:
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123!"
}
```

**Response**: `200 OK`
```json
{
  "data": {
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "name": "John Doe"
    },
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_in": 3600
  }
}
```

### Refresh Token

Get new access token using refresh token.

**Endpoint**: `POST /auth/refresh`

**Request**:
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Response**: `200 OK`
```json
{
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_in": 3600
  }
}
```

### Logout

Invalidate current token.

**Endpoint**: `POST /auth/logout`

**Headers**: `Authorization: Bearer <token>`

**Response**: `204 No Content`

## Users

### Get Current User

Get authenticated user profile.

**Endpoint**: `GET /users/me`

**Headers**: `Authorization: Bearer <token>`

**Response**: `200 OK`
```json
{
  "data": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "John Doe",
    "profile_picture": "https://...",
    "created_at": "2024-01-15T10:00:00Z",
    "updated_at": "2024-01-15T10:00:00Z"
  }
}
```

### Update User

Update user profile.

**Endpoint**: `PATCH /users/me`

**Headers**: `Authorization: Bearer <token>`

**Request**:
```json
{
  "name": "John Smith",
  "profile_picture": "https://..."
}
```

**Response**: `200 OK`

### Delete User

Delete user account.

**Endpoint**: `DELETE /users/me`

**Headers**: `Authorization: Bearer <token>`

**Response**: `204 No Content`

## Families

### List Families

Get all families for authenticated user.

**Endpoint**: `GET /families`

**Headers**: `Authorization: Bearer <token>`

**Response**: `200 OK`
```json
{
  "data": [
    {
      "id": "uuid",
      "name": "Smith Family",
      "owner_id": "uuid",
      "created_at": "2024-01-15T10:00:00Z",
      "member_count": 4
    }
  ]
}
```

### Create Family

Create a new family.

**Endpoint**: `POST /families`

**Headers**: `Authorization: Bearer <token>`

**Request**:
```json
{
  "name": "Smith Family",
  "description": "Our family household"
}
```

**Response**: `201 Created`
```json
{
  "data": {
    "id": "uuid",
    "name": "Smith Family",
    "description": "Our family household",
    "owner_id": "uuid",
    "created_at": "2024-01-15T10:00:00Z"
  }
}
```

### Get Family

Get family details.

**Endpoint**: `GET /families/{family_id}`

**Headers**: `Authorization: Bearer <token>`

**Response**: `200 OK`
```json
{
  "data": {
    "id": "uuid",
    "name": "Smith Family",
    "description": "Our family household",
    "owner_id": "uuid",
    "created_at": "2024-01-15T10:00:00Z",
    "members": [
      {
        "user_id": "uuid",
        "name": "John Doe",
        "role": "admin",
        "joined_at": "2024-01-15T10:00:00Z"
      }
    ]
  }
}
```

### Update Family

Update family details.

**Endpoint**: `PATCH /families/{family_id}`

**Headers**: `Authorization: Bearer <token>`

**Request**:
```json
{
  "name": "Updated Family Name",
  "description": "New description"
}
```

**Response**: `200 OK`

### Delete Family

Delete a family.

**Endpoint**: `DELETE /families/{family_id}`

**Headers**: `Authorization: Bearer <token>`

**Response**: `204 No Content`

### Add Member

Add member to family.

**Endpoint**: `POST /families/{family_id}/members`

**Headers**: `Authorization: Bearer <token>`

**Request**:
```json
{
  "email": "member@example.com",
  "role": "member"
}
```

**Response**: `201 Created`

### Remove Member

Remove member from family.

**Endpoint**: `DELETE /families/{family_id}/members/{user_id}`

**Headers**: `Authorization: Bearer <token>`

**Response**: `204 No Content`

## Tasks

### List Tasks

Get all tasks with optional filters.

**Endpoint**: `GET /tasks`

**Headers**: `Authorization: Bearer <token>`

**Query Parameters**:
- `family_id` (optional): Filter by family
- `assigned_to` (optional): Filter by assigned user
- `status` (optional): Filter by status (pending, in_progress, completed)
- `due_date_from` (optional): Filter by due date start
- `due_date_to` (optional): Filter by due date end
- `page` (optional): Page number (default: 1)
- `per_page` (optional): Items per page (default: 20, max: 100)

**Response**: `200 OK`
```json
{
  "data": [
    {
      "id": "uuid",
      "title": "Weekly Groceries",
      "description": "Buy groceries for the week",
      "family_id": "uuid",
      "assigned_to": "uuid",
      "due_date": "2024-01-20T18:00:00Z",
      "points": 50,
      "status": "pending",
      "priority": "high",
      "created_at": "2024-01-15T10:00:00Z",
      "updated_at": "2024-01-15T10:00:00Z"
    }
  ],
  "meta": {
    "total": 25,
    "page": 1,
    "per_page": 20,
    "total_pages": 2
  }
}
```

### Create Task

Create a new task.

**Endpoint**: `POST /tasks`

**Headers**: `Authorization: Bearer <token>`

**Request**:
```json
{
  "title": "Weekly Groceries",
  "description": "Buy groceries for the week",
  "family_id": "uuid",
  "assigned_to": "uuid",
  "due_date": "2024-01-20T18:00:00Z",
  "points": 50,
  "priority": "high"
}
```

**Response**: `201 Created`
```json
{
  "data": {
    "id": "uuid",
    "title": "Weekly Groceries",
    "description": "Buy groceries for the week",
    "family_id": "uuid",
    "assigned_to": "uuid",
    "due_date": "2024-01-20T18:00:00Z",
    "points": 50,
    "status": "pending",
    "priority": "high",
    "created_at": "2024-01-15T10:00:00Z"
  }
}
```

### Get Task

Get task details.

**Endpoint**: `GET /tasks/{task_id}`

**Headers**: `Authorization: Bearer <token>`

**Response**: `200 OK`

### Update Task

Update task details.

**Endpoint**: `PATCH /tasks/{task_id}`

**Headers**: `Authorization: Bearer <token>`

**Request**:
```json
{
  "title": "Updated Title",
  "status": "in_progress",
  "points": 75
}
```

**Response**: `200 OK`

### Delete Task

Delete a task.

**Endpoint**: `DELETE /tasks/{task_id}`

**Headers**: `Authorization: Bearer <token>`

**Response**: `204 No Content`

### Complete Task

Mark task as completed.

**Endpoint**: `POST /tasks/{task_id}/complete`

**Headers**: `Authorization: Bearer <token>`

**Response**: `200 OK`
```json
{
  "data": {
    "task": {
      "id": "uuid",
      "status": "completed",
      "completed_at": "2024-01-16T14:30:00Z"
    },
    "points_earned": 50
  }
}
```

## Points

### Get User Points

Get points summary for user.

**Endpoint**: `GET /points/me`

**Headers**: `Authorization: Bearer <token>`

**Query Parameters**:
- `family_id` (optional): Filter by family

**Response**: `200 OK`
```json
{
  "data": {
    "total_points": 1250,
    "weekly_points": 150,
    "monthly_points": 600,
    "rank": 2,
    "level": 5
  }
}
```

### Get Points History

Get points transaction history.

**Endpoint**: `GET /points/history`

**Headers**: `Authorization: Bearer <token>`

**Query Parameters**:
- `family_id` (optional): Filter by family
- `from_date` (optional): Start date
- `to_date` (optional): End date
- `page` (optional): Page number
- `per_page` (optional): Items per page

**Response**: `200 OK`
```json
{
  "data": [
    {
      "id": "uuid",
      "user_id": "uuid",
      "task_id": "uuid",
      "points": 50,
      "type": "earned",
      "description": "Completed task: Weekly Groceries",
      "created_at": "2024-01-16T14:30:00Z"
    }
  ],
  "meta": {
    "total": 45,
    "page": 1,
    "per_page": 20
  }
}
```

### Get Family Leaderboard

Get points leaderboard for family.

**Endpoint**: `GET /families/{family_id}/leaderboard`

**Headers**: `Authorization: Bearer <token>`

**Query Parameters**:
- `period` (optional): time_period (week, month, all_time) (default: month)

**Response**: `200 OK`
```json
{
  "data": [
    {
      "rank": 1,
      "user_id": "uuid",
      "name": "John Doe",
      "points": 1250,
      "tasks_completed": 25
    },
    {
      "rank": 2,
      "user_id": "uuid",
      "name": "Jane Doe",
      "points": 980,
      "tasks_completed": 20
    }
  ]
}
```

## Notifications

### List Notifications

Get all notifications for user.

**Endpoint**: `GET /notifications`

**Headers**: `Authorization: Bearer <token>`

**Query Parameters**:
- `read` (optional): Filter by read status (true/false)
- `page` (optional): Page number
- `per_page` (optional): Items per page

**Response**: `200 OK`
```json
{
  "data": [
    {
      "id": "uuid",
      "type": "task_assigned",
      "title": "New Task Assigned",
      "body": "You've been assigned: Weekly Groceries",
      "data": {
        "task_id": "uuid"
      },
      "read": false,
      "created_at": "2024-01-15T10:00:00Z"
    }
  ],
  "meta": {
    "total": 15,
    "unread_count": 5
  }
}
```

### Mark as Read

Mark notification(s) as read.

**Endpoint**: `POST /notifications/read`

**Headers**: `Authorization: Bearer <token>`

**Request**:
```json
{
  "notification_ids": ["uuid1", "uuid2"]
}
```

**Response**: `200 OK`

### Mark All as Read

Mark all notifications as read.

**Endpoint**: `POST /notifications/read-all`

**Headers**: `Authorization: Bearer <token>`

**Response**: `200 OK`

## Error Responses

### Validation Error

**Status**: `422 Unprocessable Entity`

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": [
      {
        "field": "email",
        "message": "Invalid email format"
      },
      {
        "field": "password",
        "message": "Password must be at least 8 characters"
      }
    ]
  }
}
```

### Authentication Error

**Status**: `401 Unauthorized`

```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid or expired token"
  }
}
```

### Authorization Error

**Status**: `403 Forbidden`

```json
{
  "error": {
    "code": "FORBIDDEN",
    "message": "Insufficient permissions"
  }
}
```

### Not Found Error

**Status**: `404 Not Found`

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Resource not found"
  }
}
```

### Rate Limit Error

**Status**: `429 Too Many Requests`

```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Too many requests. Try again later.",
    "retry_after": 3600
  }
}
```

## Rate Limiting

- **Rate Limit**: 1000 requests per hour per user
- **Headers**:
  - `X-RateLimit-Limit`: Maximum requests per hour
  - `X-RateLimit-Remaining`: Remaining requests
  - `X-RateLimit-Reset`: Timestamp when limit resets

## Pagination

All list endpoints support pagination:

**Request Parameters**:
- `page`: Page number (default: 1)
- `per_page`: Items per page (default: 20, max: 100)

**Response Headers**:
- `X-Total-Count`: Total number of items
- `X-Page`: Current page
- `X-Per-Page`: Items per page
- `X-Total-Pages`: Total number of pages

**Response Meta**:
```json
{
  "meta": {
    "total": 100,
    "page": 2,
    "per_page": 20,
    "total_pages": 5
  }
}
```

## Webhooks (Future)

Coming soon: Webhook support for real-time events.

## SDKs & Client Libraries

- **JavaScript/TypeScript**: `npm install @house-helper/sdk` (coming soon)
- **Dart/Flutter**: `flutter pub add house_helper_sdk` (coming soon)

## Support

- **API Status**: https://status.house-helper.com
- **Support**: api-support@house-helper.com
- **Documentation**: https://docs.house-helper.com

## Changelog

### v1.0 (2024-01-15)
- Initial API release
- Authentication endpoints
- User management
- Family management
- Task management
- Points system
- Notifications

## License

Copyright © 2024 House Helper. All rights reserved.
