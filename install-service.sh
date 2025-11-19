#!/bin/bash

# Fleet Mate Linux - Service Installation Script
# This script installs fleet-mate-linux as a systemd service

set -e

echo "================================================"
echo "Fleet Mate Linux - Service Installation"
echo "================================================"
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "ERROR: This script must be run as root (use sudo)"
    exit 1
fi

# Get the directory where the script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SERVICE_NAME="fleet-mate-linux"
SERVICE_FILE="${SCRIPT_DIR}/${SERVICE_NAME}.service"

echo "Script directory: ${SCRIPT_DIR}"
echo "Service name: ${SERVICE_NAME}"
echo ""

# Check if fleet-mate binary exists
if [ ! -f "${SCRIPT_DIR}/fleet-mate" ]; then
    echo "ERROR: fleet-mate binary not found in ${SCRIPT_DIR}"
    echo "Please build the binary first with: go build -o fleet-mate main.go"
    exit 1
fi

# Check if service file exists
if [ ! -f "${SERVICE_FILE}" ]; then
    echo "ERROR: Service file not found: ${SERVICE_FILE}"
    exit 1
fi

# Check if config.yml exists
if [ ! -f "${SCRIPT_DIR}/config.yml" ]; then
    echo "WARNING: config.yml not found in ${SCRIPT_DIR}"
    echo "Please create config.yml before starting the service"
fi

# Create log directory
LOG_DIR="/var/log/fleet-mate"
echo "Creating log directory: ${LOG_DIR}"
mkdir -p "${LOG_DIR}"
chown trainer:trainer "${LOG_DIR}"
chmod 755 "${LOG_DIR}"

# Stop service if running
if systemctl is-active --quiet ${SERVICE_NAME}; then
    echo "Stopping existing ${SERVICE_NAME} service..."
    systemctl stop ${SERVICE_NAME}
fi

# Copy service file to systemd directory
echo "Installing service file to /etc/systemd/system/${SERVICE_NAME}.service"
cp "${SERVICE_FILE}" /etc/systemd/system/${SERVICE_NAME}.service

# Reload systemd daemon
echo "Reloading systemd daemon..."
systemctl daemon-reload

# Enable service
echo "Enabling ${SERVICE_NAME} service..."
systemctl enable ${SERVICE_NAME}

echo ""
echo "================================================"
echo "Installation completed successfully!"
echo "================================================"
echo ""
echo "Service commands:"
echo "  Start service:   sudo systemctl start ${SERVICE_NAME}"
echo "  Stop service:    sudo systemctl stop ${SERVICE_NAME}"
echo "  Restart service: sudo systemctl restart ${SERVICE_NAME}"
echo "  Service status:  sudo systemctl status ${SERVICE_NAME}"
echo "  View logs:       sudo journalctl -u ${SERVICE_NAME} -f"
echo ""
echo "To start the service now, run:"
echo "  sudo systemctl start ${SERVICE_NAME}"
echo ""
