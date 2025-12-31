#!/bin/bash

# TG-Wrapped Deployment Script
# Use this to deploy updates

set -e

# Colors
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

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

APP_DIR="/opt/tg-wrapped"
APP_USER="tgwrapped"
BACKUP_DIR="/opt/tg-wrapped/backups"
SOURCE_DIR=$(dirname "$(dirname "$(readlink -f "$0")")")

# Check if running as root
if [[ $EUID -ne 0 ]]; then
    log_error "This script must be run as root (use sudo)"
    exit 1
fi

log_info "Starting deployment..."
log_info "Source: $SOURCE_DIR"
log_info "Target: $APP_DIR"

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Backup current binary
if [[ -f "$APP_DIR/tg-wrapped" ]]; then
    BACKUP_NAME="tg-wrapped-$(date +%Y%m%d-%H%M%S)"
    log_info "Backing up current binary to $BACKUP_DIR/$BACKUP_NAME"
    cp "$APP_DIR/tg-wrapped" "$BACKUP_DIR/$BACKUP_NAME"
fi

# Stop service
log_info "Stopping tg-wrapped service..."
systemctl stop tg-wrapped || true

# Build new binary
log_info "Building new binary..."
cd "$SOURCE_DIR"
export PATH=$PATH:/usr/local/go/bin
go build -o "$APP_DIR/tg-wrapped" .

# Set permissions
chown $APP_USER:$APP_USER "$APP_DIR/tg-wrapped"
chmod +x "$APP_DIR/tg-wrapped"

# Start service
log_info "Starting tg-wrapped service..."
systemctl start tg-wrapped

# Wait for startup
sleep 3

# Check if service is running
if systemctl is-active --quiet tg-wrapped; then
    log_info "============================================"
    log_info "Deployment completed successfully!"
    log_info "============================================"
    systemctl status tg-wrapped --no-pager
else
    log_error "Service failed to start!"
    log_error "Check logs with: journalctl -u tg-wrapped -n 50"
    
    # Rollback
    if [[ -f "$BACKUP_DIR/$BACKUP_NAME" ]]; then
        log_warn "Rolling back to previous version..."
        cp "$BACKUP_DIR/$BACKUP_NAME" "$APP_DIR/tg-wrapped"
        chown $APP_USER:$APP_USER "$APP_DIR/tg-wrapped"
        systemctl start tg-wrapped
    fi
    exit 1
fi

# Cleanup old backups (keep last 5)
log_info "Cleaning up old backups..."
ls -t "$BACKUP_DIR"/tg-wrapped-* 2>/dev/null | tail -n +6 | xargs -r rm

log_info ""
log_info "Deployment info:"
log_info "  Binary: $APP_DIR/tg-wrapped"
log_info "  Logs: $APP_DIR/logs/"
log_info "  Config: /etc/tg-wrapped/.env"
