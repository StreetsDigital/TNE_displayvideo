#!/bin/bash
# Performance Comparison Tool
# Compares metrics between Production (95%) and Staging (5%) environments

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "================================================"
echo "Catalyst Performance Comparison"
echo "Production (95%) vs Staging (5%)"
echo "================================================"
echo ""

# Check if running split deployment
if ! docker ps | grep -q "catalyst-staging"; then
    echo -e "${RED}ERROR: Staging container not running${NC}"
    echo "This tool requires split deployment (docker-compose-split.yml)"
    exit 1
fi

# Function to get container stats
get_container_stats() {
    local container=$1
    docker stats --no-stream --format "{{.Name}},{{.CPUPerc}},{{.MemUsage}},{{.NetIO}}" $container
}

# Function to count log entries
count_logs() {
    local container=$1
    local pattern=$2
    local minutes=${3:-60}

    docker logs --since ${minutes}m $container 2>&1 | grep -c "$pattern" || echo "0"
}

# Function to extract average duration
get_avg_duration() {
    local container=$1
    local minutes=${2:-60}

    docker logs --since ${minutes}m $container 2>&1 | \
        grep "duration_ms" | \
        grep -oP 'duration_ms":\K[0-9.]+' | \
        awk '{ sum += $1; n++ } END { if (n > 0) print sum / n; else print "0" }'
}

# Function to get error rate
get_error_rate() {
    local container=$1
    local minutes=${2:-60}

    local total=$(count_logs $container "HTTP request" $minutes)
    local errors=$(count_logs $container '"level":"error"' $minutes)

    if [ "$total" -gt 0 ]; then
        echo "scale=2; ($errors / $total) * 100" | bc
    else
        echo "0"
    fi
}

# Time period for analysis
TIME_MINUTES=60
echo -e "${BLUE}Analyzing last ${TIME_MINUTES} minutes...${NC}"
echo ""

# ================================================
# 1. RESOURCE USAGE
# ================================================
echo "================================================"
echo "1. RESOURCE USAGE"
echo "================================================"

echo -e "\n${YELLOW}Production Container:${NC}"
get_container_stats "catalyst-prod"

echo -e "\n${YELLOW}Staging Container:${NC}"
get_container_stats "catalyst-staging"

# ================================================
# 2. REQUEST VOLUME
# ================================================
echo -e "\n================================================"
echo "2. REQUEST VOLUME (Last ${TIME_MINUTES} min)"
echo "================================================"

PROD_REQUESTS=$(count_logs "catalyst-prod" "HTTP request" $TIME_MINUTES)
STAGING_REQUESTS=$(count_logs "catalyst-staging" "HTTP request" $TIME_MINUTES)
TOTAL_REQUESTS=$((PROD_REQUESTS + STAGING_REQUESTS))

echo -e "${YELLOW}Production:${NC} $PROD_REQUESTS requests"
echo -e "${YELLOW}Staging:${NC}    $STAGING_REQUESTS requests"
echo -e "${YELLOW}Total:${NC}      $TOTAL_REQUESTS requests"

if [ $TOTAL_REQUESTS -gt 0 ]; then
    PROD_PCT=$(echo "scale=1; ($PROD_REQUESTS / $TOTAL_REQUESTS) * 100" | bc)
    STAGING_PCT=$(echo "scale=1; ($STAGING_REQUESTS / $TOTAL_REQUESTS) * 100" | bc)
    echo -e "${YELLOW}Split:${NC}      ${PROD_PCT}% / ${STAGING_PCT}%"
fi

# ================================================
# 3. RESPONSE TIMES
# ================================================
echo -e "\n================================================"
echo "3. AVERAGE RESPONSE TIME (Last ${TIME_MINUTES} min)"
echo "================================================"

PROD_AVG=$(get_avg_duration "catalyst-prod" $TIME_MINUTES)
STAGING_AVG=$(get_avg_duration "catalyst-staging" $TIME_MINUTES)

echo -e "${YELLOW}Production:${NC} ${PROD_AVG}ms"
echo -e "${YELLOW}Staging:${NC}    ${STAGING_AVG}ms"

# Compare
if (( $(echo "$PROD_AVG > 0" | bc -l) && $(echo "$STAGING_AVG > 0" | bc -l) )); then
    DIFF=$(echo "scale=2; $STAGING_AVG - $PROD_AVG" | bc)
    DIFF_PCT=$(echo "scale=1; (($STAGING_AVG - $PROD_AVG) / $PROD_AVG) * 100" | bc)

    if (( $(echo "$DIFF > 0" | bc -l) )); then
        echo -e "${RED}⚠ Staging is ${DIFF}ms slower (${DIFF_PCT}% worse)${NC}"
    elif (( $(echo "$DIFF < 0" | bc -l) )); then
        DIFF=${DIFF#-}
        DIFF_PCT=${DIFF_PCT#-}
        echo -e "${GREEN}✓ Staging is ${DIFF}ms faster (${DIFF_PCT}% better)${NC}"
    else
        echo -e "${GREEN}✓ Same performance${NC}"
    fi
fi

# ================================================
# 4. ERROR RATES
# ================================================
echo -e "\n================================================"
echo "4. ERROR RATES (Last ${TIME_MINUTES} min)"
echo "================================================"

PROD_ERRORS=$(count_logs "catalyst-prod" '"level":"error"' $TIME_MINUTES)
STAGING_ERRORS=$(count_logs "catalyst-staging" '"level":"error"' $TIME_MINUTES)

echo -e "${YELLOW}Production Errors:${NC} $PROD_ERRORS"
echo -e "${YELLOW}Staging Errors:${NC}    $STAGING_ERRORS"

PROD_ERROR_RATE=$(get_error_rate "catalyst-prod" $TIME_MINUTES)
STAGING_ERROR_RATE=$(get_error_rate "catalyst-staging" $TIME_MINUTES)

echo -e "${YELLOW}Production Error Rate:${NC} ${PROD_ERROR_RATE}%"
echo -e "${YELLOW}Staging Error Rate:${NC}    ${STAGING_ERROR_RATE}%"

# Compare
if (( $(echo "$STAGING_ERROR_RATE > $PROD_ERROR_RATE" | bc -l) )); then
    echo -e "${RED}⚠ Staging has higher error rate${NC}"
elif (( $(echo "$STAGING_ERROR_RATE < $PROD_ERROR_RATE" | bc -l) )); then
    echo -e "${GREEN}✓ Staging has lower error rate${NC}"
else
    echo -e "${GREEN}✓ Same error rate${NC}"
fi

# ================================================
# 5. IVT DETECTION
# ================================================
echo -e "\n================================================"
echo "5. IVT DETECTION (Last ${TIME_MINUTES} min)"
echo "================================================"

PROD_IVT=$(count_logs "catalyst-prod" "IVT detected" $TIME_MINUTES)
STAGING_IVT=$(count_logs "catalyst-staging" "IVT detected" $TIME_MINUTES)

PROD_IVT_BLOCKED=$(count_logs "catalyst-prod" "Request blocked" $TIME_MINUTES)
STAGING_IVT_BLOCKED=$(count_logs "catalyst-staging" "Request blocked" $TIME_MINUTES)

echo -e "${YELLOW}Production IVT Detected:${NC} $PROD_IVT"
echo -e "${YELLOW}Staging IVT Detected:${NC}    $STAGING_IVT"
echo -e "${YELLOW}Production IVT Blocked:${NC}  $PROD_IVT_BLOCKED"
echo -e "${YELLOW}Staging IVT Blocked:${NC}     $STAGING_IVT_BLOCKED"

# ================================================
# 6. AUCTION METRICS
# ================================================
echo -e "\n================================================"
echo "6. AUCTION METRICS (Last ${TIME_MINUTES} min)"
echo "================================================"

PROD_AUCTIONS=$(count_logs "catalyst-prod" "Auction complete" $TIME_MINUTES)
STAGING_AUCTIONS=$(count_logs "catalyst-staging" "Auction complete" $TIME_MINUTES)

echo -e "${YELLOW}Production Auctions:${NC} $PROD_AUCTIONS"
echo -e "${YELLOW}Staging Auctions:${NC}    $STAGING_AUCTIONS"

# ================================================
# 7. REDIS MEMORY
# ================================================
echo -e "\n================================================"
echo "7. REDIS MEMORY USAGE"
echo "================================================"

REDIS_PROD=$(docker exec redis-prod redis-cli INFO memory 2>/dev/null | grep "used_memory_human" | cut -d: -f2 | tr -d '\r')
REDIS_STAGING=$(docker exec redis-staging redis-cli INFO memory 2>/dev/null | grep "used_memory_human" | cut -d: -f2 | tr -d '\r')

echo -e "${YELLOW}Production Redis:${NC} ${REDIS_PROD}"
echo -e "${YELLOW}Staging Redis:${NC}    ${REDIS_STAGING}"

# ================================================
# 8. RECOMMENDATIONS
# ================================================
echo -e "\n================================================"
echo "8. RECOMMENDATIONS"
echo "================================================"

HAS_ISSUES=false

# Check response time
if (( $(echo "$PROD_AVG > 0" | bc -l) && $(echo "$STAGING_AVG > 0" | bc -l) )); then
    if (( $(echo "$STAGING_AVG > $PROD_AVG * 1.2" | bc -l) )); then
        echo -e "${RED}⚠ Staging is 20%+ slower than production${NC}"
        echo "  Action: Investigate staging performance issues"
        HAS_ISSUES=true
    fi
fi

# Check error rate
if (( $(echo "$STAGING_ERROR_RATE > $PROD_ERROR_RATE * 1.5" | bc -l) )); then
    echo -e "${RED}⚠ Staging error rate is 50%+ higher${NC}"
    echo "  Action: Review staging error logs"
    HAS_ISSUES=true
fi

# Check if staging getting traffic
if [ $STAGING_REQUESTS -eq 0 ] && [ $PROD_REQUESTS -gt 0 ]; then
    echo -e "${RED}⚠ Staging is not receiving any traffic${NC}"
    echo "  Action: Check nginx split configuration"
    HAS_ISSUES=true
fi

# Check split ratio
if [ $TOTAL_REQUESTS -gt 100 ]; then
    ACTUAL_STAGING_PCT=$(echo "scale=1; ($STAGING_REQUESTS / $TOTAL_REQUESTS) * 100" | bc)
    if (( $(echo "$ACTUAL_STAGING_PCT < 3" | bc -l) || $(echo "$ACTUAL_STAGING_PCT > 7" | bc -l) )); then
        echo -e "${YELLOW}⚠ Traffic split is not 95/5 (currently ${PROD_PCT}%/${STAGING_PCT}%)${NC}"
        echo "  Note: This is normal with low traffic, check again with more requests"
    fi
fi

if [ "$HAS_ISSUES" = false ]; then
    echo -e "${GREEN}✓ No issues detected${NC}"
    echo -e "${GREEN}✓ Staging performance is comparable to production${NC}"
    echo ""
    echo "Consider increasing staging traffic percentage or rolling out changes."
fi

# ================================================
# SUMMARY
# ================================================
echo -e "\n================================================"
echo "SUMMARY"
echo "================================================"

echo "Time Period:     Last ${TIME_MINUTES} minutes"
echo "Total Requests:  ${TOTAL_REQUESTS}"
echo "Traffic Split:   ${PROD_PCT}% / ${STAGING_PCT}%"
echo ""
echo "For detailed logs:"
echo "  docker logs -f catalyst-prod"
echo "  docker logs -f catalyst-staging"
echo ""
echo "To adjust split ratio, edit: nginx-split.conf"

echo ""
echo "================================================"
