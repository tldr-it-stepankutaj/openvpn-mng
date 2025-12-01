# OpenVPN Manager

A web-based management system for OpenVPN users, groups, networks, and access control. Built with Go, Gin framework, and GORM ORM.

## Features

- **User Management**: Create, update, delete users with role-based access control
- **VPN User Validity**: Control user access with `is_active`, `valid_from`, `valid_to` fields
- **Static VPN IP**: Optionally assign static VPN IP addresses to users
- **Group Management**: Organize users into groups (IT, HR, Finance, etc.)
- **Network Management**: Define network segments (IP/CIDR) and assign them to groups
- **VPN Session Tracking**: Monitor active connections, traffic statistics, and usage history
- **Role-Based Access Control (RBAC)**:
  - `USER` - Can only view and edit their own profile
  - `MANAGER` - Can create and manage users assigned to them
  - `ADMIN` - Full access to all resources including networks, VPN sessions, and audit logs
- **Manager Hierarchy**: Users can be assigned to a manager for hierarchical access control
- **Audit Logging**: Track all operations (create, read, update, delete, login, logout) with full API access
- **REST API**: Full-featured API with Swagger documentation
- **Web Interface**: Bootstrap-based HTML interface for user-friendly management
- **Database Support**: PostgreSQL and MySQL support via GORM
- **JWT Authentication**: Secure token-based authentication
- **IP Filtering**: Restrict Swagger documentation access by IP/CIDR ranges
- **Flexible Logging**: Configurable output (stdout/file), format (text/JSON), and log levels

## Requirements

- Go 1.21 or higher
- PostgreSQL 12+ or MySQL 8+
- swag CLI tool (for Swagger documentation generation)

## Installation

### 1. Clone the repository

```bash
git clone https://github.com/tldr-it-stepankutaj/openvpn-mng.git
cd openvpn-mng
```

### 2. Install dependencies

```bash
go mod download
```

### 3. Install swag CLI (for Swagger docs)

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

Make sure `$GOPATH/bin` (usually `~/go/bin`) is in your `PATH`.

### 4. Generate Swagger documentation

```bash
swag init -g cmd/server/main.go -o docs
```

### 5. Setup database

Create a database in PostgreSQL or MySQL:

**PostgreSQL:**
```sql
CREATE DATABASE openvpn_mng;
```

**MySQL:**
```sql
CREATE DATABASE openvpn_mng CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 6. Configure the application

Copy and edit the configuration file:

```bash
cp config.yaml.example config.yaml
# Edit config.yaml with your settings
```

## Configuration

The application is configured via `config.yaml`:

```yaml
server:
  host: "127.0.0.1"    # Server bind address
  port: 8080           # Server port

database:
  type: "postgres"     # "postgres" or "mysql"
  host: "localhost"
  port: 5432           # 5432 for PostgreSQL, 3306 for MySQL
  username: "your_user"
  password: "your_password"
  database: "openvpn_mng"
  sslmode: "disable"   # PostgreSQL only: disable, require, verify-ca, verify-full

api:
  enabled: true              # Enable/disable REST API
  swagger_enabled: true      # Enable/disable Swagger UI
  swagger_allowed_ips:       # IP ranges allowed to access Swagger (CIDR notation)
    - "127.0.0.1/32"         # Localhost IPv4
    - "::1/128"              # Localhost IPv6
    # - "192.168.1.0/24"     # Allow entire subnet
    # - "0.0.0.0/0"          # Allow all IPs (not recommended for production)

auth:
  jwt_secret: "your-super-secret-jwt-key-change-in-production"
  token_expiry: 24     # JWT token expiry in hours
  session_expiry: 8    # Session expiry in hours

logging:
  output: "stdout"     # "stdout" (default), "file", or "both"
  path: ""             # Directory for log files (empty = current directory)
  format: "text"       # "text" (default) or "json"
  level: "info"        # "debug", "info", "warn", "error"
```

### IP Filtering for Swagger

The `swagger_allowed_ips` setting controls which IP addresses can access the Swagger documentation:

- `127.0.0.1/32` - Single IPv4 address (localhost)
- `::1/128` - Single IPv6 address (localhost)
- `192.168.1.0/24` - Entire subnet (192.168.1.0 - 192.168.1.255)
- `172.16.0.0/16` - Large subnet
- `0.0.0.0/0` - Allow all IPv4 addresses
- `::/0` - Allow all IPv6 addresses

### Logging Configuration

The `logging` section controls application logging behavior:

#### Output Options

| Value | Description |
|-------|-------------|
| `stdout` | Logs to standard output only (default, recommended for Kubernetes/Docker) |
| `file` | Logs to file only (creates `openvpn-mng-YYYY-MM-DD.log`) |
| `both` | Logs to both stdout and file |

#### Format Options

| Value | Description |
|-------|-------------|
| `text` | Human-readable format (default) |
| `json` | JSON format for machine parsing (recommended for log aggregators) |

#### Log Levels

| Level | Description |
|-------|-------------|
| `debug` | All messages including SQL queries and detailed migration info |
| `info` | Standard operational messages (default) |
| `warn` | Warnings and errors only |
| `error` | Errors only |

#### Example Log Output

**Text format:**
```
time=2025-12-01T07:54:09.499+01:00 level=INFO msg="OpenVPN Manager starting" version=1.0.0
time=2025-12-01T07:54:09.509+01:00 level=INFO msg="Connected to database" type=postgres host=localhost port=5432
time=2025-12-01T07:54:15.000+01:00 level=INFO msg="HTTP request" status=200 latency=515.667µs client_ip=127.0.0.1 method=GET path=/swagger/index.html
time=2025-12-01T07:57:55.000+01:00 level=WARN msg="HTTP request" status=401 latency=91.5µs client_ip=127.0.0.1 method=GET path=/api/v1/auth/me
```

**JSON format:**
```json
{"time":"2025-12-01T07:55:19.746321+01:00","level":"INFO","msg":"OpenVPN Manager starting","version":"1.0.0"}
{"time":"2025-12-01T07:55:19.755857+01:00","level":"INFO","msg":"Connected to database","type":"postgres","host":"localhost","port":5432}
{"time":"2025-12-01T07:55:30.000+01:00","level":"INFO","msg":"HTTP request","status":200,"latency":"515.667µs","client_ip":"127.0.0.1","method":"GET","path":"/swagger/index.html"}
```

## Building

### Development build

```bash
go build -o openvpn-mng ./cmd/server
```

### Production build (with optimizations)

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o openvpn-mng ./cmd/server
```

### Build for different platforms

**Windows:**
```bash
GOOS=windows GOARCH=amd64 go build -o openvpn-mng.exe ./cmd/server
```

**macOS (Intel):**
```bash
GOOS=darwin GOARCH=amd64 go build -o openvpn-mng ./cmd/server
```

**macOS (Apple Silicon):**
```bash
GOOS=darwin GOARCH=arm64 go build -o openvpn-mng ./cmd/server
```

## Running

### Start the server

```bash
./openvpn-mng
```

Or with a custom config path:

```bash
./openvpn-mng -config /path/to/config.yaml
```

### Default admin credentials

On first run, a default admin user is created:
- **Username:** `admin`
- **Password:** `admin123`

**Important:** Change the default password immediately after first login!

## Docker / Kubernetes Deployment

For containerized deployments, use these recommended settings:

```yaml
logging:
  output: "stdout"    # Let the container runtime collect logs
  format: "json"      # Structured logs for log aggregators (Fluentd, Loki, ELK)
  level: "info"
```

Environment variables can override config file settings if needed.

## API Endpoints

### Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/login` | Login and get JWT token |
| POST | `/api/v1/auth/logout` | Logout (requires auth) |
| GET | `/api/v1/auth/me` | Get current user info |

**Login Validation:**
- Checks `is_active` - returns "User account is inactive" if false
- Checks `valid_from` - returns "User account is not yet valid" if before date
- Checks `valid_to` - returns "User account has expired" if after date

### Users

| Method | Endpoint | Description | Required Role |
|--------|----------|-------------|---------------|
| GET | `/api/v1/users` | List users (filtered by role) | Any |
| POST | `/api/v1/users` | Create user | MANAGER, ADMIN |
| GET | `/api/v1/users/:id` | Get user details | Own / Subordinates / ADMIN |
| PUT | `/api/v1/users/:id` | Update user | Subordinates / ADMIN |
| DELETE | `/api/v1/users/:id` | Delete user | ADMIN |
| PUT | `/api/v1/users/profile` | Update own profile | Any |
| PUT | `/api/v1/users/password` | Change own password | Any |

**User Fields:**
- `is_active` (bool) - If false, user cannot login
- `valid_from` (date, optional) - User valid from this date
- `valid_to` (date, optional) - User valid until this date
- `vpn_ip` (string, optional) - Static VPN IP address

**Note:**
- `USER` role can only see themselves in the list
- `MANAGER` role sees only their subordinates
- `ADMIN` role sees all users

### Groups

| Method | Endpoint | Description | Required Role |
|--------|----------|-------------|---------------|
| GET | `/api/v1/groups` | List all groups | Any |
| POST | `/api/v1/groups` | Create group | ADMIN |
| GET | `/api/v1/groups/:id` | Get group details | Any |
| PUT | `/api/v1/groups/:id` | Update group | ADMIN |
| DELETE | `/api/v1/groups/:id` | Delete group | ADMIN |
| GET | `/api/v1/groups/:id/users` | Get group members (filtered) | Any |
| POST | `/api/v1/groups/:id/users` | Add user to group | MANAGER, ADMIN |
| DELETE | `/api/v1/groups/:id/users/:user_id` | Remove user from group | MANAGER, ADMIN |

### Networks

| Method | Endpoint | Description | Required Role |
|--------|----------|-------------|---------------|
| GET | `/api/v1/networks` | List all networks | ADMIN |
| POST | `/api/v1/networks` | Create network | ADMIN |
| GET | `/api/v1/networks/:id` | Get network details | ADMIN |
| PUT | `/api/v1/networks/:id` | Update network | ADMIN |
| DELETE | `/api/v1/networks/:id` | Delete network | ADMIN |
| GET | `/api/v1/networks/:id/groups` | Get groups assigned to network | ADMIN |
| POST | `/api/v1/networks/:id/groups` | Add group to network | ADMIN |
| DELETE | `/api/v1/networks/:id/groups/:group_id` | Remove group from network | ADMIN |

**Network CIDR Examples:**
- `192.168.1.0/24` - Subnet (192.168.1.0 - 192.168.1.255)
- `10.0.0.1/32` - Single IP address
- `10.0.0.1` - Single IP (automatically converted to /32)

### VPN Sessions

| Method | Endpoint | Description | Required Role |
|--------|----------|-------------|---------------|
| POST | `/api/v1/vpn/sessions` | Create session (VPN server calls this) | Any authenticated |
| PUT | `/api/v1/vpn/sessions/:id/disconnect` | Update session with disconnect info | Any authenticated |
| POST | `/api/v1/vpn/traffic-stats` | Create traffic stats entry | Any authenticated |
| GET | `/api/v1/vpn/sessions` | List all sessions | ADMIN |
| GET | `/api/v1/vpn/sessions/active` | Get active sessions | ADMIN |
| GET | `/api/v1/vpn/sessions/:id` | Get session details | ADMIN |
| GET | `/api/v1/vpn/stats` | Get aggregated usage statistics | ADMIN |
| GET | `/api/v1/vpn/stats/users` | Get usage statistics per user | ADMIN |
| GET | `/api/v1/vpn/traffic-stats` | List traffic stats | ADMIN |

**Session Fields:**
- `user_id` - User who connected
- `vpn_ip` - VPN IP address assigned
- `client_ip` - Client's real IP address
- `connected_at`, `disconnected_at` - Connection timestamps
- `bytes_received`, `bytes_sent` - Traffic statistics
- `disconnect_reason` - USER_REQUEST, TIMEOUT, SERVER_SHUTDOWN, ERROR, ADMIN_ACTION

### Audit Logs

| Method | Endpoint | Description | Required Role |
|--------|----------|-------------|---------------|
| GET | `/api/v1/audit` | List audit logs | ADMIN |
| GET | `/api/v1/audit/:id` | Get audit log details | ADMIN |
| GET | `/api/v1/audit/actions` | Get available audit actions | ADMIN |
| GET | `/api/v1/audit/entity-types` | Get logged entity types | ADMIN |
| GET | `/api/v1/audit/stats` | Get audit statistics | ADMIN |
| GET | `/api/v1/audit/entity/:type/:id` | Get logs for specific entity | ADMIN |
| GET | `/api/v1/audit/user/:id` | Get logs for specific user | ADMIN |

**Audit Actions:**
- `CREATE`, `READ`, `UPDATE`, `DELETE`, `LOGIN`, `LOGOUT`

**Note:** Audit logs are append-only - no update or delete endpoints exist.

### Swagger Documentation

Access Swagger UI at: `http://localhost:8080/swagger/index.html`

## Web Interface

| URL | Description |
|-----|-------------|
| `/` | Index (redirects to login) |
| `/login` | Login page |
| `/dashboard` | Dashboard (requires auth) |
| `/users` | User management |
| `/users/:id` | User details |
| `/groups` | Group management |
| `/profile` | Current user profile |

## Project Structure

```
openvpn-mng/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration loading
│   ├── database/
│   │   └── database.go          # Database initialization
│   ├── dto/
│   │   ├── auth.go              # Auth DTOs
│   │   ├── audit.go             # Audit DTOs
│   │   ├── group.go             # Group DTOs
│   │   ├── network.go           # Network DTOs
│   │   ├── user.go              # User DTOs
│   │   └── vpn_session.go       # VPN session DTOs
│   ├── handlers/
│   │   ├── auth_handler.go      # Auth endpoints
│   │   ├── audit_handler.go     # Audit endpoints
│   │   ├── group_handler.go     # Group endpoints
│   │   ├── network_handler.go   # Network endpoints
│   │   ├── user_handler.go      # User endpoints
│   │   ├── vpn_session_handler.go # VPN session endpoints
│   │   └── web_handler.go       # HTML page handlers
│   ├── logger/
│   │   ├── logger.go            # Application logger
│   │   └── gin.go               # Gin HTTP request logger
│   ├── middleware/
│   │   ├── auth.go              # JWT authentication
│   │   ├── audit.go             # Audit logging
│   │   └── ip_filter.go         # IP-based access control
│   ├── models/
│   │   ├── audit.go             # Audit log model
│   │   ├── group.go             # Group model
│   │   ├── network.go           # Network model
│   │   ├── role.go              # Role enum
│   │   ├── user.go              # User model
│   │   ├── vpn_session.go       # VPN session model
│   │   └── vpn_traffic_stats.go # VPN traffic stats model
│   ├── routes/
│   │   └── routes.go            # Route definitions
│   └── services/
│       ├── auth_service.go      # Auth business logic
│       ├── audit_service.go     # Audit business logic
│       ├── group_service.go     # Group business logic
│       ├── network_service.go   # Network business logic
│       ├── user_service.go      # User business logic
│       └── vpn_session_service.go # VPN session business logic
├── web/
│   ├── static/
│   │   ├── css/
│   │   │   └── style.css        # Custom styles
│   │   └── js/
│   │       └── app.js           # Frontend JavaScript
│   └── templates/
│       ├── dashboard.html
│       ├── error.html
│       ├── groups.html
│       ├── index.html
│       ├── login.html
│       ├── profile.html
│       ├── user_detail.html
│       └── users.html
├── docs/                         # Swagger generated docs
├── config.yaml                   # Configuration file
├── config.yaml.example           # Configuration template
├── go.mod
├── go.sum
└── README.md
```

## Database Schema

### Users Table
- `id` (UUID) - Primary key
- `username` (string) - Unique username
- `password` (string) - Hashed password
- `manager_id` (UUID, nullable) - Reference to manager user
- `first_name`, `middle_name`, `last_name` (string)
- `email` (string) - Unique email
- `telephone` (string)
- `role` (enum) - USER, MANAGER, ADMIN
- `is_active` (bool) - Whether user can login (default: true)
- `valid_from` (date, nullable) - User valid from this date
- `valid_to` (date, nullable) - User valid until this date
- `vpn_ip` (string, nullable) - Static VPN IP address
- `created_at`, `updated_at`, `deleted_at` - Timestamps
- `created_by`, `updated_by` (UUID) - Audit fields

### Groups Table
- `id` (UUID) - Primary key
- `name` (string) - Unique group name
- `description` (string)
- `created_at`, `updated_at`, `deleted_at` - Timestamps
- `created_by`, `updated_by` (UUID) - Audit fields

### Networks Table
- `id` (UUID) - Primary key
- `name` (string) - Unique network name
- `cidr` (string) - IP address or CIDR range
- `description` (string)
- `created_at`, `updated_at`, `deleted_at` - Timestamps
- `created_by`, `updated_by` (UUID) - Audit fields

### VPN Sessions Table
- `id` (UUID) - Primary key
- `user_id` (UUID) - Reference to user
- `vpn_ip` (string) - Assigned VPN IP address
- `client_ip` (string) - Client's real IP address
- `connected_at` (timestamp) - Connection start time
- `disconnected_at` (timestamp, nullable) - Connection end time
- `bytes_received` (int64) - Total bytes received
- `bytes_sent` (int64) - Total bytes sent
- `disconnect_reason` (enum) - USER_REQUEST, TIMEOUT, SERVER_SHUTDOWN, ERROR, ADMIN_ACTION

### VPN Traffic Stats Table
- `id` (UUID) - Primary key
- `session_id` (UUID) - Reference to VPN session
- `timestamp` (timestamp) - Measurement timestamp
- `bytes_received_delta` (int64) - Bytes received since last measurement
- `bytes_sent_delta` (int64) - Bytes sent since last measurement

### Junction Tables
- `user_groups` - Many-to-many: Users <-> Groups
- `network_groups` - Many-to-many: Networks <-> Groups

### Audit Logs Table
- `id` (UUID) - Primary key
- `user_id` (UUID) - Who performed the action
- `action` (enum) - CREATE, READ, UPDATE, DELETE, LOGIN, LOGOUT
- `entity_type` (string) - user, group, network, etc.
- `entity_id` (UUID) - Affected entity
- `old_values`, `new_values` (JSON) - Change details
- `ip_address`, `user_agent` (string) - Request info
- `created_at` - Timestamp

## OpenVPN Integration

The API is designed to integrate with OpenVPN server scripts:

### Connect Script (client-connect)
```bash
#!/bin/bash
# Called when client connects
curl -X POST "http://localhost:8080/api/v1/vpn/sessions" \
  -H "Authorization: Bearer $API_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": \"$USER_ID\",
    \"vpn_ip\": \"$ifconfig_pool_remote_ip\",
    \"client_ip\": \"$trusted_ip\",
    \"connected_at\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\"
  }"
```

### Disconnect Script (client-disconnect)
```bash
#!/bin/bash
# Called when client disconnects
curl -X PUT "http://localhost:8080/api/v1/vpn/sessions/$SESSION_ID/disconnect" \
  -H "Authorization: Bearer $API_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"disconnected_at\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",
    \"bytes_received\": $bytes_received,
    \"bytes_sent\": $bytes_sent,
    \"disconnect_reason\": \"USER_REQUEST\"
  }"
```

## Security Considerations

1. **Change default credentials**: Always change the default admin password
2. **Use strong JWT secret**: Generate a secure random string for `jwt_secret`
3. **Enable SSL/TLS**: Use a reverse proxy (nginx, traefik) with HTTPS in production
4. **Database security**: Use strong passwords and restrict database access
5. **IP filtering**: Restrict Swagger access to trusted IPs in production
6. **Audit logging**: Monitor audit logs for suspicious activity
7. **Log security**: In production with `output: "file"`, ensure log directory has proper permissions
8. **User validity**: Use `valid_from`/`valid_to` for temporary access
9. **VPN API authentication**: Use dedicated service account for VPN server integration

## License

MIT License - see LICENSE file for details.
