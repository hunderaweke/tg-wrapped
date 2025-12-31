#!/bin/bash

# TG-Wrapped Nginx Configuration Script
# This script configures Nginx as a reverse proxy

set -e

# Colors for output
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

# Check if running as root
if [[ $EUID -ne 0 ]]; then
    log_error "This script must be run as root (use sudo)"
    exit 1
fi

# Get domain name
read -p "Enter your domain name (e.g., api.example.com): " DOMAIN_NAME

if [[ -z "$DOMAIN_NAME" ]]; then
    log_error "Domain name is required"
    exit 1
fi

# Get backend port
read -p "Enter backend port [7000]: " BACKEND_PORT
BACKEND_PORT=${BACKEND_PORT:-7000}

# Ask about SSL
read -p "Do you want to configure SSL with Let's Encrypt? (y/n) [y]: " SETUP_SSL
SETUP_SSL=${SETUP_SSL:-y}

log_info "Configuring Nginx for domain: $DOMAIN_NAME"

# Create Nginx configuration
cat > /etc/nginx/sites-available/tg-wrapped << EOF
# TG-Wrapped API Configuration
# Generated on $(date)

# Rate limiting zone
limit_req_zone \$binary_remote_addr zone=tgwrapped_limit:10m rate=10r/s;

# Upstream backend
upstream tg_wrapped_backend {
    server 127.0.0.1:${BACKEND_PORT};
    keepalive 32;
}

server {
    listen 80;
    listen [::]:80;
    server_name ${DOMAIN_NAME};

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    # Logging
    access_log /var/log/nginx/tg-wrapped-access.log;
    error_log /var/log/nginx/tg-wrapped-error.log;

    # Max body size for requests
    client_max_body_size 10M;

    # Health check endpoint (no rate limiting)
    location /health {
        proxy_pass http://tg_wrapped_backend;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }

    # API endpoints
    location / {
        # Rate limiting
        limit_req zone=tgwrapped_limit burst=20 nodelay;
        limit_req_status 429;

        proxy_pass http://tg_wrapped_backend;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_set_header Connection "";

        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 120s;
        proxy_read_timeout 120s;

        # Buffering
        proxy_buffering on;
        proxy_buffer_size 4k;
        proxy_buffers 8 4k;
    }

    # Analytics endpoint (longer timeout)
    location /analytics {
        # Rate limiting (stricter for analytics)
        limit_req zone=tgwrapped_limit burst=5 nodelay;
        limit_req_status 429;

        proxy_pass http://tg_wrapped_backend;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_set_header Connection "";

        # Longer timeouts for analytics processing
        proxy_connect_timeout 60s;
        proxy_send_timeout 300s;
        proxy_read_timeout 300s;
    }

    # Deny access to hidden files
    location ~ /\. {
        deny all;
    }
}
EOF

# Enable the site
ln -sf /etc/nginx/sites-available/tg-wrapped /etc/nginx/sites-enabled/

# Remove default site if exists
rm -f /etc/nginx/sites-enabled/default

# Test Nginx configuration
log_info "Testing Nginx configuration..."
nginx -t

if [[ $? -ne 0 ]]; then
    log_error "Nginx configuration test failed"
    exit 1
fi

# Reload Nginx
log_info "Reloading Nginx..."
systemctl reload nginx

# Setup SSL with Let's Encrypt
if [[ "$SETUP_SSL" == "y" || "$SETUP_SSL" == "Y" ]]; then
    log_info "Setting up SSL with Let's Encrypt..."
    
    read -p "Enter your email for Let's Encrypt notifications: " SSL_EMAIL
    
    if [[ -z "$SSL_EMAIL" ]]; then
        log_warn "No email provided, skipping SSL setup"
    else
        certbot --nginx -d "$DOMAIN_NAME" --non-interactive --agree-tos -m "$SSL_EMAIL" --redirect
        
        if [[ $? -eq 0 ]]; then
            log_info "SSL certificate installed successfully"
            
            # Setup auto-renewal
            systemctl enable certbot.timer
            systemctl start certbot.timer
            log_info "SSL auto-renewal enabled"
        else
            log_error "SSL setup failed. You can retry manually with:"
            log_error "certbot --nginx -d $DOMAIN_NAME"
        fi
    fi
fi

log_info "============================================"
log_info "Nginx configuration completed!"
log_info "============================================"
log_info ""
log_info "Your API is available at:"
if [[ "$SETUP_SSL" == "y" || "$SETUP_SSL" == "Y" ]]; then
    log_info "  https://${DOMAIN_NAME}"
else
    log_info "  http://${DOMAIN_NAME}"
fi
log_info ""
log_info "Test with:"
log_info "  curl http://${DOMAIN_NAME}/health"
