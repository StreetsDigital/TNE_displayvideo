# TNE Catalyst - Pre-Deployment Checklist

**Domain**: catalyst.springwire.ai
**Date**: ______________
**Deployed By**: ______________

Use this checklist to ensure everything is ready before production deployment.

---

## Pre-Deployment Checklist

### Infrastructure Setup

- [ ] **Server Provisioned**
  - [ ] SSH access working
  - [ ] Minimum 4 CPU cores
  - [ ] Minimum 8GB RAM
  - [ ] 50GB+ disk space
  - [ ] Ubuntu 20.04+ or equivalent

- [ ] **Docker Installed**
  - [ ] Docker version 20.10+ installed
  - [ ] Docker Compose v2+ installed
  - [ ] User added to docker group
  - [ ] Test: `docker run hello-world` works

- [ ] **PostgreSQL Database**
  - [ ] PostgreSQL 13+ installed and running
  - [ ] Database created: `catalyst_production`
  - [ ] User created: `catalyst_prod`
  - [ ] Password set (strong password)
  - [ ] Permissions granted
  - [ ] Test connection successful
  - [ ] Database accessible from Docker network

- [ ] **DNS Configuration**
  - [ ] catalyst.springwire.ai points to server IP
  - [ ] DNS propagation complete (check with `dig`)
  - [ ] A record or CNAME configured
  - [ ] TTL set appropriately

- [ ] **Firewall Configuration**
  - [ ] Port 80 open (HTTP)
  - [ ] Port 443 open (HTTPS)
  - [ ] Port 22 open (SSH)
  - [ ] Port 8000 blocked (internal only)
  - [ ] Port 6379 blocked (Redis internal only)
  - [ ] Port 5432 blocked (PostgreSQL internal only)

---

### SSL Certificates

- [ ] **SSL Setup Method Chosen**
  - [ ] Option A: Certbot (Let's Encrypt) - Free, auto-renewal
  - [ ] Option B: Custom certificates - Manual management

- [ ] **Certificates Obtained**
  - [ ] fullchain.pem exists
  - [ ] privkey.pem exists
  - [ ] Certificates valid (not expired)
  - [ ] Certificates match domain (catalyst.springwire.ai)

- [ ] **Certificates Installed**
  - [ ] Copied to `/opt/catalyst/ssl/`
  - [ ] Permissions correct (readable by docker)
  - [ ] Test: `openssl x509 -in ssl/fullchain.pem -text -noout`

- [ ] **Auto-Renewal Configured** (if using Certbot)
  - [ ] Certbot renewal cron job added
  - [ ] Test renewal: `certbot renew --dry-run`
  - [ ] Deploy hook configured to restart nginx

---

### File Setup

- [ ] **Repository Cloned**
  - [ ] Cloned from: https://github.com/thenexusengine/tne_springwire.git
  - [ ] Latest code pulled: `git pull origin main`
  - [ ] Correct branch checked out

- [ ] **Deployment Directory Created**
  - [ ] `/opt/catalyst` directory created
  - [ ] Correct ownership: `chown $USER:$USER /opt/catalyst`
  - [ ] Deployment files copied to `/opt/catalyst`

- [ ] **Required Directories Created**
  - [ ] `ssl/` directory exists
  - [ ] `nginx-logs/` directory created
  - [ ] Permissions correct

---

### Configuration Files

- [ ] **Environment File Configured**
  - [ ] Copied: `cp .env.production .env`
  - [ ] Edited: `nano .env`
  - [ ] All placeholders replaced

- [ ] **Critical Settings Updated**
  - [ ] `PBS_HOST_URL=https://catalyst.springwire.ai`
  - [ ] `DB_PASSWORD` changed (not default)
  - [ ] `REDIS_PASSWORD` set (strong password)
  - [ ] `CORS_ALLOWED_ORIGINS` set to publisher domains (not *)
  - [ ] `DB_HOST` set correctly (container name or IP)
  - [ ] `REDIS_HOST` set correctly (container name)

- [ ] **Optional Settings Reviewed**
  - [ ] `IDR_ENABLED=false` (start disabled)
  - [ ] `IVT_BLOCKING_ENABLED=false` (monitor first)
  - [ ] `LOG_LEVEL=info` (not debug)
  - [ ] `LOG_FORMAT=json` (for log aggregation)
  - [ ] `RATE_LIMIT_GENERAL` appropriate for traffic
  - [ ] `RATE_LIMIT_AUCTION` appropriate for traffic
  - [ ] `AUCTION_TIMEOUT` reasonable (2s recommended)
  - [ ] `AUCTION_MAX_BIDDERS` appropriate (15 recommended)

---

### Docker Compose

- [ ] **Correct Compose File Selected**
  - [ ] `docker-compose.yml` for regular deployment (100% traffic)
  - [ ] `docker-compose-split.yml` for traffic splitting (95/5)

- [ ] **Resource Limits Reviewed**
  - [ ] CPU limits appropriate for server
  - [ ] Memory limits appropriate for server
  - [ ] Redis memory limit set (1024mb default)

- [ ] **Volumes Configured**
  - [ ] Redis persistence enabled (redis-data volume)
  - [ ] SSL volume mounted correctly
  - [ ] Nginx logs volume mounted

---

### Nginx Configuration

- [ ] **Correct Nginx Config Selected**
  - [ ] `nginx.conf` for regular deployment
  - [ ] `nginx-split.conf` for traffic splitting

- [ ] **Rate Limits Appropriate**
  - [ ] `limit_req_zone` general set (100r/s default)
  - [ ] `limit_req_zone` auction set (50r/s default)
  - [ ] Burst values reasonable

- [ ] **SSL Configuration Verified**
  - [ ] SSL certificate paths correct
  - [ ] TLS 1.2+ only
  - [ ] Strong ciphers configured

- [ ] **Security Headers Enabled**
  - [ ] HSTS header present
  - [ ] X-Frame-Options set
  - [ ] X-Content-Type-Options set
  - [ ] CORS headers configured

- [ ] **Timeouts Appropriate**
  - [ ] Auction timeout: 10s (default)
  - [ ] General timeout: 30s (default)
  - [ ] Connection timeout: 10s (default)

---

### Security Review

- [ ] **Passwords & Secrets**
  - [ ] All default passwords changed
  - [ ] Strong passwords used (20+ characters)
  - [ ] Passwords not in git history
  - [ ] Passwords documented securely (password manager)

- [ ] **CORS Configuration**
  - [ ] NOT using `CORS_ALLOWED_ORIGINS=*` in production
  - [ ] All publisher domains listed
  - [ ] Wildcard subdomains used carefully

- [ ] **Database Security**
  - [ ] Database user has minimal required permissions
  - [ ] SSL connection enabled (`DB_SSL_MODE=require`)
  - [ ] Database not exposed to internet
  - [ ] Strong database password

- [ ] **Redis Security**
  - [ ] Redis password set
  - [ ] Redis not exposed to internet
  - [ ] Redis persistence enabled (AOF)

- [ ] **Debug Features Disabled**
  - [ ] `PPROF_ENABLED=false`
  - [ ] `DEBUG_ENDPOINTS=false`
  - [ ] `FEATURE_DEBUG_ENDPOINTS=false`

- [ ] **Firewall Rules**
  - [ ] Only ports 80, 443, 22 exposed
  - [ ] Internal services blocked from internet

---

### Testing (Pre-Deployment)

- [ ] **Configuration Validation**
  - [ ] `.env` file has no syntax errors
  - [ ] `docker-compose.yml` validates: `docker compose config`
  - [ ] `nginx.conf` validates: `nginx -t` (in container)

- [ ] **Dry Run** (if possible)
  - [ ] Test on staging server first
  - [ ] Verify all containers start
  - [ ] Verify health checks pass

---

### Deployment

- [ ] **Services Started**
  - [ ] Run: `docker compose up -d`
  - [ ] No errors in startup
  - [ ] All containers running: `docker compose ps`

- [ ] **Health Checks**
  - [ ] Catalyst container shows "(healthy)"
  - [ ] Redis container shows "(healthy)"
  - [ ] Nginx container shows "(healthy)"

- [ ] **Logs Reviewed**
  - [ ] No critical errors in logs
  - [ ] Catalyst started successfully
  - [ ] Redis connected
  - [ ] Database connected

---

### Post-Deployment Verification

- [ ] **HTTP Redirect Working**
  - [ ] Test: `curl -I http://catalyst.springwire.ai`
  - [ ] Should return: `301 Moved Permanently`
  - [ ] Location header: `https://catalyst.springwire.ai`

- [ ] **HTTPS Working**
  - [ ] Test: `curl -I https://catalyst.springwire.ai`
  - [ ] Should return: `200 OK`
  - [ ] No SSL errors

- [ ] **Health Endpoint**
  - [ ] Test: `curl https://catalyst.springwire.ai/health`
  - [ ] Should return: `{"status":"ok"}` or similar
  - [ ] Response time < 100ms

- [ ] **Auction Endpoint**
  - [ ] Test with sample bid request (see DEPLOYMENT_GUIDE.md)
  - [ ] Should return: Valid OpenRTB response
  - [ ] No timeout errors
  - [ ] Response time < 2s

- [ ] **CORS Headers**
  - [ ] Test from browser console on publisher site
  - [ ] No CORS errors
  - [ ] Preflight requests successful

- [ ] **SSL Certificate**
  - [ ] Test: `openssl s_client -connect catalyst.springwire.ai:443`
  - [ ] Certificate valid
  - [ ] Certificate matches domain
  - [ ] Not expired

- [ ] **Response Headers**
  - [ ] HSTS header present
  - [ ] Security headers present
  - [ ] No sensitive information exposed

---

### Monitoring Setup

- [ ] **Log Monitoring**
  - [ ] Know how to view logs: `docker compose logs -f`
  - [ ] Nginx logs accessible: `tail -f nginx-logs/access.log`
  - [ ] Error logs accessible: `tail -f nginx-logs/error.log`

- [ ] **Performance Monitoring**
  - [ ] Know how to check stats: `docker stats`
  - [ ] Performance script tested: `./compare-performance.sh`
  - [ ] Response time monitoring in place

- [ ] **Error Alerting** (optional but recommended)
  - [ ] Set up error log monitoring
  - [ ] Email alerts configured
  - [ ] Slack/Discord webhooks configured

---

### Documentation

- [ ] **Credentials Documented**
  - [ ] Database credentials stored securely
  - [ ] Redis password stored securely
  - [ ] SSL certificate renewal process documented

- [ ] **Runbook Created**
  - [ ] How to restart services
  - [ ] How to view logs
  - [ ] How to rollback
  - [ ] Emergency contacts listed

- [ ] **Access Documentation**
  - [ ] SSH access instructions
  - [ ] Docker access instructions
  - [ ] Database access instructions

---

### Backup & Recovery

- [ ] **Backup Strategy Defined**
  - [ ] Database backup frequency decided
  - [ ] Configuration backup plan
  - [ ] SSL certificate backup plan

- [ ] **Initial Backups Taken**
  - [ ] Database backed up
  - [ ] Configuration files backed up
  - [ ] SSL certificates backed up

- [ ] **Recovery Tested**
  - [ ] Know how to restore database
  - [ ] Know how to restore configuration
  - [ ] Know how to rollback deployment

---

### Traffic Splitting (if using)

Only if deploying with `docker-compose-split.yml`:

- [ ] **Staging Configuration**
  - [ ] `.env.staging` configured
  - [ ] Separate Redis instance configured
  - [ ] Different settings for testing

- [ ] **Split Verification**
  - [ ] Test split ratio: send 100 requests, check X-Backend header
  - [ ] ~95 should go to production
  - [ ] ~5 should go to staging

- [ ] **Monitoring Tools**
  - [ ] `compare-performance.sh` executable
  - [ ] Performance comparison tested
  - [ ] Both containers being monitored

---

### Integration Testing

- [ ] **Prebid.js Integration**
  - [ ] Test page created with Prebid.js
  - [ ] Catalyst configured as bidder
  - [ ] Test auction runs successfully
  - [ ] Bids received and rendered

- [ ] **Publisher Integration**
  - [ ] Publisher domains added to CORS
  - [ ] Test on actual publisher page
  - [ ] No console errors
  - [ ] Ads rendering correctly

- [ ] **Demand Partners** (if configured)
  - [ ] Bidder adapters configured
  - [ ] Test auctions with real bidders
  - [ ] Bids returning successfully
  - [ ] Win notifications working

---

### Performance Validation

- [ ] **Response Times**
  - [ ] Average auction response < 2s
  - [ ] Health check response < 100ms
  - [ ] No timeout errors

- [ ] **Resource Usage**
  - [ ] CPU usage < 70% average
  - [ ] Memory usage stable (not growing)
  - [ ] Disk space sufficient

- [ ] **Error Rates**
  - [ ] Error rate < 1%
  - [ ] No critical errors
  - [ ] Warnings reviewed

---

### 24-Hour Monitoring

After deployment, monitor for 24 hours:

- [ ] **Hour 1-4: Intensive Monitoring**
  - [ ] Check logs every 30 minutes
  - [ ] Monitor error rates
  - [ ] Watch for anomalies

- [ ] **Hour 4-12: Regular Monitoring**
  - [ ] Check logs every 2 hours
  - [ ] Verify stability
  - [ ] Review metrics

- [ ] **Hour 12-24: Stability Verification**
  - [ ] Check logs every 4 hours
  - [ ] Verify no memory leaks
  - [ ] Confirm no degradation

- [ ] **24-Hour Review**
  - [ ] Total error count acceptable
  - [ ] No critical issues
  - [ ] Performance stable
  - [ ] Ready for normal operation

---

### Post-Deployment Tasks

- [ ] **Announce Go-Live**
  - [ ] Notify stakeholders
  - [ ] Provide monitoring links
  - [ ] Share emergency contacts

- [ ] **Update Documentation**
  - [ ] Record deployment date
  - [ ] Document any issues encountered
  - [ ] Update runbook with learnings

- [ ] **Schedule First Review**
  - [ ] 1-week performance review
  - [ ] Identify optimization opportunities
  - [ ] Plan next steps (IVT, IDR, etc.)

---

## Sign-Off

**Deployment Completed**:

- Date: ______________
- Time: ______________
- Deployed By: ______________
- Verified By: ______________

**24-Hour Monitoring Completed**:

- Date: ______________
- No Critical Issues: [ ] Yes [ ] No
- Performance Acceptable: [ ] Yes [ ] No
- Ready for Production: [ ] Yes [ ] No

**Notes**:
_________________________________________________________________
_________________________________________________________________
_________________________________________________________________

---

## Emergency Rollback

If critical issues arise:

```bash
# Stop current deployment
docker compose down

# Restore previous configuration
cp .env.backup .env

# Start with previous version
docker compose up -d

# Verify rollback successful
curl https://catalyst.springwire.ai/health
```

**Rollback Contact**: ______________
**Rollback Decision Authority**: ______________

---

**Last Updated**: 2025-01-13
**Version**: 1.0.0
**Domain**: catalyst.springwire.ai
