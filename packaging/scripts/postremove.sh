#!/bin/bash
# Post-removal script for openvpn-mng

set -e

# Reload systemd
systemctl daemon-reload

# Note: We don't remove user, group, config files, or data directories
# to preserve data in case of reinstallation

echo ""
echo "OpenVPN Manager has been removed."
echo ""
echo "Note: Configuration, data, and log files have been preserved:"
echo "  - /etc/openvpn-mng/"
echo "  - /var/lib/openvpn-mng/"
echo "  - /var/log/openvpn-mng/"
echo ""
echo "To completely remove all data, run:"
echo "  sudo rm -rf /etc/openvpn-mng /var/lib/openvpn-mng /var/log/openvpn-mng"
echo "  sudo userdel openvpn-mng"
echo "  sudo groupdel openvpn-mng"
echo ""

exit 0
