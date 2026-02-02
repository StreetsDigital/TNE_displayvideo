# Traffic Splitting: 95% Production / 5% Staging

## Purpose

Enable **canary deployments** by sending a small percentage (5%) of production traffic to a staging environment for testing new features or configurations before full rollout.

## What Is This?

```
Internet Traffic (100%)
        ↓
ads.thenexusengine.com
        ↓
    Nginx (splits traffic)
    ├─ 95% → Production Catalyst
    └─  5% → Staging Catalyst
```

**Same domain, two backends** - Users don't know which backend serves their request.

---

## Why Use Traffic Splitting?

### Traditional Approach (Risky)
```
1. Test on localhost ✓
2. Deploy to production ✗
3. Hope nothing breaks 🤞
4. If it breaks, everyone affected 💥
```

### Canary Approach (Safe)
```
1. Test on localhost ✓
2. Deploy to staging (5% traffic) ✓
3. Monitor for errors/performance issues ✓
4. If problems: only 5% affected
5. If good: increase to 100%
```

**Use cases:**
- Testing new bidder adapters
- Testing configuration changes
- Testing performance optimizations
- A/B testing different IVT rules

---

## Architecture

### Two Environments on One Server

```
┌─────────────────────────────────────────────────┐
│                Server                            │
│                                                  │
│  ┌──────────┐                                   │
│  │  Nginx   │  (splits traffic)                 │
│  │  :443    │                                   │
│  └────┬─────┘                                   │
│       │                                         │
│       ├─ 95% ─┐          ┌─ 5% ─┐              │
│       │       ▼          ▼       │              │
│   ┌──────────────┐  ┌──────────────┐           │
│   │ Catalyst     │  │ Catalyst     │           │
│   │ Production   │  │ Staging      │           │
│   │ :8000        │  │ :8001        │           │
│   └──────┬───────┘  └──────┬───────┘           │
│          │                  │                   │
│   ┌──────▼───────┐  ┌──────▼───────┐           │
│   │ Redis Prod   │  │ Redis Stage  │           │
│   │ :6379        │  │ :6380        │           │
│   └──────────────┘  └──────────────┘           │
│                                                  │
└─────────────────────────────────────────────────┘
```

**Key Points:**
- ✅ Same server (no extra hardware needed)
- ✅ Same domain (ads.thenexusengine.com)
- ✅ Separate Redis instances (data isolation)
- ✅ Different resource limits (staging gets less)
- ✅ Independent configurations (.env.production vs .env.staging)

---

## Files Involved

### Use These Files for Traffic Splitting

```
/opt/catalyst/
├── docker-compose-split.yml  ← Use instead of docker-compose.yml
├── nginx-split.conf          ← Use instead of nginx.conf
├── .env.production           ← 95% traffic config
├── .env.staging              ← 5% traffic config
└── ssl/                      ← SSL certs (same for both)
```

### Regular Deployment (No Split)

```
/opt/catalyst/
├── docker-compose.yml   ← Regular, single environment
├── nginx.conf           ← No traffic splitting
├── .env.production      ← 100% traffic config
└── ssl/
```

---

## How Traffic Splitting Works

### Algorithm: split_clients

```nginx
split_clients "${remote_addr}${request_uri}" $backend {
    95%     prod;
    *       staging;
}
```

**What this does:**
1. Takes client IP + request URI
2. Hashes them together
3. Based on hash, assigns to prod (95%) or staging (5%)
4. **Same client + same URL = same backend** (sticky)

**Example:**
```
Client IP: 203.0.113.42
Request: /openrtb2/auction

Hash: md5("203.0.113.42/openrtb2/auction") = abc123...
If hash % 100 < 95 → Production
Else → Staging

This client will ALWAYS hit the same backend
```

### Stickiness

**Important**: A single user consistently hits the same backend.

```
User A → Makes 10 requests → All go to Production
User B → Makes 10 requests → All go to Staging
User C → Makes 10 requests → All go to Production
```

**Why this matters:**
- ✅ Consistent experience per user
- ✅ Easier to debug (user either sees prod or staging, not mixed)
- ✅ Simpler A/B testing

---

## Deployment

### Step 1: Create Environment Files

```bash
cd /opt/catalyst

# Create production config (95% traffic)
nano .env.production
# Configure for production (see README-environments.md)

# Create staging config (5% traffic)
nano .env.staging
# Configure for staging (can test new features here)
```

**Key Difference: IVT Blocking**
```bash
# .env.production
IVT_BLOCKING_ENABLED=false  # Safe, monitoring only

# .env.staging (test blocking here first)
IVT_BLOCKING_ENABLED=true   # Test aggressive blocking
```

### Step 2: Start Split Deployment

```bash
# Use split version of docker-compose
docker compose -f docker-compose-split.yml up -d

# Check status
docker compose -f docker-compose-split.yml ps

# Expected output:
# catalyst-prod      Up (healthy)
# catalyst-staging   Up (healthy)
# redis-prod         Up (healthy)
# redis-staging      Up (healthy)
# nginx              Up (healthy)
```

### Step 3: Verify Traffic Split

```bash
# Make 20 requests and check backend header
for i in {1..20}; do
  curl -I https://ads.thenexusengine.com/health 2>&1 | grep X-Backend
done

# Expected output (approximately):
# X-Backend: prod    (appears ~19 times = 95%)
# X-Backend: staging (appears ~1 time = 5%)
```

### Step 4: Monitor Both Backends

```bash
# Production logs (most traffic)
docker compose -f docker-compose-split.yml logs -f catalyst-prod

# Staging logs (5% traffic)
docker compose -f docker-compose-split.yml logs -f catalyst-staging

# Compare error rates
docker compose -f docker-compose-split.yml logs catalyst-prod | grep -i error | wc -l
docker compose -f docker-compose-split.yml logs catalyst-staging | grep -i error | wc -l
```

---

## Monitoring & Validation

### Check Split Status Endpoint

```bash
curl https://ads.thenexusengine.com/admin/split-status

# Returns:
{
  "split_enabled": true,
  "production_percentage": 95,
  "staging_percentage": 5,
  "algorithm": "split_clients"
}
```

### Nginx Access Logs

Logs include backend info:
```
203.0.113.42 - [13/Jan/2025:10:30:15 +0000] "POST /openrtb2/auction" 200
backend=prod rt=0.045
```

**Analyze traffic distribution:**
```bash
# Count requests per backend
grep "backend=prod" /opt/catalyst/nginx-logs/access.log | wc -l
grep "backend=staging" /opt/catalyst/nginx-logs/access.log | wc -l

# Should be approximately 95/5 ratio
```

### Container Resource Usage

```bash
# Check CPU/memory for both
docker stats catalyst-prod catalyst-staging

# Production should use more (95% traffic)
# Staging should use less (5% traffic)
```

### Compare Error Rates

```bash
# Production errors
docker compose -f docker-compose-split.yml logs catalyst-prod | grep "level\":\"error\"" | wc -l

# Staging errors
docker compose -f docker-compose-split.yml logs catalyst-staging | grep "level\":\"error\"" | wc -l

# If staging has significantly more errors, don't roll it out!
```

---

## Adjusting Split Percentage

### Change from 95/5 to 90/10

Edit `nginx-split.conf`:
```nginx
split_clients "${remote_addr}${request_uri}" $backend {
    90%     prod;      # Changed from 95%
    *       staging;   # Now gets 10%
}
```

Reload nginx:
```bash
docker compose -f docker-compose-split.yml exec nginx nginx -s reload
```

### Gradual Rollout Plan

```
Week 1: 95% prod / 5% staging   (test with minimal impact)
Week 2: 90% prod / 10% staging  (increase if no issues)
Week 3: 80% prod / 20% staging  (more confidence)
Week 4: 50% prod / 50% staging  (equal split for A/B)
Week 5: 0% prod / 100% staging  (full rollout)
```

Then **flip the configs**:
```bash
# Make current staging the new production
mv .env.production .env.production.old
mv .env.staging .env.production
```

---

## Switching Between Modes

### From Regular to Split

```bash
# Stop regular deployment
docker compose -f docker-compose.yml down

# Start split deployment
docker compose -f docker-compose-split.yml up -d
```

### From Split to Regular (Rollback)

```bash
# Stop split deployment
docker compose -f docker-compose-split.yml down

# Start regular deployment (100% production)
docker compose -f docker-compose.yml up -d
```

---

## Resource Allocation

### Production Container
```yaml
resources:
  limits:
    cpus: '2.0'
    memory: 4G
  reservations:
    cpus: '0.5'
    memory: 1G
```
**Why**: Gets 95% of traffic, needs full resources

### Staging Container
```yaml
resources:
  limits:
    cpus: '1.0'
    memory: 2G
  reservations:
    cpus: '0.25'
    memory: 512M
```
**Why**: Only gets 5% of traffic, needs less resources

### Total Server Requirements

```
Production:  2 CPU, 4GB RAM
Staging:     1 CPU, 2GB RAM
Redis Prod:  0.5 CPU, 1GB RAM
Redis Stage: 0.25 CPU, 512MB RAM
Nginx:       0.5 CPU, 512MB RAM
─────────────────────────────
Total:       4.25 CPU, 8GB RAM
```

**Minimum Server**: 6 CPU, 12GB RAM (leaves headroom)

---

## Testing Scenarios

### Scenario 1: Test New IVT Rules

```bash
# Production: Monitoring only
# .env.production
IVT_BLOCKING_ENABLED=false

# Staging: Block aggressively (5% of users affected)
# .env.staging
IVT_BLOCKING_ENABLED=true
IVT_ALLOWED_COUNTRIES=US,GB,CA  # Stricter geo filtering

# Monitor staging for false positives
docker compose -f docker-compose-split.yml logs -f catalyst-staging | grep "IVT"
```

### Scenario 2: Test New Bidder Adapter

```bash
# Add new bidder to staging only
# Staging has newbidder adapter enabled
# Production uses existing adapters

# Monitor staging auction logs
docker compose -f docker-compose-split.yml logs -f catalyst-staging | grep "newbidder"

# Check success rate
# If good → Add to production
```

### Scenario 3: Performance Comparison

```bash
# Production: Current config
# Staging: New optimization (e.g., increased Redis timeout)

# Compare latencies
docker compose -f docker-compose-split.yml logs catalyst-prod | grep "duration_ms" | awk '{print $X}' | avg
docker compose -f docker-compose-split.yml logs catalyst-staging | grep "duration_ms" | awk '{print $X}' | avg

# If staging is faster → Roll out to production
```

---

## Troubleshooting

### Problem: Staging getting no traffic

**Check nginx split configuration:**
```bash
docker compose -f docker-compose-split.yml exec nginx cat /etc/nginx/nginx.conf | grep split_clients
```

**Test locally:**
```bash
for i in {1..100}; do
  curl -I https://ads.thenexusengine.com/health 2>&1 | grep X-Backend
done | sort | uniq -c

# Should see roughly:
#  95 X-Backend: prod
#   5 X-Backend: staging
```

### Problem: Staging container unhealthy

**Check logs:**
```bash
docker compose -f docker-compose-split.yml logs catalyst-staging
```

**Check environment:**
```bash
docker compose -f docker-compose-split.yml exec catalyst-staging env | grep PBS_
```

**Common issue**: Wrong .env file loaded

### Problem: Can't tell which backend served request

**Add debug header in nginx:**
```nginx
add_header X-Backend $backend always;
add_header X-Container-ID $hostname always;
```

**Check with:**
```bash
curl -I https://ads.thenexusengine.com/health
```

### Problem: Uneven traffic distribution

**Expected**: Not exactly 95/5 every minute
**Actual**: Over thousands of requests, averages to 95/5

**Check over larger sample:**
```bash
# Last 1000 requests
tail -1000 /opt/catalyst/nginx-logs/access.log | grep backend=prod | wc -l
tail -1000 /opt/catalyst/nginx-logs/access.log | grep backend=staging | wc -l
```

---

## Security Considerations

### Staging Contains Real Traffic

⚠️ **Important**: Staging receives 5% of **real production traffic**

**Implications:**
- Real user data
- Real bid requests
- Real publishers

**Must have:**
- ✅ Same privacy compliance (GDPR, CCPA)
- ✅ Same security headers
- ✅ Same rate limiting
- ✅ Separate Redis (don't pollute production data)

### Data Isolation

```
Production Redis → Production data only
Staging Redis → Staging data only

Never share Redis between environments!
```

---

## When to Use Split vs Regular

### Use Traffic Splitting When:
- ✅ Testing risky changes (IVT rules, new bidders)
- ✅ Validating performance improvements
- ✅ A/B testing configurations
- ✅ Need real traffic for testing

### Use Regular Deployment When:
- ✅ Stable production (no testing needed)
- ✅ Low server resources (can't run both)
- ✅ Testing locally is sufficient
- ✅ Simple configuration changes

---

## Commands Reference

```bash
# Start split deployment
docker compose -f docker-compose-split.yml up -d

# Stop split deployment
docker compose -f docker-compose-split.yml down

# View logs (production)
docker compose -f docker-compose-split.yml logs -f catalyst-prod

# View logs (staging)
docker compose -f docker-compose-split.yml logs -f catalyst-staging

# Check status
docker compose -f docker-compose-split.yml ps

# Restart staging only
docker compose -f docker-compose-split.yml restart catalyst-staging

# Reload nginx config
docker compose -f docker-compose-split.yml exec nginx nginx -s reload

# Test nginx config
docker compose -f docker-compose-split.yml exec nginx nginx -t

# View split status
curl https://ads.thenexusengine.com/admin/split-status
```

---

## Migration Path

### Phase 1: Regular Deployment (Now)
```bash
docker compose -f docker-compose.yml up -d
# 100% production traffic
```

### Phase 2: Enable Splitting (Later)
```bash
docker compose -f docker-compose.yml down
docker compose -f docker-compose-split.yml up -d
# 95% prod / 5% staging
```

### Phase 3: Validate & Adjust
```bash
# Monitor for 1 week
# If good: increase staging percentage
# If bad: switch back to regular
```

### Phase 4: Full Rollout
```bash
# When staging is proven:
# 1. Update .env.production with staging config
# 2. Switch back to regular deployment
docker compose -f docker-compose-split.yml down
docker compose -f docker-compose.yml up -d
```

---

## Summary

**Traffic splitting enables safe deployments by:**
- ✅ Testing with real traffic (5%)
- ✅ Minimizing risk (95% unaffected)
- ✅ Validating changes before full rollout
- ✅ Easy rollback (switch to regular deployment)

**Start with**: Regular deployment (100% production)
**Add later**: Traffic splitting when you need to test risky changes

---

**Last Updated**: 2025-01-13
**Split Ratio**: 95% Production / 5% Staging
**Algorithm**: Nginx split_clients
**Deployment**: ads.thenexusengine.com
