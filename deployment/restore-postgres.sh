#!/bin/bash
# PostgreSQL Restore Script
# Restores database from a backup file

set -euo pipefail

# Configuration
BACKUP_FILE="${1:-}"
POSTGRES_HOST="${POSTGRES_HOST:-catalyst-postgres}"
POSTGRES_PORT="${POSTGRES_PORT:-5432}"
POSTGRES_DB="${POSTGRES_DB:-catalyst}"
POSTGRES_USER="${POSTGRES_USER:-catalyst}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-changeme}"
FORCE="${FORCE:-false}"

# Check if backup file is provided
if [ -z "${BACKUP_FILE}" ]; then
    echo "❌ Error: Backup file not specified"
    echo ""
    echo "Usage: $0 <backup_file> [FORCE=true]"
    echo ""
    echo "Examples:"
    echo "  $0 /backups/latest/latest.sql.gz"
    echo "  $0 /backups/daily/daily_20260119.sql.gz"
    echo "  FORCE=true $0 /backups/monthly/monthly_202601.sql.gz"
    echo ""
    echo "Available backups:"
    echo ""
    if [ -d "/backups/latest" ]; then
        echo "Latest:"
        ls -lh /backups/latest/*.sql.gz 2>/dev/null || echo "  (none)"
    fi
    if [ -d "/backups/daily" ]; then
        echo ""
        echo "Daily (last 7):"
        ls -lht /backups/daily/*.sql.gz 2>/dev/null | head -7 || echo "  (none)"
    fi
    exit 1
fi

# Check if backup file exists
if [ ! -f "${BACKUP_FILE}" ]; then
    echo "❌ Error: Backup file not found: ${BACKUP_FILE}"
    exit 1
fi

BACKUP_SIZE=$(du -h "${BACKUP_FILE}" | cut -f1)
echo "=== PostgreSQL Restore Started: $(date) ==="
echo "Backup file: ${BACKUP_FILE}"
echo "Backup size: ${BACKUP_SIZE}"
echo "Target database: ${POSTGRES_DB}"
echo "Target host: ${POSTGRES_HOST}"
echo ""

# Verify backup integrity before restore
echo "Verifying backup integrity..."
if ! pg_restore --list "${BACKUP_FILE}" > /dev/null 2>&1; then
    echo "❌ Error: Backup file is corrupted or invalid"
    exit 1
fi
echo "✅ Backup integrity verified"
echo ""

# Warning if not forcing
if [ "${FORCE}" != "true" ]; then
    echo "⚠️  WARNING: This will OVERWRITE the current database!"
    echo ""
    echo "Database: ${POSTGRES_DB}"
    echo "Host: ${POSTGRES_HOST}"
    echo ""
    echo "To proceed, set FORCE=true:"
    echo "  FORCE=true $0 ${BACKUP_FILE}"
    echo ""
    exit 1
fi

echo "🔥 FORCE mode enabled - proceeding with restore..."
echo ""

# Set password for pg_restore
export PGPASSWORD="${POSTGRES_PASSWORD}"

# Check if database exists
echo "Checking database connection..."
if ! psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d postgres -c '\q' 2>/dev/null; then
    echo "❌ Error: Cannot connect to PostgreSQL server"
    exit 1
fi
echo "✅ Connected to PostgreSQL server"

# Terminate existing connections to the database
echo "Terminating existing connections..."
psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d postgres <<EOF
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE datname = '${POSTGRES_DB}'
  AND pid <> pg_backend_pid();
EOF

# Drop and recreate database
echo "Dropping database: ${POSTGRES_DB}..."
psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d postgres -c "DROP DATABASE IF EXISTS ${POSTGRES_DB};"

echo "Creating database: ${POSTGRES_DB}..."
psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d postgres -c "CREATE DATABASE ${POSTGRES_DB} OWNER ${POSTGRES_USER};"

# Restore the backup
echo "Restoring backup..."
if pg_restore -h "${POSTGRES_HOST}" \
             -p "${POSTGRES_PORT}" \
             -U "${POSTGRES_USER}" \
             -d "${POSTGRES_DB}" \
             --verbose \
             --no-owner \
             --no-acl \
             "${BACKUP_FILE}"; then
    echo "✅ Restore completed successfully"
else
    echo "⚠️  Restore completed with warnings (this is often normal)"
fi

# Verify restore
echo ""
echo "Verifying restore..."
TABLE_COUNT=$(psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';")
echo "Tables restored: ${TABLE_COUNT}"

if [ "${TABLE_COUNT}" -gt 0 ]; then
    echo "✅ Restore verification passed"
else
    echo "⚠️  Warning: No tables found in restored database"
fi

# Show table sizes
echo ""
echo "Restored table sizes:"
psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" <<EOF
SELECT
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC
LIMIT 10;
EOF

echo ""
echo "=== Restore Completed: $(date) ==="
