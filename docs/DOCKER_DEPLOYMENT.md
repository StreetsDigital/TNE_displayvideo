# Docker Deployment Guide: catalyst.springwire.ai

Complete step-by-step guide to deploy TNE Catalyst using Docker Compose.

## Table of Contents
1. [Prerequisites](#prerequisites)
2. [Quick Start](#quick-start)
3. [Detailed Installation](#detailed-installation)
4. [Configuration](#configuration)
5. [SSL Setup](#ssl-setup)
6. [Operations](#operations)
7. [Monitoring](#monitoring)
8. [Troubleshooting](#troubleshooting)

---

## Prerequisites

### Server Requirements
- **Server**: VPS or dedicated server
- **OS**: Ubuntu 20.04+ or Debian 11+
- **RAM**: 4GB minimum (8GB recommended)
- **CPU**: 2+ cores
- **Storage**: 20GB+ SSD
- **Domain**: catalyst.springwire.ai pointing to server IP

### Software to Install
Only Docker is needed! Everything else runs in containers.

---

## Quick Start

### Step 1: Install Docker (5 minutes)

```bash
# Install Docker Engine
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Install Docker Compose plugin
sudo apt update
sudo apt install docker-compose-plugin -y

# Add your user to docker group (so you don't need sudo)
sudo usermod -aG docker $USER

# IMPORTANT: Log out and back in for group membership to take effect
exit
# Then SSH back in
```

Verify installation:
```bash
docker --version
docker compose version
```

### Step 2: Download Deployment Files (2 minutes)

```bash
# Create deployment directory
mkdir -p /opt/catalyst
cd /opt/catalyst

# Download configuration files
# (Upload these files from the repository to your server)
# Files needed:
# - docker-compose.yml
# - nginx.conf
# - .env
```

### Step 3: Configure Environment (3 minutes)

```bash
# Copy example environment file
cp .env.example .env

# Edit configuration
nano .env

# At minimum, update these:
# - CORS_ALLOWED_ORIGINS (add your publisher domains)
# - Change default values as needed
```

### Step 4: Start Services (1 minute)

```bash
cd /opt/catalyst
docker compose up -d
```

### Step 5: Get SSL Certificate (5 minutes)

```bash
# First, ensure DNS is pointing to your server
# Then get certificate from Let's Encrypt
docker compose run --rm certbot certonly --webroot \
  --webroot-path=/var/www/certbot \
  --email your-email@example.com \
  --agree-tos \
  --no-eff-email \
  -d catalyst.springwire.ai

# Restart nginx to load certificate
docker compose restart nginx
```

### Step 6: Verify (1 minute)

```bash
# Check all containers are running
docker compose ps

# Test health endpoint
curl http://localhost:8000/health

# Test through nginx (after SSL setup)
curl https://catalyst.springwire.ai/health
```

**Done!** Your auction server is live at `https://catalyst.springwire.ai`

---

## Detailed Installation

### 1. Server Preparation

#### Update System
```bash
sudo apt update
sudo apt upgrade -y
```

#### Configure Firewall
```bash
# Install UFW if not already installed
sudo apt install ufw -y

# Allow SSH (IMPORTANT: Do this first!)
sudo ufw allow 22/tcp

# Allow HTTP and HTTPS
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Enable firewall
sudo ufw enable

# Check status
sudo ufw status
```

#### Set Up DNS
Ensure your domain points to the server:
```bash
# Check DNS propagation
dig catalyst.springwire.ai +short

# Should return your server's IP address
```

### 2. Install Docker

#### Method 1: Official Docker Script (Recommended)
```bash
# Download and run Docker installation script
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Clean up
rm get-docker.sh
```

#### Method 2: Manual Installation (Ubuntu)
```bash
# Install prerequisites
sudo apt update
sudo apt install ca-certificates curl gnupg -y

# Add Docker's official GPG key
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

# Set up repository
echo \
  "deb [arch="$(dpkg --print-architecture)" signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  "$(. /etc/os-release && echo "$VERSION_CODENAME")" stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# Install Docker
sudo apt update
sudo apt install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin -y
```

#### Configure Docker User
```bash
# Add your user to docker group
sudo usermod -aG docker $USER

# Verify group membership
groups $USER

# IMPORTANT: Log out and back in
exit
# SSH back in

# Test Docker without sudo
docker run hello-world
```

### 3. Deploy Catalyst

#### Create Directory Structure
```bash
# Create deployment directory
sudo mkdir -p /opt/catalyst
sudo chown $USER:$USER /opt/catalyst
cd /opt/catalyst

# Create subdirectories
mkdir -p ssl nginx-logs redis-data
```

#### Upload Configuration Files

Upload these files to `/opt/catalyst/`:
- `docker-compose.yml` (provided below)
- `nginx.conf` (provided below)
- `.env` (copy from `.env.example` and customize)

**Option 1: Using SCP from your local machine:**
```bash
# From your local machine where you have the files
scp docker-compose.yml your-user@catalyst.springwire.ai:/opt/catalyst/
scp nginx.conf your-user@catalyst.springwire.ai:/opt/catalyst/
scp .env.example your-user@catalyst.springwire.ai:/opt/catalyst/.env
```

**Option 2: Using Git:**
```bash
cd /opt/catalyst
git clone https://github.com/thenexusengine/tne_springwire.git temp
cp temp/deployment/docker-compose.yml .
cp temp/deployment/nginx.conf .
cp temp/deployment/.env.example .env
rm -rf temp
```

**Option 3: Manual creation** (copy from files below)

#### Configure Environment Variables
```bash
cd /opt/catalyst
nano .env
```

Update these critical values:
```bash
# Your publisher domains (comma-separated)
CORS_ALLOWED_ORIGINS=https://yourpublisher.com,https://*.yourpublisher.com

# Set to true after testing in monitoring mode
IVT_BLOCKING_ENABLED=false

# Add your publisher IDs (optional, can use Redis later)
REGISTERED_PUBLISHERS=pub-123:yourpublisher.com

# Keep IDR disabled initially
IDR_ENABLED=false
```

#### Start Services
```bash
cd /opt/catalyst

# Pull images
docker compose pull

# Start in detached mode
docker compose up -d

# Check status
docker compose ps

# View logs
docker compose logs -f catalyst
```

---

## Configuration

### Environment Variables Reference

See `.env.example` for all available options. Key variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `PBS_HOST_URL` | Public URL | `https://catalyst.springwire.ai` |
| `CORS_ALLOWED_ORIGINS` | Publisher domains | Empty (configure!) |
| `IDR_ENABLED` | Enable ML routing | `false` |
| `IVT_BLOCKING_ENABLED` | Block invalid traffic | `false` (monitoring mode) |
| `PUBLISHER_AUTH_ENABLED` | Validate publishers | `true` |
| `LOG_LEVEL` | Logging verbosity | `info` |

### Adding Publishers

#### Method 1: Environment Variable (Simple, for testing)
```bash
# Edit .env
nano /opt/catalyst/.env

# Add line:
REGISTERED_PUBLISHERS=pub-123:example.com,pub-456:another.com

# Restart
docker compose restart catalyst
```

#### Method 2: Redis (Flexible, for production)
```bash
# Connect to Redis container
docker compose exec redis redis-cli

# Add publisher
HSET tne_catalyst:publishers pub-123 "example.com|*.example.com"

# List all publishers
HGETALL tne_catalyst:publishers

# Remove publisher
HDEL tne_catalyst:publishers pub-123

# Exit Redis
exit
```

### Updating Configuration

After changing `.env`:
```bash
docker compose down
docker compose up -d
```

Or for quicker changes (doesn't recreate containers):
```bash
docker compose restart catalyst
```

---

## SSL Setup

### Option 1: Let's Encrypt (Recommended - Free)

#### First Time Setup
```bash
cd /opt/catalyst

# Get certificate
docker compose run --rm certbot certonly --webroot \
  --webroot-path=/var/www/certbot \
  --email your-email@example.com \
  --agree-tos \
  --no-eff-email \
  -d catalyst.springwire.ai

# Restart nginx
docker compose restart nginx
```

#### Automatic Renewal (Set up cron job)
```bash
# Create renewal script
cat > /opt/catalyst/renew-cert.sh << 'EOF'
#!/bin/bash
cd /opt/catalyst
docker compose run --rm certbot renew
docker compose restart nginx
echo "Certificate renewal completed at $(date)" >> /opt/catalyst/renewal.log
EOF

# Make executable
chmod +x /opt/catalyst/renew-cert.sh

# Add to crontab (runs monthly)
(crontab -l 2>/dev/null; echo "0 3 1 * * /opt/catalyst/renew-cert.sh") | crontab -
```

### Option 2: Existing SSL Certificate

If you have existing SSL certificates:
```bash
# Copy certificates to ssl directory
sudo cp /path/to/fullchain.pem /opt/catalyst/ssl/
sudo cp /path/to/privkey.pem /opt/catalyst/ssl/

# Set permissions
sudo chown $USER:$USER /opt/catalyst/ssl/*.pem
sudo chmod 600 /opt/catalyst/ssl/*.pem

# Restart nginx
docker compose restart nginx
```

### Testing SSL
```bash
# Test SSL connection
curl -I https://catalyst.springwire.ai

# Check certificate
echo | openssl s_client -servername catalyst.springwire.ai -connect catalyst.springwire.ai:443 2>/dev/null | openssl x509 -noout -dates
```

---

## Operations

### Daily Operations

#### View Logs
```bash
# All services
docker compose logs -f

# Catalyst only
docker compose logs -f catalyst

# Last 100 lines
docker compose logs --tail=100 catalyst

# Nginx access logs
docker compose logs -f nginx | grep "catalyst.springwire.ai"
```

#### Check Status
```bash
# Container status
docker compose ps

# Resource usage
docker stats

# Health check
curl http://localhost:8000/health
curl https://catalyst.springwire.ai/health/ready
```

#### Restart Services
```bash
# Restart all
docker compose restart

# Restart specific service
docker compose restart catalyst
docker compose restart nginx
docker compose restart redis
```

### Updating Catalyst

#### Update to Latest Version
```bash
cd /opt/catalyst

# Pull new image
docker compose pull catalyst

# Restart with new image
docker compose up -d catalyst

# Check logs
docker compose logs -f catalyst
```

#### Rollback to Previous Version
```bash
# Stop current version
docker compose down catalyst

# Edit docker-compose.yml to specify version
# Change: image: ghcr.io/streetsdigital/tne-catalyst:latest
# To: image: ghcr.io/streetsdigital/tne-catalyst:v1.0.0

# Start previous version
docker compose up -d catalyst
```

### Backup and Restore

#### Backup Redis Data
```bash
# Create backup
docker compose exec redis redis-cli SAVE
docker cp catalyst-redis:/data/dump.rdb /opt/catalyst/backups/redis-$(date +%Y%m%d).rdb

# Automated daily backup
cat > /opt/catalyst/backup-redis.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="/opt/catalyst/backups"
mkdir -p $BACKUP_DIR
docker compose -f /opt/catalyst/docker-compose.yml exec -T redis redis-cli SAVE
docker cp catalyst-redis:/data/dump.rdb $BACKUP_DIR/redis-$(date +%Y%m%d-%H%M%S).rdb
# Keep last 7 days
find $BACKUP_DIR -name "redis-*.rdb" -mtime +7 -delete
EOF

chmod +x /opt/catalyst/backup-redis.sh

# Add to crontab (daily at 2 AM)
(crontab -l 2>/dev/null; echo "0 2 * * * /opt/catalyst/backup-redis.sh") | crontab -
```

#### Restore Redis Data
```bash
# Stop Redis
docker compose stop redis

# Restore backup
docker cp /opt/catalyst/backups/redis-20250113.rdb catalyst-redis:/data/dump.rdb

# Start Redis
docker compose start redis
```

### Scaling

#### Vertical Scaling (More Resources)
```bash
# Edit docker-compose.yml
nano docker-compose.yml

# Under catalyst service, add:
#   deploy:
#     resources:
#       limits:
#         cpus: '2.0'
#         memory: 4G
#       reservations:
#         cpus: '1.0'
#         memory: 2G

# Apply changes
docker compose up -d
```

#### Horizontal Scaling (Multiple Instances)
```bash
# Scale to 3 instances
docker compose up -d --scale catalyst=3

# Add nginx load balancing (edit nginx.conf)
```

---

## Monitoring

### Health Checks

#### Basic Health
```bash
curl http://localhost:8000/health
```

Expected response:
```json
{
  "status": "healthy",
  "timestamp": "2025-01-13T10:30:00Z",
  "version": "1.0.0"
}
```

#### Readiness Check (with dependencies)
```bash
curl http://localhost:8000/health/ready
```

Expected response:
```json
{
  "ready": true,
  "timestamp": "2025-01-13T10:30:00Z",
  "checks": {
    "redis": {"status": "healthy"},
    "idr": {"status": "disabled"}
  }
}
```

### Prometheus Metrics
```bash
# View metrics
curl http://localhost:8000/metrics

# Key metrics to watch:
# - catalyst_auctions_total
# - catalyst_auction_duration_ms
# - catalyst_ivt_flagged_total
# - catalyst_bidder_timeouts_total
```

### Performance Monitoring
```bash
# Container resource usage
docker stats catalyst-catalyst-1

# Detailed stats
docker compose exec catalyst sh -c 'top -b -n 1'

# Network connections
docker compose exec catalyst ss -tunap
```

### Log Analysis
```bash
# Count requests per minute
docker compose logs --since 1h catalyst | grep "HTTP request" | wc -l

# Find errors
docker compose logs --since 1h catalyst | grep "level\":\"error\""

# IVT detections
docker compose logs catalyst | grep "IVT detected"

# Slow requests (>100ms)
docker compose logs catalyst | grep "duration_ms" | awk -F'duration_ms":' '{print $2}' | awk '{print $1}' | awk '$1 > 100'
```

---

## Troubleshooting

### Common Issues

#### Issue: Containers won't start
```bash
# Check logs
docker compose logs

# Check for port conflicts
sudo netstat -tlnp | grep -E ':(80|443|8000|6379)'

# Remove old containers
docker compose down
docker compose up -d
```

#### Issue: Can't reach https://catalyst.springwire.ai
```bash
# Check nginx is running
docker compose ps nginx

# Check nginx logs
docker compose logs nginx

# Verify DNS
dig catalyst.springwire.ai

# Test direct connection
curl http://catalyst.springwire.ai

# Check firewall
sudo ufw status
```

#### Issue: SSL certificate errors
```bash
# Check certificate exists
ls -la /opt/catalyst/ssl/

# Verify certificate
docker compose exec nginx nginx -t

# Re-create certificate
docker compose run --rm certbot certonly --webroot \
  --webroot-path=/var/www/certbot \
  --email your-email@example.com \
  --agree-tos \
  -d catalyst.springwire.ai --force-renewal

docker compose restart nginx
```

#### Issue: Redis connection failed
```bash
# Check Redis is running
docker compose ps redis

# Test Redis connection
docker compose exec redis redis-cli ping

# Check Redis logs
docker compose logs redis

# Restart Redis
docker compose restart redis
```

#### Issue: High memory usage
```bash
# Check which container is using memory
docker stats

# Check Redis memory
docker compose exec redis redis-cli INFO memory

# Flush Redis cache (CAUTION: clears all data)
docker compose exec redis redis-cli FLUSHDB
```

#### Issue: Auction endpoint returns errors
```bash
# Check Catalyst logs
docker compose logs -f catalyst

# Test with simple request
curl -X POST http://localhost:8000/openrtb2/auction \
  -H "Content-Type: application/json" \
  -d '{"id":"test","imp":[{"id":"1","banner":{"w":300,"h":250}}],"site":{"domain":"test.com"}}'

# Check publisher authentication
docker compose exec redis redis-cli HGETALL tne_catalyst:publishers
```

### Debug Mode

Enable debug logging:
```bash
# Edit .env
nano /opt/catalyst/.env

# Change:
LOG_LEVEL=debug

# Restart
docker compose restart catalyst

# View detailed logs
docker compose logs -f catalyst
```

### Getting Help

**Collect diagnostic information:**
```bash
# Create diagnostic report
cat > /opt/catalyst/diagnostic-report.txt << EOF
=== System Info ===
$(uname -a)
$(docker --version)
$(docker compose version)

=== Container Status ===
$(docker compose ps)

=== Resource Usage ===
$(docker stats --no-stream)

=== Recent Catalyst Logs ===
$(docker compose logs --tail=50 catalyst)

=== Recent Nginx Logs ===
$(docker compose logs --tail=50 nginx)

=== Redis Status ===
$(docker compose exec -T redis redis-cli INFO server)

=== Configuration ===
$(cat /opt/catalyst/.env | grep -v "PASSWORD\|SECRET\|KEY")
EOF

# View report
cat /opt/catalyst/diagnostic-report.txt
```

**Support Channels:**
- GitHub Issues: https://github.com/thenexusengine/tne_springwire/issues
- Documentation: https://docs.thenexusengine.com

---

## Security Best Practices

### Checklist

- [ ] Firewall configured (UFW or iptables)
- [ ] SSL certificate installed and auto-renewing
- [ ] SSH key authentication enabled
- [ ] SSH password authentication disabled
- [ ] Docker running rootless (optional but recommended)
- [ ] Redis password configured (optional for localhost)
- [ ] Regular backups configured
- [ ] Log rotation configured
- [ ] Rate limiting enabled in nginx
- [ ] CORS properly configured
- [ ] Keep Docker and images updated

### Hardening Steps

#### Disable SSH Password Authentication
```bash
sudo nano /etc/ssh/sshd_config

# Set these values:
PasswordAuthentication no
PermitRootLogin no
PubkeyAuthentication yes

# Restart SSH
sudo systemctl restart sshd
```

#### Configure Log Rotation
```bash
# Create logrotate config
sudo tee /etc/logrotate.d/catalyst << EOF
/opt/catalyst/nginx-logs/*.log {
    daily
    rotate 14
    compress
    delaycompress
    notifempty
    create 0640 www-data adm
    sharedscripts
    postrotate
        docker compose -f /opt/catalyst/docker-compose.yml restart nginx > /dev/null
    endscript
}
EOF
```

#### Set Up Fail2Ban (Optional)
```bash
# Install fail2ban
sudo apt install fail2ban -y

# Create nginx jail
sudo tee /etc/fail2ban/jail.d/nginx.conf << EOF
[nginx-limit-req]
enabled = true
filter = nginx-limit-req
logpath = /opt/catalyst/nginx-logs/*.log
maxretry = 10
findtime = 60
bantime = 3600
EOF

# Restart fail2ban
sudo systemctl restart fail2ban
```

---

## Appendix

### Complete File Checklist

Files needed in `/opt/catalyst/`:
- `docker-compose.yml` - Container orchestration
- `nginx.conf` - Reverse proxy configuration
- `.env` - Environment variables
- `ssl/` - SSL certificates directory
- `nginx-logs/` - Nginx log files
- `redis-data/` - Redis persistence

### Useful Commands Reference

```bash
# Start services
docker compose up -d

# Stop services
docker compose down

# Restart service
docker compose restart catalyst

# View logs
docker compose logs -f

# Execute command in container
docker compose exec catalyst sh

# Update images
docker compose pull

# Clean up old images
docker image prune -a

# Full cleanup (WARNING: removes all data)
docker compose down -v
docker system prune -a
```

### Port Reference

| Port | Service | Access |
|------|---------|--------|
| 80 | Nginx HTTP | External |
| 443 | Nginx HTTPS | External |
| 8000 | Catalyst | Internal only |
| 6379 | Redis | Internal only |

---

**Deployment Date**: _____________
**Version**: 1.0.0
**Last Updated**: 2025-01-13
