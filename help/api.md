# OpenVPN Manager - REST API Documentation

Complete REST API documentation for OpenVPN Manager. All endpoints require authentication unless otherwise noted.

## Table of Contents

- [Authentication](#authentication)
- [Users](#users)
- [User Groups Management](#user-groups-management)
- [Groups](#groups)
- [Group Networks Management](#group-networks-management)
- [Networks](#networks)
- [VPN Sessions](#vpn-sessions)
- [Audit Logs](#audit-logs)
- [Error Responses](#error-responses)
- [OpenVPN Integration](#openvpn-integration)

---

## Authentication

All protected endpoints require a JWT token in the Authorization header:

```
Authorization: Bearer <token>
```

### Login

**POST** `/api/v1/auth/login`

Authenticate and receive a JWT token.

**Request Body:**
```json
{
  "username": "admin",
  "password": "admin123"
}
```

**Response (200 OK):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2025-12-02T10:00:00Z",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "admin",
    "first_name": "Admin",
    "last_name": "User",
    "email": "admin@example.com",
    "role": "ADMIN",
    "is_active": true
  }
}
```

**Login Validation:**
- Checks `is_active` - returns "User account is inactive" if false
- Checks `valid_from` - returns "User account is not yet valid" if current date is before valid_from
- Checks `valid_to` - returns "User account has expired" if current date is after valid_to

**Error Responses:**
- `401 Unauthorized` - Invalid credentials or inactive account

---

### Logout

**POST** `/api/v1/auth/logout`

Invalidate the current session.

**Response (200 OK):**
```json
{
  "message": "Logged out successfully"
}
```

---

### Get Current User

**GET** `/api/v1/auth/me`

Get information about the authenticated user.

**Response (200 OK):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "admin",
  "first_name": "Admin",
  "middle_name": "",
  "last_name": "User",
  "email": "admin@example.com",
  "telephone": "+420123456789",
  "role": "ADMIN",
  "is_active": true,
  "valid_from": null,
  "valid_to": null,
  "vpn_ip": "10.8.0.100",
  "created_at": "2025-12-01T08:00:00Z"
}
```

---

## Users

### List Users

**GET** `/api/v1/users`

List users based on the authenticated user's role:
- `USER` - Can only see themselves
- `MANAGER` - Sees only their subordinates
- `ADMIN` - Sees all users

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `page` | int | Page number (default: 1) |
| `page_size` | int | Items per page (default: 10, max: 100) |
| `search` | string | Search in username, first_name, last_name, email |
| `role` | string | Filter by role (USER, MANAGER, ADMIN) |
| `is_active` | bool | Filter by active status |

**Response (200 OK):**
```json
{
  "users": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "username": "john.doe",
      "first_name": "John",
      "last_name": "Doe",
      "email": "john@example.com",
      "role": "USER",
      "is_active": true,
      "valid_from": "2025-01-01T00:00:00Z",
      "valid_to": "2025-12-31T23:59:59Z",
      "vpn_ip": "10.8.0.101",
      "created_at": "2025-12-01T08:00:00Z"
    }
  ],
  "total": 50,
  "page": 1,
  "page_size": 10,
  "total_pages": 5
}
```

---

### Create User

**POST** `/api/v1/users`

Create a new user. Requires `MANAGER` or `ADMIN` role.

**Request Body:**
```json
{
  "username": "john.doe",
  "password": "securepassword123",
  "first_name": "John",
  "middle_name": "",
  "last_name": "Doe",
  "email": "john@example.com",
  "telephone": "+420123456789",
  "role": "USER",
  "is_active": true,
  "valid_from": "2025-01-01",
  "valid_to": "2025-12-31",
  "vpn_ip": "10.8.0.101",
  "manager_id": "550e8400-e29b-41d4-a716-446655440001"
}
```

**Field Descriptions:**
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `username` | string | Yes | Unique username (3-100 chars) |
| `password` | string | Yes | Password (min 8 chars) |
| `first_name` | string | Yes | First name (max 100 chars) |
| `middle_name` | string | No | Middle name (max 100 chars) |
| `last_name` | string | Yes | Last name (max 100 chars) |
| `email` | string | Yes | Valid email address |
| `telephone` | string | No | Phone number (max 50 chars) |
| `role` | string | Yes | USER, MANAGER, or ADMIN |
| `is_active` | bool | No | Active status (default: true) |
| `valid_from` | date | No | Account valid from date (YYYY-MM-DD) |
| `valid_to` | date | No | Account valid until date (YYYY-MM-DD) |
| `vpn_ip` | string | No | Static VPN IP (max 45 chars) |
| `manager_id` | UUID | No | Manager's user ID |

**Response (201 Created):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440002",
  "username": "john.doe",
  "first_name": "John",
  "last_name": "Doe",
  "email": "john@example.com",
  "role": "USER",
  "is_active": true,
  "created_at": "2025-12-01T10:00:00Z"
}
```

---

### Get User

**GET** `/api/v1/users/:id`

Get user details. Access control:
- `USER` - Can only get their own profile
- `MANAGER` - Can get their subordinates
- `ADMIN` - Can get any user

**Response (200 OK):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "john.doe",
  "manager_id": "550e8400-e29b-41d4-a716-446655440001",
  "manager": {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "username": "manager",
    "first_name": "Jane",
    "last_name": "Manager"
  },
  "first_name": "John",
  "middle_name": "",
  "last_name": "Doe",
  "email": "john@example.com",
  "telephone": "+420123456789",
  "role": "USER",
  "is_active": true,
  "valid_from": "2025-01-01T00:00:00Z",
  "valid_to": "2025-12-31T23:59:59Z",
  "vpn_ip": "10.8.0.101",
  "created_at": "2025-12-01T08:00:00Z",
  "updated_at": "2025-12-01T10:00:00Z",
  "created_by": "550e8400-e29b-41d4-a716-446655440001",
  "updated_by": "550e8400-e29b-41d4-a716-446655440001"
}
```

---

### Update User

**PUT** `/api/v1/users/:id`

Update user details. Access control same as Get User.

**Request Body (all fields optional):**
```json
{
  "username": "john.doe.updated",
  "password": "newpassword123",
  "first_name": "John",
  "last_name": "Doe",
  "email": "john.updated@example.com",
  "role": "MANAGER",
  "is_active": true,
  "valid_from": "2025-01-01",
  "valid_to": "2026-12-31",
  "vpn_ip": "10.8.0.102"
}
```

**Response (200 OK):** Updated user object

---

### Delete User

**DELETE** `/api/v1/users/:id`

Soft delete a user. Requires `ADMIN` role.

**Response (200 OK):**
```json
{
  "message": "User deleted successfully"
}
```

---

### Update Own Profile

**PUT** `/api/v1/users/profile`

Update the authenticated user's own profile. Available to all roles.

**Request Body:**
```json
{
  "first_name": "John",
  "middle_name": "William",
  "last_name": "Doe",
  "email": "john@example.com",
  "telephone": "+420123456789"
}
```

---

### Change Password

**PUT** `/api/v1/users/password`

Change the authenticated user's password.

**Request Body:**
```json
{
  "current_password": "oldpassword123",
  "new_password": "newsecurepassword123"
}
```

**Response (200 OK):**
```json
{
  "message": "Password updated successfully"
}
```

---

## User Groups Management

### Get User's Groups

**GET** `/api/v1/users/:id/groups`

Get all groups a user belongs to, including the networks assigned to each group.

**Response (200 OK):**
```json
{
  "groups": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440010",
      "name": "IT Department",
      "description": "IT team members",
      "networks": [
        {
          "id": "550e8400-e29b-41d4-a716-446655440020",
          "name": "Server Network",
          "cidr": "192.168.1.0/24",
          "description": "Internal servers"
        },
        {
          "id": "550e8400-e29b-41d4-a716-446655440021",
          "name": "Database Network",
          "cidr": "10.0.0.0/24",
          "description": "Database servers"
        }
      ]
    }
  ]
}
```

---

### Add User to Group

**POST** `/api/v1/users/:id/groups`

Add a user to a group. Requires `MANAGER` or `ADMIN` role.

**Request Body:**
```json
{
  "group_id": "550e8400-e29b-41d4-a716-446655440010"
}
```

**Response (200 OK):**
```json
{
  "message": "User added to group successfully"
}
```

---

### Remove User from Group

**DELETE** `/api/v1/users/:id/groups/:group_id`

Remove a user from a group. Requires `MANAGER` or `ADMIN` role.

**Response (200 OK):**
```json
{
  "message": "User removed from group successfully"
}
```

---

## Groups

### List Groups

**GET** `/api/v1/groups`

List all groups. Available to all authenticated users.

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `page` | int | Page number (default: 1) |
| `page_size` | int | Items per page (default: 10, max: 100) |
| `search` | string | Search in name and description |

**Response (200 OK):**
```json
{
  "groups": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440010",
      "name": "IT Department",
      "description": "IT team members",
      "created_at": "2025-12-01T08:00:00Z"
    }
  ],
  "total": 10,
  "page": 1,
  "page_size": 10,
  "total_pages": 1
}
```

---

### Create Group

**POST** `/api/v1/groups`

Create a new group. Requires `ADMIN` role.

**Request Body:**
```json
{
  "name": "Finance Department",
  "description": "Finance team members"
}
```

**Response (201 Created):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440011",
  "name": "Finance Department",
  "description": "Finance team members",
  "created_at": "2025-12-01T10:00:00Z"
}
```

---

### Get Group

**GET** `/api/v1/groups/:id`

Get group details.

**Response (200 OK):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440010",
  "name": "IT Department",
  "description": "IT team members",
  "created_at": "2025-12-01T08:00:00Z",
  "updated_at": "2025-12-01T10:00:00Z",
  "created_by": "550e8400-e29b-41d4-a716-446655440000",
  "updated_by": "550e8400-e29b-41d4-a716-446655440000"
}
```

---

### Update Group

**PUT** `/api/v1/groups/:id`

Update group details. Requires `ADMIN` role.

**Request Body:**
```json
{
  "name": "IT & DevOps Department",
  "description": "IT and DevOps team members"
}
```

---

### Delete Group

**DELETE** `/api/v1/groups/:id`

Delete a group. Requires `ADMIN` role. Users will be removed from the group.

**Response (200 OK):**
```json
{
  "message": "Group deleted successfully"
}
```

---

### Get Group Users

**GET** `/api/v1/groups/:id/users`

Get all users in a group. Results are filtered by the authenticated user's role.

**Response (200 OK):**
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "john.doe",
    "first_name": "John",
    "last_name": "Doe",
    "email": "john@example.com",
    "role": "USER"
  }
]
```

---

### Add User to Group

**POST** `/api/v1/groups/:id/users`

Add a user to a group. Requires `MANAGER` or `ADMIN` role.

**Request Body:**
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

---

### Remove User from Group

**DELETE** `/api/v1/groups/:id/users/:user_id`

Remove a user from a group. Requires `MANAGER` or `ADMIN` role.

---

## Group Networks Management

### Get Group Networks

**GET** `/api/v1/groups/:id/networks`

Get all networks assigned to a group.

**Response (200 OK):**
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440020",
    "name": "Server Network",
    "cidr": "192.168.1.0/24",
    "description": "Internal servers"
  }
]
```

---

### Add Network to Group

**POST** `/api/v1/groups/:id/networks`

Assign a network to a group. Requires `ADMIN` role.

**Request Body:**
```json
{
  "network_id": "550e8400-e29b-41d4-a716-446655440020"
}
```

**Response (200 OK):**
```json
{
  "message": "Network added to group successfully"
}
```

---

### Remove Network from Group

**DELETE** `/api/v1/groups/:id/networks/:network_id`

Remove a network from a group. Requires `ADMIN` role.

**Response (200 OK):**
```json
{
  "message": "Network removed from group successfully"
}
```

---

## Networks

All network endpoints require `ADMIN` role.

### List Networks

**GET** `/api/v1/networks`

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `page` | int | Page number (default: 1) |
| `page_size` | int | Items per page (default: 10, max: 100) |
| `search` | string | Search in name, cidr, description |

**Response (200 OK):**
```json
{
  "networks": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440020",
      "name": "Server Network",
      "cidr": "192.168.1.0/24",
      "description": "Internal servers",
      "created_at": "2025-12-01T08:00:00Z"
    }
  ],
  "total": 5,
  "page": 1,
  "page_size": 10,
  "total_pages": 1
}
```

---

### Create Network

**POST** `/api/v1/networks`

**Request Body:**
```json
{
  "name": "Database Network",
  "cidr": "10.0.0.0/24",
  "description": "Database servers"
}
```

**CIDR Examples:**
- `192.168.1.0/24` - Subnet (192.168.1.0 - 192.168.1.255)
- `10.0.0.1/32` - Single IP address
- `10.0.0.1` - Single IP (automatically converted to /32)

**Response (201 Created):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440021",
  "name": "Database Network",
  "cidr": "10.0.0.0/24",
  "description": "Database servers",
  "created_at": "2025-12-01T10:00:00Z"
}
```

---

### Get Network

**GET** `/api/v1/networks/:id`

---

### Update Network

**PUT** `/api/v1/networks/:id`

**Request Body:**
```json
{
  "name": "Database Network v2",
  "cidr": "10.0.0.0/16",
  "description": "All database servers"
}
```

---

### Delete Network

**DELETE** `/api/v1/networks/:id`

---

### Get Network Groups

**GET** `/api/v1/networks/:id/groups`

Get all groups that have this network assigned.

**Response (200 OK):**
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440010",
    "name": "IT Department",
    "description": "IT team members"
  }
]
```

---

### Add Group to Network

**POST** `/api/v1/networks/:id/groups`

**Request Body:**
```json
{
  "group_id": "550e8400-e29b-41d4-a716-446655440010"
}
```

---

### Remove Group from Network

**DELETE** `/api/v1/networks/:id/groups/:group_id`

---

## VPN Sessions

### Create Session

**POST** `/api/v1/vpn/sessions`

Called by OpenVPN server when a client connects. Available to any authenticated user.

**Request Body:**
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "vpn_ip": "10.8.0.101",
  "client_ip": "203.0.113.50",
  "connected_at": "2025-12-01T10:00:00Z"
}
```

**Response (201 Created):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440100",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "vpn_ip": "10.8.0.101",
  "client_ip": "203.0.113.50",
  "connected_at": "2025-12-01T10:00:00Z"
}
```

---

### Disconnect Session

**PUT** `/api/v1/vpn/sessions/:id/disconnect`

Called by OpenVPN server when a client disconnects. Available to any authenticated user.

**Request Body:**
```json
{
  "disconnected_at": "2025-12-01T12:00:00Z",
  "bytes_received": 104857600,
  "bytes_sent": 52428800,
  "disconnect_reason": "USER_REQUEST"
}
```

**Disconnect Reasons:**
- `USER_REQUEST` - User initiated disconnect
- `TIMEOUT` - Connection timeout
- `SERVER_SHUTDOWN` - VPN server shutdown
- `ERROR` - Connection error
- `ADMIN_ACTION` - Administrator disconnected user

---

### Create Traffic Stats

**POST** `/api/v1/vpn/traffic-stats`

Record periodic traffic statistics. Available to any authenticated user.

**Request Body:**
```json
{
  "session_id": "550e8400-e29b-41d4-a716-446655440100",
  "timestamp": "2025-12-01T10:05:00Z",
  "bytes_received_delta": 1048576,
  "bytes_sent_delta": 524288
}
```

---

### List Sessions (Admin Only)

**GET** `/api/v1/vpn/sessions`

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `page` | int | Page number |
| `page_size` | int | Items per page |
| `user_id` | UUID | Filter by user |
| `is_active` | bool | Filter active/closed sessions |
| `from` | datetime | Sessions after this time |
| `to` | datetime | Sessions before this time |

---

### Get Active Sessions (Admin Only)

**GET** `/api/v1/vpn/sessions/active`

Returns only currently active VPN connections.

**Response (200 OK):**
```json
{
  "sessions": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440100",
      "user": {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "username": "john.doe",
        "first_name": "John",
        "last_name": "Doe"
      },
      "vpn_ip": "10.8.0.101",
      "client_ip": "203.0.113.50",
      "connected_at": "2025-12-01T10:00:00Z",
      "disconnected_at": null,
      "bytes_received": 104857600,
      "bytes_sent": 52428800
    }
  ],
  "total": 5
}
```

---

### Get Session (Admin Only)

**GET** `/api/v1/vpn/sessions/:id`

---

### Get VPN Stats (Admin Only)

**GET** `/api/v1/vpn/stats`

Get aggregated VPN usage statistics.

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `from` | datetime | Stats from this time |
| `to` | datetime | Stats until this time |

**Response (200 OK):**
```json
{
  "total_sessions": 1500,
  "active_sessions": 25,
  "total_bytes_received": 10737418240,
  "total_bytes_sent": 5368709120,
  "unique_users": 50,
  "avg_session_duration": 7200
}
```

---

### Get User Stats (Admin Only)

**GET** `/api/v1/vpn/stats/users`

Get per-user VPN usage statistics.

**Response (200 OK):**
```json
{
  "users": [
    {
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "username": "john.doe",
      "total_sessions": 30,
      "total_bytes_received": 214748364,
      "total_bytes_sent": 107374182,
      "total_duration": 216000,
      "last_session": "2025-12-01T10:00:00Z"
    }
  ]
}
```

---

### List Traffic Stats (Admin Only)

**GET** `/api/v1/vpn/traffic-stats`

Get detailed traffic statistics entries.

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `session_id` | UUID | Filter by session |
| `from` | datetime | Stats from this time |
| `to` | datetime | Stats until this time |

---

## Audit Logs

All audit endpoints require `ADMIN` role. Audit logs are append-only.

### List Audit Logs

**GET** `/api/v1/audit`

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `page` | int | Page number |
| `page_size` | int | Items per page |
| `user_id` | UUID | Filter by user who performed action |
| `action` | string | Filter by action type |
| `entity_type` | string | Filter by entity type |
| `from` | datetime | Logs after this time |
| `to` | datetime | Logs before this time |

**Response (200 OK):**
```json
{
  "logs": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440200",
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "user": {
        "username": "admin"
      },
      "action": "CREATE",
      "entity_type": "user",
      "entity_id": "550e8400-e29b-41d4-a716-446655440001",
      "old_values": null,
      "new_values": {
        "username": "john.doe",
        "email": "john@example.com"
      },
      "ip_address": "192.168.1.100",
      "user_agent": "Mozilla/5.0...",
      "description": "Created new user",
      "created_at": "2025-12-01T10:00:00Z"
    }
  ],
  "total": 1000,
  "page": 1,
  "page_size": 10,
  "total_pages": 100
}
```

---

### Get Audit Log

**GET** `/api/v1/audit/:id`

---

### Get Available Actions

**GET** `/api/v1/audit/actions`

**Response (200 OK):**
```json
{
  "actions": ["CREATE", "READ", "UPDATE", "DELETE", "LOGIN", "LOGOUT"]
}
```

---

### Get Entity Types

**GET** `/api/v1/audit/entity-types`

**Response (200 OK):**
```json
{
  "entity_types": ["user", "group", "network", "vpn_session", "user_group", "network_group"]
}
```

---

### Get Audit Stats

**GET** `/api/v1/audit/stats`

**Response (200 OK):**
```json
{
  "total_logs": 10000,
  "by_action": {
    "CREATE": 2000,
    "READ": 5000,
    "UPDATE": 1500,
    "DELETE": 500,
    "LOGIN": 800,
    "LOGOUT": 200
  },
  "by_entity_type": {
    "user": 3000,
    "group": 1000,
    "network": 500,
    "vpn_session": 5500
  }
}
```

---

### Get Logs by Entity

**GET** `/api/v1/audit/entity/:entity_type/:entity_id`

Get all audit logs for a specific entity.

---

### Get Logs by User

**GET** `/api/v1/audit/user/:user_id`

Get all audit logs for actions performed by a specific user.

---

## Error Responses

All endpoints return consistent error responses:

```json
{
  "error": "Error Type",
  "message": "Detailed error message",
  "code": 400
}
```

**Common HTTP Status Codes:**
| Code | Description |
|------|-------------|
| 400 | Bad Request - Invalid input data |
| 401 | Unauthorized - Missing or invalid token |
| 403 | Forbidden - Insufficient permissions |
| 404 | Not Found - Resource doesn't exist |
| 409 | Conflict - Duplicate entry (username, email) |
| 500 | Internal Server Error |

---

## OpenVPN Integration

The API is designed to integrate with OpenVPN server scripts.

### Client Connect Script

Save as `/etc/openvpn/scripts/client-connect.sh`:

```bash
#!/bin/bash

# Configuration
API_URL="http://localhost:8080"
API_TOKEN="your-service-account-token"

# Get user ID from certificate CN or username
USER_ID=$(get_user_id_from_cn "$common_name")

# Create session
SESSION_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/vpn/sessions" \
  -H "Authorization: Bearer $API_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": \"$USER_ID\",
    \"vpn_ip\": \"$ifconfig_pool_remote_ip\",
    \"client_ip\": \"$trusted_ip\",
    \"connected_at\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\"
  }")

# Save session ID for disconnect script
SESSION_ID=$(echo "$SESSION_RESPONSE" | jq -r '.id')
echo "$SESSION_ID" > "/tmp/openvpn-session-$common_name"

exit 0
```

### Client Disconnect Script

Save as `/etc/openvpn/scripts/client-disconnect.sh`:

```bash
#!/bin/bash

# Configuration
API_URL="http://localhost:8080"
API_TOKEN="your-service-account-token"

# Get session ID
SESSION_ID=$(cat "/tmp/openvpn-session-$common_name" 2>/dev/null)

if [ -n "$SESSION_ID" ]; then
  # Update session with disconnect info
  curl -s -X PUT "$API_URL/api/v1/vpn/sessions/$SESSION_ID/disconnect" \
    -H "Authorization: Bearer $API_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
      \"disconnected_at\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",
      \"bytes_received\": $bytes_received,
      \"bytes_sent\": $bytes_sent,
      \"disconnect_reason\": \"USER_REQUEST\"
    }"

  # Cleanup
  rm -f "/tmp/openvpn-session-$common_name"
fi

exit 0
```

### Periodic Traffic Stats Script

Save as `/etc/openvpn/scripts/traffic-stats.sh` and run via cron:

```bash
#!/bin/bash

# Configuration
API_URL="http://localhost:8080"
API_TOKEN="your-service-account-token"
STATUS_FILE="/var/log/openvpn/status.log"

# Parse status file and send stats for each active client
while read line; do
  # Parse client data from status file
  # Send to API
  curl -s -X POST "$API_URL/api/v1/vpn/traffic-stats" \
    -H "Authorization: Bearer $API_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
      \"session_id\": \"$SESSION_ID\",
      \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",
      \"bytes_received_delta\": $BYTES_IN,
      \"bytes_sent_delta\": $BYTES_OUT
    }"
done < "$STATUS_FILE"
```

### OpenVPN Server Configuration

Add to your OpenVPN server config:

```
# Script settings
script-security 2
client-connect /etc/openvpn/scripts/client-connect.sh
client-disconnect /etc/openvpn/scripts/client-disconnect.sh

# Status file for traffic monitoring
status /var/log/openvpn/status.log 10
```

---

## Swagger Documentation

Interactive API documentation is available at:

```
http://localhost:8080/swagger/index.html
```

Access can be restricted by IP in the configuration file.

---

## Rate Limiting

Currently, no rate limiting is implemented. Consider using a reverse proxy (nginx, traefik) for rate limiting in production.

---

## Versioning

The API is versioned via URL path (`/api/v1/`). Breaking changes will be introduced in new versions (`/api/v2/`).
