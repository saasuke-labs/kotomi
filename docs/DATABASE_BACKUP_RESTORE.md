# Database Backup and Restore Guide

This guide covers backup strategies, procedures, and disaster recovery for Kotomi's SQLite database.

## Table of Contents

1. [Overview](#overview)
2. [Backup Strategies](#backup-strategies)
3. [Manual Backup](#manual-backup)
4. [Automated Backup](#automated-backup)
5. [Restore Procedures](#restore-procedures)
6. [Disaster Recovery](#disaster-recovery)
7. [Best Practices](#best-practices)

## Overview

Kotomi uses SQLite as its database, storing all data in a single file. This makes backups straightforward but requires careful planning for production use.

### What Data is Stored

The SQLite database (`kotomi.db`) contains:
- Sites configuration
- Pages information
- Comments and their metadata
- Reactions (likes/dislikes)
- User information
- Moderation status
- Analytics data

### Backup Requirements

For beta testing:
- **Frequency**: Daily backups recommended
- **Retention**: Keep last 7 days minimum
- **Location**: Store backups off-server
- **Testing**: Verify restore procedures monthly

For production (future):
- Frequency: Continuous/hourly backups
- Retention: 30 days + monthly archives
- Location: Multi-region storage
- Testing: Weekly restore drills

## Backup Strategies

### 1. File-Based Backup (Simplest)

Copy the database file while the application is running.

**Pros**:
- Simple to implement
- Works with any storage solution
- Easy to restore

**Cons**:
- May capture inconsistent state during writes
- Requires SQLite WAL mode (which Kotomi uses)

**When to Use**: Beta testing, development, low-traffic sites

### 2. SQLite Backup API (Recommended)

Use SQLite's online backup API for consistent backups.

**Pros**:
- Consistent snapshot
- Safe during writes
- Built-in SQLite feature

**Cons**:
- Requires script/tool
- Application must support it

**When to Use**: Production, high-traffic sites

### 3. Volume Snapshots (Cloud)

Use cloud provider's volume snapshot features.

**Pros**:
- Automated by platform
- Point-in-time recovery
- No application downtime

**Cons**:
- Cloud-specific
- May cost extra
- Slower to restore

**When to Use**: Cloud Run, GKE, managed hosting

## Manual Backup

### Docker Deployment

#### Using Docker Volume

```bash
# Create a backup directory
mkdir -p ~/kotomi-backups

# Copy database from Docker volume
docker run --rm \
  -v kotomi-data:/data \
  -v ~/kotomi-backups:/backup \
  alpine \
  cp /data/kotomi.db /backup/kotomi-$(date +%Y%m%d-%H%M%S).db

# Verify backup
ls -lh ~/kotomi-backups/
```

#### Using Docker Exec

```bash
# Copy from running container
docker cp kotomi:/app/data/kotomi.db ~/kotomi-backups/kotomi-$(date +%Y%m%d-%H%M%S).db

# Or use docker-compose
docker-compose cp kotomi:/app/data/kotomi.db ~/kotomi-backups/kotomi-$(date +%Y%m%d-%H%M%S).db
```

### Cloud Run Deployment

Cloud Run containers are stateless, but if you're using persistent storage:

```bash
# Connect to Cloud Run instance
gcloud run services proxy kotomi --region us-central1

# Copy database via kubectl (if using GKE)
kubectl cp kotomi-pod:/app/data/kotomi.db ./kotomi-backup-$(date +%Y%m%d).db
```

**Note**: For Cloud Run, use Cloud SQL or persistent disks with automated snapshots instead.

### Binary Deployment

```bash
# Simple file copy
cp /path/to/kotomi.db /path/to/backups/kotomi-$(date +%Y%m%d-%H%M%S).db

# With compression
tar -czf /path/to/backups/kotomi-$(date +%Y%m%d-%H%M%S).tar.gz /path/to/kotomi.db

# With verification
cp /path/to/kotomi.db /path/to/backups/kotomi-$(date +%Y%m%d-%H%M%S).db
sqlite3 /path/to/backups/kotomi-$(date +%Y%m%d-%H%M%S).db "PRAGMA integrity_check;"
```

## Automated Backup

### Bash Script for Automated Backups

Create `/usr/local/bin/backup-kotomi.sh`:

```bash
#!/bin/bash

# Configuration
BACKUP_DIR="/var/backups/kotomi"
RETENTION_DAYS=7
DB_PATH="/app/data/kotomi.db"
DATE=$(date +%Y%m%d-%H%M%S)
BACKUP_FILE="$BACKUP_DIR/kotomi-$DATE.db"

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Perform backup
if [ -f "$DB_PATH" ]; then
    # Use SQLite online backup
    sqlite3 "$DB_PATH" ".backup '$BACKUP_FILE'"
    
    # Verify backup
    if sqlite3 "$BACKUP_FILE" "PRAGMA integrity_check;" | grep -q "ok"; then
        echo "Backup successful: $BACKUP_FILE"
        
        # Compress backup
        gzip "$BACKUP_FILE"
        
        # Upload to cloud storage (optional)
        # gsutil cp "$BACKUP_FILE.gz" gs://your-bucket/kotomi-backups/
        # aws s3 cp "$BACKUP_FILE.gz" s3://your-bucket/kotomi-backups/
    else
        echo "Backup verification failed!"
        exit 1
    fi
else
    echo "Database file not found: $DB_PATH"
    exit 1
fi

# Cleanup old backups
find "$BACKUP_DIR" -name "kotomi-*.db.gz" -mtime +$RETENTION_DAYS -delete

echo "Backup completed and old backups cleaned up"
```

Make it executable:

```bash
chmod +x /usr/local/bin/backup-kotomi.sh
```

### Cron Job Setup

```bash
# Edit crontab
crontab -e

# Add daily backup at 2 AM
0 2 * * * /usr/local/bin/backup-kotomi.sh >> /var/log/kotomi-backup.log 2>&1

# Or hourly backups
0 * * * * /usr/local/bin/backup-kotomi.sh >> /var/log/kotomi-backup.log 2>&1
```

### Docker Compose with Backup Service

Add a backup service to `docker-compose.yml`:

```yaml
version: '3.8'

services:
  kotomi:
    image: gcr.io/your-project/kotomi:0.1.0-beta.1
    # ... existing configuration ...
    volumes:
      - kotomi-data:/app/data

  backup:
    image: alpine:latest
    volumes:
      - kotomi-data:/data:ro
      - ./backups:/backups
    command: >
      sh -c "
      while true; do
        cp /data/kotomi.db /backups/kotomi-$$(date +%Y%m%d-%H%M%S).db
        find /backups -name 'kotomi-*.db' -mtime +7 -delete
        sleep 86400
      done
      "
    restart: unless-stopped

volumes:
  kotomi-data:
```

### Cloud Storage Integration

#### Google Cloud Storage

```bash
#!/bin/bash
# backup-to-gcs.sh

BACKUP_FILE="/tmp/kotomi-$(date +%Y%m%d-%H%M%S).db"
BUCKET="gs://your-kotomi-backups"

# Create backup
sqlite3 /app/data/kotomi.db ".backup '$BACKUP_FILE'"

# Upload to GCS
gsutil cp "$BACKUP_FILE" "$BUCKET/"

# Cleanup local copy
rm "$BACKUP_FILE"

# Lifecycle: Delete backups older than 30 days (configure in GCS)
```

#### AWS S3

```bash
#!/bin/bash
# backup-to-s3.sh

BACKUP_FILE="/tmp/kotomi-$(date +%Y%m%d-%H%M%S).db"
BUCKET="s3://your-kotomi-backups"

# Create backup
sqlite3 /app/data/kotomi.db ".backup '$BACKUP_FILE'"

# Upload to S3
aws s3 cp "$BACKUP_FILE" "$BUCKET/"

# Cleanup local copy
rm "$BACKUP_FILE"
```

## Restore Procedures

### Restore from Backup File

#### Docker Deployment

```bash
# Stop the container
docker-compose down

# Restore backup to volume
docker run --rm \
  -v kotomi-data:/data \
  -v ~/kotomi-backups:/backup \
  alpine \
  cp /backup/kotomi-YYYYMMDD-HHMMSS.db /data/kotomi.db

# Start the container
docker-compose up -d

# Verify
docker-compose logs -f kotomi
```

#### Binary Deployment

```bash
# Stop Kotomi service
systemctl stop kotomi  # or however you run it

# Backup current database (just in case)
cp /path/to/kotomi.db /path/to/kotomi.db.backup-before-restore

# Restore from backup
cp /path/to/backups/kotomi-YYYYMMDD-HHMMSS.db /path/to/kotomi.db

# Verify integrity
sqlite3 /path/to/kotomi.db "PRAGMA integrity_check;"

# Start service
systemctl start kotomi

# Check logs
journalctl -u kotomi -f
```

### Restore from Compressed Backup

```bash
# Decompress
gunzip kotomi-YYYYMMDD-HHMMSS.db.gz

# Then follow normal restore procedure
```

### Restore from Cloud Storage

```bash
# Download from GCS
gsutil cp gs://your-kotomi-backups/kotomi-YYYYMMDD-HHMMSS.db /tmp/restore.db

# Or from S3
aws s3 cp s3://your-kotomi-backups/kotomi-YYYYMMDD-HHMMSS.db /tmp/restore.db

# Then follow normal restore procedure
```

### Verify Restore

After restoring, verify:

```bash
# Check database integrity
sqlite3 /path/to/kotomi.db "PRAGMA integrity_check;"

# Check tables exist
sqlite3 /path/to/kotomi.db ".tables"

# Check record counts
sqlite3 /path/to/kotomi.db "
SELECT 
  (SELECT COUNT(*) FROM sites) as sites,
  (SELECT COUNT(*) FROM pages) as pages,
  (SELECT COUNT(*) FROM comments) as comments,
  (SELECT COUNT(*) FROM reactions) as reactions;
"

# Test application
curl http://localhost:8080/healthz
```

## Disaster Recovery

### Complete Data Loss Scenario

If the database is completely lost:

1. **Stop the application** to prevent further damage
2. **Identify latest backup** from your backup location
3. **Restore from backup** following procedures above
4. **Verify restore** with integrity checks
5. **Start application** and test functionality
6. **Review logs** for any issues
7. **Notify users** if there's data loss

### Partial Data Corruption

If the database is corrupted but recoverable:

```bash
# Try to recover
sqlite3 kotomi.db ".dump" > kotomi-dump.sql

# Create new database
mv kotomi.db kotomi.db.corrupted
sqlite3 kotomi.db < kotomi-dump.sql

# Verify
sqlite3 kotomi.db "PRAGMA integrity_check;"
```

### Rolling Back After Failed Update

If an update causes issues:

1. Stop the application
2. Restore database from before update
3. Revert application to previous version
4. Start application
5. Investigate issue before re-attempting update

## Best Practices

### Backup Best Practices

1. **Automate backups**: Don't rely on manual processes
2. **Test restores**: Verify backups work regularly
3. **Multiple locations**: Store backups in different locations
4. **Encrypt backups**: Protect sensitive data
5. **Monitor backups**: Alert on backup failures
6. **Document procedures**: Keep this guide updated

### Retention Strategy

**Recommended retention**:
- Keep last 7 daily backups
- Keep last 4 weekly backups (Sunday)
- Keep last 12 monthly backups (1st of month)
- Archive annually

**Implementation**:
```bash
# In your backup script
BACKUP_DIR="/var/backups/kotomi"
DATE=$(date +%Y%m%d)
DAY=$(date +%A)

# Daily backup
sqlite3 kotomi.db ".backup '$BACKUP_DIR/daily/kotomi-$DATE.db'"

# Weekly backup (Sundays)
if [ "$DAY" = "Sunday" ]; then
    cp "$BACKUP_DIR/daily/kotomi-$DATE.db" "$BACKUP_DIR/weekly/kotomi-$DATE.db"
fi

# Monthly backup (1st of month)
if [ "$(date +%d)" = "01" ]; then
    cp "$BACKUP_DIR/daily/kotomi-$DATE.db" "$BACKUP_DIR/monthly/kotomi-$DATE.db"
fi

# Cleanup old backups
find "$BACKUP_DIR/daily" -name "kotomi-*.db" -mtime +7 -delete
find "$BACKUP_DIR/weekly" -name "kotomi-*.db" -mtime +28 -delete
find "$BACKUP_DIR/monthly" -name "kotomi-*.db" -mtime +365 -delete
```

### Security Considerations

1. **Encrypt at rest**: Use encryption for backup storage
2. **Access control**: Limit who can access backups
3. **Secure transfer**: Use encrypted channels (HTTPS, SSH)
4. **Audit logs**: Track backup/restore operations
5. **Test security**: Verify backup encryption works

### Monitoring

Monitor your backup system:

```bash
# Check backup age
#!/bin/bash
LATEST_BACKUP=$(ls -t /var/backups/kotomi/*.db | head -1)
BACKUP_AGE=$(( ($(date +%s) - $(stat -c %Y "$LATEST_BACKUP")) / 3600 ))

if [ $BACKUP_AGE -gt 24 ]; then
    echo "ALERT: Latest backup is $BACKUP_AGE hours old!"
    # Send alert (email, Slack, PagerDuty, etc.)
fi
```

## Troubleshooting

### Backup Fails with "Database is Locked"

```bash
# Wait for ongoing operations
sleep 5

# Or use timeout
timeout 60 sqlite3 kotomi.db ".backup 'backup.db'"
```

### Restore Fails with "Database Disk Image is Malformed"

```bash
# Try recovery
sqlite3 corrupted.db ".dump" | sqlite3 new.db

# Or use recover command
sqlite3 corrupted.db ".recover" | sqlite3 new.db
```

### Backup File is Too Large

```bash
# Use compression
sqlite3 kotomi.db ".backup 'backup.db'"
gzip backup.db

# Or use vacuum before backup
sqlite3 kotomi.db "VACUUM;"
```

## Future Enhancements

Planned improvements:
- Point-in-time recovery
- Incremental backups
- Automatic backup testing
- Backup monitoring dashboard
- One-click restore from UI
- PostgreSQL support for better backup options

---

**Version**: 0.1.0-beta.1
**Last Updated**: February 4, 2026
