#!/bin/bash

# TG-Wrapped Uninstall Script
# Removes the application and optionally its data

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# Check if running as root
if [[ $EUID -ne 0 ]]; then
    echo -e "${RED}[ERROR]${NC} This script must be run as root (use sudo)"
    exit 1
fi

echo "============================================"
echo "TG-Wrapped Uninstall"
echo "============================================"
echo ""

read -p "Are you sure you want to uninstall TG-Wrapped? (y/n): " CONFIRM
if [[ "$CONFIRM" != "y" && "$CONFIRM" != "Y" ]]; then
    echo "Uninstall cancelled"
    exit 0
fi

read -p "Do you want to remove all data (logs, sessions, etc.)? (y/n): " REMOVE_DATA

# Stop services
log_info "Stopping services..."
systemctl stop tg-wrapped 2>/dev/null || true
systemctl disable tg-wrapped 2>/dev/null || true

# Remove systemd service
log_info "Removing systemd service..."
rm -f /etc/systemd/system/tg-wrapped.service
systemctl daemon-reload

# Remove Nginx configuration
log_info "Removing Nginx configuration..."
rm -f /etc/nginx/sites-enabled/tg-wrapped
rm -f /etc/nginx/sites-available/tg-wrapped
systemctl reload nginx 2>/dev/null || true

# Remove application files
log_info "Removing application files..."
rm -rf /opt/tg-wrapped

# Remove configuration
if [[ "$REMOVE_DATA" == "y" || "$REMOVE_DATA" == "Y" ]]; then
    log_info "Removing configuration and data..."
    rm -rf /etc/tg-wrapped
    
    # Remove user
    userdel tgwrapped 2>/dev/null || true
fi

log_info "============================================"
log_info "Uninstall completed!"
log_info "============================================"
log_info ""
log_warn "Note: Redis and MinIO were not removed."
log_warn "Remove them manually if no longer needed:"
log_warn "  apt remove redis-server"
log_warn "  systemctl stop minio && rm /usr/local/bin/minio"
