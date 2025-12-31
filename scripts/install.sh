#!/bin/bash

# TG-Wrapped Installation Script
# This script installs all dependencies and sets up the application

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_error "This script must be run as root (use sudo)"
        exit 1
    fi
}

# Install system dependencies
install_dependencies() {
    log_info "Updating package lists..."
    apt-get update

    log_info "Installing system dependencies..."
    apt-get install -y \
        curl \
        wget \
        git \
        nginx \
        certbot \
        python3-certbot-nginx \
        redis-server \
        supervisor
}

# Install Go
install_go() {
    GO_VERSION="1.21.5"
    
    if command -v go &> /dev/null; then
        log_info "Go is already installed: $(go version)"
        return
    fi

    log_info "Installing Go ${GO_VERSION}..."
    wget -q "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -O /tmp/go.tar.gz
    rm -rf /usr/local/go
    tar -C /usr/local -xzf /tmp/go.tar.gz
    rm /tmp/go.tar.gz

    # Add Go to PATH
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile.d/go.sh
    source /etc/profile.d/go.sh

    log_info "Go installed successfully: $(go version)"
}

# Install MinIO
install_minio() {
    if command -v minio &> /dev/null; then
        log_info "MinIO is already installed"
        return
    fi

    log_info "Installing MinIO..."
    wget -q https://dl.min.io/server/minio/release/linux-amd64/minio -O /usr/local/bin/minio
    chmod +x /usr/local/bin/minio

    # Create MinIO user and directories
    useradd -r minio-user -s /sbin/nologin || true
    mkdir -p /data/minio
    chown -R minio-user:minio-user /data/minio

    log_info "MinIO installed successfully"
}

# Create application user and directories
setup_app_user() {
    APP_USER="tgwrapped"
    APP_DIR="/opt/tg-wrapped"

    log_info "Setting up application user and directories..."

    # Create user if doesn't exist
    id -u $APP_USER &>/dev/null || useradd -r -s /sbin/nologin $APP_USER

    # Create directories
    mkdir -p $APP_DIR
    mkdir -p $APP_DIR/logs
    mkdir -p $APP_DIR/data
    mkdir -p /etc/tg-wrapped

    chown -R $APP_USER:$APP_USER $APP_DIR
    chown -R $APP_USER:$APP_USER /etc/tg-wrapped

    log_info "Application directories created at $APP_DIR"
}

# Create environment file template
create_env_template() {
    log_info "Creating environment file template..."

    cat > /etc/tg-wrapped/.env.example << 'EOF'
# Telegram API Credentials
APP_ID=your_app_id
APP_HASH=your_app_hash
APP_SESSION_STORAGE=/opt/tg-wrapped/data/session.json

# Server Configuration
SERVER_PORT=7000
ENV=production
LOG_DIR=/opt/tg-wrapped/logs

# Redis Configuration
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=

# MinIO Configuration
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_ID=minioadmin
MINIO_SECRET_ID=minioadmin
MINIO_TOKEN=
MINIO_BUCKET=channel-profiles
EOF

    log_info "Environment template created at /etc/tg-wrapped/.env.example"
    log_warn "Copy to /etc/tg-wrapped/.env and fill in your values"
}

# Create systemd service files
create_systemd_services() {
    log_info "Creating systemd service files..."

    # TG-Wrapped service
    cat > /etc/systemd/system/tg-wrapped.service << 'EOF'
[Unit]
Description=TG-Wrapped Analytics Service
After=network.target redis.service minio.service
Wants=redis.service minio.service

[Service]
Type=simple
User=tgwrapped
Group=tgwrapped
WorkingDirectory=/opt/tg-wrapped
EnvironmentFile=/etc/tg-wrapped/.env
ExecStart=/opt/tg-wrapped/tg-wrapped
Restart=always
RestartSec=5
StandardOutput=append:/opt/tg-wrapped/logs/stdout.log
StandardError=append:/opt/tg-wrapped/logs/stderr.log

# Security hardening
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/tg-wrapped

[Install]
WantedBy=multi-user.target
EOF

    # MinIO service
    cat > /etc/systemd/system/minio.service << 'EOF'
[Unit]
Description=MinIO Object Storage
After=network.target
Documentation=https://docs.min.io

[Service]
Type=simple
User=minio-user
Group=minio-user
Environment="MINIO_ROOT_USER=minioadmin"
Environment="MINIO_ROOT_PASSWORD=minioadmin"
ExecStart=/usr/local/bin/minio server /data/minio --console-address ":9001"
Restart=always
RestartSec=5
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    log_info "Systemd services created"
}

# Enable and start services
enable_services() {
    log_info "Enabling and starting services..."

    systemctl enable redis-server
    systemctl start redis-server

    systemctl enable minio
    systemctl start minio

    systemctl enable nginx
    systemctl start nginx

    log_info "Services enabled and started"
}

# Build application
build_app() {
    APP_DIR="/opt/tg-wrapped"
    SOURCE_DIR=$(dirname "$(dirname "$(readlink -f "$0")")")

    log_info "Building application from $SOURCE_DIR..."

    cd "$SOURCE_DIR"
    
    export PATH=$PATH:/usr/local/go/bin
    go build -o "$APP_DIR/tg-wrapped" .

    chown tgwrapped:tgwrapped "$APP_DIR/tg-wrapped"
    chmod +x "$APP_DIR/tg-wrapped"

    log_info "Application built successfully"
}

# Main installation
main() {
    log_info "Starting TG-Wrapped installation..."

    check_root
    install_dependencies
    install_go
    install_minio
    setup_app_user
    create_env_template
    create_systemd_services
    enable_services
    build_app

    log_info "============================================"
    log_info "Installation completed successfully!"
    log_info "============================================"
    log_info ""
    log_info "Next steps:"
    log_info "1. Copy and edit the environment file:"
    log_info "   cp /etc/tg-wrapped/.env.example /etc/tg-wrapped/.env"
    log_info "   nano /etc/tg-wrapped/.env"
    log_info ""
    log_info "2. Configure Nginx (run scripts/configure-nginx.sh)"
    log_info ""
    log_info "3. Start the application:"
    log_info "   systemctl start tg-wrapped"
    log_info ""
    log_info "4. Check status:"
    log_info "   systemctl status tg-wrapped"
}

main "$@"
