# Catalyst MAI Publisher Integration - Deployment Guide

## Overview

This guide covers the deployment process for the Catalyst server-side bidding integration with MAI Publisher.

## Architecture Summary

- **Client SDK**: `catalyst-sdk.js` (< 50KB gzipped)
- **API Endpoint**: `POST /v1/bid` (JSON request/response)
- **Server Timeout**: 2500ms internal processing
- **Client Timeout**: 2800ms (includes network overhead)

## Pre-Deployment Checklist

### 1. Code Review
- [ ] Review `internal/endpoints/catalyst_bid_handler.go`
- [ ] Review `assets/catalyst-sdk.js`
- [ ] Review `cmd/server/server.go` endpoint registration
- [ ] Verify all tests pass: `go test ./tests/catalyst_*`

### 2. Configuration
- [ ] Verify `PBS_HOST_URL` environment variable is set
- [ ] Verify CORS is enabled (default: all origins)
- [ ] Verify exchange timeout supports 2500ms
- [ ] Verify rate limiting is configured appropriately

### 3. Infrastructure
- [ ] CDN configuration for SDK delivery
- [ ] Load balancer health checks configured
- [ ] Monitoring dashboards created
- [ ] Alert rules configured

### 4. Testing
- [ ] Unit tests pass: `go test ./tests/catalyst_bid_test.go`
- [ ] Integration tests pass: `go test ./tests/catalyst_integration_test.go`
- [ ] Load tests complete: 100 req/s sustained
- [ ] Browser SDK test: Open `tests/catalyst_sdk_test.html`

## Staging Deployment

### Step 1: Build and Test

```bash
# Build the server
go build -o catalyst-server cmd/server/main.go

# Run unit tests
go test -v ./tests/catalyst_bid_test.go

# Run integration tests
go test -v ./tests/catalyst_integration_test.go

# Run load tests
go test -v -run=TestCatalystIntegration_HighLoad ./tests/
```

### Step 2: Deploy to Staging

```bash
# Set environment variables
export PBS_PORT=8000
export PBS_HOST_URL=https://staging-ads.thenexusengine.com
export PBS_TIMEOUT=2500ms

# Start server
./catalyst-server
```

### Step 3: Upload SDK to Staging CDN

```bash
# Upload SDK to CDN
aws s3 cp assets/catalyst-sdk.js s3://staging-cdn/assets/catalyst-sdk.js \
  --content-type "application/javascript" \
  --cache-control "public, max-age=3600" \
  --content-encoding gzip

# Verify SDK is accessible
curl https://staging-cdn.thenexusengine.com/assets/catalyst-sdk.js
```

### Step 4: Verify Staging Endpoints

```bash
# Test health endpoint
curl https://staging-ads.thenexusengine.com/health

# Test SDK endpoint
curl https://staging-ads.thenexusengine.com/assets/catalyst-sdk.js

# Test bid endpoint
curl -X POST https://staging-ads.thenexusengine.com/v1/bid \
  -H "Content-Type: application/json" \
  -d '{
    "accountId": "staging-test",
    "timeout": 2800,
    "slots": [{
      "divId": "test-slot",
      "sizes": [[728, 90]]
    }]
  }'
```

### Step 5: Run Staging Tests

```bash
# Run browser test
open tests/catalyst_sdk_test.html

# Expected results:
# - SDK loads in < 500ms
# - All 5 tests pass
# - Response times < 2500ms
# - biddersReady() callback fires
```

### Step 6: MAI Publisher Staging Integration

1. Provide MAI Publisher with staging URLs:
   - SDK URL: `https://staging-cdn.thenexusengine.com/assets/catalyst-sdk.js`
   - API Endpoint: `https://staging-ads.thenexusengine.com/v1/bid`
   - Test Account ID: `mai-staging-test`

2. Coordinate integration testing session
3. Monitor staging metrics during testing
4. Address any issues found

## Production Deployment

### Prerequisites

- [ ] Staging deployment successful
- [ ] MAI Publisher staging integration successful
- [ ] All tests passing
- [ ] Performance metrics meet SLA targets
- [ ] Monitoring and alerts configured

### Step 1: Production Build

```bash
# Build with production optimizations
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags="-w -s" \
  -o catalyst-server \
  cmd/server/main.go

# Verify binary
./catalyst-server --version
```

### Step 2: Deploy to Production

```bash
# Set production environment variables
export PBS_PORT=8000
export PBS_HOST_URL=https://ads.thenexusengine.com
export PBS_TIMEOUT=2500ms
export PBS_DATABASE_HOST=prod-db.thenexusengine.com
export PBS_REDIS_URL=redis://prod-redis.thenexusengine.com:6379

# Deploy using your deployment tool (e.g., Kubernetes, ECS, etc.)
kubectl apply -f k8s/catalyst-deployment.yaml

# Verify deployment
kubectl get pods -l app=catalyst-server
kubectl logs -l app=catalyst-server --tail=50
```

### Step 3: Upload SDK to Production CDN

```bash
# Gzip the SDK
gzip -k -f assets/catalyst-sdk.js

# Upload to production CDN
aws s3 cp assets/catalyst-sdk.js.gz s3://prod-cdn/assets/catalyst-sdk.js \
  --content-type "application/javascript" \
  --cache-control "public, max-age=3600" \
  --content-encoding gzip

# Invalidate CDN cache
aws cloudfront create-invalidation \
  --distribution-id EXAMPLEID \
  --paths "/assets/catalyst-sdk.js"

# Verify CDN
curl -I https://cdn.thenexusengine.com/assets/catalyst-sdk.js
```

### Step 4: Verify Production Endpoints

```bash
# Health check
curl https://ads.thenexusengine.com/health

# Ready check
curl https://ads.thenexusengine.com/health/ready

# SDK endpoint
curl -I https://cdn.thenexusengine.com/assets/catalyst-sdk.js

# Bid endpoint (with test request)
curl -X POST https://ads.thenexusengine.com/v1/bid \
  -H "Content-Type: application/json" \
  -d '{
    "accountId": "production-smoke-test",
    "timeout": 2800,
    "slots": [{
      "divId": "smoke-test-slot",
      "sizes": [[728, 90]]
    }]
  }'
```

### Step 5: Production Smoke Tests

```bash
# Run load test (light)
ab -n 100 -c 10 -p bid_request.json \
  -T application/json \
  https://ads.thenexusengine.com/v1/bid

# Expected results:
# - 0% failed requests
# - P95 latency < 2500ms
# - No 5xx errors
```

### Step 6: Monitor Production Metrics

Access Prometheus metrics at: `https://ads.thenexusengine.com/metrics`

Key metrics to monitor:

```promql
# Request rate
rate(catalyst_bid_requests_total[5m])

# Error rate
rate(catalyst_bid_requests_total{status="error"}[5m]) / rate(catalyst_bid_requests_total[5m])

# P95 latency
histogram_quantile(0.95, rate(catalyst_bid_latency_seconds_bucket[5m]))

# Timeout rate
rate(catalyst_bid_timeouts_total[5m]) / rate(catalyst_bid_requests_total[5m])
```

### Step 7: MAI Publisher Production Integration

1. Provide MAI Publisher with production URLs:
   - SDK URL: `https://cdn.thenexusengine.com/assets/catalyst-sdk.js`
   - API Endpoint: `https://ads.thenexusengine.com/v1/bid`
   - Production Account ID: `mai-publisher-12345`

2. MAI Publisher integrates Catalyst bidder into their wrapper
3. Monitor metrics during rollout
4. Confirm coordination callback working correctly

## Monitoring and Alerting

### Key Metrics

1. **Request Rate**
   - Metric: `catalyst_bid_requests_total`
   - Alert: < 10 req/min (service down)

2. **Error Rate**
   - Metric: `catalyst_bid_requests_total{status="error"}`
   - Alert: > 1% (exceeds SLA)

3. **Latency**
   - Metric: `catalyst_bid_latency_seconds`
   - Alert: P95 > 2500ms (exceeds SLA)

4. **Timeout Rate**
   - Metric: `catalyst_bid_timeouts_total`
   - Alert: > 5% (exceeds SLA)

5. **SDK Load Time**
   - Monitor via browser analytics
   - Alert: P95 > 500ms

### Grafana Dashboards

Create dashboards for:

1. **Catalyst Overview**
   - Request rate
   - Error rate
   - Latency percentiles
   - Timeout rate

2. **Catalyst Details**
   - Per-account metrics
   - Slot distribution
   - Bid fill rates
   - Revenue tracking

3. **SLA Compliance**
   - 99.9% uptime
   - < 1% error rate
   - < 2500ms P95 latency
   - < 5% timeout rate

### Alert Rules

```yaml
# Prometheus alert rules
groups:
  - name: catalyst
    rules:
      - alert: CatalystHighErrorRate
        expr: rate(catalyst_bid_requests_total{status="error"}[5m]) / rate(catalyst_bid_requests_total[5m]) > 0.01
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Catalyst error rate exceeds 1%"

      - alert: CatalystHighLatency
        expr: histogram_quantile(0.95, rate(catalyst_bid_latency_seconds_bucket[5m])) > 2.5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Catalyst P95 latency exceeds 2500ms"

      - alert: CatalystHighTimeoutRate
        expr: rate(catalyst_bid_timeouts_total[5m]) / rate(catalyst_bid_requests_total[5m]) > 0.05
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Catalyst timeout rate exceeds 5%"

      - alert: CatalystServiceDown
        expr: rate(catalyst_bid_requests_total[5m]) < 0.16
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Catalyst service appears down (< 10 req/min)"
```

## Rollback Procedure

If issues are encountered in production:

### Step 1: Assess Impact

```bash
# Check error rate
curl -s https://ads.thenexusengine.com/metrics | grep catalyst_bid_requests_total

# Check recent logs
kubectl logs -l app=catalyst-server --tail=100

# Check health
curl https://ads.thenexusengine.com/health/ready
```

### Step 2: Disable MAI Publisher Integration

Contact MAI Publisher to:
1. Remove Catalyst from `enabled_bidders` array
2. Stop calling `catalyst.requestBids()`
3. Confirm traffic has stopped

### Step 3: Investigate Root Cause

```bash
# Check server logs
kubectl logs -l app=catalyst-server --tail=1000

# Check metrics
curl https://ads.thenexusengine.com/admin/circuit-breaker

# Check database
# Check Redis
# Check bidder adapters
```

### Step 4: Apply Fix

```bash
# Deploy hotfix
kubectl set image deployment/catalyst-server \
  catalyst-server=catalyst-server:v1.0.1-hotfix

# Verify deployment
kubectl rollout status deployment/catalyst-server
```

### Step 5: Re-enable Integration

After verifying fix:
1. Notify MAI Publisher
2. Re-enable Catalyst bidder
3. Monitor metrics closely
4. Confirm SLA compliance

## Scaling

### Horizontal Scaling

```bash
# Scale up replicas
kubectl scale deployment catalyst-server --replicas=5

# Auto-scaling based on CPU
kubectl autoscale deployment catalyst-server \
  --min=3 --max=10 --cpu-percent=70
```

### Performance Optimization

1. **Connection Pooling**
   - Default: 100 connections per bidder
   - Adjust: `PBS_HTTP_CLIENT_MAX_CONNECTIONS_PER_HOST`

2. **Circuit Breakers**
   - Default: 5 failures in 10s triggers open
   - Monitor: `/admin/circuit-breaker`

3. **Caching**
   - Redis for bid response caching (optional)
   - CDN for SDK (1 hour cache)

4. **Database**
   - Connection pool: 10-50 connections
   - Query timeout: 1000ms

## Troubleshooting

### High Latency

**Symptoms**: P95 > 2500ms

**Checks**:
```bash
# Check bidder latencies
curl https://ads.thenexusengine.com/admin/circuit-breaker

# Check database
curl https://ads.thenexusengine.com/health/ready
```

**Solutions**:
- Increase timeout for slow bidders
- Disable slow bidders
- Scale horizontally
- Optimize database queries

### High Error Rate

**Symptoms**: Error rate > 1%

**Checks**:
```bash
# Check logs
kubectl logs -l app=catalyst-server --tail=100 | grep ERROR

# Check metrics
curl https://ads.thenexusengine.com/metrics | grep error
```

**Solutions**:
- Check bidder configurations
- Verify database connectivity
- Check Redis connectivity
- Review recent code changes

### SDK Load Failures

**Symptoms**: SDK not loading in browser

**Checks**:
```bash
# Check CDN
curl -I https://cdn.thenexusengine.com/assets/catalyst-sdk.js

# Check CORS
curl -I -H "Origin: https://example.com" \
  https://cdn.thenexusengine.com/assets/catalyst-sdk.js
```

**Solutions**:
- Verify CDN cache
- Check CORS headers
- Invalidate CDN cache
- Re-upload SDK

## Success Metrics

### Day 1-7 (Initial Launch)
- No critical errors
- < 1% error rate
- P95 latency < 2500ms
- Successful MAI Publisher integration

### Week 2-4 (Stabilization)
- 99.9% uptime achieved
- Performance optimizations deployed
- Monitoring dashboards refined
- Alert noise reduced

### Month 2+ (Optimization)
- P95 latency < 2000ms
- Error rate < 0.5%
- Revenue tracking enabled
- A/B testing framework deployed

## Support Contacts

- **Technical Issues**: tech-support@thenexusengine.com
- **MAI Publisher Integration**: mai-support@thenexusengine.com
- **On-Call**: pagerduty@thenexusengine.com
- **Slack Channel**: #catalyst-integration

## Additional Resources

- [Integration Specification](./BB_NEXUS-ENGINE-INTEGRATION-SPEC.md)
- [API Documentation](./CATALYST_API_REFERENCE.md)
- [Grafana Dashboards](https://grafana.thenexusengine.com/d/catalyst)
- [Runbook](https://wiki.thenexusengine.com/catalyst-runbook)
