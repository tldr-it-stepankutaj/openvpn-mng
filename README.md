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
- **Audit Logging**: Track all operations (create, read, update, delete, login, logout)
- **REST API**: Full-featured API with Swagger documentation (see [API Documentation](help/api.md))
- **Web Interface**: Bootstrap-based HTML interface for user-friendly management
- **Database Support**: PostgreSQL and MySQL support via GORM
- **JWT Authentication**: Secure token-based authentication
- **IP Filtering**: Restrict Swagger documentation access by IP/CIDR ranges
- **Flexible Logging**: Configurable output (stdout/file), format (text/JSON), and log levels

## Requirements

- Go 1.21 or higher
- PostgreSQL 12+ or MySQL 8+
- swag CLI tool (for Swagger documentation generation)

## Quick Start

### 1. Clone and install

```bash
git clone https://github.com/tldr-it-stepankutaj/openvpn-mng.git
cd openvpn-mng
go mod download
```

### 2. Install swag CLI

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

Make sure `$GOPATH/bin` (usually `~/go/bin`) is in your `PATH`.

### 3. Generate Swagger documentation

```bash
swag init -g cmd/server/main.go -o docs
```

### 4. Setup database

**PostgreSQL:**
```sql
CREATE DATABASE openvpn_mng;
```

**MySQL:**
```sql
CREATE DATABASE openvpn_mng CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 5. Configure and run

```bash
cp config.yaml.example config.yaml
# Edit config.yaml with your settings
./openvpn-mng
```

### Default admin credentials

- **Username:** `admin`
- **Password:** `admin123`

**Important:** Change the default password immediately after first login!

## Configuration

The application is configured via `config.yaml`:

```yaml
server:
  host: "127.0.0.1"
  port: 8080

database:
  type: "postgres"       # "postgres" or "mysql"
  host: "localhost"
  port: 5432
  username: "your_user"
  password: "your_password"
  database: "openvpn_mng"
  sslmode: "disable"

api:
  enabled: true
  swagger_enabled: true
  swagger_allowed_ips:
    - "127.0.0.1/32"
    - "::1/128"

auth:
  jwt_secret: "your-super-secret-jwt-key-change-in-production"
  token_expiry: 24
  session_expiry: 8

logging:
  output: "stdout"       # "stdout", "file", or "both"
  path: ""
  format: "text"         # "text" or "json"
  level: "info"          # "debug", "info", "warn", "error"
```

### Logging Options

| Option | Values | Description |
|--------|--------|-------------|
| `output` | stdout, file, both | Where to send logs |
| `format` | text, json | Log format (JSON recommended for aggregators) |
| `level` | debug, info, warn, error | Minimum log level |

## Building

```bash
# Development
go build -o openvpn-mng ./cmd/server

# Production (Linux)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o openvpn-mng ./cmd/server

# Windows
GOOS=windows GOARCH=amd64 go build -o openvpn-mng.exe ./cmd/server

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o openvpn-mng ./cmd/server
```

## Web Interface

| URL | Description |
|-----|-------------|
| `/login` | Login page |
| `/dashboard` | Dashboard |
| `/users` | User management |
| `/groups` | Group management |
| `/networks` | Network management (Admin only) |
| `/audit` | Audit logs (Admin only) |
| `/sessions` | VPN session history (Admin only) |
| `/profile` | User profile |

## Documentation

- **[API Documentation](help/api.md)** - Complete REST API reference with examples
- **Swagger UI** - Interactive API docs at `http://localhost:8080/swagger/index.html`

## Project Structure

```
openvpn-mng/
├── cmd/server/           # Application entry point
├── internal/
│   ├── config/           # Configuration loading
│   ├── database/         # Database initialization
│   ├── dto/              # Data transfer objects
│   ├── handlers/         # HTTP handlers
│   ├── logger/           # Logging utilities
│   ├── middleware/       # Auth, audit, IP filtering
│   ├── models/           # Database models
│   ├── routes/           # Route definitions
│   └── services/         # Business logic
├── web/
│   ├── static/           # CSS, JS assets
│   └── templates/        # HTML templates
├── docs/                 # Swagger generated docs
├── help/                 # Documentation
│   └── api.md            # API documentation
└── config.yaml           # Configuration file
```

## Database Schema

### Core Tables

- **users** - User accounts with VPN settings (is_active, valid_from, valid_to, vpn_ip)
- **groups** - User groups (IT, HR, Finance, etc.)
- **networks** - Network definitions (CIDR ranges)
- **vpn_sessions** - VPN connection history
- **vpn_traffic_stats** - Traffic statistics
- **audit_logs** - Audit trail

### Junction Tables

- **user_groups** - Many-to-many: Users <-> Groups
- **network_groups** - Many-to-many: Networks <-> Groups

## Security Considerations

1. **Change default credentials** - Always change the default admin password
2. **Use strong JWT secret** - Generate a secure random string for `jwt_secret`
3. **Enable SSL/TLS** - Use a reverse proxy with HTTPS in production
4. **Database security** - Use strong passwords and restrict database access
5. **IP filtering** - Restrict Swagger access to trusted IPs in production
6. **Audit logging** - Monitor audit logs for suspicious activity
7. **User validity** - Use `valid_from`/`valid_to` for temporary access
8. **VPN API authentication** - Use dedicated service account for VPN server integration

## Docker / Kubernetes

For containerized deployments:

```yaml
logging:
  output: "stdout"
  format: "json"
  level: "info"
```

## License

MIT License - see LICENSE file for details.
