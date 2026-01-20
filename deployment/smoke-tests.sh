#!/bin/bash
# Smoke Tests for Production Deployment
# Runs after deployment to verify critical functionality

set -euo pipefail

# Configuration
BASE_URL="${BASE_URL:-http://localhost:8000}"
TIMEOUT=5

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Counters
PASSED=0
FAILED=0

echo "========================================="
echo "Production Smoke Tests"
echo "========================================="
echo "Testing: ${BASE_URL}"
echo ""

test_pass() {
    echo -e "${GREEN}✅ PASS${NC}: $1"
    ((PASSED++))
}

test_fail() {
    echo -e "${RED}❌ FAIL${NC}: $1"
    ((FAILED++))
}

# Test 1: Basic health check
echo "Test 1: Basic Health Check"
if curl -sf --max-time ${TIMEOUT} "${BASE_URL}/health" > /dev/null; then
    RESPONSE=$(curl -s "${BASE_URL}/health")
    if echo "${RESPONSE}" | grep -q "healthy"; then
        test_pass "Health endpoint responding correctly"
    else
        test_fail "Health endpoint returned unexpected response: ${RESPONSE}"
    fi
else
    test_fail "Health endpoint not reachable"
fi

# Test 2: Readiness check (includes database, redis, IDR)
echo ""
echo "Test 2: Readiness Check"
if curl -sf --max-time ${TIMEOUT} "${BASE_URL}/health/ready" > /dev/null; then
    RESPONSE=$(curl -s "${BASE_URL}/health/ready")
    if echo "${RESPONSE}" | grep -q '"ready":true'; then
        test_pass "Readiness check passed (database, redis, IDR healthy)"
    else
        test_fail "Readiness check failed: ${RESPONSE}"
    fi
else
    test_fail "Readiness endpoint not reachable"
fi

# Test 3: Metrics endpoint
echo ""
echo "Test 3: Metrics Endpoint"
if curl -sf --max-time ${TIMEOUT} "${BASE_URL}/metrics" > /dev/null; then
    RESPONSE=$(curl -s "${BASE_URL}/metrics")
    if echo "${RESPONSE}" | grep -q "catalyst_"; then
        test_pass "Metrics endpoint returning Prometheus metrics"
    else
        test_fail "Metrics endpoint not returning expected format"
    fi
else
    test_fail "Metrics endpoint not reachable"
fi

# Test 4: Auction endpoint (with invalid request - should return 400, not 500)
echo ""
echo "Test 4: Auction Endpoint (Error Handling)"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" --max-time ${TIMEOUT} \
    -X POST \
    -H "Content-Type: application/json" \
    -d '{"invalid":"request"}' \
    "${BASE_URL}/openrtb2/auction" || echo "000")

if [ "${HTTP_CODE}" = "400" ] || [ "${HTTP_CODE}" = "401" ]; then
    test_pass "Auction endpoint properly rejects invalid requests (HTTP ${HTTP_CODE})"
elif [ "${HTTP_CODE}" = "500" ]; then
    test_fail "Auction endpoint returned 500 error"
else
    test_fail "Auction endpoint returned unexpected status: ${HTTP_CODE}"
fi

# Test 5: Database connectivity (via health check)
echo ""
echo "Test 5: Database Connectivity"
READY_RESPONSE=$(curl -s "${BASE_URL}/health/ready")
if echo "${READY_RESPONSE}" | grep -q '"database":{"status":"healthy"}'; then
    test_pass "Database connectivity verified"
else
    test_fail "Database not healthy: ${READY_RESPONSE}"
fi

# Test 6: Redis connectivity (via health check)
echo ""
echo "Test 6: Redis Connectivity"
if echo "${READY_RESPONSE}" | grep -q '"redis":{"status":"healthy"}'; then
    test_pass "Redis connectivity verified"
else
    test_fail "Redis not healthy: ${READY_RESPONSE}"
fi

# Test 7: CORS headers (if applicable)
echo ""
echo "Test 7: CORS Headers"
CORS_HEADERS=$(curl -sI -X OPTIONS "${BASE_URL}/health" | grep -i "access-control" || echo "")
if [ -n "${CORS_HEADERS}" ]; then
    test_pass "CORS headers present"
else
    echo -e "${YELLOW}⚠️  INFO${NC}: No CORS headers found (may be intentional)"
fi

# Test 8: Response time check
echo ""
echo "Test 8: Response Time"
START_TIME=$(date +%s%N)
curl -sf --max-time ${TIMEOUT} "${BASE_URL}/health" > /dev/null
END_TIME=$(date +%s%N)
RESPONSE_TIME_MS=$(( (END_TIME - START_TIME) / 1000000 ))

if [ ${RESPONSE_TIME_MS} -lt 100 ]; then
    test_pass "Response time: ${RESPONSE_TIME_MS}ms (excellent)"
elif [ ${RESPONSE_TIME_MS} -lt 500 ]; then
    test_pass "Response time: ${RESPONSE_TIME_MS}ms (acceptable)"
else
    test_fail "Response time: ${RESPONSE_TIME_MS}ms (too slow)"
fi

# Test 9: Check Docker containers running
echo ""
echo "Test 9: Docker Containers Status"
if command -v docker &> /dev/null; then
    if docker-compose ps | grep -q "Up"; then
        RUNNING_CONTAINERS=$(docker-compose ps --services --filter "status=running" | wc -l)
        test_pass "${RUNNING_CONTAINERS} Docker containers running"
    else
        test_fail "No Docker containers running"
    fi
else
    echo -e "${YELLOW}⚠️  INFO${NC}: Docker not available (may be remote deployment)"
fi

# Test 10: Check backup service
echo ""
echo "Test 10: Backup Service"
if command -v docker &> /dev/null; then
    if docker ps | grep -q "catalyst-backup"; then
        test_pass "Backup service container running"
    else
        test_fail "Backup service not running"
    fi
else
    echo -e "${YELLOW}⚠️  INFO${NC}: Cannot check backup service (Docker not available)"
fi

# Summary
echo ""
echo "========================================="
echo "Smoke Test Summary"
echo "========================================="
echo -e "${GREEN}Passed:${NC} ${PASSED}"
echo -e "${RED}Failed:${NC} ${FAILED}"
echo "========================================="
echo ""

if [ ${FAILED} -eq 0 ]; then
    echo -e "${GREEN}✅ All smoke tests passed!${NC}"
    echo -e "${GREEN}Deployment appears healthy.${NC}"
    exit 0
else
    echo -e "${RED}❌ ${FAILED} smoke tests failed!${NC}"
    echo -e "${RED}Investigate failures before continuing.${NC}"
    exit 1
fi
