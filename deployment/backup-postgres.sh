#!/bin/bash
# PostgreSQL Automated Backup Script
# Performs automated backups with retention policy

set -euo pipefail

# Configuration
BACKUP_DIR="${BACKUP_DIR:-/backups}"
S3_BUCKET="${S3_BUCKET:-}"
POSTGRES_HOST="${POSTGRES_HOST:-catalyst-postgres}"
POSTGRES_PORT="${POSTGRES_PORT:-5432}"
POSTGRES_DB="${POSTGRES_DB:-catalyst}"
POSTGRES_USER="${POSTGRES_USER:-catalyst}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-changeme}"
RETENTION_DAYS="${RETENTION_DAYS:-7}"
RETENTION_WEEKS="${RETENTION_WEEKS:-4}"
RETENTION_MONTHS="${RETENTION_MONTHS:-3}"

# Timestamp for backup filename
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
DATE=$(date +%Y%m%d)
BACKUP_FILE="catalyst_backup_${TIMESTAMP}.sql.gz"
BACKUP_PATH="${BACKUP_DIR}/${BACKUP_FILE}"

# Create backup directory structure
mkdir -p "${BACKUP_DIR}/daily"
mkdir -p "${BACKUP_DIR}/weekly"
mkdir -p "${BACKUP_DIR}/monthly"
mkdir -p "${BACKUP_DIR}/latest"

echo "=== PostgreSQL Backup Started: $(date) ==="
echo "Database: ${POSTGRES_DB}"
echo "Host: ${POSTGRES_HOST}"
echo "Backup file: ${BACKUP_FILE}"

# Perform the backup using pg_dump
export PGPASSWORD="${POSTGRES_PASSWORD}"

if pg_dump -h "${POSTGRES_HOST}" \
           -p "${POSTGRES_PORT}" \
           -U "${POSTGRES_USER}" \
           -d "${POSTGRES_DB}" \
           --format=custom \
           --compress=9 \
           --verbose \
           --file="${BACKUP_PATH}"; then
    echo "✅ Backup completed successfully: ${BACKUP_FILE}"
else
    echo "❌ Backup failed!"
    exit 1
fi

# Verify backup file exists and is not empty
if [ ! -s "${BACKUP_PATH}" ]; then
    echo "❌ Backup file is empty or doesn't exist!"
    exit 1
fi

BACKUP_SIZE=$(du -h "${BACKUP_PATH}" | cut -f1)
echo "Backup size: ${BACKUP_SIZE}"

# Copy to latest
cp "${BACKUP_PATH}" "${BACKUP_DIR}/latest/latest.sql.gz"

# Copy to daily backups
cp "${BACKUP_PATH}" "${BACKUP_DIR}/daily/daily_${DATE}.sql.gz"

# Weekly backup (every Sunday)
if [ "$(date +%u)" -eq 7 ]; then
    WEEK_NUM=$(date +%V)
    cp "${BACKUP_PATH}" "${BACKUP_DIR}/weekly/weekly_${WEEK_NUM}_${DATE}.sql.gz"
    echo "📅 Weekly backup created"
fi

# Monthly backup (first day of month)
if [ "$(date +%d)" -eq "01" ]; then
    MONTH=$(date +%Y%m)
    cp "${BACKUP_PATH}" "${BACKUP_DIR}/monthly/monthly_${MONTH}.sql.gz"
    echo "📅 Monthly backup created"
fi

# Verify backup integrity
echo "Verifying backup integrity..."
if pg_restore --list "${BACKUP_PATH}" > /dev/null 2>&1; then
    echo "✅ Backup integrity verified"
else
    echo "⚠️  Backup verification failed - file may be corrupted"
fi

# Apply retention policy
echo "Applying retention policy..."

# Remove daily backups older than RETENTION_DAYS
if [ -d "${BACKUP_DIR}/daily" ]; then
    find "${BACKUP_DIR}/daily" -name "daily_*.sql.gz" -mtime "+${RETENTION_DAYS}" -delete
    echo "Cleaned daily backups older than ${RETENTION_DAYS} days"
fi

# Remove weekly backups older than RETENTION_WEEKS weeks
if [ -d "${BACKUP_DIR}/weekly" ]; then
    find "${BACKUP_DIR}/weekly" -name "weekly_*.sql.gz" -mtime "+$((RETENTION_WEEKS * 7))" -delete
    echo "Cleaned weekly backups older than ${RETENTION_WEEKS} weeks"
fi

# Remove monthly backups older than RETENTION_MONTHS months
if [ -d "${BACKUP_DIR}/monthly" ]; then
    find "${BACKUP_DIR}/monthly" -name "monthly_*.sql.gz" -mtime "+$((RETENTION_MONTHS * 30))" -delete
    echo "Cleaned monthly backups older than ${RETENTION_MONTHS} months"
fi

# Upload to S3 if configured
if [ -n "${S3_BUCKET}" ]; then
    echo "Uploading to S3: s3://${S3_BUCKET}/catalyst-backups/"

    if command -v aws &> /dev/null; then
        # Upload daily backup
        aws s3 cp "${BACKUP_PATH}" "s3://${S3_BUCKET}/catalyst-backups/daily/${BACKUP_FILE}" \
            --storage-class STANDARD_IA

        # Upload latest
        aws s3 cp "${BACKUP_DIR}/latest/latest.sql.gz" \
            "s3://${S3_BUCKET}/catalyst-backups/latest/latest.sql.gz" \
            --storage-class STANDARD

        echo "✅ Uploaded to S3"

        # Clean up old S3 backups
        echo "Cleaning S3 backups older than ${RETENTION_DAYS} days..."
        CUTOFF_DATE=$(date -d "${RETENTION_DAYS} days ago" +%Y-%m-%d 2>/dev/null || date -v-${RETENTION_DAYS}d +%Y-%m-%d)
        aws s3 ls "s3://${S3_BUCKET}/catalyst-backups/daily/" | while read -r line; do
            BACKUP_DATE=$(echo "$line" | awk '{print $1}')
            if [[ "${BACKUP_DATE}" < "${CUTOFF_DATE}" ]]; then
                BACKUP_NAME=$(echo "$line" | awk '{print $4}')
                aws s3 rm "s3://${S3_BUCKET}/catalyst-backups/daily/${BACKUP_NAME}"
                echo "Deleted old S3 backup: ${BACKUP_NAME}"
            fi
        done
    else
        echo "⚠️  AWS CLI not found - skipping S3 upload"
    fi
fi

# Summary
echo ""
echo "=== Backup Summary ==="
echo "Timestamp: $(date)"
echo "Backup file: ${BACKUP_FILE}"
echo "Size: ${BACKUP_SIZE}"
echo "Location: ${BACKUP_PATH}"
if [ -n "${S3_BUCKET}" ]; then
    echo "S3 Location: s3://${S3_BUCKET}/catalyst-backups/daily/${BACKUP_FILE}"
fi
echo ""
echo "Current backup inventory:"
echo "Daily backups: $(find "${BACKUP_DIR}/daily" -name "daily_*.sql.gz" 2>/dev/null | wc -l)"
echo "Weekly backups: $(find "${BACKUP_DIR}/weekly" -name "weekly_*.sql.gz" 2>/dev/null | wc -l)"
echo "Monthly backups: $(find "${BACKUP_DIR}/monthly" -name "monthly_*.sql.gz" 2>/dev/null | wc -l)"

echo "=== Backup Completed: $(date) ==="
