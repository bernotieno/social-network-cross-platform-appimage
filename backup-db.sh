#!/bin/bash

# Database Backup Script for Social Network Application
# This script creates backups of the SQLite database

set -e

# Configuration
BACKUP_DIR="./backups"
DB_PATH="./data/social_network.db"
RETENTION_DAYS=30

# Create backup directory if it doesn't exist
mkdir -p "$BACKUP_DIR"

# Generate timestamp
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/social_network_$TIMESTAMP.db"

echo "🗄️  Starting database backup..."

# Check if database exists
if [ ! -f "$DB_PATH" ]; then
    echo "❌ Error: Database file not found at $DB_PATH"
    exit 1
fi

# Create backup
cp "$DB_PATH" "$BACKUP_FILE"

# Verify backup
if [ -f "$BACKUP_FILE" ]; then
    BACKUP_SIZE=$(du -h "$BACKUP_FILE" | cut -f1)
    echo "✅ Backup created successfully: $BACKUP_FILE ($BACKUP_SIZE)"
else
    echo "❌ Error: Backup failed"
    exit 1
fi

# Clean up old backups (keep only last 30 days)
echo "🧹 Cleaning up old backups (keeping last $RETENTION_DAYS days)..."
find "$BACKUP_DIR" -name "social_network_*.db" -type f -mtime +$RETENTION_DAYS -delete

# Count remaining backups
BACKUP_COUNT=$(find "$BACKUP_DIR" -name "social_network_*.db" -type f | wc -l)
echo "📊 Total backups: $BACKUP_COUNT"

echo "🎉 Backup completed successfully!"
