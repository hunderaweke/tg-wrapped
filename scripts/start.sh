#!/bin/bash

# TG-Wrapped Start Script
# Use this for manual startup or debugging

set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

APP_DIR="/opt/tg-wrapped"
ENV_FILE="/etc/tg-wrapped/.env"

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if env file exists
if [[ ! -f "$ENV_FILE" ]]; then
    log_error "Environment file not found at $ENV_FILE"
    log_error "Please copy and configure:"
    log_error "  cp /etc/tg-wrapped/.env.example /etc/tg-wrapped/.env"
    exit 1
fi

# Load environment variables
set -a
source "$ENV_FILE"
set +a

# Check if binary exists
if [[ ! -f "$APP_DIR/tg-wrapped" ]]; then
    log_error "Application binary not found at $APP_DIR/tg-wrapped"
    log_error "Please run the install script first"
    exit 1
fi

# Create log directory if not exists
mkdir -p "$APP_DIR/logs"

log_info "Starting TG-Wrapped..."
log_info "Environment: ${ENV:-development}"
log_info "Port: ${SERVER_PORT:-7000}"
log_info "Logs: $APP_DIR/logs"

cd "$APP_DIR"
exec ./tg-wrapped
