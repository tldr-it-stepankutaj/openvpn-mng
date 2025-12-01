#!/bin/bash
# Post-installation script for openvpn-mng

set -e

# Create system user and group if they don't exist
if ! getent group openvpn-mng > /dev/null 2>&1; then
    groupadd --system openvpn-mng
fi

if ! getent passwd openvpn-mng > /dev/null 2>&1; then
    useradd --system --gid openvpn-mng --home-dir /var/lib/openvpn-mng \
        --shell /usr/sbin/nologin --comment "OpenVPN Manager" openvpn-mng
fi

# Create directories
mkdir -p /var/lib/openvpn-mng
mkdir -p /var/log/openvpn-mng
mkdir -p /etc/openvpn-mng

# Set ownership
chown -R openvpn-mng:openvpn-mng /var/lib/openvpn-mng
chown -R openvpn-mng:openvpn-mng /var/log/openvpn-mng
chown -R root:openvpn-mng /etc/openvpn-mng
chmod 750 /etc/openvpn-mng

# Create default config if it doesn't exist
if [ ! -f /etc/openvpn-mng/config.yaml ]; then
    cp /usr/share/openvpn-mng/config.yaml.example /etc/openvpn-mng/config.yaml
    chmod 640 /etc/openvpn-mng/config.yaml
    chown root:openvpn-mng /etc/openvpn-mng/config.yaml
    echo ""
    echo "=========================================="
    echo "OpenVPN Manager installed successfully!"
    echo "=========================================="
    echo ""
    echo "Configuration file created at: /etc/openvpn-mng/config.yaml"
    echo ""
    echo "IMPORTANT: Before starting the service, you must:"
    echo "  1. Generate secrets:"
    echo "     openssl rand -hex 32  # for jwt_secret"
    echo "     openssl rand -hex 32  # for vpn_token (optional)"
    echo ""
    echo "  2. Edit the configuration file:"
    echo "     sudo nano /etc/openvpn-mng/config.yaml"
    echo ""
    echo "  3. Configure database connection settings"
    echo ""
    echo "  4. Start and enable the service:"
    echo "     sudo systemctl enable --now openvpn-mng.service"
    echo ""
    echo "Documentation: https://github.com/tldr-it-stepankutaj/openvpn-mng"
    echo "=========================================="
fi

# Reload systemd (only if systemd is running)
if command -v systemctl > /dev/null 2>&1 && systemctl --version > /dev/null 2>&1; then
    systemctl daemon-reload || true
fi

exit 0
