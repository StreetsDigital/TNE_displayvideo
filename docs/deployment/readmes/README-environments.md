# Environment Strategy: Dev, Staging, Production

## Overview

This deployment is designed for **ads.thenexusengine.com** which is **PRODUCTION**.

We support 3 environments with a simple `.env` file strategy.

## Environment Definitions

### 1. Development (Local Machine)
**Purpose**: Local testing and development
**Location**: Your laptop/desktop
**Domain**: `localhost:8000`
**Database**: Local Redis
**SSL**: Not needed (HTTP only)

### 2. Staging (Optional)
**Purpose**: Pre-production testing
**Location**: Separate server OR subdomain
**Domain**: `staging.springwire.ai` OR `catalyst-staging.springwire.ai`
**Database**: Separate Redis instance
**SSL**: Yes (Let's Encrypt)

### 3. Production (ads.thenexusengine.com)
**Purpose**: Live traffic
**Location**: Your colleague's server
**Domain**: `ads.thenexusengine.com`
**Database**: Production Redis
**SSL**: Yes (managed by colleague)

---

## Simple Strategy: Multiple .env Files

We use different `.env` files for each environment:

```
/opt/catalyst/
├── .env              ← Current active config (production)
├── .env.dev          ← Development template
├── .env.staging      ← Staging config (if used)
├── .env.production   ← Production config (backup)
├── docker-compose.yml ← Same for all environments
└── nginx.conf        ← Same for all environments
```

### How It Works

**Switching environments** is as simple as:
```bash
# Switch to staging
cp .env.staging .env
docker compose restart

# Switch to production
cp .env.production .env
docker compose restart
```

---

## Environment Configuration Files

### .env.dev (Development - Local)

```bash
# Development Environment
# For local testing on your machine

# Server Configuration
PBS_PORT=8000
PBS_HOST_URL=http://localhost:8000
HOST=0.0.0.0
LOG_LEVEL=debug

# CORS - Allow all for dev
CORS_ALLOWED_ORIGINS=*
CORS_ENABLED=true

# Redis - Use local Redis or Docker
REDIS_URL=redis://localhost:6379/0

# IDR - Disabled or local instance
IDR_ENABLED=false
IDR_URL=http://localhost:5050

# IVT Detection - Monitoring only
IVT_MONITORING_ENABLED=true
IVT_BLOCKING_ENABLED=false
IVT_CHECK_UA=true
IVT_CHECK_REFERER=false  # Dev requests often have no referer

# Privacy - Relaxed for testing
PBS_ENFORCE_GDPR=false
PBS_ENFORCE_CCPA=false
PBS_ENFORCE_COPPA=false
PBS_DISABLE_GDPR_ENFORCEMENT=true

# Publisher Auth - Permissive
PUBLISHER_AUTH_ENABLED=true
PUBLISHER_ALLOW_UNREGISTERED=true
PUBLISHER_VALIDATE_DOMAIN=false
REGISTERED_PUBLISHERS=test-pub:localhost,dev-pub:127.0.0.1
```

### .env.staging (Staging - Pre-Production Server)

```bash
# Staging Environment
# For testing before production deployment

# Server Configuration
PBS_PORT=8000
PBS_HOST_URL=https://staging.springwire.ai
HOST=0.0.0.0
LOG_LEVEL=info

# CORS - Restrict to test domains
CORS_ALLOWED_ORIGINS=https://staging-publisher.com,https://*.staging-publisher.com
CORS_ENABLED=true

# Redis - Staging Redis instance
REDIS_URL=redis://staging-redis:6379/0

# IDR - Can test IDR here
IDR_ENABLED=false
IDR_URL=https://staging-idr.springwire.ai

# IVT Detection - Test blocking mode
IVT_MONITORING_ENABLED=true
IVT_BLOCKING_ENABLED=true  # Test blocking behavior
IVT_CHECK_UA=true
IVT_CHECK_REFERER=true

# Privacy - Enforce like production
PBS_ENFORCE_GDPR=true
PBS_ENFORCE_CCPA=true
PBS_ENFORCE_COPPA=true

# Publisher Auth - Test strict mode
PUBLISHER_AUTH_ENABLED=true
PUBLISHER_ALLOW_UNREGISTERED=false
PUBLISHER_VALIDATE_DOMAIN=true
REGISTERED_PUBLISHERS=staging-pub:staging-publisher.com
```

### .env.production (Production - ads.thenexusengine.com)

```bash
# Production Environment
# Live traffic on ads.thenexusengine.com

# Server Configuration
PBS_PORT=8000
PBS_HOST_URL=https://ads.thenexusengine.com
HOST=0.0.0.0
LOG_LEVEL=info

# CORS - Restrict to real publisher domains
CORS_ALLOWED_ORIGINS=https://yourpublisher.com,https://*.yourpublisher.com
CORS_ENABLED=true

# Redis - Production Redis
REDIS_URL=redis://redis:6379/0

# IDR - Disabled initially, enable after testing
IDR_ENABLED=false
IDR_URL=https://idr.thenexusengine.com

# IVT Detection - Start monitoring, enable blocking later
IVT_MONITORING_ENABLED=true
IVT_BLOCKING_ENABLED=false  # Set to true after 1-2 weeks monitoring
IVT_CHECK_UA=true
IVT_CHECK_REFERER=true
IVT_ALLOWED_COUNTRIES=US,GB,CA,AU,NZ,DE,FR,IT,ES

# Privacy - Full enforcement
PBS_ENFORCE_GDPR=true
PBS_ENFORCE_CCPA=true
PBS_ENFORCE_COPPA=true

# Publisher Auth - Strict mode
PUBLISHER_AUTH_ENABLED=true
PUBLISHER_ALLOW_UNREGISTERED=false
PUBLISHER_VALIDATE_DOMAIN=true
REGISTERED_PUBLISHERS=prod-pub-123:yourpublisher.com,prod-pub-456:anotherpub.com
```

---

## Recommended Workflow

### Phase 1: Local Development (Your Machine)

```bash
# On your laptop
cd ~/projects/tne_springwire
cp deployment/.env.dev .env

# Start with docker-compose locally
docker compose -f deployment/docker-compose.yml up -d

# Test
curl http://localhost:8000/health

# Make changes, test, commit to GitHub
```

### Phase 2: Deploy to Production (Colleague's Server)

```bash
# On production server
cd /opt/catalyst
cp .env.production .env

# Customize for real publishers
nano .env
# Update CORS_ALLOWED_ORIGINS
# Update REGISTERED_PUBLISHERS

# Deploy
docker compose up -d

# Monitor
docker compose logs -f
```

### Phase 3 (Optional): Add Staging

If you need a staging environment:

```bash
# On staging server (or subdomain)
cd /opt/catalyst-staging
cp .env.staging .env

# Deploy
docker compose up -d
```

---

## Key Differences Between Environments

| Feature | Development | Staging | Production |
|---------|------------|---------|-----------|
| **Domain** | localhost | staging.springwire.ai | ads.thenexusengine.com |
| **SSL** | ❌ HTTP | ✅ HTTPS | ✅ HTTPS |
| **Log Level** | debug | info | info |
| **CORS** | * (all) | Test domains | Real publishers |
| **IVT Blocking** | ❌ Off | ✅ On (for testing) | ⚠️ Monitor first |
| **GDPR** | ❌ Off | ✅ On | ✅ On |
| **Publisher Auth** | Permissive | Strict | Strict |
| **IDR** | ❌ Disabled | ⚠️ Optional | ⚠️ Enable later |

---

## Switching Environments

### On Development Machine
```bash
# Use dev settings
docker compose down
cp deployment/.env.dev .env
docker compose up -d
```

### On Production Server
```bash
# During initial setup
cd /opt/catalyst
cp .env.production .env

# To update configuration
nano .env
docker compose restart catalyst
```

### Testing Configuration Before Applying
```bash
# Verify environment variables will load correctly
docker compose config

# Dry-run (shows what will happen)
docker compose up --no-start
```

---

## Environment-Specific Considerations

### Development
- **No nginx needed** - can run Catalyst directly
- **No SSL** - HTTP is fine
- **Sample data** - Use test publishers and fake bid requests
- **Fast iteration** - No need to restart on code changes (use `go run`)

### Staging (Optional)
- **Mirrors production** - Same config as prod
- **Safe testing** - Test IVT blocking, GDPR enforcement
- **Real data** - Can use real bid requests
- **Before deployment** - Test new features here first

### Production
- **Strict security** - All protections enabled
- **Monitoring mode first** - Start with IVT monitoring only
- **Gradual rollout**:
  - Week 1: IVT monitoring only
  - Week 2: Enable IVT blocking
  - Week 3: Enable strict publisher auth
  - Week 4: Enable IDR (optional)

---

## For ads.thenexusengine.com

Since **ads.thenexusengine.com is PRODUCTION**, your colleague should:

1. **Use `.env.production`** as the template
2. **Customize** for real publishers:
   ```bash
   CORS_ALLOWED_ORIGINS=https://realpublisher.com
   REGISTERED_PUBLISHERS=pub-123:realpublisher.com
   ```
3. **Start conservative**:
   - IVT_BLOCKING_ENABLED=false (monitor first)
   - PUBLISHER_ALLOW_UNREGISTERED=true (test first)
   - IDR_ENABLED=false (add later)
4. **Enable features gradually** based on logs

---

## Do You Need Staging?

### Skip staging if:
- ✅ Small deployment (1-2 publishers)
- ✅ Low traffic (<1000 QPS)
- ✅ Can test locally
- ✅ Quick rollback available

### Use staging if:
- ⚠️ Many publishers (5+)
- ⚠️ High traffic (>5000 QPS)
- ⚠️ Can't afford downtime
- ⚠️ Need to test integrations

**For ads.thenexusengine.com**: Probably don't need staging initially. Start with production, monitor closely.

---

## Quick Environment Check

To see what environment you're in:

```bash
# Check current config
grep PBS_HOST_URL /opt/catalyst/.env

# Should show:
# Development: PBS_HOST_URL=http://localhost:8000
# Staging: PBS_HOST_URL=https://staging.springwire.ai
# Production: PBS_HOST_URL=https://ads.thenexusengine.com
```

---

## Summary

**Simple approach for this deployment:**

1. **Development** = Your local machine (localhost)
2. **Production** = ads.thenexusengine.com (colleague's server)
3. **(Optional) Staging** = If needed later

**One docker-compose.yml** works for all environments
**Different .env files** customize behavior

**Switching is simple**: Copy the right `.env` file and restart.

---

**Decision for this deployment**:
- ✅ Production only (ads.thenexusengine.com)
- ✅ Use `.env.production` template
- ✅ Add staging later if needed
- ✅ Development on local machines

---

**Last Updated**: 2025-01-13
**Strategy**: Simple .env file switching
**Current Target**: Production (ads.thenexusengine.com)
