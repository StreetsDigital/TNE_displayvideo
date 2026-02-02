# TNE Catalyst Deployment Script Guide

Quick guide for deploying to **thenexusengine.com** using the automated deployment script.

## Prerequisites

Before running the deployment script, ensure you have:

### 1. Server Requirements
- Ubuntu 20.04+ or Debian 11+ (recommended)
- Minimum 2GB RAM, 2 CPU cores
- 20GB+ available disk space
- Root or sudo access

### 2. DNS Configuration
Configure these A records in your DNS provider:
```
ads.thenexusengine.com         →  [YOUR_SERVER_IP]
staging.thenexusengine.com     →  [YOUR_SERVER_IP]
grafana.thenexusengine.com     →  [YOUR_SERVER_IP]
prometheus.thenexusengine.com  →  [YOUR_SERVER_IP]
```

### 3. Firewall/Security Group
Open these ports:
- **22** - SSH (for management)
- **80** - HTTP (for Let's Encrypt)
- **443** - HTTPS (for services)

### 4. Domain Email
Have access to: **ops@thenexusengine.io** (for SSL certificate notifications)

## Quick Start

### On Your Server:

```bash
# 1. Clone the repository
git clone https://github.com/thenexusengine/tne_springwire.git
cd tne_springwire/deployment

# 2. Run the deployment script
sudo ./deploy-to-thenexusengine.sh
```

That's it! The script will guide you through the entire deployment process.

## What the Script Does

The deployment script automates these steps:

### ✅ Step 1: Prerequisites Check
- Verifies Docker, Docker Compose, Git, Certbot are installed
- Checks system requirements

### ✅ Step 2: DNS Configuration Check
- Verifies DNS records are configured
- Shows server IP address
- Tests DNS resolution

### ✅ Step 3: Deployment Directory Setup
- Creates `/opt/catalyst/` directory
- Copies all deployment files
- Sets up directory structure

### ✅ Step 4: SSL Certificate Setup
- Obtains Let's Encrypt certificates for all domains
- Configures auto-renewal
- Places certificates in correct locations

### ✅ Step 5: Environment Configuration
- Creates `.env` from `.env.production`
- Generates secure secrets (if needed)
- Validates configuration

### ✅ Step 6: Database Setup
- Verifies database migrations exist
- Prepares database initialization

### ✅ Step 7: Docker Deployment
- Pulls latest Docker images
- Starts all services
- Configures networking

### ✅ Step 8: Health Checks
- Verifies all services are running
- Tests health endpoints
- Confirms deployment success

### ✅ Step 9: Monitoring Setup
- Deploys Grafana and Prometheus
- Configures dashboards
- Sets up metrics collection

### ✅ Step 10: Firewall Configuration
- Configures UFW firewall rules
- Secures the server
- Enables firewall

## Deployment Modes

The script supports two deployment modes:

### Mode 1: Standard Deployment (Default)
- Single production environment
- 100% traffic to production
- Simpler, recommended for most use cases

### Mode 2: Traffic Splitting
- Dual environment (production + staging)
- 95% traffic to production, 5% to staging
- For A/B testing and canary deployments

## After Deployment

### Verify Deployment

```bash
# Check service status
docker-compose ps

# Test health endpoint
curl https://ads.thenexusengine.com/health

# View logs
docker-compose logs -f
```

### Access Services

- **Main API**: https://ads.thenexusengine.com
- **Staging**: https://staging.thenexusengine.com
- **Grafana**: https://grafana.thenexusengine.com (admin/admin)
- **Prometheus**: https://prometheus.thenexusengine.com

### Important Next Steps

1. **Change Grafana Password**
   ```bash
   # Login to Grafana and change from admin/admin
   ```

2. **Configure Backups**
   ```bash
   cd /opt/catalyst
   ./setup-s3-backups.sh
   ```

3. **Test SSL Renewal**
   ```bash
   certbot renew --dry-run
   ```

4. **Add Publishers/Bidders**
   ```bash
   ./manage-publishers.sh
   ./manage-bidders.sh
   ```

## Troubleshooting

### DNS Not Resolving
**Problem**: DNS records not resolving to server IP

**Solution**:
```bash
# Check DNS propagation
dig ads.thenexusengine.com

# Wait 5-15 minutes for DNS propagation
# Re-run the deployment script
```

### SSL Certificate Failed
**Problem**: Let's Encrypt certificate generation failed

**Solution**:
```bash
# Ensure port 80 is open
sudo ufw allow 80/tcp

# Ensure DNS is pointing to server
dig ads.thenexusengine.com

# Manually obtain certificate
certbot certonly --standalone -d ads.thenexusengine.com
```

### Services Not Starting
**Problem**: Docker containers failing to start

**Solution**:
```bash
# Check logs
docker-compose logs

# Verify secrets are set
grep CHANGE_ME .env

# Restart services
docker-compose down
docker-compose up -d
```

### Port Already in Use
**Problem**: Port 80 or 443 already in use

**Solution**:
```bash
# Find process using port
sudo lsof -i :80
sudo lsof -i :443

# Stop conflicting service
sudo systemctl stop apache2  # if Apache is running
sudo systemctl stop nginx    # if Nginx is running

# Re-run deployment script
```

## Manual Deployment Steps

If you prefer manual deployment instead of using the script:

1. **Install Dependencies**
   ```bash
   sudo apt update
   sudo apt install -y docker.io docker-compose git certbot
   ```

2. **Setup Directory**
   ```bash
   sudo mkdir -p /opt/catalyst
   cd /opt/catalyst
   git clone https://github.com/thenexusengine/tne_springwire.git .
   cd deployment
   ```

3. **Generate Secrets**
   ```bash
   ./generate-secrets.sh
   ```

4. **Obtain SSL Certificates**
   ```bash
   sudo certbot certonly --standalone \
     -d ads.thenexusengine.com \
     -d staging.thenexusengine.com \
     -d grafana.thenexusengine.com \
     -d prometheus.thenexusengine.com
   ```

5. **Deploy Services**
   ```bash
   cp .env.production .env
   docker-compose up -d
   ```

## Updating Deployment

To update to a new version:

```bash
cd /opt/catalyst
git pull origin main
docker-compose pull
docker-compose up -d
```

## Rolling Back

If something goes wrong:

```bash
cd /opt/catalyst
docker-compose down
git checkout [previous-version-tag]
docker-compose up -d
```

## Security Best Practices

- ✅ Store secrets in password manager
- ✅ Enable automatic SSL renewal
- ✅ Configure S3 backups
- ✅ Set up log rotation
- ✅ Enable firewall
- ✅ Rotate secrets every 90 days
- ✅ Monitor logs for suspicious activity
- ✅ Keep Docker images updated

## Support

- **Documentation**: `/docs/deployment/`
- **Email**: ops@thenexusengine.io
- **Issues**: GitHub Issues

## Files Created by Script

After successful deployment, these files will exist:

```
/opt/catalyst/
├── .env                        # Environment configuration
├── .env.production.backup      # Backup of original config
├── docker-compose.yml          # Docker services definition
├── ssl/                        # SSL certificates
│   ├── fullchain.pem
│   └── privkey.pem
├── DEPLOYMENT_INFO.txt         # Deployment summary
├── SECRETS-BACKUP.txt          # Secrets backup (DELETE after storing)
└── ... (other deployment files)
```

## Quick Reference Commands

```bash
# Start services
docker-compose up -d

# Stop services
docker-compose down

# View logs
docker-compose logs -f

# Restart service
docker-compose restart catalyst

# Update services
git pull && docker-compose pull && docker-compose up -d

# Check health
curl https://ads.thenexusengine.com/health

# Run smoke tests
./smoke-tests.sh
```

---

**Last Updated**: 2026-02-02
**Script Version**: 1.0.0
**Domain**: ads.thenexusengine.com
