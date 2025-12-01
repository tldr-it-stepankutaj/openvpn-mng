#!/bin/bash
# Pre-removal script for openvpn-mng

set -e

# Stop and disable service if running
if systemctl is-active --quiet openvpn-mng.service 2>/dev/null; then
    systemctl stop openvpn-mng.service
fi

if systemctl is-enabled --quiet openvpn-mng.service 2>/dev/null; then
    systemctl disable openvpn-mng.service
fi

exit 0
