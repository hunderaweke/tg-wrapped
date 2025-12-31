#!/bin/bash

# TG-Wrapped Status Script
# Check status of all services

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

print_status() {
    local service=$1
    if systemctl is-active --quiet "$service"; then
        echo -e "  $service: ${GREEN}running${NC}"
    else
        echo -e "  $service: ${RED}stopped${NC}"
    fi
}

echo "============================================"
echo "TG-Wrapped System Status"
echo "============================================"
echo ""

echo "Services:"
print_status "tg-wrapped"
print_status "nginx"
print_status "redis-server"
print_status "minio"

echo ""
echo "Disk Usage:"
df -h /opt/tg-wrapped 2>/dev/null | tail -1 | awk '{print "  App directory: " $3 " used / " $2 " total (" $5 " used)"}'
df -h /data/minio 2>/dev/null | tail -1 | awk '{print "  MinIO data: " $3 " used / " $2 " total (" $5 " used)"}'

echo ""
echo "Memory Usage:"
free -h | grep Mem | awk '{print "  Total: " $2 ", Used: " $3 ", Free: " $4}'

echo ""
echo "Recent Logs (last 10 lines):"
if [[ -f /opt/tg-wrapped/logs/stdout.log ]]; then
    echo "  --- stdout.log ---"
    tail -5 /opt/tg-wrapped/logs/stdout.log 2>/dev/null | sed 's/^/  /'
fi

echo ""
echo "Health Check:"
HEALTH=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:7000/health 2>/dev/null)
if [[ "$HEALTH" == "200" ]]; then
    echo -e "  API Health: ${GREEN}OK${NC}"
else
    echo -e "  API Health: ${RED}FAILED (HTTP $HEALTH)${NC}"
fi

echo ""
echo "============================================"
echo "Commands:"
echo "  View logs:     journalctl -u tg-wrapped -f"
echo "  Restart:       systemctl restart tg-wrapped"
echo "  Stop:          systemctl stop tg-wrapped"
echo "  Deploy update: sudo ./scripts/deploy.sh"
echo "============================================"
