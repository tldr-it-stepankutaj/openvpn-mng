# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- VPN Auth API with token-based authentication for OpenVPN server integration
- New endpoints: `/api/v1/vpn-auth/*` for VPN server communication
- `vpn_token` configuration option for VPN Auth API
- DEB package support (Debian, Ubuntu)
- RPM package support (RHEL, AlmaLinux, Rocky Linux)
- Docker multi-architecture support (amd64, arm64)
- Systemd service file with security hardening
- Comprehensive installation guide (`help/install.md`)
- GitHub Actions workflows for CI/CD
- GoReleaser configuration for automated releases

### Changed
- Improved configuration file with better documentation
- Updated README with VPN Auth API documentation

### Security
- Systemd service runs with restricted privileges
- Configuration file protected with proper permissions

## [1.0.0] - 2024-XX-XX

### Added
- Initial release
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

[Unreleased]: https://github.com/tldr-it-stepankutaj/openvpn-mng/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/tldr-it-stepankutaj/openvpn-mng/releases/tag/v1.0.0
