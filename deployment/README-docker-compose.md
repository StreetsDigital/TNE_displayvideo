# docker-compose.yml - Container Orchestration

## Purpose
This file defines all the containers (services) needed to run Catalyst and orchestrates how they work together.

## What This File Does

Defines 3 services that run together:
1. **catalyst** - The auction server (Go application)
2. **redis** - Cache and data store
3. **nginx** - Reverse proxy and SSL termination

## Architecture

```
┌─────────────────────────────────────┐
│     Docker Compose Network          │
│                                     │
│  ┌──────────┐    ┌─────────────┐  │
│  │  Nginx   │───▶│  Catalyst   │  │
│  │  :443    │    │   :8000     │  │
│  └──────────┘    └──────┬──────┘  │
│       │                  │         │
│       │           ┌──────▼──────┐  │
│       │           │   Redis     │  │
│       │           │   :6379     │  │
│       │           └─────────────┘  │
└───────┼─────────────────────────────┘
        │
    Internet
```

## Service Breakdown

### 1. Catalyst Service

```yaml
catalyst:
  build:
    context: https://github.com/thenexusengine/tne_springwire.git
    dockerfile: Dockerfile
```

**Decision**: Build from GitHub repository
**Why**:
- ✅ Always get latest code
- ✅ Easy updates: just `docker compose pull && docker compose up -d --build`
- ✅ No need to manually build/push images
- ✅ Source of truth is GitHub

**Alternative Considered**: Pre-built image from registry
**Why Not**: Extra step to build/push, this is simpler for single deployment

---

```yaml
ports:
  - "127.0.0.1:8000:8000"
```

**Decision**: Port 8000 only accessible from localhost
**Why**:
- ✅ Security - not exposed to internet directly
- ✅ All traffic goes through Nginx (rate limiting, SSL, logging)
- ✅ Can still test directly from server: `curl http://localhost:8000/health`

**What this means**:
- `127.0.0.1:8000` = Server's localhost only
- `:8000` inside container
- Nginx (inside Docker network) can reach it as `catalyst:8000`
- External traffic CANNOT reach port 8000 directly

---

```yaml
env_file:
  - .env
```

**Decision**: All environment variables in `.env` file
**Why**:
- ✅ Keep secrets out of docker-compose.yml
- ✅ Easy to change without editing compose file
- ✅ Can have `.env.dev`, `.env.staging`, `.env.prod`

---

```yaml
depends_on:
  redis:
    condition: service_healthy
```

**Decision**: Wait for Redis to be healthy before starting Catalyst
**Why**:
- ✅ Prevents Catalyst from crashing if Redis isn't ready
- ✅ Graceful startup order
- ✅ Health check confirms Redis is actually responding, not just running

---

```yaml
healthcheck:
  test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8000/health"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 20s
```

**Decision**: Check health every 30 seconds
**Why**:
- ✅ Docker knows if Catalyst is actually working
- ✅ Nginx won't start until Catalyst is healthy
- ✅ Can see health status: `docker compose ps`

**What it checks**:
- Hits `/health` endpoint every 30 seconds
- Waits 20 seconds before first check (startup grace period)
- After 3 failures, marks as unhealthy

---

```yaml
deploy:
  resources:
    limits:
      cpus: '2.0'
      memory: 4G
    reservations:
      cpus: '0.5'
      memory: 1G
```

**Decision**: Limit Catalyst to 2 CPUs and 4GB RAM
**Why**:
- ✅ Prevents one container from using all server resources
- ✅ Guarantees minimum 0.5 CPU and 1GB RAM
- ✅ Room for Redis and Nginx

**To Adjust**:
- **Low traffic** (<100 QPS): 1 CPU, 2GB RAM is enough
- **High traffic** (>2000 QPS): Increase to 4 CPU, 8GB RAM

---

```yaml
logging:
  driver: "json-file"
  options:
    max-size: "10m"
    max-file: "3"
```

**Decision**: Rotate logs at 10MB, keep 3 files
**Why**:
- ✅ Prevents disk from filling with logs
- ✅ Max 30MB of logs per container (10MB × 3)
- ✅ Automatic - no cron jobs needed

**What happens**:
```
catalyst.log (current) → 10MB
↓ Rotates to catalyst.log.1
catalyst.log (new current) → 10MB
↓ Rotates to catalyst.log.2
catalyst.log.2 → 10MB
↓ Gets deleted
```

---

### 2. Redis Service

```yaml
redis:
  image: redis:7-alpine
```

**Decision**: Use official Redis 7 Alpine image
**Why**:
- ✅ Alpine = small image (~30MB vs ~100MB)
- ✅ Redis 7 = latest stable version
- ✅ Official image = well-maintained

---

```yaml
command: redis-server --appendonly yes --maxmemory 1024mb --maxmemory-policy allkeys-lru
```

**Decision**: Configure Redis with persistence and memory limit

**`--appendonly yes`**:
- Enables AOF (Append-Only File) persistence
- ✅ Redis saves data to disk
- ✅ Survives container restarts
- ✅ Data stored in `redis-data` volume

**`--maxmemory 1024mb`**:
- Limit Redis to 1GB RAM
- ✅ Prevents Redis from using all memory
- **Decided**: Changed from 512MB to 1GB for higher traffic capacity

**`--maxmemory-policy allkeys-lru`**:
- When memory is full, evict least-recently-used keys
- ✅ Keeps hot data in cache
- ✅ Old data automatically removed

---

```yaml
volumes:
  - redis-data:/data
```

**Decision**: Persist Redis data to named volume
**Why**:
- ✅ Data survives container restarts
- ✅ Easy to backup: `docker cp catalyst-redis:/data/dump.rdb backup.rdb`
- ✅ Managed by Docker (automatic cleanup when removing volumes)

**Where it's stored**:
- Logical: `redis-data` volume
- Physical: `/var/lib/docker/volumes/catalyst_redis-data/_data/`

---

### 3. Nginx Service

```yaml
nginx:
  image: nginx:alpine
```

**Decision**: Official Nginx Alpine image
**Why**: Lightweight, secure, well-maintained

---

```yaml
ports:
  - "80:80"
  - "443:443"
```

**Decision**: Expose HTTP and HTTPS to internet
**Why**:
- Port 80: HTTP → HTTPS redirect + Let's Encrypt challenges
- Port 443: HTTPS traffic (main entry point)

---

```yaml
volumes:
  - ./nginx.conf:/etc/nginx/nginx.conf:ro
  - ./ssl:/etc/nginx/ssl:ro
  - ./nginx-logs:/var/log/nginx
```

**Decision**: Mount config and SSL certs as read-only (`:ro`)
**Why**:
- ✅ Security - container can't modify config or certs
- ✅ Easy to update: edit file, restart container
- ✅ Logs written to host for easy access

**What maps where**:
```
Host                          → Container
/opt/catalyst/nginx.conf      → /etc/nginx/nginx.conf (read-only)
/opt/catalyst/ssl/            → /etc/nginx/ssl/ (read-only)
/opt/catalyst/nginx-logs/     → /var/log/nginx/ (writable)
```

---

## Networks

```yaml
networks:
  catalyst-network:
    driver: bridge
```

**Decision**: All containers on same bridge network
**Why**:
- ✅ Containers can talk to each other by name
- ✅ Isolated from other Docker networks
- ✅ Standard Docker networking

**How it works**:
- Nginx can reach Catalyst as `http://catalyst:8000`
- Catalyst can reach Redis as `redis://redis:6379`
- No need for IP addresses

---

## Volumes

```yaml
volumes:
  redis-data:
    driver: local
```

**Decision**: Local volume for Redis persistence
**Why**:
- ✅ Data survives container restarts
- ✅ Managed by Docker
- ✅ Easy to backup/restore

**Lifecycle**:
- Created: `docker compose up`
- Persists: Even when containers stop
- Removed: `docker compose down -v` (explicit)

---

## Commands Reference

### Start Everything
```bash
cd /opt/catalyst
docker compose up -d
```

### Stop Everything
```bash
docker compose down
```

### View Logs
```bash
docker compose logs -f          # All services
docker compose logs -f catalyst # Just Catalyst
```

### Restart Single Service
```bash
docker compose restart catalyst
docker compose restart redis
docker compose restart nginx
```

### Check Status
```bash
docker compose ps
```

Expected output:
```
NAME                IMAGE               STATUS
catalyst            tne_springwire      Up (healthy)
catalyst-redis      redis:7-alpine      Up (healthy)
catalyst-nginx      nginx:alpine        Up (healthy)
```

### Update Catalyst Code
```bash
# Pull latest code and rebuild
docker compose build --no-cache catalyst
docker compose up -d catalyst

# Or force rebuild from scratch
docker compose down
docker compose up -d --build
```

### Check Resource Usage
```bash
docker stats
```

### Access Container Shell
```bash
docker compose exec catalyst sh      # Catalyst shell
docker compose exec redis redis-cli  # Redis CLI
docker compose exec nginx sh         # Nginx shell
```

---

## Troubleshooting

### Problem: Container won't start
```bash
# Check logs
docker compose logs catalyst

# Common issues:
# - Redis not ready (check depends_on)
# - Port already in use (check with: sudo netstat -tlnp)
# - .env file missing or invalid
```

### Problem: "unhealthy" status
```bash
# Check what health check is failing
docker compose ps

# Check health check logs
docker inspect catalyst | grep -A 10 Health

# Manually test health endpoint
docker compose exec catalyst wget -O- http://localhost:8000/health
```

### Problem: Can't connect to Redis
```bash
# Is Redis running?
docker compose ps redis

# Test Redis connection
docker compose exec redis redis-cli ping
# Should return: PONG

# Check Catalyst can reach Redis
docker compose exec catalyst sh -c "wget -O- http://redis:6379"
```

### Problem: Out of memory
```bash
# Check which container is using too much
docker stats

# Increase limits in docker-compose.yml
# Then:
docker compose down
docker compose up -d
```

### Problem: Need to reset everything
```bash
# CAUTION: This deletes all data
docker compose down -v  # -v removes volumes too
docker compose up -d
```

---

## Scaling Considerations

### Current Capacity
- **Throughput**: ~2000 QPS
- **Memory**: 4GB Catalyst + 1GB Redis = 5GB total
- **Storage**: Redis data (varies by usage)

### To Handle More Traffic

#### Vertical Scaling (Bigger Server)
```yaml
# Increase resource limits
catalyst:
  deploy:
    resources:
      limits:
        cpus: '4.0'
        memory: 8G

redis:
  command: redis-server --maxmemory 2048mb ...
```

#### Horizontal Scaling (Multiple Catalyst Instances)
```bash
# Run 3 Catalyst instances
docker compose up -d --scale catalyst=3

# Then update nginx.conf to load balance:
upstream catalyst {
    server catalyst:8000;
    server catalyst:8000;
    server catalyst:8000;
}
```

---

## Security Notes

### What's Secured
- ✅ Catalyst not exposed to internet (via localhost binding)
- ✅ Redis not exposed to internet (internal network only)
- ✅ Config files read-only
- ✅ Containers run as non-root user (defined in Dockerfile)

### What to Add for Production
- [ ] Redis password (add to command: `--requirepass YOUR_PASSWORD`)
- [ ] Network encryption between containers (Docker secrets)
- [ ] Log aggregation (ship logs to external service)
- [ ] Backup automation (cron job for Redis backups)

---

## Environment Variables

This file uses `.env` file for all configuration.
See `README-env.md` for detailed explanation of each variable.

Quick reference:
- **PBS_HOST_URL**: Public domain (catalyst.springwire.ai)
- **CORS_ALLOWED_ORIGINS**: Publisher domains
- **IDR_ENABLED**: false (start disabled)
- **IVT_BLOCKING_ENABLED**: false (monitoring mode first)

---

## Related Files

- **nginx.conf**: Nginx configuration (see README-nginx.md)
- **.env**: Environment variables (see README-env.md)
- **Dockerfile**: How Catalyst image is built (in repo root)
- **ssl/**: SSL certificates (your colleague provides)

---

## Quick Start Checklist

Before running `docker compose up`:

- [ ] Docker and Docker Compose installed
- [ ] Directory created: `/opt/catalyst/`
- [ ] Files present: `docker-compose.yml`, `nginx.conf`, `.env`
- [ ] SSL directory created: `/opt/catalyst/ssl/`
- [ ] SSL certificates present: `fullchain.pem`, `privkey.pem`
- [ ] `.env` file customized (CORS origins, etc.)
- [ ] DNS pointing to server: `catalyst.springwire.ai`
- [ ] Ports 80, 443 open in firewall

Then:
```bash
cd /opt/catalyst
docker compose up -d
docker compose logs -f
```

---

**Last Updated**: 2025-01-13
**Deployment**: catalyst.springwire.ai
**Maintainer**: The Nexus Engine / Springwire
