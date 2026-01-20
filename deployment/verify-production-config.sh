#!/bin/bash
# Production Configuration Verification Script
# Validates all critical settings before deployment

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Counters
PASSED=0
FAILED=0
WARNINGS=0

# Environment file to check
ENV_FILE="${ENV_FILE:-.env.production}"

echo "========================================="
echo "Production Configuration Verification"
echo "========================================="
echo "Checking: ${ENV_FILE}"
echo ""

# Helper functions
check_pass() {
    echo -e "${GREEN}✅ PASS${NC}: $1"
    ((PASSED++))
}

check_fail() {
    echo -e "${RED}❌ FAIL${NC}: $1"
    ((FAILED++))
}

check_warn() {
    echo -e "${YELLOW}⚠️  WARN${NC}: $1"
    ((WARNINGS++))
}

# Check if env file exists
if [ ! -f "${ENV_FILE}" ]; then
    echo -e "${RED}❌ Environment file not found: ${ENV_FILE}${NC}"
    exit 1
fi

# Source environment file
set -a
source "${ENV_FILE}"
set +a

echo "=== Security Checks ==="
echo ""

# Check Redis password
if [ -z "${REDIS_PASSWORD:-}" ] || [ "${REDIS_PASSWORD}" = "CHANGE_ME_REDIS_PASSWORD" ]; then
    check_fail "Redis password not configured or using default"
else
    if [ ${#REDIS_PASSWORD} -lt 16 ]; then
        check_warn "Redis password is short (< 16 chars)"
    else
        check_pass "Redis password configured (${#REDIS_PASSWORD} chars)"
    fi
fi

# Check PostgreSQL password
if [ "${DB_PASSWORD:-changeme}" = "changeme" ] || [ "${DB_PASSWORD}" = "CHANGE_ME" ]; then
    check_fail "PostgreSQL password still set to default 'changeme'"
else
    if [ ${#DB_PASSWORD} -lt 16 ]; then
        check_warn "PostgreSQL password is short (< 16 chars)"
    else
        check_pass "PostgreSQL password configured (${#DB_PASSWORD} chars)"
    fi
fi

# Check JWT secret (if applicable)
if [ -z "${JWT_SECRET:-}" ]; then
    check_warn "JWT_SECRET not set (may not be required)"
elif [ ${#JWT_SECRET} -lt 32 ]; then
    check_warn "JWT_SECRET is short (< 32 chars)"
else
    check_pass "JWT_SECRET configured (${#JWT_SECRET} chars)"
fi

# Check for any remaining CHANGE_ME values
CHANGE_ME_COUNT=$(grep -c "CHANGE_ME" "${ENV_FILE}" || true)
if [ ${CHANGE_ME_COUNT} -gt 0 ]; then
    check_fail "Found ${CHANGE_ME_COUNT} CHANGE_ME placeholders in ${ENV_FILE}"
    grep -n "CHANGE_ME" "${ENV_FILE}" | while read -r line; do
        echo "   Line: $line"
    done
else
    check_pass "No CHANGE_ME placeholders found"
fi

echo ""
echo "=== Privacy & Compliance Checks ==="
echo ""

# Check GDPR enforcement
if [ "${PBS_ENFORCE_GDPR:-false}" != "true" ]; then
    check_fail "GDPR enforcement not enabled (PBS_ENFORCE_GDPR)"
else
    check_pass "GDPR enforcement enabled"
fi

# Check CCPA enforcement
if [ "${PBS_ENFORCE_CCPA:-false}" != "true" ]; then
    check_fail "CCPA enforcement not enabled (PBS_ENFORCE_CCPA)"
else
    check_pass "CCPA enforcement enabled"
fi

# Check IP anonymization
if [ "${PBS_ANONYMIZE_IP:-false}" != "true" ]; then
    check_warn "IP anonymization not enabled (recommended for GDPR)"
else
    check_pass "IP anonymization enabled"
fi

# Check geo enforcement
if [ "${PBS_GEO_ENFORCEMENT:-false}" != "true" ]; then
    check_warn "Geo enforcement not enabled (recommended)"
else
    check_pass "Geo enforcement enabled"
fi

echo ""
echo "=== Backup Configuration Checks ==="
echo ""

# Check S3 bucket configured
if [ -z "${BACKUP_S3_BUCKET:-}" ]; then
    check_warn "S3 backup bucket not configured (local backups only)"
else
    check_pass "S3 backup bucket configured: ${BACKUP_S3_BUCKET}"
fi

# Check AWS credentials (if S3 configured)
if [ -n "${BACKUP_S3_BUCKET:-}" ]; then
    if [ -z "${AWS_ACCESS_KEY_ID:-}" ] || [ -z "${AWS_SECRET_ACCESS_KEY:-}" ]; then
        check_fail "S3 bucket configured but AWS credentials missing"
    else
        check_pass "AWS credentials configured for S3 backup"
    fi
fi

# Check backup retention
if [ -z "${BACKUP_RETENTION_DAYS:-}" ]; then
    check_warn "Backup retention not set, using default (7 days)"
else
    check_pass "Backup retention: ${BACKUP_RETENTION_DAYS} days"
fi

echo ""
echo "=== Monitoring & Logging Checks ==="
echo ""

# Check log level
if [ "${LOG_LEVEL:-debug}" = "debug" ]; then
    check_warn "Log level set to DEBUG (should be INFO or WARN in production)"
elif [ "${LOG_LEVEL:-info}" = "info" ] || [ "${LOG_LEVEL:-warn}" = "warn" ]; then
    check_pass "Log level appropriately set: ${LOG_LEVEL}"
else
    check_warn "Unexpected log level: ${LOG_LEVEL}"
fi

# Check Prometheus enabled
if [ "${ENABLE_METRICS:-true}" = "true" ]; then
    check_pass "Metrics collection enabled"
else
    check_warn "Metrics collection disabled (recommended for production)"
fi

echo ""
echo "=== Infrastructure Checks ==="
echo ""

# Check Docker Compose syntax
if docker-compose -f docker-compose.yml config >/dev/null 2>&1; then
    check_pass "docker-compose.yml syntax is valid"
else
    check_fail "docker-compose.yml has syntax errors"
fi

# Check SSL certificates exist
if [ -d "ssl" ] && [ -f "ssl/server.crt" ] && [ -f "ssl/server.key" ]; then
    check_pass "SSL certificates found"

    # Check certificate expiry
    if command -v openssl &> /dev/null; then
        EXPIRY=$(openssl x509 -enddate -noout -in ssl/server.crt | cut -d= -f2)
        EXPIRY_EPOCH=$(date -d "${EXPIRY}" +%s 2>/dev/null || date -j -f "%b %d %T %Y %Z" "${EXPIRY}" +%s 2>/dev/null || echo "0")
        NOW_EPOCH=$(date +%s)
        DAYS_UNTIL_EXPIRY=$(( (EXPIRY_EPOCH - NOW_EPOCH) / 86400 ))

        if [ ${DAYS_UNTIL_EXPIRY} -lt 30 ]; then
            check_warn "SSL certificate expires in ${DAYS_UNTIL_EXPIRY} days"
        else
            check_pass "SSL certificate valid for ${DAYS_UNTIL_EXPIRY} days"
        fi
    fi
else
    check_warn "SSL certificates not found in ssl/ directory"
fi

# Check migrations directory
if [ -d "migrations" ] && [ "$(ls -A migrations)" ]; then
    MIGRATION_COUNT=$(ls migrations/*.sql 2>/dev/null | wc -l)
    check_pass "Database migrations found (${MIGRATION_COUNT} files)"
else
    check_warn "No database migrations found"
fi

echo ""
echo "=== Performance Checks ==="
echo ""

# Check if resource limits are configured in docker-compose
if grep -q "limits:" docker-compose.yml; then
    check_pass "Resource limits configured in docker-compose.yml"
else
    check_warn "No resource limits found in docker-compose.yml"
fi

# Check connection pool settings
if [ -n "${DB_MAX_CONNS:-}" ]; then
    check_pass "Database connection pool configured: ${DB_MAX_CONNS}"
else
    check_warn "Database connection pool not configured (using defaults)"
fi

# Check rate limiting
if [ -n "${RATE_LIMIT_REQUESTS:-}" ]; then
    check_pass "Rate limiting configured: ${RATE_LIMIT_REQUESTS} req/min"
else
    check_warn "Rate limiting not configured"
fi

echo ""
echo "========================================="
echo "Verification Summary"
echo "========================================="
echo -e "${GREEN}Passed:${NC}   ${PASSED}"
echo -e "${YELLOW}Warnings:${NC} ${WARNINGS}"
echo -e "${RED}Failed:${NC}   ${FAILED}"
echo "========================================="
echo ""

if [ ${FAILED} -eq 0 ]; then
    if [ ${WARNINGS} -eq 0 ]; then
        echo -e "${GREEN}✅ All checks passed! Ready for production deployment.${NC}"
        exit 0
    else
        echo -e "${YELLOW}⚠️  All critical checks passed, but ${WARNINGS} warnings found.${NC}"
        echo -e "${YELLOW}Review warnings before deploying to production.${NC}"
        exit 0
    fi
else
    echo -e "${RED}❌ ${FAILED} critical checks failed!${NC}"
    echo -e "${RED}Fix all failures before deploying to production.${NC}"
    exit 1
fi
