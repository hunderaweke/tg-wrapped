# TG-Wrapped Deployment Guide

This guide covers deploying TG-Wrapped on a Linux server (Ubuntu/Debian).

## Prerequisites

- Ubuntu 20.04+ or Debian 11+
- Root access (sudo)
- Domain name pointed to your server
- Telegram API credentials (APP_ID and APP_HASH from https://my.telegram.org)

## Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/hunderaweke/tg-wrapped.git
cd tg-wrapped
```

### 2. Run Installation Script

```bash
sudo chmod +x scripts/*.sh
sudo ./scripts/install.sh
```

This will:

- Install Go, Redis, MinIO, and Nginx
- Create systemd services
- Build the application
- Set up directory structure

### 3. Configure Environment

```bash
sudo cp /etc/tg-wrapped/.env.example /etc/tg-wrapped/.env
sudo nano /etc/tg-wrapped/.env
```

Fill in your Telegram API credentials:

```env
APP_ID=your_app_id
APP_HASH=your_app_hash
```

### 4. Configure Nginx

```bash
sudo ./scripts/configure-nginx.sh
```

Follow the prompts to:

- Enter your domain name
- Optionally set up SSL with Let's Encrypt

### 5. Start the Application

```bash
sudo systemctl start tg-wrapped
sudo systemctl status tg-wrapped
```

## Scripts Overview

| Script                       | Description                            |
| ---------------------------- | -------------------------------------- |
| `scripts/install.sh`         | Full installation of all dependencies  |
| `scripts/configure-nginx.sh` | Configure Nginx reverse proxy with SSL |
| `scripts/start.sh`           | Manual start script for debugging      |
| `scripts/deploy.sh`          | Deploy updates with automatic rollback |
| `scripts/status.sh`          | Check status of all services           |
| `scripts/uninstall.sh`       | Remove the application                 |

## Service Management

```bash
# Start/Stop/Restart
sudo systemctl start tg-wrapped
sudo systemctl stop tg-wrapped
sudo systemctl restart tg-wrapped

# View logs
sudo journalctl -u tg-wrapped -f

# Check status
./scripts/status.sh
```

## Updating

To deploy a new version:

```bash
git pull origin master
sudo ./scripts/deploy.sh
```

The deploy script will:

1. Backup the current binary
2. Build the new version
3. Restart the service
4. Automatically rollback if startup fails

## Directory Structure

```
/opt/tg-wrapped/
├── tg-wrapped          # Application binary
├── logs/               # Application logs
│   ├── tg-wrapped-YYYY-MM-DD.log
│   ├── stdout.log
│   └── stderr.log
├── data/               # Application data
│   └── session.json    # Telegram session
└── backups/            # Binary backups

/etc/tg-wrapped/
├── .env                # Environment configuration
└── .env.example        # Configuration template

/data/minio/            # MinIO object storage
```

## Environment Variables

| Variable              | Description                          | Default                             |
| --------------------- | ------------------------------------ | ----------------------------------- |
| `APP_ID`              | Telegram API ID                      | Required                            |
| `APP_HASH`            | Telegram API Hash                    | Required                            |
| `APP_SESSION_STORAGE` | Session file path                    | `/opt/tg-wrapped/data/session.json` |
| `SERVER_PORT`         | HTTP server port                     | `7000`                              |
| `ENV`                 | Environment (development/production) | `production`                        |
| `LOG_DIR`             | Log files directory                  | `/opt/tg-wrapped/logs`              |
| `REDIS_ADDR`          | Redis server address                 | `localhost:6379`                    |
| `REDIS_PASSWORD`      | Redis password                       | Empty                               |
| `MINIO_ENDPOINT`      | MinIO server address                 | `localhost:9000`                    |
| `MINIO_ACCESS_ID`     | MinIO access key                     | `minioadmin`                        |
| `MINIO_SECRET_ID`     | MinIO secret key                     | `minioadmin`                        |
| `MINIO_BUCKET`        | MinIO bucket name                    | `channel-profiles`                  |

## Nginx Configuration

The Nginx configuration includes:

- Rate limiting (10 requests/second with burst of 20)
- Security headers
- Longer timeouts for `/analytics` endpoint (5 minutes)
- SSL with Let's Encrypt (optional)

### Custom Nginx Config

Edit `/etc/nginx/sites-available/tg-wrapped` for custom settings.

## Troubleshooting

### Service won't start

```bash
# Check logs
sudo journalctl -u tg-wrapped -n 100

# Check configuration
cat /etc/tg-wrapped/.env

# Test manually
sudo -u tgwrapped /opt/tg-wrapped/tg-wrapped
```

### Connection errors

```bash
# Check if Redis is running
redis-cli ping

# Check if MinIO is running
curl http://localhost:9000/minio/health/live

# Check if ports are open
ss -tlnp | grep -E '7000|6379|9000'
```

### SSL Issues

```bash
# Renew certificate manually
sudo certbot renew

# Test certificate
sudo certbot certificates
```

## Security Recommendations

1. **Firewall**: Only expose ports 80 and 443

   ```bash
   sudo ufw allow 80/tcp
   sudo ufw allow 443/tcp
   sudo ufw enable
   ```

2. **Change MinIO credentials** in `/etc/systemd/system/minio.service`

3. **Set Redis password** if exposed externally

4. **Regular updates**:
   ```bash
   sudo apt update && sudo apt upgrade
   ```

## Monitoring

### Health Check

```bash
curl http://localhost:7000/health
```

### Logs

```bash
# Application logs
tail -f /opt/tg-wrapped/logs/tg-wrapped-*.log

# Nginx access logs
tail -f /var/log/nginx/tg-wrapped-access.log

# System logs
sudo journalctl -u tg-wrapped -f
```

## Backup

### Database (Redis)

```bash
redis-cli BGSAVE
cp /var/lib/redis/dump.rdb /backup/redis-$(date +%Y%m%d).rdb
```

### MinIO Data

```bash
tar -czf /backup/minio-$(date +%Y%m%d).tar.gz /data/minio
```

### Session File

```bash
cp /opt/tg-wrapped/data/session.json /backup/
```
