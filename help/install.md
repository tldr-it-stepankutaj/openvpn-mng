# OpenVPN Manager Installation Guide

This guide covers all installation methods for OpenVPN Manager.

![Login Page](images/login.png)

## Table of Contents

- [Prerequisites](#prerequisites)
- [Installation Methods](#installation-methods)
  - [DEB Package (Debian, Ubuntu)](#deb-package-debian-ubuntu)
  - [RPM Package (RHEL, AlmaLinux, Rocky Linux)](#rpm-package-rhel-almalinux-rocky-linux)
  - [Docker](#docker)
  - [Docker Compose](#docker-compose)
  - [From Source](#from-source)
- [Post-Installation Configuration](#post-installation-configuration)
- [Starting the Service](#starting-the-service)
- [Upgrading](#upgrading)
- [Uninstallation](#uninstallation)

---

## Prerequisites

### Database

OpenVPN Manager requires a PostgreSQL or MySQL database.

**PostgreSQL (recommended):**

```bash
# Debian/Ubuntu
sudo apt install postgresql postgresql-contrib

# RHEL/AlmaLinux/Rocky
sudo dnf install postgresql-server postgresql-contrib
sudo postgresql-setup --initdb
sudo systemctl enable --now postgresql
```

Create database:

```sql
sudo -u postgres psql

CREATE USER openvpn WITH PASSWORD 'your-secure-password';
CREATE DATABASE openvpn_mng OWNER openvpn;
GRANT ALL PRIVILEGES ON DATABASE openvpn_mng TO openvpn;
\q
```

**MySQL/MariaDB:**

```bash
# Debian/Ubuntu
sudo apt install mariadb-server

# RHEL/AlmaLinux/Rocky
sudo dnf install mariadb-server
sudo systemctl enable --now mariadb
```

Create database:

```sql
sudo mysql

CREATE DATABASE openvpn_mng CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'openvpn'@'localhost' IDENTIFIED BY 'your-secure-password';
GRANT ALL PRIVILEGES ON openvpn_mng.* TO 'openvpn'@'localhost';
FLUSH PRIVILEGES;
EXIT;
```

### Generate Secrets

Before installation, generate the required secrets:

```bash
# JWT Secret (required) - for user authentication
openssl rand -hex 32

# VPN Token (optional) - for OpenVPN server integration
openssl rand -hex 32
```

Save these values - you'll need them during configuration.

---

## Installation Methods

### DEB Package (Debian, Ubuntu)

Supported: Debian 11+, Ubuntu 20.04+

**Download and install:**

```bash
# Download latest release (replace VERSION with actual version)
VERSION="1.1.0"
wget https://github.com/tldr-it-stepankutaj/openvpn-mng/releases/download/v${VERSION}/openvpn-mng_${VERSION}_linux_amd64.deb

# Verify checksum (recommended)
wget https://github.com/tldr-it-stepankutaj/openvpn-mng/releases/download/v${VERSION}/checksums.txt
sha256sum -c checksums.txt --ignore-missing

# Install
sudo dpkg -i openvpn-mng_${VERSION}_linux_amd64.deb

# If dependencies are missing
sudo apt-get install -f
```

**ARM64 (Raspberry Pi, etc.):**

```bash
wget https://github.com/tldr-it-stepankutaj/openvpn-mng/releases/download/v${VERSION}/openvpn-mng_${VERSION}_linux_arm64.deb
sudo dpkg -i openvpn-mng_${VERSION}_linux_arm64.deb
```

**What gets installed:**

| Path | Description |
|------|-------------|
| `/usr/bin/openvpn-mng` | Main binary |
| `/etc/openvpn-mng/config.yaml` | Configuration file |
| `/usr/lib/systemd/system/openvpn-mng.service` | Systemd service |
| `/usr/share/openvpn-mng/web/` | Web assets (templates, static files) |
| `/var/lib/openvpn-mng/` | Data directory |
| `/var/log/openvpn-mng/` | Log directory |

---

### RPM Package (RHEL, AlmaLinux, Rocky Linux)

Supported: RHEL 8+, AlmaLinux 8+, Rocky Linux 8+, Fedora 38+

**Download and install:**

```bash
# Download latest release
VERSION="1.1.0"
wget https://github.com/tldr-it-stepankutaj/openvpn-mng/releases/download/v${VERSION}/openvpn-mng_${VERSION}_linux_amd64.rpm

# Verify checksum (recommended)
wget https://github.com/tldr-it-stepankutaj/openvpn-mng/releases/download/v${VERSION}/checksums.txt
sha256sum -c checksums.txt --ignore-missing

# Install
sudo rpm -i openvpn-mng_${VERSION}_linux_amd64.rpm

# Or with dnf (handles dependencies)
sudo dnf install ./openvpn-mng_${VERSION}_linux_amd64.rpm
```

**ARM64:**

```bash
wget https://github.com/tldr-it-stepankutaj/openvpn-mng/releases/download/v${VERSION}/openvpn-mng_${VERSION}_linux_arm64.rpm
sudo dnf install ./openvpn-mng_${VERSION}_linux_arm64.rpm
```

---

### Docker

**Pull image:**

```bash
# Latest version
docker pull tldr/openvpn-mng:latest

# Specific version
docker pull tldr/openvpn-mng:1.1.0
```

**Run container:**

```bash
# Generate secrets first
JWT_SECRET=$(openssl rand -hex 32)
VPN_TOKEN=$(openssl rand -hex 32)

docker run -d \
  --name openvpn-mng \
  -p 8080:8080 \
  -e DB_TYPE=postgres \
  -e DB_HOST=your-db-host \
  -e DB_PORT=5432 \
  -e DB_USERNAME=openvpn \
  -e DB_PASSWORD=your-db-password \
  -e DB_DATABASE=openvpn_mng \
  -e AUTH_JWT_SECRET=$JWT_SECRET \
  -e API_VPN_TOKEN=$VPN_TOKEN \
  tldr/openvpn-mng:latest
```

**With custom config file:**

```bash
docker run -d \
  --name openvpn-mng \
  -p 8080:8080 \
  -v /path/to/config.yaml:/app/config.yaml:ro \
  tldr/openvpn-mng:latest
```

**Multi-architecture support:**

The Docker image supports both `linux/amd64` and `linux/arm64` architectures. Docker will automatically pull the correct image for your platform.

---

### Docker Compose

Create `docker-compose.yaml`:

```yaml
version: '3.8'

services:
  db:
    image: postgres:16-alpine
    container_name: openvpn-mng-db
    restart: unless-stopped
    environment:
      POSTGRES_USER: openvpn
      POSTGRES_PASSWORD: change-me-in-production
      POSTGRES_DB: openvpn_mng
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U openvpn -d openvpn_mng"]
      interval: 10s
      timeout: 5s
      retries: 5

  app:
    image: tldr/openvpn-mng:latest
    container_name: openvpn-mng
    restart: unless-stopped
    depends_on:
      db:
        condition: service_healthy
    ports:
      - "8080:8080"
    environment:
      SERVER_HOST: "0.0.0.0"
      SERVER_PORT: "8080"
      DB_TYPE: postgres
      DB_HOST: db
      DB_PORT: "5432"
      DB_USERNAME: openvpn
      DB_PASSWORD: change-me-in-production
      DB_DATABASE: openvpn_mng
      DB_SSLMODE: disable
      # Generate with: openssl rand -hex 32
      AUTH_JWT_SECRET: change-me-generate-with-openssl-rand-hex-32
      AUTH_TOKEN_EXPIRY: "24"
      AUTH_SESSION_EXPIRY: "8"
      API_ENABLED: "true"
      API_SWAGGER_ENABLED: "true"
      # Generate with: openssl rand -hex 32 (optional)
      API_VPN_TOKEN: ""
      LOG_OUTPUT: stdout
      LOG_FORMAT: json
      LOG_LEVEL: info

volumes:
  postgres_data:
```

**Start services:**

```bash
# Generate secrets and update docker-compose.yaml
JWT_SECRET=$(openssl rand -hex 32)
echo "Generated JWT_SECRET: $JWT_SECRET"

# Start
docker-compose up -d

# View logs
docker-compose logs -f app

# Stop
docker-compose down
```

---

### From Source

**Requirements:**

- Go 1.22 or higher
- Git
- swag CLI (for Swagger documentation)

**Clone and build:**

```bash
# Clone repository
git clone https://github.com/tldr-it-stepankutaj/openvpn-mng.git
cd openvpn-mng

# Install swag CLI
go install github.com/swaggo/swag/cmd/swag@latest

# Make sure ~/go/bin is in PATH
export PATH=$PATH:$(go env GOPATH)/bin

# Generate Swagger documentation
swag init -g cmd/server/main.go -o docs

# Build
go build -ldflags="-s -w" -o openvpn-mng ./cmd/server
```

**Install manually:**

```bash
# Create user and group
sudo groupadd --system openvpn-mng
sudo useradd --system --gid openvpn-mng --home-dir /var/lib/openvpn-mng \
    --shell /usr/sbin/nologin --comment "OpenVPN Manager" openvpn-mng

# Create directories
sudo mkdir -p /etc/openvpn-mng
sudo mkdir -p /var/lib/openvpn-mng
sudo mkdir -p /var/log/openvpn-mng
sudo mkdir -p /usr/share/openvpn-mng

# Copy files
sudo cp openvpn-mng /usr/bin/
sudo chmod 755 /usr/bin/openvpn-mng

sudo cp config.yaml.example /etc/openvpn-mng/config.yaml
sudo chmod 640 /etc/openvpn-mng/config.yaml
sudo chown root:openvpn-mng /etc/openvpn-mng/config.yaml

sudo cp -r web /usr/share/openvpn-mng/

# Set ownership
sudo chown -R openvpn-mng:openvpn-mng /var/lib/openvpn-mng
sudo chown -R openvpn-mng:openvpn-mng /var/log/openvpn-mng
sudo chown -R root:openvpn-mng /etc/openvpn-mng
```

**Create systemd service:**

Create `/etc/systemd/system/openvpn-mng.service`:

```ini
[Unit]
Description=OpenVPN Manager - Web-based management system for OpenVPN
Documentation=https://github.com/tldr-it-stepankutaj/openvpn-mng
After=network.target postgresql.service mysql.service
Wants=network-online.target

[Service]
Type=simple
User=openvpn-mng
Group=openvpn-mng
ExecStart=/usr/bin/openvpn-mng -config /etc/openvpn-mng/config.yaml
WorkingDirectory=/var/lib/openvpn-mng
Restart=on-failure
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=openvpn-mng

# Security hardening
NoNewPrivileges=yes
PrivateTmp=yes
ProtectSystem=strict
ProtectHome=yes
ReadWritePaths=/var/lib/openvpn-mng /var/log/openvpn-mng
ProtectKernelTunables=yes
ProtectKernelModules=yes
ProtectControlGroups=yes

[Install]
WantedBy=multi-user.target
```

**Reload systemd:**

```bash
sudo systemctl daemon-reload
```

---

## Post-Installation Configuration

After installation, edit the configuration file:

```bash
sudo nano /etc/openvpn-mng/config.yaml
```

**Minimal required changes:**

1. **Database connection:**
   ```yaml
   database:
     type: "postgres"
     host: "localhost"
     port: 5432
     username: "openvpn"
     password: "your-database-password"
     database: "openvpn_mng"
   ```

2. **JWT secret (required):**
   ```yaml
   auth:
     jwt_secret: "paste-your-generated-jwt-secret-here"
   ```

3. **VPN token (optional, for OpenVPN integration):**
   ```yaml
   api:
     vpn_token: "paste-your-generated-vpn-token-here"
   ```

4. **Server binding (for production):**
   ```yaml
   server:
     host: "0.0.0.0"  # Listen on all interfaces
     port: 8080
   ```

**Full configuration example:**

```yaml
server:
  host: "0.0.0.0"
  port: 8080

database:
  type: "postgres"
  host: "localhost"
  port: 5432
  username: "openvpn"
  password: "your-secure-db-password"
  database: "openvpn_mng"
  sslmode: "disable"

api:
  enabled: true
  swagger_enabled: true
  swagger_allowed_ips:
    - "127.0.0.1/32"
    - "192.168.1.0/24"
  vpn_token: "your-generated-vpn-token"

auth:
  jwt_secret: "your-generated-jwt-secret"
  token_expiry: 24
  session_expiry: 8

logging:
  output: "both"
  path: "/var/log/openvpn-mng"
  format: "json"
  level: "info"

security:
  rate_limit_enabled: true
  rate_limit_requests: 5
  rate_limit_window: 60
  rate_limit_burst: 10
  lockout_max_attempts: 5
  lockout_duration: 15
```

---

## Starting the Service

**Important:** The service does NOT start automatically after installation. You must explicitly enable and start it.

```bash
# Enable service to start on boot
sudo systemctl enable openvpn-mng.service

# Start service
sudo systemctl start openvpn-mng.service

# Or combine both commands
sudo systemctl enable --now openvpn-mng.service

# Check status
sudo systemctl status openvpn-mng.service

# View logs
sudo journalctl -u openvpn-mng.service -f
```

**Verify installation:**

```bash
# Check if service is running
curl http://localhost:8080/

# Check API (if enabled)
curl http://localhost:8080/api/v1/auth/me

# Access web interface
# Open in browser: http://localhost:8080/login
```

**Default credentials:**

- Username: `admin`
- Password: `admin123`

**Important:** Change the default password immediately after first login!

After login, you'll see the admin dashboard:

![Admin Dashboard](images/admin_dashboard.png)

---

## Upgrading

### DEB Package

```bash
# Download new version
wget https://github.com/tldr-it-stepankutaj/openvpn-mng/releases/download/v${NEW_VERSION}/openvpn-mng_${NEW_VERSION}_linux_amd64.deb

# Stop service
sudo systemctl stop openvpn-mng.service

# Upgrade (preserves configuration)
sudo dpkg -i openvpn-mng_${NEW_VERSION}_linux_amd64.deb

# Start service
sudo systemctl start openvpn-mng.service
```

### RPM Package

```bash
# Download new version
wget https://github.com/tldr-it-stepankutaj/openvpn-mng/releases/download/v${NEW_VERSION}/openvpn-mng_${NEW_VERSION}_linux_amd64.rpm

# Stop service
sudo systemctl stop openvpn-mng.service

# Upgrade
sudo rpm -U openvpn-mng_${NEW_VERSION}_linux_amd64.rpm

# Start service
sudo systemctl start openvpn-mng.service
```

### Docker

```bash
# Pull new image
docker pull tldr/openvpn-mng:${NEW_VERSION}

# Stop and remove old container
docker stop openvpn-mng
docker rm openvpn-mng

# Start new container with same configuration
docker run -d ... tldr/openvpn-mng:${NEW_VERSION}
```

### Docker Compose

```bash
# Update image tag in docker-compose.yaml
# Or use :latest tag

# Pull and restart
docker-compose pull
docker-compose up -d
```

---

## Uninstallation

### DEB Package

```bash
# Stop service
sudo systemctl stop openvpn-mng.service
sudo systemctl disable openvpn-mng.service

# Remove package
sudo dpkg -r openvpn-mng

# Remove package and configuration
sudo dpkg -P openvpn-mng

# Remove data (optional - destructive!)
sudo rm -rf /etc/openvpn-mng
sudo rm -rf /var/lib/openvpn-mng
sudo rm -rf /var/log/openvpn-mng
sudo userdel openvpn-mng
sudo groupdel openvpn-mng
```

### RPM Package

```bash
# Stop service
sudo systemctl stop openvpn-mng.service
sudo systemctl disable openvpn-mng.service

# Remove package
sudo rpm -e openvpn-mng

# Remove data (optional - destructive!)
sudo rm -rf /etc/openvpn-mng
sudo rm -rf /var/lib/openvpn-mng
sudo rm -rf /var/log/openvpn-mng
sudo userdel openvpn-mng
sudo groupdel openvpn-mng
```

### Docker

```bash
# Stop and remove container
docker stop openvpn-mng
docker rm openvpn-mng

# Remove image
docker rmi tldr/openvpn-mng:latest

# Remove volumes (optional - destructive!)
docker volume rm openvpn-mng_data
```

---

## Troubleshooting

### Service won't start

```bash
# Check logs
sudo journalctl -u openvpn-mng.service -n 50

# Common issues:
# - Database connection failed: Check database credentials
# - Port already in use: Change port or stop conflicting service
# - Permission denied: Check file ownership and permissions
```

### Database connection issues

```bash
# Test PostgreSQL connection
psql -h localhost -U openvpn -d openvpn_mng

# Test MySQL connection
mysql -h localhost -u openvpn -p openvpn_mng
```

### Permission issues

```bash
# Fix ownership
sudo chown -R openvpn-mng:openvpn-mng /var/lib/openvpn-mng
sudo chown -R openvpn-mng:openvpn-mng /var/log/openvpn-mng
sudo chown root:openvpn-mng /etc/openvpn-mng/config.yaml
sudo chmod 640 /etc/openvpn-mng/config.yaml
```

### Check configuration

```bash
# Validate config file syntax
cat /etc/openvpn-mng/config.yaml | python3 -c "import yaml,sys; yaml.safe_load(sys.stdin)"
```

---

## Next Steps

After installation:

1. [Configure users and groups](api.md)
2. [Set up OpenVPN integration](client.md)
3. [Configure networks and access control](api.md#networks)

### User Management

Create and manage users from the Users page:

![User Management](images/admin_users.png)

Edit user details including VPN IP assignment:

![User Edit](images/admin_user_edit.png)

### Group Management

Organize users into groups:

![Group Management](images/admin_groups.png)

### Network Management

Define network segments and assign them to groups:

![Network Management](images/admin_networks.png)

For more information, see the [API Documentation](api.md) and [Client Integration Guide](client.md).

---

## OpenVPN Client Integration

After installing OpenVPN Manager, you need to install **[openvpn-client](https://github.com/tldr-it-stepankutaj/openvpn-client)** on the OpenVPN server to connect the two systems. The client provides hook scripts that OpenVPN calls during user authentication, connection, and disconnection events.

1. Install openvpn-client on the OpenVPN server — see the [openvpn-client installation guide](https://github.com/tldr-it-stepankutaj/openvpn-client#installation)
2. Configure it to point to this OpenVPN Manager instance using the VPN Auth API token
3. Set up the OpenVPN server hooks (`auth-user-pass-verify`, `client-connect`, `client-disconnect`) to call the openvpn-client binaries
