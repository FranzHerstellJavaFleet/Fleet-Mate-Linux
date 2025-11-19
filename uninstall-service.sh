#!/bin/bash

# Fleet Mate Linux - Service Uninstallation Script
# This script removes the fleet-mate-linux systemd service

set -e

echo "================================================"
echo "Fleet Mate Linux - Service Uninstallation"
echo "================================================"
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "ERROR: This script must be run as root (use sudo)"
    exit 1
fi

SERVICE_NAME="fleet-mate-linux"

# Stop service if running
if systemctl is-active --quiet ${SERVICE_NAME}; then
    echo "Stopping ${SERVICE_NAME} service..."
    systemctl stop ${SERVICE_NAME}
fi

# Disable service
if systemctl is-enabled --quiet ${SERVICE_NAME} 2>/dev/null; then
    echo "Disabling ${SERVICE_NAME} service..."
    systemctl disable ${SERVICE_NAME}
fi

# Remove service file
if [ -f "/etc/systemd/system/${SERVICE_NAME}.service" ]; then
    echo "Removing service file..."
    rm /etc/systemd/system/${SERVICE_NAME}.service
fi

# Reload systemd daemon
echo "Reloading systemd daemon..."
systemctl daemon-reload

echo ""
echo "================================================"
echo "Uninstallation completed!"
echo "================================================"
echo ""
echo "Note: Log files in /var/log/fleet-mate have been kept."
echo "To remove them manually, run:"
echo "  sudo rm -rf /var/log/fleet-mate"
echo ""
