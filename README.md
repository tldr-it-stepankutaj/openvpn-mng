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
- **VPN Auth API**: Dedicated API endpoints for OpenVPN server integration with token-based authentication
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

### 5. Generate secrets

```bash
# Generate JWT secret (required)
openssl rand -hex 32

# Generate VPN token (optional, for OpenVPN integration)
openssl rand -hex 32
```

### 6. Configure and run

```bash
cp config.yaml.example config.yaml
# Edit config.yaml with your settings (add generated secrets)
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
  # VPN token for OpenVPN server integration
  # Generate with: openssl rand -hex 32
  vpn_token: ""          # Leave empty to disable VPN Auth API

auth:
  # Generate with: openssl rand -hex 32
  jwt_secret: "your-super-secret-jwt-key-change-in-production"
  token_expiry: 24       # hours
  session_expiry: 8      # hours

logging:
  output: "stdout"       # "stdout", "file", or "both"
  path: ""
  format: "text"         # "text" or "json"
  level: "info"          # "debug", "info", "warn", "error"
```

### Generating Secrets

Always use cryptographically secure random strings for secrets:

```bash
# Generate JWT secret (64 characters hex = 256 bits)
openssl rand -hex 32

# Example output: a1b2c3d4e5f6...
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

## Testing

Run the test suite:

```bash
# Run all tests
make test

# Run with race detection
make test-race

# Run specific test categories
make test-services      # Service layer tests
make test-handlers      # HTTP handler tests
make test-middleware    # Middleware tests
make test-integration   # Integration tests

# Generate coverage report
make test-coverage-html
```

For detailed testing documentation, see **[Testing Guide](help/test.md)**.

## VPN Auth API (OpenVPN Integration)

The VPN Auth API provides dedicated endpoints for OpenVPN server integration using token-based authentication instead of a service account.

### Enable VPN Auth API

1. Generate a VPN token:
   ```bash
   openssl rand -hex 32
   ```

2. Add to `config.yaml`:
   ```yaml
   api:
     vpn_token: "your-generated-token"
   ```

3. Restart the application

### VPN Auth Endpoints

All endpoints require the `X-VPN-Token` header.

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/vpn-auth/authenticate` | POST | Validate VPN user credentials |
| `/api/v1/vpn-auth/users` | GET | List all active VPN users |
| `/api/v1/vpn-auth/users/{id}` | GET | Get user by ID |
| `/api/v1/vpn-auth/users/{id}/routes` | GET | Get user's network routes |
| `/api/v1/vpn-auth/users/by-username/{username}` | GET | Get user by username |
| `/api/v1/vpn-auth/sessions` | POST | Create VPN session |
| `/api/v1/vpn-auth/sessions/{id}/disconnect` | PUT | End VPN session |

### Example Usage

```bash
# Authenticate user
curl -X POST http://localhost:8080/api/v1/vpn-auth/authenticate \
  -H "X-VPN-Token: your-token" \
  -H "Content-Type: application/json" \
  -d '{"username": "user", "password": "pass"}'

# Get user routes
curl http://localhost:8080/api/v1/vpn-auth/users/by-username/testuser \
  -H "X-VPN-Token: your-token"
```

For complete OpenVPN integration guide, see **[Client Integration Guide](help/client.md)**.

## Documentation

- **[API Documentation](help/api.md)** - Complete REST API reference with examples
- **[Client Integration Guide](help/client.md)** - OpenVPN server integration guide
- **[Testing Guide](help/test.md)** - Test suite structure and usage
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
├── test/
│   ├── testutil/         # Test utilities and helpers
│   ├── services/         # Service layer tests
│   ├── handlers/         # HTTP handler tests
│   ├── middleware/       # Middleware tests
│   ├── dto/              # DTO tests
│   └── integration/      # Integration tests
├── web/
│   ├── static/           # CSS, JS assets
│   └── templates/        # HTML templates
├── docs/                 # Swagger generated docs
├── help/                 # Documentation
│   ├── api.md            # API documentation
│   └── test.md           # Testing guide
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
2. **Use strong secrets** - Generate secure random strings:
   ```bash
   # For jwt_secret and vpn_token
   openssl rand -hex 32
   ```
3. **Enable SSL/TLS** - Use a reverse proxy with HTTPS in production
4. **Database security** - Use strong passwords and restrict database access
5. **IP filtering** - Restrict Swagger access to trusted IPs in production
6. **Audit logging** - Monitor audit logs for suspicious activity
7. **User validity** - Use `valid_from`/`valid_to` for temporary access
8. **VPN API authentication** - Use VPN token instead of service account for OpenVPN integration

## Docker

### Quick Start with Docker Compose

```bash
# Start with PostgreSQL
docker-compose up -d

# View logs
docker-compose logs -f app

# Stop
docker-compose down
```

The application will be available at `http://localhost:8080`.

### Environment Variables

The application supports configuration via environment variables (useful for Docker/Kubernetes):

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_HOST` | Server bind address | `0.0.0.0` |
| `SERVER_PORT` | Server port | `8080` |
| `DB_TYPE` | Database type (`postgres` or `mysql`) | `postgres` |
| `DB_HOST` | Database host | - |
| `DB_PORT` | Database port | `5432` (postgres) / `3306` (mysql) |
| `DB_USERNAME` | Database username | - |
| `DB_PASSWORD` | Database password | - |
| `DB_DATABASE` | Database name | - |
| `DB_SSLMODE` | PostgreSQL SSL mode | `disable` |
| `AUTH_JWT_SECRET` | JWT signing secret (generate with `openssl rand -hex 32`) | - |
| `AUTH_TOKEN_EXPIRY` | Token expiry in hours | `24` |
| `AUTH_SESSION_EXPIRY` | Session expiry in hours | `8` |
| `API_ENABLED` | Enable REST API | `true` |
| `API_SWAGGER_ENABLED` | Enable Swagger UI | `true` |
| `API_SWAGGER_ALLOWED_IPS` | Comma-separated CIDR list | - |
| `API_VPN_TOKEN` | VPN Auth API token (generate with `openssl rand -hex 32`) | - |
| `LOG_OUTPUT` | Log output (`stdout`, `file`, `both`) | `stdout` |
| `LOG_FORMAT` | Log format (`text`, `json`) | `text` |
| `LOG_LEVEL` | Log level (`debug`, `info`, `warn`, `error`) | `info` |

Environment variables override values from `config.yaml`.

### Building Docker Image

```bash
# Build image
docker build -t openvpn-mng:latest .

# Generate secrets first
JWT_SECRET=$(openssl rand -hex 32)
VPN_TOKEN=$(openssl rand -hex 32)

# Run standalone (requires external database)
docker run -p 8080:8080 \
  -e DB_HOST=your-db-host \
  -e DB_USERNAME=your-user \
  -e DB_PASSWORD=your-password \
  -e DB_DATABASE=openvpn_mng \
  -e AUTH_JWT_SECRET=$JWT_SECRET \
  -e API_VPN_TOKEN=$VPN_TOKEN \
  openvpn-mng:latest
```

### Kubernetes

For Kubernetes deployments, use JSON logging format:

```yaml
env:
  - name: LOG_OUTPUT
    value: "stdout"
  - name: LOG_FORMAT
    value: "json"
  - name: LOG_LEVEL
    value: "info"
```

## License

MIT License - see LICENSE file for details.
