# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.1] - 2025-12-02

### Added
- Automatic VPN IP allocation for new users
- VPN network configuration in `config.yaml` (`vpn.network`, `vpn.server_ip`)
- New API endpoints for VPN IP management:
  - `GET /api/v1/vpn/next-ip` - Get next available VPN IP
  - `GET /api/v1/vpn/network-info` - Get VPN network information
  - `POST /api/v1/vpn/validate-ip` - Validate VPN IP address
  - `GET /api/v1/vpn/used-ips` - List all used VPN IPs
- Web form auto-fills VPN IP with next available address
- VPN IP validation against configured network range
- `.dockerignore` file for optimized Docker builds

### Changed
- User creation now auto-assigns VPN IP if not provided
- VPN IP field in web forms is read-only by default with edit option

## [1.0.0] - 2025-12-01

### Added
- Initial release
- VPN Auth API with token-based authentication for OpenVPN server integration
- New endpoints: `/api/v1/vpn-auth/*` for VPN server communication
- `vpn_token` configuration option for VPN Auth API
- User management with RBAC (User, Manager, Admin roles)
- Group management
- Network management with CIDR support
- VPN session tracking
- Audit logging
- REST API with Swagger documentation
- Web interface with Bootstrap
- PostgreSQL and MySQL support
- JWT authentication
- Environment variable configuration support
- Docker support
- IP filtering for Swagger access
- Configurable logging (stdout, file, JSON format)
- DEB package support (Debian, Ubuntu)
- RPM package support (RHEL, AlmaLinux, Rocky Linux)
- Docker multi-architecture support (amd64, arm64)
- Systemd service file with security hardening
- Comprehensive installation guide (`help/install.md`)
- GitHub Actions workflows for CI/CD
- GoReleaser configuration for automated releases

### Security
- Systemd service runs with restricted privileges
- Configuration file protected with proper permissions

[1.0.1]: https://github.com/tldr-it-stepankutaj/openvpn-mng/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/tldr-it-stepankutaj/openvpn-mng/releases/tag/v1.0.0
