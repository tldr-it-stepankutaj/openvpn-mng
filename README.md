# OpenVPN Manager

A web-based management system for OpenVPN users, groups, and access control. Built with Go, Gin framework, and GORM ORM.

## Features

- **User Management**: Create, update, delete users with role-based access control
- **Group Management**: Organize users into groups (IT, HR, Finance, etc.)
- **Role-Based Access Control (RBAC)**:
  - `USER` - Can only view and edit their own profile
  - `MANAGER` - Can manage users assigned to them
  - `ADMIN` - Full access to all resources
- **Manager Hierarchy**: Users can be assigned to a manager for hierarchical access control
- **Audit Logging**: Track all operations (create, read, update, delete, login, logout)
- **REST API**: Full-featured API with Swagger documentation
- **Web Interface**: Bootstrap-based HTML interface for user-friendly management
- **Database Support**: PostgreSQL and MySQL support via GORM
- **JWT Authentication**: Secure token-based authentication
- **IP Filtering**: Restrict Swagger documentation access by IP/CIDR ranges

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
```

### IP Filtering for Swagger

The `swagger_allowed_ips` setting controls which IP addresses can access the Swagger documentation:

- `127.0.0.1/32` - Single IPv4 address (localhost)
- `::1/128` - Single IPv6 address (localhost)
- `192.168.1.0/24` - Entire subnet (192.168.1.0 - 192.168.1.255)
- `172.16.0.0/16` - Large subnet
- `0.0.0.0/0` - Allow all IPv4 addresses
- `::/0` - Allow all IPv6 addresses

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

## API Endpoints

### Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/login` | Login and get JWT token |
| POST | `/api/v1/auth/logout` | Logout (requires auth) |
| GET | `/api/v1/auth/me` | Get current user info |

### Users

| Method | Endpoint | Description | Required Role |
|--------|----------|-------------|---------------|
| GET | `/api/v1/users` | List all users | MANAGER, ADMIN |
| POST | `/api/v1/users` | Create user | ADMIN |
| GET | `/api/v1/users/:id` | Get user details | Any (own) / MANAGER, ADMIN |
| PUT | `/api/v1/users/:id` | Update user | MANAGER, ADMIN |
| DELETE | `/api/v1/users/:id` | Delete user | ADMIN |
| PUT | `/api/v1/users/profile` | Update own profile | Any |
| PUT | `/api/v1/users/password` | Change own password | Any |

### Groups

| Method | Endpoint | Description | Required Role |
|--------|----------|-------------|---------------|
| GET | `/api/v1/groups` | List all groups | Any |
| POST | `/api/v1/groups` | Create group | ADMIN |
| GET | `/api/v1/groups/:id` | Get group details | Any |
| PUT | `/api/v1/groups/:id` | Update group | ADMIN |
| DELETE | `/api/v1/groups/:id` | Delete group | ADMIN |
| GET | `/api/v1/groups/:id/users` | Get group members | Any |
| POST | `/api/v1/groups/:id/users` | Add user to group | MANAGER, ADMIN |
| DELETE | `/api/v1/groups/:id/users/:user_id` | Remove user from group | MANAGER, ADMIN |

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
│       └── main.go           # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go         # Configuration loading
│   ├── database/
│   │   └── database.go       # Database initialization
│   ├── dto/
│   │   ├── auth.go           # Auth DTOs
│   │   ├── audit.go          # Audit DTOs
│   │   ├── group.go          # Group DTOs
│   │   └── user.go           # User DTOs
│   ├── handlers/
│   │   ├── auth_handler.go   # Auth endpoints
│   │   ├── group_handler.go  # Group endpoints
│   │   ├── user_handler.go   # User endpoints
│   │   └── web_handler.go    # HTML page handlers
│   ├── middleware/
│   │   ├── auth.go           # JWT authentication
│   │   ├── audit.go          # Audit logging
│   │   └── ip_filter.go      # IP-based access control
│   ├── models/
│   │   ├── audit.go          # Audit log model
│   │   ├── group.go          # Group model
│   │   ├── role.go           # Role enum
│   │   └── user.go           # User model
│   ├── routes/
│   │   └── routes.go         # Route definitions
│   └── services/
│       ├── auth_service.go   # Auth business logic
│       ├── group_service.go  # Group business logic
│       └── user_service.go   # User business logic
├── web/
│   ├── static/
│   │   ├── css/
│   │   │   └── style.css     # Custom styles
│   │   └── js/
│   │       └── app.js        # Frontend JavaScript
│   └── templates/
│       ├── dashboard.html
│       ├── error.html
│       ├── groups.html
│       ├── index.html
│       ├── login.html
│       ├── profile.html
│       ├── user_detail.html
│       └── users.html
├── docs/                      # Swagger generated docs
├── config.yaml               # Configuration file
├── go.mod
├── go.sum
└── README.md
```

## Security Considerations

1. **Change default credentials**: Always change the default admin password
2. **Use strong JWT secret**: Generate a secure random string for `jwt_secret`
3. **Enable SSL/TLS**: Use a reverse proxy (nginx, traefik) with HTTPS in production
4. **Database security**: Use strong passwords and restrict database access
5. **IP filtering**: Restrict Swagger access to trusted IPs in production
6. **Audit logging**: Monitor audit logs for suspicious activity

## License

MIT License - see LICENSE file for details.
