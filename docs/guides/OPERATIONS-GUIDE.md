# Operations Guide - TNE Catalyst

Comprehensive guide for operating, monitoring, and troubleshooting TNE Catalyst in production.

## Table of Contents
1. [Server Logging](#server-logging)
2. [Failure Scenarios](#failure-scenarios)
3. [Troubleshooting Runbook](#troubleshooting-runbook)
4. [Monitoring & Alerts](#monitoring--alerts)
5. [Performance Tuning](#performance-tuning)
6. [Security Operations](#security-operations)

---

## Server Logging

### Logging Framework

TNE Catalyst uses **zerolog** for structured JSON logging with the following levels:

| Level | Use Case | Examples |
|-------|----------|----------|
| **DEBUG** | Development, detailed tracing | Request/response bodies, detailed timing |
| **INFO** | Normal operations | Startup, config loaded, request processed |
| **WARN** | Recoverable issues | Redis unavailable (fallback used), slow bidder |
| **ERROR** | Service degradation | Database timeout, bidder error, validation failure |
| **FATAL** | Unrecoverable failures | Cannot bind port, critical service unavailable |

### Log Format

All logs are structured JSON with standard fields:

```json
{
  "level": "error",
  "time": "2026-01-16T18:00:00Z",
  "message": "Bidder timeout",
  "request_id": "req_abc123",
  "publisher_id": "pub_xyz",
  "bidder_code": "rubicon",
  "duration_ms": 150,
  "error": "context deadline exceeded"
}
```

### Standard Fields

Every log entry includes:
- `level`: Log level (debug/info/warn/error/fatal)
- `time`: ISO 8601 timestamp
- `message`: Human-readable message
- `request_id`: Unique request identifier (for tracing)

Context-specific fields:
- `publisher_id`: Publisher making request
- `bidder_code`: Bidder adapter involved
- `duration_ms`: Operation duration
- `error`: Error message (if applicable)
- `ip`: Client IP (anonymized if GDPR applies)
- `user_agent`: Client user agent
- `status_code`: HTTP response code

### Log Locations

**Docker Deployment:**
```bash
# Container logs
docker logs catalyst

# Follow logs
docker logs -f catalyst

# Last 100 lines
docker logs --tail=100 catalyst
```

**Direct Deployment:**
```bash
# Standard output (if using systemd)
journalctl -u catalyst -f

# File-based logging (if configured)
tail -f /var/log/catalyst/app.log
```

**Kubernetes:**
```bash
kubectl logs -f deployment/catalyst -n production
kubectl logs -f deployment/catalyst -n production --previous  # Previous crash
```

### Log Queries

**Find errors in last hour:**
```bash
docker logs catalyst --since 1h | jq 'select(.level=="error")'
```

**Track specific request:**
```bash
docker logs catalyst | jq 'select(.request_id=="req_abc123")'
```

**Count errors by bidder:**
```bash
docker logs catalyst | jq -r 'select(.level=="error" and .bidder_code) | .bidder_code' | sort | uniq -c
```

**Monitor slow requests (>200ms):**
```bash
docker logs catalyst | jq 'select(.duration_ms > 200)'
```

---

## Failure Scenarios

### 1. Database (PostgreSQL) Failures

#### Scenario: Connection Refused
```json
{
  "level": "fatal",
  "message": "Failed to connect to PostgreSQL",
  "error": "dial tcp 127.0.0.1:5432: connect: connection refused",
  "db_host": "localhost",
  "db_port": "5432"
}
```

**Causes:**
- PostgreSQL not running
- Wrong credentials (DB_PASSWORD)
- Network connectivity issue
- Firewall blocking port 5432

**Impact:** Server won't start (FATAL error)

**Resolution:**
1. Check PostgreSQL status: `systemctl status postgresql`
2. Verify credentials in environment variables
3. Test connection: `psql -h localhost -U catalyst -d catalyst`
4. Check firewall: `sudo ufw status`
5. Review PostgreSQL logs: `tail -f /var/log/postgresql/postgresql-*.log`

#### Scenario: Query Timeout
```json
{
  "level": "error",
  "message": "Database query timeout",
  "error": "pq: canceling statement due to user request",
  "query": "SELECT * FROM publishers WHERE id = $1",
  "duration_ms": 5000
}
```

**Causes:**
- Missing database indexes
- Large table scans
- Database under heavy load
- Connection pool exhausted

**Impact:** Request fails with 500 error

**Resolution:**
1. Check slow query log
2. Add indexes: `CREATE INDEX idx_publishers_id ON publishers(id);`
3. Increase connection pool: `DB_MAX_OPEN_CONNS=50`
4. Monitor active connections: `SELECT count(*) FROM pg_stat_activity;`

#### Scenario: Connection Pool Exhausted
```json
{
  "level": "warn",
  "message": "Connection pool wait timeout",
  "error": "timed out waiting for connection",
  "pool_size": 25,
  "active_connections": 25
}
```

**Causes:**
- Connection leak (connections not closed)
- Traffic spike
- Pool size too small

**Impact:** Requests queued, increased latency

**Resolution:**
1. Check for connection leaks in code
2. Increase pool: `DB_MAX_OPEN_CONNS=50`
3. Reduce idle timeout: `DB_CONN_MAX_LIFETIME=300s`
4. Monitor: `SELECT count(*), state FROM pg_stat_activity GROUP BY state;`

---

### 2. Redis Failures

#### Scenario: Redis Unavailable (CRITICAL)
```json
{
  "level": "error",
  "message": "Redis connection failed, falling back to PostgreSQL",
  "error": "dial tcp 127.0.0.1:6379: connect: connection refused",
  "fallback": "database"
}
```

**Causes:**
- Redis not running
- Network issue
- Redis out of memory
- Wrong REDIS_HOST/PORT

**Impact:**
- ⚠️ CRITICAL: Without fallback, ALL publisher auth fails
- Performance degradation (DB slower than Redis)

**Current State:** ❌ No fallback implemented - SINGLE POINT OF FAILURE

**Mitigation Needed:**
```go
// Recommended: Add fallback in publisher_auth.go
func (m *PublisherAuthMiddleware) getPublisher(id string) (*Publisher, error) {
    // Try Redis first
    pub, err := m.redis.Get(ctx, "publisher:"+id)
    if err != nil {
        log.Warn().Err(err).Msg("Redis unavailable, falling back to database")
        // Fallback to PostgreSQL
        return m.db.GetPublisher(id)
    }
    return pub, nil
}
```

**Immediate Resolution:**
1. Check Redis status: `redis-cli ping`
2. Restart Redis: `systemctl restart redis`
3. Check memory: `redis-cli INFO memory`
4. If critical, disable Redis cache temporarily: `PUBLISHER_AUTH_USE_REDIS=false`

#### Scenario: Redis Out of Memory
```json
{
  "level": "error",
  "message": "Redis OOM error",
  "error": "OOM command not allowed when used memory > 'maxmemory'",
  "memory_used": "4.2GB",
  "maxmemory": "4GB"
}
```

**Causes:**
- Memory limit too low
- Memory leak (keys not expiring)
- TTL not set on keys

**Impact:** Cannot write to Redis, reads still work

**Resolution:**
1. Check memory: `redis-cli INFO memory`
2. Increase maxmemory: `redis-cli CONFIG SET maxmemory 8gb`
3. Check key TTLs: `redis-cli --scan --pattern '*' | xargs -L 1 redis-cli TTL`
4. Flush old data: `redis-cli FLUSHDB` (⚠️ USE WITH CAUTION)
5. Verify TTL settings: `REDIS_AUCTION_TTL=300`

---

### 3. IDR (Intelligent Demand Router) Failures

#### Scenario: IDR Timeout
```json
{
  "level": "warn",
  "message": "IDR request timeout, proceeding with all bidders",
  "error": "context deadline exceeded",
  "duration_ms": 150,
  "idr_url": "https://idr.thenexusengine.com"
}
```

**Causes:**
- IDR service slow/overloaded
- Network latency
- IDR_TIMEOUT_MS too aggressive

**Impact:** Falls back to all bidders (no routing optimization)

**Resolution:**
1. Check IDR circuit breaker state: `curl localhost:8000/admin/circuit-breaker`
2. Increase timeout: `IDR_TIMEOUT_MS=200`
3. Monitor IDR health: `curl https://idr.thenexusengine.com/health`
4. Temporary disable if critical: `IDR_ENABLED=false`

#### Scenario: Circuit Breaker Open
```json
{
  "level": "warn",
  "message": "IDR circuit breaker OPEN, skipping IDR call",
  "failure_rate": "75%",
  "total_requests": 50,
  "failed_requests": 38
}
```

**Causes:**
- IDR service degraded
- Network issues
- Sustained high error rate

**Impact:** All auctions bypass IDR (performance degradation)

**Resolution:**
1. Check IDR service status
2. Circuit breaker auto-recovers after 30s
3. Monitor recovery: `curl localhost:8000/admin/circuit-breaker`
4. If persistent, investigate IDR logs

---

### 4. External Bidder Failures

#### Scenario: Bidder Timeout
```json
{
  "level": "warn",
  "message": "Bidder request timeout",
  "bidder_code": "rubicon",
  "duration_ms": 150,
  "timeout_ms": 150,
  "endpoint": "https://fastlane.rubiconproject.com/a/api/fastlane.json"
}
```

**Causes:**
- Bidder SSP slow/overloaded
- Network latency
- Timeout too aggressive

**Impact:** Lower bid density, potential revenue loss

**Resolution:**
1. Check bidder timeout: Default 150ms per bidder
2. Increase if needed (but max 200ms to avoid auction delay)
3. Monitor bidder performance metrics
4. ⚠️ **No circuit breaker** - recommended to add

**Recommended Enhancement:**
```go
// Per-bidder circuit breaker (not yet implemented)
// Track failures per bidder over 1-minute sliding window
// Open circuit if >50% failure rate with >10 requests
```

#### Scenario: Bidder Returns Invalid Response
```json
{
  "level": "error",
  "message": "Failed to parse bidder response",
  "bidder_code": "pubmatic",
  "error": "invalid character 'x' looking for beginning of value",
  "response_body": "<html>503 Service Unavailable</html>"
}
```

**Causes:**
- Bidder SSP returned HTML error page
- Bidder API changed format
- Bidder returned 5xx error

**Impact:** Single bidder excluded from auction

**Resolution:**
1. Check bidder endpoint health manually
2. Review bidder adapter code for API changes
3. Contact bidder support if persistent
4. Temporarily disable problematic bidder

---

### 5. Privacy & Compliance Failures

#### Scenario: Invalid GDPR Consent
```json
{
  "level": "warn",
  "message": "Rejecting request due to invalid GDPR consent",
  "consent_string": "CO...(truncated)",
  "error": "TCF string version mismatch",
  "ip": "192.168.1.0",  // anonymized
  "country": "DE",
  "regulation": "GDPR"
}
```

**Causes:**
- Malformed TCF consent string
- Outdated consent string version
- Missing required consent

**Impact:** Request rejected with 400 status code

**Resolution:**
1. Verify consent string format with publisher
2. Check TCF vendor list is up to date
3. If strict mode issue: `PBS_PRIVACY_STRICT_MODE=false` (strips PII instead)
4. Review: [GEO-CONSENT-GUIDE.md](../GEO-CONSENT-GUIDE.md)

#### Scenario: Geo-Location Mismatch
```json
{
  "level": "warn",
  "message": "GDPR enforcement triggered by user.geo (device.geo missing)",
  "device_geo": null,
  "user_geo_country": "FR",
  "regulation_detected": "GDPR"
}
```

**Causes:**
- device.geo not provided in request
- Only user.geo available

**Impact:** GDPR enforcement applied correctly (audit fix prevents bypass)

**Note:** This is EXPECTED behavior after Jan 2026 audit fix. Both device.geo and user.geo are now checked.

---

### 6. Authentication Failures

#### Scenario: Publisher Not Found
```json
{
  "level": "warn",
  "message": "Publisher authentication failed",
  "publisher_id": "unknown_pub",
  "error": "publisher not found",
  "allow_unregistered": false
}
```

**Causes:**
- Publisher not registered in database
- Wrong publisher ID in request header
- Typo in X-Publisher-ID header

**Impact:** Request rejected with 401 status

**Resolution:**
1. Check publisher exists: `docker exec postgres psql -U catalyst -d catalyst -c "SELECT * FROM publishers WHERE id='unknown_pub';"`
2. Register publisher if needed
3. For testing: `PUBLISHER_ALLOW_UNREGISTERED=true`

#### Scenario: Suspicious Publisher ID (Audit Fix)
```json
{
  "level": "warn",
  "message": "Rejecting suspicious X-Publisher-ID (too short)",
  "publisher_id": "abc",
  "minimum_length": 8
}
```

**Causes:**
- Short/invalid publisher ID (possible injection attempt)
- Publisher ID less than 8 characters

**Impact:** Request rejected (security measure)

**Note:** This is security hardening from Jan 2026 audit. Minimum 8 characters required.

---

### 7. Startup Failures

#### Scenario: Port Already in Use
```json
{
  "level": "fatal",
  "message": "Failed to bind server to port",
  "error": "bind: address already in use",
  "port": "8000"
}
```

**Causes:**
- Another process using port 8000
- Previous instance not cleanly shut down

**Resolution:**
```bash
# Find process using port
lsof -i :8000
# OR
netstat -tulpn | grep 8000

# Kill process
kill -9 <PID>

# Or change port
PBS_PORT=8001
```

#### Scenario: Missing Required Configuration
```json
{
  "level": "fatal",
  "message": "Missing required configuration",
  "error": "DB_PASSWORD environment variable not set"
}
```

**Causes:**
- Environment variables not loaded
- .env file not sourced
- Typo in variable name

**Resolution:**
1. Check .env file exists: `ls -la .env`
2. Source environment: `export $(cat .env | xargs)`
3. Verify: `echo $DB_PASSWORD`
4. Review required vars in [README.md](../README.md#configuration)

---

### 8. WAF (ModSecurity) Failures

#### Scenario: Request Blocked by WAF
```json
{
  "level": "warn",
  "message": "WAF blocked request",
  "rule_id": "942100",
  "rule_msg": "SQL Injection Attack Detected via libinjection",
  "uri": "/openrtb2/auction",
  "client_ip": "203.0.113.42"
}
```

**Causes:**
- Malicious request (SQL injection, XSS)
- False positive on legitimate request
- Paranoia level too high

**Impact:** Request blocked with 403

**Resolution:**
1. Review WAF logs: `docker logs modsecurity-waf`
2. If false positive, whitelist rule
3. Adjust paranoia: `PARANOIA_LEVEL=1` (less strict)
4. Review: [deployment/WAF-README.md](../deployment/WAF-README.md)

---

## Troubleshooting Runbook

### High CPU Usage

**Symptoms:**
- CPU usage consistently >80%
- Slow response times
- Requests timing out

**Investigation:**
```bash
# Check CPU usage
top
htop

# Profile Go application
curl http://localhost:8000/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof cpu.prof

# Check for runaway goroutines
curl http://localhost:8000/debug/pprof/goroutine > goroutines.txt
```

**Common Causes:**
1. Too many concurrent bidder requests
2. Goroutine leak
3. Inefficient regex in IVT detection
4. JSON marshaling bottleneck

**Resolution:**
- Limit concurrent bidders
- Fix goroutine leaks
- Optimize hot paths
- Consider caching

### High Memory Usage

**Symptoms:**
- Memory usage growing over time
- OOM kills
- Swap usage increasing

**Investigation:**
```bash
# Check memory
free -h
docker stats catalyst

# Heap profile
curl http://localhost:8000/debug/pprof/heap > heap.prof
go tool pprof -alloc_space heap.prof

# Check for leaks
curl http://localhost:8000/debug/pprof/heap?seconds=30 > heap-live.prof
```

**Common Causes:**
1. Rate limiter map growth (audit fix addressed this)
2. Connection leaks
3. Large request/response bodies cached
4. Goroutine leak holding references

**Resolution:**
- Implement cleanup routines
- Add memory limits
- Review caching strategy
- Fix connection leaks

### Auction Latency Spikes

**Symptoms:**
- P99 latency >300ms
- Timeouts increasing
- User complaints about slowness

**Investigation:**
```bash
# Check logs for slow requests
docker logs catalyst | jq 'select(.duration_ms > 300)'

# Check bidder performance
docker logs catalyst | jq -r 'select(.bidder_code) | "\(.bidder_code): \(.duration_ms)ms"' | sort

# Check database query times
docker logs catalyst | jq 'select(.query and .duration_ms > 100)'
```

**Common Causes:**
1. Slow bidder (no circuit breaker)
2. Database slow query
3. Redis unavailable (no fallback)
4. IDR timeout

**Resolution:**
- Implement bidder circuit breakers (HIGH priority)
- Optimize database queries
- Add Redis fallback (HIGH priority)
- Tune timeouts

---

## Monitoring & Alerts

### Recommended Metrics

**System Metrics:**
- CPU usage (target: <70%)
- Memory usage (target: <80%)
- Disk I/O
- Network bandwidth

**Application Metrics:**
- Request rate (requests/sec)
- Response time (P50, P95, P99)
- Error rate (5xx errors)
- Auction success rate

**Service Health:**
- PostgreSQL connection count
- Redis hit rate
- IDR circuit breaker state
- Bidder timeout rate

**Business Metrics:**
- Auctions per minute
- Bid density (bids per auction)
- Revenue per auction
- Publisher distribution

### Alert Thresholds

| Metric | Warning | Critical |
|--------|---------|----------|
| Error rate | >1% | >5% |
| P99 latency | >300ms | >500ms |
| Redis unavailable | N/A | Any downtime |
| DB connections | >80% pool | >95% pool |
| Bidder timeout rate | >10% | >25% |
| Memory usage | >80% | >90% |

### Prometheus Metrics Endpoint

```bash
# Metrics available at:
curl http://localhost:8000/metrics

# Example metrics:
catalyst_requests_total{method="POST",path="/openrtb2/auction",status="200"} 1250
catalyst_request_duration_seconds{quantile="0.99"} 0.245
catalyst_bidder_timeouts_total{bidder="rubicon"} 15
```

---

## Performance Tuning

### Connection Pools

**PostgreSQL:**
```bash
DB_MAX_OPEN_CONNS=25      # Total connections
DB_MAX_IDLE_CONNS=5       # Idle connections
DB_CONN_MAX_LIFETIME=300s # Connection lifetime
```

**Redis:**
```bash
REDIS_POOL_SIZE=50
REDIS_IDLE_TIMEOUT=300s
REDIS_POOL_TIMEOUT=4s
```

### Timeout Configuration

```bash
# IDR
IDR_TIMEOUT_MS=150        # Fast failure

# Bidders (per-bidder, not configurable via env)
# Default: 150ms per bidder adapter
# Can be overridden in adapter code

# Database
DB_QUERY_TIMEOUT=5s      # Prevent long-running queries
```

### Concurrency Limits

```bash
# HTTP Server
MAX_HEADER_BYTES=1048576   # 1MB
READ_TIMEOUT=5s
WRITE_TIMEOUT=10s

# Request body size
SIZE_LIMITER_MAX_BODY=10485760  # 10MB
```

---

## Security Operations

### Log Sanitization

**PII Handling:**
- IPs anonymized when GDPR applies
- User IDs hashed in logs
- Email addresses never logged
- Full requests only in DEBUG mode

**Sensitive Data:**
- Database passwords masked
- Redis auth tokens masked
- Bidder credentials never logged

### Incident Response

**Security Incident Detected:**
1. Check WAF logs for attack patterns
2. Review authentication failures
3. Check for privilege escalation attempts
4. Review admin endpoint access logs

**Data Breach Response:**
1. Identify scope (which publishers affected)
2. Check privacy compliance logs
3. Review consent string handling
4. Notify affected parties per GDPR Article 33

---

## Quick Reference

### Essential Commands

```bash
# Check all services healthy
curl http://localhost:8000/health

# View real-time logs
docker logs -f catalyst

# Check error rate
docker logs --since 1h catalyst | jq -r '.level' | grep error | wc -l

# Monitor slow requests
docker logs -f catalyst | jq 'select(.duration_ms > 200)'

# Check specific bidder
docker logs catalyst | grep rubicon

# Export logs for analysis
docker logs catalyst > catalyst.log
```

### Emergency Contacts

- **On-Call Engineer:** [Configure in your system]
- **Database Team:** [Configure in your system]
- **Infrastructure Team:** [Configure in your system]
- **Security Team:** [Configure in your system]

---

## Version History

- **v2.0** (2026-01-16): Initial operations guide
  - Documented all failure scenarios
  - Added troubleshooting runbooks
  - Identified critical gaps (Redis fallback, bidder circuit breakers)
