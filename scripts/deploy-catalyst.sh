#!/bin/bash
set -e

# Catalyst Deployment Script
# Deploys Catalyst server with bidder mapping to ads.thenexusengine.com

SERVER="user@ads.thenexusengine.com"
DEPLOY_DIR="/opt/catalyst"
PACKAGE="build/catalyst-deployment.tar.gz"

echo "=== Catalyst Deployment ==="
echo ""

# Check package exists
if [ ! -f "$PACKAGE" ]; then
    echo "Error: Deployment package not found at $PACKAGE"
    echo "Run: tar czf build/catalyst-deployment.tar.gz build/catalyst-server assets/catalyst-sdk.js assets/tne-ads.js config/bizbudding-all-bidders-mapping.json"
    exit 1
fi

echo "Step 1: Uploading deployment package..."
scp "$PACKAGE" "$SERVER:/tmp/catalyst-deployment.tar.gz"
echo "✓ Upload complete"
echo ""

echo "Step 2: Deploying on server..."
ssh "$SERVER" << 'REMOTE_SCRIPT'
set -e

cd /opt/catalyst

# Stop service
echo "  - Stopping catalyst service..."
sudo systemctl stop catalyst

# Backup current version
echo "  - Backing up current version..."
if [ -f catalyst-server ]; then
    sudo cp catalyst-server catalyst-server.backup.$(date +%Y%m%d-%H%M%S)
fi

# Extract new version
echo "  - Extracting new version..."
sudo tar xzf /tmp/catalyst-deployment.tar.gz --strip-components=1

# Set permissions
echo "  - Setting permissions..."
sudo chmod +x build/catalyst-server
sudo mv build/catalyst-server ./catalyst-server

# Verify files exist
echo "  - Verifying deployment..."
ls -lh catalyst-server config/bizbudding-all-bidders-mapping.json

# Start service
echo "  - Starting catalyst service..."
sudo systemctl start catalyst

# Wait for startup
sleep 3

# Health check
echo "  - Running health check..."
curl -s http://localhost:8000/health || echo "Warning: Health check failed"

echo ""
echo "✓ Deployment complete!"
REMOTE_SCRIPT

echo ""
echo "Step 3: Verifying deployment..."
echo ""

# Test endpoints
echo "Testing health endpoint..."
curl -s https://ads.thenexusengine.com/health | jq . || echo "Warning: Health check failed"
echo ""

echo "Testing SDK endpoint..."
curl -s -I https://ads.thenexusengine.com/assets/catalyst-sdk.js | head -5
echo ""

echo "Testing bid endpoint..."
curl -s -X POST https://ads.thenexusengine.com/v1/bid \
  -H "Content-Type: application/json" \
  -d '{
    "accountId": "icisic-media",
    "timeout": 2800,
    "slots": [{
      "divId": "test-div",
      "sizes": [[728, 90]],
      "adUnitPath": "totalprosports.com/leaderboard"
    }]
  }' | jq . || echo "Warning: Bid endpoint test failed"
echo ""

echo ""
echo "=== Deployment Summary ==="
echo "✓ Binary deployed to /opt/catalyst/catalyst-server"
echo "✓ Mapping deployed to /opt/catalyst/config/bizbudding-all-bidders-mapping.json"
echo "✓ Service restarted"
echo ""
echo "Next steps:"
echo "1. Monitor logs: ssh $SERVER 'sudo journalctl -u catalyst -f'"
echo "2. Test in browser: https://ads.thenexusengine.com/test-magnite.html"
echo "3. Check metrics: https://ads.thenexusengine.com/metrics"
