# Lightsail + Docker Deployment

Deploy Catalyst to AWS Lightsail using Docker containers.

## Why Lightsail?

- ✅ **Simpler** than EC2 - fixed pricing, easy management
- ✅ **Cheaper** - $20/month for 2GB RAM instance
- ✅ **Docker-ready** - runs containers out of the box
- ✅ **Static IP included** - free static IP address
- ✅ **Predictable costs** - no surprise bills

## What Gets Deployed

**Lightsail Instance:**
- Location: us-east-1 (N. Virginia) - close to major SSPs
- Plan: Medium (2GB RAM, 2 vCPU, 60GB SSD)
- Cost: $20/month

**Docker Containers:**
- `catalyst-server` - Go application
- `nginx` - Reverse proxy with SSL
- `certbot` - Automatic SSL renewal

## Prerequisites

1. AWS CLI installed and configured
2. Docker installed locally (for building image)
3. DNS access to update ads.thenexusengine.com

## Quick Start (3 Steps)

### Step 1: Create Lightsail Instance

```bash
cd infrastructure/lightsail
./deploy.sh
```

This will:
- Create Lightsail instance in us-east-1
- Allocate static IP
- Configure firewall (ports 22, 80, 443, 8000)
- Download SSH key

Time: ~3 minutes

### Step 2: Update DNS

Point your domain to the static IP:
```
A record: ads.thenexusengine.com → <IP_FROM_OUTPUT>
```

Wait 5-10 minutes for DNS propagation.

### Step 3: Deploy Docker

```bash
./deploy-docker.sh
```

This will:
- Install Docker on the instance
- Build Catalyst Docker image
- Upload and start containers
- Configure Nginx

Time: ~5 minutes

### Step 4: Setup SSL (after DNS propagation)

```bash
# SSH to server
ssh -i ~/.ssh/lightsail-us-east-1.pem ec2-user@<IP>

# Run SSL setup
cd ~/catalyst
docker-compose run --rm certbot certonly \
  --webroot \
  --webroot-path=/var/www/certbot \
  --email your@email.com \
  --agree-tos \
  --no-eff-email \
  -d ads.thenexusengine.com

# Update nginx config for HTTPS
# Edit nginx/conf.d/catalyst.conf to add SSL server block

# Restart nginx
docker-compose restart nginx
```

## Manual Deployment

If you prefer step-by-step control:

### 1. Create Instance via AWS Console

Go to: https://lightsail.aws.amazon.com

1. Click "Create instance"
2. Select "Linux/Unix"
3. Select "OS Only" → "Amazon Linux 2"
4. Choose "Medium" plan ($20/month)
5. Name it "catalyst-server"
6. Click "Create instance"

### 2. Configure Networking

1. Click on instance → "Networking"
2. Create static IP and attach it
3. Add firewall rules:
   - SSH (22)
   - HTTP (80)
   - HTTPS (443)
   - Custom (8000)

### 3. Connect and Deploy

```bash
# Download SSH key from Lightsail console
# Then deploy Docker:
cd infrastructure/lightsail
./deploy-docker.sh <STATIC_IP>
```

## Managing Your Instance

### SSH Access

```bash
ssh -i ~/.ssh/lightsail-us-east-1.pem ec2-user@<IP>
```

### Docker Commands

```bash
cd ~/catalyst

# View logs
docker-compose logs -f catalyst
docker-compose logs -f nginx

# Restart services
docker-compose restart catalyst
docker-compose restart nginx

# Stop all
docker-compose down

# Start all
docker-compose up -d

# Rebuild after code changes
docker-compose build
docker-compose up -d
```

### Update Catalyst

To deploy new version:

```bash
# On local machine - build new image
docker build -t catalyst-server:latest .
docker save catalyst-server:latest | gzip > /tmp/catalyst-image.tar.gz

# Upload to server
scp -i ~/.ssh/lightsail-us-east-1.pem /tmp/catalyst-image.tar.gz ec2-user@<IP>:~/catalyst/

# On server - reload and restart
ssh -i ~/.ssh/lightsail-us-east-1.pem ec2-user@<IP>
cd ~/catalyst
docker load < catalyst-image.tar.gz
docker-compose up -d
```

## Monitoring

### View Logs

```bash
# All logs
docker-compose logs -f

# Just Catalyst
docker-compose logs -f catalyst

# Just Nginx
docker-compose logs -f nginx

# Last 100 lines
docker-compose logs --tail=100 catalyst
```

### Check Health

```bash
# From server
curl http://localhost:8000/health

# From internet
curl https://ads.thenexusengine.com/health
```

### Resource Usage

```bash
# Container stats
docker stats

# Disk usage
df -h

# Memory usage
free -h
```

## Costs

| Item | Cost |
|------|------|
| Lightsail Medium (2GB) | $20/month |
| Static IP (attached) | $0 |
| Data transfer (1TB) | Included |
| Backups (optional) | $1/snapshot |
| **Total** | **$20/month** |

Way cheaper than EC2! 🎉

## Backups

### Create Snapshot

```bash
aws lightsail create-instance-snapshot \
  --instance-name catalyst-server \
  --instance-snapshot-name catalyst-backup-$(date +%Y%m%d) \
  --region us-east-1
```

### Restore from Snapshot

1. Go to Lightsail console
2. Click "Snapshots"
3. Select snapshot
4. Click "Create new instance"

## Upgrading Instance Size

If you need more resources:

```bash
# Create snapshot first
aws lightsail create-instance-snapshot \
  --instance-name catalyst-server \
  --instance-snapshot-name before-upgrade \
  --region us-east-1

# Create new larger instance from snapshot
# Via console or CLI
```

Available plans:
- Small (1GB): $10/month
- Medium (2GB): $20/month ← recommended
- Large (4GB): $40/month
- XLarge (8GB): $80/month

## Troubleshooting

### Can't connect to instance
```bash
# Check instance status
aws lightsail get-instance --instance-name catalyst-server --region us-east-1

# Check firewall
aws lightsail get-instance-port-states --instance-name catalyst-server --region us-east-1
```

### Docker not working
```bash
# Check Docker status
sudo systemctl status docker

# Restart Docker
sudo systemctl restart docker
```

### Nginx errors
```bash
# Check nginx logs
docker-compose logs nginx

# Test nginx config
docker-compose exec nginx nginx -t

# Restart nginx
docker-compose restart nginx
```

### SSL issues
```bash
# Check certificates
docker-compose exec nginx ls -la /etc/letsencrypt/live/ads.thenexusengine.com/

# Test SSL
curl -vI https://ads.thenexusengine.com
```

## Advantages over EC2

- ✅ Fixed pricing (no surprise bills)
- ✅ Easier to manage
- ✅ Free static IP
- ✅ Simpler interface
- ✅ 1TB data transfer included
- ✅ Better for predictable workloads

## When to Upgrade to EC2

Consider EC2 if you need:
- Auto-scaling
- Load balancing across multiple instances
- VPC peering
- Complex networking
- Spot instances for cost savings

For most ad servers, Lightsail is perfect! 🚀

## Support

For issues:
1. Check Lightsail console
2. View Docker logs: `docker-compose logs`
3. SSH and investigate: `ssh -i ~/.ssh/lightsail-us-east-1.pem ec2-user@<IP>`
