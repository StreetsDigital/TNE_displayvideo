# TNE Catalyst - Quick Deployment Guide

**Target**: catalyst.springwire.ai
**Method**: Docker Compose
**Time**: ~30 minutes

This is a quick-start guide. For detailed documentation, see `deployment/README.md`.

---

## Prerequisites

What you need before starting:

- [ ] SSH access to your server
- [ ] Docker and Docker Compose installed
- [ ] PostgreSQL database running
- [ ] Domain pointing to your server (catalyst.springwire.ai)
- [ ] SSL certificates (via Certbot or similar)

---

## Step 1: Install Docker (if needed)

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Add your user to docker group
sudo usermod -aG docker $USER

# Log out and back in for group changes to take effect

# Verify installation
docker --version
docker compose version
```

---

## Step 2: Setup Directory Structure

```bash
# Create deployment directory
sudo mkdir -p /opt/catalyst
sudo chown $USER:$USER /opt/catalyst
cd /opt/catalyst

# Create required directories
mkdir -p ssl nginx-logs
```

---

## Step 3: Clone Repository

```bash
# Clone the repository
git clone https://github.com/thenexusengine/tne_springwire.git
cd tne_springwire

# Copy deployment files to /opt/catalyst
cp deployment/* /opt/catalyst/
cd /opt/catalyst
```

---

## Step 4: Setup SSL Certificates

**Option A: Using Certbot (Recommended)**

```bash
# Install Certbot
sudo apt install certbot -y

# Generate certificates (HTTP challenge)
sudo certbot certonly --standalone \
  -d catalyst.springwire.ai \
  --email your@email.com \
  --agree-tos

# Copy certificates to deployment directory
sudo cp /etc/letsencrypt/live/catalyst.springwire.ai/fullchain.pem ./ssl/
sudo cp /etc/letsencrypt/live/catalyst.springwire.ai/privkey.pem ./ssl/
sudo chown $USER:$USER ./ssl/*
```

**Option B: Using Existing Certificates**

```bash
# Copy your existing certificates
cp /path/to/your/fullchain.pem ./ssl/
cp /path/to/your/privkey.pem ./ssl/

# Verify files exist
ls -la ./ssl/
# Should show: fullchain.pem and privkey.pem
```

---

## Step 5: Setup PostgreSQL Database

**If PostgreSQL not installed:**

```bash
sudo apt install postgresql postgresql-contrib -y
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

**Create Database:**

```bash
# Switch to postgres user
sudo -u postgres psql

# In PostgreSQL prompt:
CREATE DATABASE catalyst_production;
CREATE USER catalyst_prod WITH PASSWORD 'YOUR_STRONG_PASSWORD_HERE';
GRANT ALL PRIVILEGES ON DATABASE catalyst_production TO catalyst_prod;
\q
```

**Verify connection:**

```bash
psql -h localhost -U catalyst_prod -d catalyst_production
# Enter password when prompted
# Type \q to exit
```

---

## Step 6: Configure Environment File

```bash
cd /opt/catalyst

# Copy production template
cp .env.production .env

# Edit configuration
nano .env
```

**CRITICAL: Update these values:**

```bash
# Database credentials (from Step 5)
DB_PASSWORD=YOUR_STRONG_PASSWORD_HERE

# Redis password (choose a strong password)
REDIS_PASSWORD=YOUR_STRONG_REDIS_PASSWORD

# CORS - Add your publisher domains
CORS_ALLOWED_ORIGINS=https://yourpublisher.com,https://*.yourpublisher.com
```

**Save and exit**: Ctrl+X, then Y, then Enter

---

## Step 7: Start Services

```bash
cd /opt/catalyst

# Start services
docker compose up -d

# Expected output:
# ✔ Network catalyst-network  Created
# ✔ Container redis-prod      Started
# ✔ Container catalyst-prod   Started
# ✔ Container catalyst-nginx  Started
```

---

## Step 8: Verify Deployment

**Check container status:**

```bash
docker compose ps

# All containers should show "Up" and "healthy"
```

**Check logs:**

```bash
# Catalyst logs
docker compose logs -f catalyst-prod

# Look for:
# "Server starting on :8000"
# "Health check passed"
```

**Test health endpoint:**

```bash
# Local test (should return OK)
curl http://localhost:8000/health

# Public test (should return OK)
curl https://catalyst.springwire.ai/health

# Check headers
curl -I https://catalyst.springwire.ai/health
# Should show: HTTP/2 200
```

**Test auction endpoint:**

```bash
curl -X POST https://catalyst.springwire.ai/openrtb2/auction \
  -H "Content-Type: application/json" \
  -d '{
    "id": "test-auction-001",
    "imp": [{
      "id": "1",
      "banner": {
        "w": 300,
        "h": 250
      }
    }],
    "site": {
      "domain": "test.com",
      "page": "https://test.com/page"
    },
    "device": {
      "ua": "Mozilla/5.0",
      "ip": "203.0.113.1"
    }
  }'
```

**Expected response:** JSON with bid responses or empty seatbid array.

---

## Step 9: Monitor Logs

```bash
# Follow all logs
docker compose logs -f

# Follow specific service
docker compose logs -f catalyst-prod

# View recent errors only
docker compose logs --tail=100 catalyst-prod | grep error

# Exit logs: Ctrl+C
```

---

## Step 10: Setup Auto-Renewal (SSL)

```bash
# Add Certbot renewal to crontab
sudo crontab -e

# Add this line (runs twice daily):
0 0,12 * * * certbot renew --quiet --deploy-hook "docker compose -f /opt/catalyst/docker-compose.yml restart nginx"
```

---

## Common Issues & Solutions

### Issue: Container won't start

**Check logs:**
```bash
docker compose logs catalyst-prod
```

**Common causes:**
- Database connection failed → Check DB_PASSWORD in .env
- Redis connection failed → Check REDIS_PASSWORD in .env
- Port already in use → Check if another service uses port 80/443

### Issue: "Connection refused" errors

**Verify containers are running:**
```bash
docker compose ps
```

**Restart services:**
```bash
docker compose restart
```

### Issue: CORS errors in browser console

**Check CORS configuration:**
```bash
grep CORS_ALLOWED_ORIGINS .env
```

**Update to include your publisher domain:**
```bash
nano .env
# Update CORS_ALLOWED_ORIGINS=https://yourpublisher.com
docker compose restart catalyst-prod
```

### Issue: SSL certificate errors

**Verify certificates exist:**
```bash
ls -la ./ssl/
# Should show: fullchain.pem and privkey.pem
```

**Check certificate validity:**
```bash
openssl x509 -in ./ssl/fullchain.pem -text -noout | grep "Not After"
```

### Issue: High memory usage

**Check container stats:**
```bash
docker stats
```

**Adjust Redis memory limit in docker-compose.yml if needed.**

---

## Maintenance Commands

```bash
# View running containers
docker compose ps

# View logs
docker compose logs -f

# Restart all services
docker compose restart

# Restart specific service
docker compose restart catalyst-prod

# Stop all services
docker compose down

# Update to latest code
git pull origin main
docker compose pull
docker compose up -d

# View resource usage
docker stats

# Clean up old images
docker system prune -a
```

---

## Performance Monitoring

**Run comparison tool:**

```bash
cd /opt/catalyst
./compare-performance.sh
```

**Check response times:**

```bash
docker logs --since 60m catalyst-prod 2>&1 | \
  grep "duration_ms" | \
  grep -oP 'duration_ms":\K[0-9.]+' | \
  awk '{ sum += $1; n++ } END { print "Avg:", sum/n, "ms" }'
```

**Check error rate:**

```bash
docker logs --since 60m catalyst-prod 2>&1 | grep -c '"level":"error"'
```

---

## Advanced: Traffic Splitting (95/5)

**When to use:** Test new features with 5% of traffic before full rollout.

**Switch to traffic splitting:**

```bash
# Stop regular deployment
docker compose down

# Start split deployment
docker compose -f docker-compose-split.yml up -d

# Verify both containers running
docker compose -f docker-compose-split.yml ps
```

**Monitor both environments:**

```bash
# Compare performance
./compare-performance.sh

# View production logs
docker compose -f docker-compose-split.yml logs -f catalyst-prod

# View staging logs
docker compose -f docker-compose-split.yml logs -f catalyst-staging
```

**Rollback to regular deployment:**

```bash
docker compose -f docker-compose-split.yml down
docker compose up -d
```

**See:** `deployment/README-traffic-splitting.md` for full guide.

---

## Security Checklist

After deployment, verify:

- [ ] Changed default passwords (DB_PASSWORD, REDIS_PASSWORD)
- [ ] CORS_ALLOWED_ORIGINS set to specific domains (not *)
- [ ] SSL certificates are valid and auto-renewing
- [ ] Firewall configured (allow 80, 443; block 8000)
- [ ] Database not exposed to internet
- [ ] Regular security updates scheduled
- [ ] Log monitoring configured
- [ ] Backup strategy in place

---

## Backup & Recovery

**Backup PostgreSQL:**

```bash
# Create backup
docker exec -i postgres pg_dump -U catalyst_prod catalyst_production > backup_$(date +%Y%m%d).sql

# Restore backup
docker exec -i postgres psql -U catalyst_prod catalyst_production < backup_20250113.sql
```

**Backup Redis (optional):**

```bash
# Redis persists to ./redis-data volume
docker compose exec redis-prod redis-cli BGSAVE

# Copy Redis data
sudo cp -r /var/lib/docker/volumes/catalyst_redis-data/_data ./redis-backup/
```

**Backup configuration:**

```bash
# Backup all configs
tar -czf catalyst-config-backup.tar.gz \
  .env \
  docker-compose.yml \
  nginx.conf \
  ssl/
```

---

## Getting Help

**Documentation:**
- Full deployment guide: `deployment/README.md`
- Environment variables: `deployment/README-env.md`
- Docker Compose: `deployment/README-docker-compose.md`
- Nginx configuration: `deployment/README-nginx.md`
- Traffic splitting: `deployment/README-traffic-splitting.md`
- Monitoring: `deployment/README-monitoring.md`

**Logs:**
```bash
# Catalyst application logs
docker compose logs catalyst-prod

# Nginx access logs
tail -f /opt/catalyst/nginx-logs/access.log

# Nginx error logs
tail -f /opt/catalyst/nginx-logs/error.log
```

**Check health:**
```bash
curl https://catalyst.springwire.ai/health
```

---

## Next Steps

After successful deployment:

1. **Monitor for 24 hours**
   - Watch logs for errors
   - Check response times
   - Verify auctions completing

2. **Validate with real traffic**
   - Integrate Prebid.js on test page
   - Verify bid responses
   - Check CORS working

3. **Setup monitoring**
   - Configure alerting for errors
   - Monitor resource usage
   - Track auction success rate

4. **Performance tuning**
   - Adjust rate limits if needed
   - Optimize database queries
   - Review Redis memory usage

5. **Consider enhancements**
   - Enable IVT blocking (after validation)
   - Deploy IDR service (optional ML routing)
   - Setup traffic splitting for testing

---

**Deployment Date**: _________________
**Deployed By**: _________________
**Version**: _________________

**Last Updated**: 2025-01-13
**Repository**: https://github.com/thenexusengine/tne_springwire
**Domain**: catalyst.springwire.ai
