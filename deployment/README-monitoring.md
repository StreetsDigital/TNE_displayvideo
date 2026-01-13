# Monitoring & Performance Comparison

## Purpose

This guide shows you how to **compare performance** between Production (95%) and Staging (5%) environments during traffic splitting.

## Quick Comparison Tool

### Run Automated Comparison

```bash
cd /opt/catalyst
chmod +x compare-performance.sh
./compare-performance.sh
```

**Output includes:**
- Resource usage (CPU, memory)
- Request volume and traffic split ratio
- Average response times
- Error rates
- IVT detection stats
- Auction metrics
- Redis memory usage
- Recommendations

**Example Output:**
```
================================================
Catalyst Performance Comparison
Production (95%) vs Staging (5%)
================================================

1. RESOURCE USAGE
Production: CPU 45%, Memory 1.2GB
Staging:    CPU 8%, Memory 350MB

2. REQUEST VOLUME (Last 60 min)
Production: 5700 requests
Staging:    300 requests
Total:      6000 requests
Split:      95.0% / 5.0%

3. AVERAGE RESPONSE TIME
Production: 42.5ms
Staging:    38.2ms
✓ Staging is 4.3ms faster (10.1% better)

4. ERROR RATES
Production: 12 errors (0.21%)
Staging:    1 error (0.33%)
✓ Same error rate

8. RECOMMENDATIONS
✓ No issues detected
✓ Staging performance is comparable to production
```

---

## Manual Performance Comparison

### 1. Response Time Comparison

#### Production Average Response Time
```bash
docker logs --since 60m catalyst-prod 2>&1 | \
  grep "duration_ms" | \
  grep -oP 'duration_ms":\K[0-9.]+' | \
  awk '{ sum += $1; n++ } END { if (n > 0) print "Avg:", sum / n, "ms"; }'
```

#### Staging Average Response Time
```bash
docker logs --since 60m catalyst-staging 2>&1 | \
  grep "duration_ms" | \
  grep -oP 'duration_ms":\K[0-9.]+' | \
  awk '{ sum += $1; n++ } END { if (n > 0) print "Avg:", sum / n, "ms"; }'
```

#### Response Time Distribution (Production)
```bash
docker logs --since 60m catalyst-prod 2>&1 | \
  grep "duration_ms" | \
  grep -oP 'duration_ms":\K[0-9.]+' | \
  awk '
    { times[NR] = $1; sum += $1; n++ }
    END {
      asort(times)
      print "Count:", n
      print "Min:", times[1], "ms"
      print "P50:", times[int(n*0.5)], "ms"
      print "P95:", times[int(n*0.95)], "ms"
      print "P99:", times[int(n*0.99)], "ms"
      print "Max:", times[n], "ms"
      print "Avg:", sum/n, "ms"
    }
  '
```

#### Side-by-Side Response Time Comparison
```bash
echo "=== PRODUCTION ===" && \
docker logs --since 60m catalyst-prod 2>&1 | \
  grep "duration_ms" | grep -oP 'duration_ms":\K[0-9.]+' | \
  awk '{ sum += $1; n++; if ($1 > max) max = $1; if (min == 0 || $1 < min) min = $1 }
       END { print "Avg:", sum/n "ms", "Min:", min "ms", "Max:", max "ms" }'

echo "=== STAGING ===" && \
docker logs --since 60m catalyst-staging 2>&1 | \
  grep "duration_ms" | grep -oP 'duration_ms":\K[0-9.]+' | \
  awk '{ sum += $1; n++; if ($1 > max) max = $1; if (min == 0 || $1 < min) min = $1 }
       END { print "Avg:", sum/n "ms", "Min:", min "ms", "Max:", max "ms" }'
```

---

### 2. Error Rate Comparison

#### Count Errors
```bash
echo "Production errors:"
docker logs --since 60m catalyst-prod 2>&1 | grep -c '"level":"error"'

echo "Staging errors:"
docker logs --since 60m catalyst-staging 2>&1 | grep -c '"level":"error"'
```

#### Error Rate Percentage
```bash
# Production
PROD_TOTAL=$(docker logs --since 60m catalyst-prod 2>&1 | grep -c "HTTP request")
PROD_ERRORS=$(docker logs --since 60m catalyst-prod 2>&1 | grep -c '"level":"error"')
echo "Production: $PROD_ERRORS / $PROD_TOTAL = $(echo "scale=2; ($PROD_ERRORS / $PROD_TOTAL) * 100" | bc)%"

# Staging
STAGE_TOTAL=$(docker logs --since 60m catalyst-staging 2>&1 | grep -c "HTTP request")
STAGE_ERRORS=$(docker logs --since 60m catalyst-staging 2>&1 | grep -c '"level":"error"')
echo "Staging: $STAGE_ERRORS / $STAGE_TOTAL = $(echo "scale=2; ($STAGE_ERRORS / $STAGE_TOTAL) * 100" | bc)%"
```

#### View Recent Errors
```bash
# Production errors
echo "=== PRODUCTION ERRORS ==="
docker logs --since 60m catalyst-prod 2>&1 | grep '"level":"error"' | tail -5

# Staging errors
echo "=== STAGING ERRORS ==="
docker logs --since 60m catalyst-staging 2>&1 | grep '"level":"error"' | tail -5
```

---

### 3. Traffic Distribution

#### Request Count by Backend
```bash
# From nginx logs
echo "Production requests:"
grep "backend=prod" /opt/catalyst/nginx-logs/access.log | wc -l

echo "Staging requests:"
grep "backend=staging" /opt/catalyst/nginx-logs/access.log | wc -l
```

#### Traffic Split Percentage
```bash
PROD=$(grep "backend=prod" /opt/catalyst/nginx-logs/access.log | wc -l)
STAGE=$(grep "backend=staging" /opt/catalyst/nginx-logs/access.log | wc -l)
TOTAL=$((PROD + STAGE))

echo "Production: $PROD ($((PROD * 100 / TOTAL))%)"
echo "Staging: $STAGE ($((STAGE * 100 / TOTAL))%)"
echo "Total: $TOTAL"
```

#### Test Traffic Distribution
```bash
# Send 100 test requests
for i in {1..100}; do
  curl -s -I https://catalyst.springwire.ai/health 2>&1 | grep X-Backend
done | sort | uniq -c

# Expected output:
#  95 X-Backend: prod
#   5 X-Backend: staging
```

---

### 4. Resource Usage Comparison

#### Real-Time Stats
```bash
docker stats catalyst-prod catalyst-staging
```

#### CPU Usage Over Time
```bash
# Monitor for 60 seconds
docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemPerc}}" \
  catalyst-prod catalyst-staging
```

#### Memory Usage
```bash
echo "Production:"
docker stats --no-stream catalyst-prod | awk 'NR==2 {print $4}'

echo "Staging:"
docker stats --no-stream catalyst-staging | awk 'NR==2 {print $4}'
```

---

### 5. Auction Metrics Comparison

#### Auction Success Rate
```bash
# Production
PROD_AUCTIONS=$(docker logs --since 60m catalyst-prod 2>&1 | grep -c "Auction complete")
PROD_REQUESTS=$(docker logs --since 60m catalyst-prod 2>&1 | grep -c "/openrtb2/auction")
echo "Production: $PROD_AUCTIONS / $PROD_REQUESTS = $((PROD_AUCTIONS * 100 / PROD_REQUESTS))%"

# Staging
STAGE_AUCTIONS=$(docker logs --since 60m catalyst-staging 2>&1 | grep -c "Auction complete")
STAGE_REQUESTS=$(docker logs --since 60m catalyst-staging 2>&1 | grep -c "/openrtb2/auction")
echo "Staging: $STAGE_AUCTIONS / $STAGE_REQUESTS = $((STAGE_AUCTIONS * 100 / STAGE_REQUESTS))%"
```

#### Average Bid Count
```bash
# Production
docker logs --since 60m catalyst-prod 2>&1 | \
  grep "bidder_count" | \
  grep -oP 'bidder_count":\K[0-9]+' | \
  awk '{ sum += $1; n++ } END { print "Avg bidders:", sum/n }'

# Staging
docker logs --since 60m catalyst-staging 2>&1 | \
  grep "bidder_count" | \
  grep -oP 'bidder_count":\K[0-9]+' | \
  awk '{ sum += $1; n++ } END { print "Avg bidders:", sum/n }'
```

---

### 6. IVT Detection Comparison

#### IVT Detection Count
```bash
echo "Production IVT detected:"
docker logs --since 60m catalyst-prod 2>&1 | grep -c "IVT detected"

echo "Staging IVT detected:"
docker logs --since 60m catalyst-staging 2>&1 | grep -c "IVT detected"
```

#### IVT Block Rate (if blocking enabled)
```bash
# Production
PROD_IVT=$(docker logs --since 60m catalyst-prod 2>&1 | grep -c "IVT detected")
PROD_BLOCKED=$(docker logs --since 60m catalyst-prod 2>&1 | grep -c "Request blocked")
echo "Production: $PROD_BLOCKED / $PROD_IVT blocked"

# Staging
STAGE_IVT=$(docker logs --since 60m catalyst-staging 2>&1 | grep -c "IVT detected")
STAGE_BLOCKED=$(docker logs --since 60m catalyst-staging 2>&1 | grep -c "Request blocked")
echo "Staging: $STAGE_BLOCKED / $STAGE_IVT blocked"
```

---

### 7. Redis Performance

#### Memory Usage
```bash
echo "Production Redis:"
docker exec redis-prod redis-cli INFO memory | grep used_memory_human

echo "Staging Redis:"
docker exec redis-staging redis-cli INFO memory | grep used_memory_human
```

#### Connection Count
```bash
echo "Production Redis connections:"
docker exec redis-prod redis-cli INFO clients | grep connected_clients

echo "Staging Redis connections:"
docker exec redis-staging redis-cli INFO clients | grep connected_clients
```

#### Key Count
```bash
echo "Production Redis keys:"
docker exec redis-prod redis-cli DBSIZE

echo "Staging Redis keys:"
docker exec redis-staging redis-cli DBSIZE
```

---

## Real-Time Monitoring

### Watch Logs Side-by-Side

#### Using tmux (Recommended)
```bash
# Install tmux if not available
sudo apt install tmux -y

# Start split screen monitoring
tmux new-session -d -s catalyst
tmux split-window -v
tmux select-pane -t 0
tmux send-keys "docker logs -f catalyst-prod" C-m
tmux select-pane -t 1
tmux send-keys "docker logs -f catalyst-staging" C-m
tmux attach -t catalyst

# Exit with: Ctrl+B then D
```

#### Using screen
```bash
# Production
screen -S prod -dm docker logs -f catalyst-prod

# Staging
screen -S staging -dm docker logs -f catalyst-staging

# Attach to view
screen -r prod    # View production
# Ctrl+A then D to detach
screen -r staging # View staging
```

### Continuous Stats Monitoring
```bash
watch -n 5 'docker stats --no-stream catalyst-prod catalyst-staging'
```

---

## Nginx Access Log Analysis

### Response Time Distribution by Backend
```bash
# Production requests
cat /opt/catalyst/nginx-logs/access.log | \
  grep "backend=prod" | \
  awk '{print $NF}' | \
  cut -d= -f2 | \
  awk '{ sum += $1; n++; if ($1 > max) max = $1 }
       END { print "Prod: Avg:", sum/n, "Max:", max }'

# Staging requests
cat /opt/catalyst/nginx-logs/access.log | \
  grep "backend=staging" | \
  awk '{print $NF}' | \
  cut -d= -f2 | \
  awk '{ sum += $1; n++; if ($1 > max) max = $1 }
       END { print "Stage: Avg:", sum/n, "Max:", max }'
```

### Status Code Distribution
```bash
echo "Production status codes:"
grep "backend=prod" /opt/catalyst/nginx-logs/access.log | \
  awk '{print $9}' | sort | uniq -c

echo "Staging status codes:"
grep "backend=staging" /opt/catalyst/nginx-logs/access.log | \
  awk '{print $9}' | sort | uniq -c
```

---

## Decision Criteria for Rollout

### ✅ Safe to Rollout When:

1. **Response Time**: Staging within ±10% of production
2. **Error Rate**: Staging ≤ production error rate
3. **Auction Success**: Similar success rates (±5%)
4. **No Critical Errors**: No new critical errors in staging
5. **Stable for 24+ hours**: Metrics consistent over time

### ⚠️ Need Investigation When:

1. **Response Time**: Staging >20% slower
2. **Error Rate**: Staging >50% higher error rate
3. **High Resource Usage**: Staging using >2x memory
4. **New Errors**: Seeing error types not in production

### 🛑 Rollback Immediately When:

1. **Critical Errors**: Database connection failures
2. **Memory Leak**: Memory usage constantly growing
3. **High Error Rate**: >5% error rate in staging
4. **Complete Failure**: Staging health checks failing

---

## Automated Monitoring Script

### Run Comparison Every 5 Minutes

Create cron job:
```bash
# Edit crontab
crontab -e

# Add line:
*/5 * * * * /opt/catalyst/compare-performance.sh >> /opt/catalyst/monitoring.log 2>&1
```

### Email Alerts on Issues

```bash
# Install mail utilities
sudo apt install mailutils -y

# Create alert script
cat > /opt/catalyst/alert-if-issues.sh << 'EOF'
#!/bin/bash
OUTPUT=$(/opt/catalyst/compare-performance.sh)

if echo "$OUTPUT" | grep -q "⚠"; then
    echo "$OUTPUT" | mail -s "Catalyst Staging Issue Detected" your@email.com
fi
EOF

chmod +x /opt/catalyst/alert-if-issues.sh

# Add to cron (every 10 minutes)
*/10 * * * * /opt/catalyst/alert-if-issues.sh
```

---

## Long-Term Metrics

### Export Metrics to CSV

```bash
# Create CSV with timestamp
cat > /opt/catalyst/export-metrics.sh << 'EOF'
#!/bin/bash
TIMESTAMP=$(date +"%Y-%m-%d %H:%M:%S")

PROD_AVG=$(docker logs --since 60m catalyst-prod 2>&1 | \
  grep "duration_ms" | grep -oP 'duration_ms":\K[0-9.]+' | \
  awk '{ sum += $1; n++ } END { if (n > 0) print sum / n; else print "0" }')

STAGE_AVG=$(docker logs --since 60m catalyst-staging 2>&1 | \
  grep "duration_ms" | grep -oP 'duration_ms":\K[0-9.]+' | \
  awk '{ sum += $1; n++ } END { if (n > 0) print sum / n; else print "0" }')

PROD_ERRORS=$(docker logs --since 60m catalyst-prod 2>&1 | grep -c '"level":"error"')
STAGE_ERRORS=$(docker logs --since 60m catalyst-staging 2>&1 | grep -c '"level":"error"')

echo "$TIMESTAMP,$PROD_AVG,$STAGE_AVG,$PROD_ERRORS,$STAGE_ERRORS" >> /opt/catalyst/metrics.csv
EOF

chmod +x /opt/catalyst/export-metrics.sh

# Run every hour
0 * * * * /opt/catalyst/export-metrics.sh
```

View trends:
```bash
# Add header
echo "timestamp,prod_avg_ms,stage_avg_ms,prod_errors,stage_errors" > /tmp/metrics-with-header.csv
cat /opt/catalyst/metrics.csv >> /tmp/metrics-with-header.csv

# View
column -t -s, /tmp/metrics-with-header.csv
```

---

## Prometheus Metrics (Optional)

If you add Prometheus to Catalyst:

```bash
# Compare metrics
curl -s http://localhost:8000/metrics | grep catalyst_auction_duration
curl -s http://localhost:8001/metrics | grep catalyst_auction_duration
```

---

## Summary

### Quick Commands

```bash
# Run comparison tool
./compare-performance.sh

# Watch real-time stats
docker stats catalyst-prod catalyst-staging

# Compare response times
docker logs --since 60m catalyst-prod 2>&1 | grep duration_ms | tail -20
docker logs --since 60m catalyst-staging 2>&1 | grep duration_ms | tail -20

# Check error counts
docker logs --since 60m catalyst-prod 2>&1 | grep -c '"level":"error"'
docker logs --since 60m catalyst-staging 2>&1 | grep -c '"level":"error"'

# Test traffic split
curl -I https://catalyst.springwire.ai/health | grep X-Backend
```

### Decision Flow

```
Run comparison tool
        ↓
Check recommendations
        ↓
    Issues found?
    ├─ No → Monitor for 24h → Increase staging % or rollout
    └─ Yes → Investigate logs → Fix issue → Re-test
```

---

**Last Updated**: 2025-01-13
**Tool**: compare-performance.sh
**Deployment**: catalyst.springwire.ai
