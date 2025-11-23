#!/bin/bash
BACKUP_DIR="/root/backups"
DATE=$(date +%Y%m%d-%H%M%S)

mkdir -p $BACKUP_DIR

# Backup database
docker compose -f /opt/avt-service/docker-compose.prod.yml exec -T timescaledb \
  pg_dump -U telemetry_user telemetry_prod | gzip > $BACKUP_DIR/backup-$DATE.sql.gz

# Keep only last 7 days
find $BACKUP_DIR -name "backup-*.sql.gz" -mtime +7 -delete

echo "Backup completed: backup-$DATE.sql.gz"