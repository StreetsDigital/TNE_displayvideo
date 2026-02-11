# Catalyst Staging Deployment Guide

## Quick Start

This package contains everything needed for staging deployment:

- `catalyst-server` - Production-optimized binary (18MB)
- `catalyst-sdk.js` - JavaScript SDK (2.9KB gzipped)
- `tne-ads.js` - TNE Ads SDK (1.9KB gzipped)
- `.env` - Staging environment configuration

## Pre-Deployment Checklist

- [x] Server binary built and tested
- [x] All unit tests passing (10/10)
- [x] All integration tests passing (3/3)
- [x] SDK files validated (< 50KB gzipped)
- [ ] Staging environment configured
- [ ] Database accessible
- [ ] Redis accessible
- [ ] CDN configured

## Option 1: Quick Local Staging Test

Test the server locally before deploying:

```bash
# 1. Start server locally with staging config
cd build/staging-package
chmod +x catalyst-server
source .env
./catalyst-server

# Server will start on port 8000

# 2. Test health endpoint (in another terminal)
curl http://localhost:8000/health

# 3. Test Catalyst SDK endpoint
curl http://localhost:8000/assets/catalyst-sdk.js

# 4. Test bid endpoint
curl -X POST http://localhost:8000/v1/bid \
  -H "Content-Type: application/json" \
  -d '{
    "accountId": "staging-test",
    "timeout": 2800,
    "slots": [{
      "divId": "test-slot",
      "sizes": [[728, 90]]
    }]
  }'

# Expected: {"bids":[],"responseTime":0} (empty bids OK if no bidders configured)

# 5. Open browser test page
# Navigate to: http://localhost:8000/../../tests/catalyst_sdk_test.html
# Or: python3 -m http.server 8080 in tests/ directory
```

## Option 2: Deploy to Staging Server

### Step 1: Upload Package to Server

```bash
# From your local machine
cd /Users/andrewstreets/tnevideo/build

# Create deployment archive
tar czf catalyst-staging.tar.gz staging-package/

# Upload to staging server (replace with your server address)
scp catalyst-staging.tar.gz user@staging.thenexusengine.com:/tmp/

# SSH to server
ssh user@staging.thenexusengine.com
```

### Step 2: Extract and Configure on Server

```bash
# On staging server
sudo mkdir -p /opt/catalyst-staging
sudo tar xzf /tmp/catalyst-staging.tar.gz -C /opt/catalyst-staging --strip-components=1

cd /opt/catalyst-staging

# Update .env with server-specific values
sudo nano .env

# Key variables to check:
# - PBS_HOST_URL=https://staging-ads.thenexusengine.com
# - DB_HOST=<your-db-host>
# - DB_PASSWORD=<your-db-password>
# - REDIS_HOST=<your-redis-host>
```

### Step 3: Create Systemd Service (Linux)

```bash
# Create systemd service
sudo cat > /etc/systemd/system/catalyst-staging.service << 'EOF'
[Unit]
Description=Catalyst Staging Server
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=catalyst
WorkingDirectory=/opt/catalyst-staging
EnvironmentFile=/opt/catalyst-staging/.env
ExecStart=/opt/catalyst-staging/catalyst-server
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=catalyst-staging

[Install]
WantedBy=multi-user.target
EOF

# Create user
sudo useradd -r -s /bin/false catalyst

# Set permissions
sudo chown -R catalyst:catalyst /opt/catalyst-staging

# Enable and start service
sudo systemctl enable catalyst-staging
sudo systemctl start catalyst-staging

# Check status
sudo systemctl status catalyst-staging

# View logs
sudo journalctl -u catalyst-staging -f
```

### Step 4: Configure Nginx Reverse Proxy

```bash
# Create nginx config
sudo cat > /etc/nginx/sites-available/catalyst-staging << 'EOF'
server {
    listen 80;
    server_name staging-ads.thenexusengine.com;

    location / {
        proxy_pass http://localhost:8000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # CORS headers (already set by server, but nginx can add as backup)
        add_header Access-Control-Allow-Origin * always;
    }
}
EOF

# Enable site
sudo ln -s /etc/nginx/sites-available/catalyst-staging /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### Step 5: Upload SDK to CDN

If using AWS CloudFront:

```bash
# From local machine
cd /Users/andrewstreets/tnevideo

# Gzip SDK
gzip -k assets/catalyst-sdk.js

# Upload to S3
aws s3 cp assets/catalyst-sdk.js.gz \
  s3://staging-cdn-bucket/assets/catalyst-sdk.js \
  --content-type "application/javascript" \
  --content-encoding gzip \
  --cache-control "public, max-age=3600"

# Invalidate CloudFront cache
aws cloudfront create-invalidation \
  --distribution-id YOUR_STAGING_DIST_ID \
  --paths "/assets/catalyst-sdk.js"

# Verify
curl -I https://staging-cdn.thenexusengine.com/assets/catalyst-sdk.js
```

### Step 6: Verify Deployment

```bash
# Health check
curl https://staging-ads.thenexusengine.com/health

# Readiness check
curl https://staging-ads.thenexusengine.com/health/ready

# SDK endpoint
curl https://staging-ads.thenexusengine.com/assets/catalyst-sdk.js | head -5

# Bid endpoint
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

# Metrics
curl https://staging-ads.thenexusengine.com/metrics | grep catalyst
```

### Step 7: Run Browser Tests

1. Open: `https://staging-ads.thenexusengine.com/../../tests/catalyst_sdk_test.html`
2. Or upload test HTML to your staging environment
3. Click "Run All Tests"
4. Verify all tests pass

## Option 3: Docker Deployment

Using existing docker-compose setup:

```bash
# From project root
cd /Users/andrewstreets/tnevideo

# Build Docker image
docker build -t catalyst-server:staging .

# Update docker-compose with catalyst-sdk.js
# Then deploy
cd deployment
docker-compose -f docker-compose.yml up -d

# Check logs
docker-compose logs -f catalyst
```

## Verification

After deployment, verify these endpoints work:

1. **Health**: `https://staging-ads.thenexusengine.com/health`
2. **SDK**: `https://staging-ads.thenexusengine.com/assets/catalyst-sdk.js`
3. **Bid API**: `https://staging-ads.thenexusengine.com/v1/bid`
4. **Metrics**: `https://staging-ads.thenexusengine.com/metrics`

## Next Steps

After successful staging deployment:

1. Share staging URLs with MAI Publisher:
   - SDK: `https://staging-cdn.thenexusengine.com/assets/catalyst-sdk.js`
   - API: `https://staging-ads.thenexusengine.com/v1/bid`
   - Account: `mai-staging-test`

2. Coordinate integration testing session

3. Monitor metrics during testing

4. Address any issues found

5. Proceed to production deployment

## Rollback

If issues occur:

```bash
# Stop service
sudo systemctl stop catalyst-staging

# Check logs
sudo journalctl -u catalyst-staging -n 100

# Restart with previous version
# (Keep previous binary as catalyst-server.backup)
```

## Support

- Deployment Guide: `../../docs/integrations/CATALYST_DEPLOYMENT_GUIDE.md`
- Integration Spec: `../../docs/integrations/BB_NEXUS-ENGINE-INTEGRATION-SPEC.md`
- Test Page: `../../tests/catalyst_sdk_test.html`
