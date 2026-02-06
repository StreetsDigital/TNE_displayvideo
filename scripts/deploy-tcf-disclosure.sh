#!/bin/bash
# Automated deployment script for TCF Device Storage Disclosure

set -e

# Configuration
SERVER="18.209.163.224"
USER="ec2-user"
KEY="$HOME/.ssh/lightsail-catalyst.pem"
REMOTE_DIR="~/catalyst"

echo "=========================================="
echo "TCF Disclosure Deployment Script"
echo "=========================================="
echo ""

# Check if key file exists
if [ ! -f "$KEY" ]; then
    echo "❌ Error: SSH key not found at $KEY"
    exit 1
fi

echo "Step 1: Creating directories and uploading files to server..."
echo ""

# Create scripts directory if it doesn't exist
ssh -i "$KEY" "${USER}@${SERVER}" "mkdir -p ${REMOTE_DIR}/scripts"

# Upload TCF disclosure JSON
echo "📤 Uploading assets/tcf-disclosure.json..."
scp -i "$KEY" assets/tcf-disclosure.json "${USER}@${SERVER}:${REMOTE_DIR}/assets/"

# Upload TCF handler
echo "📤 Uploading internal/endpoints/tcf_disclosure.go..."
scp -i "$KEY" internal/endpoints/tcf_disclosure.go "${USER}@${SERVER}:${REMOTE_DIR}/internal/endpoints/"

# Upload updated server.go
echo "📤 Uploading cmd/server/server.go..."
scp -i "$KEY" cmd/server/server.go "${USER}@${SERVER}:${REMOTE_DIR}/cmd/server/"

# Upload test script
echo "📤 Uploading scripts/test-tcf-disclosure.sh..."
scp -i "$KEY" scripts/test-tcf-disclosure.sh "${USER}@${SERVER}:${REMOTE_DIR}/scripts/test-tcf-disclosure.sh"

echo ""
echo "✅ All files uploaded successfully"
echo ""

echo "Step 2: Building and restarting service on server..."
echo ""

# Build and restart on server
ssh -i "$KEY" "${USER}@${SERVER}" << 'ENDSSH'
cd ~/catalyst

echo "🔨 Building catalyst-server..."
make build

echo "🔄 Restarting catalyst service..."
docker-compose restart catalyst

echo "⏳ Waiting for service to start..."
sleep 5

echo "📋 Checking service logs..."
docker-compose logs --tail=20 catalyst | grep -i "tcf\|listening"

echo ""
echo "✅ Service restarted successfully"
ENDSSH

echo ""
echo "Step 3: Running validation tests..."
echo ""

# Run test script remotely
ssh -i "$KEY" "${USER}@${SERVER}" << 'ENDSSH'
cd ~/catalyst
chmod +x scripts/test-tcf-disclosure.sh
./scripts/test-tcf-disclosure.sh https://ads.thenexusengine.com
ENDSSH

echo ""
echo "=========================================="
echo "✅ TCF Disclosure Deployment Complete!"
echo "=========================================="
echo ""
echo "Next steps:"
echo "1. Validate at IAB tool: https://iabeurope.eu/vendorjson"
echo "2. Update publisher Sourcepoint CMP configuration"
echo "3. Monitor cookie sync success rates at /cookie_sync"
echo ""
echo "Endpoints available:"
echo "- https://ads.thenexusengine.com/.well-known/tcf-disclosure.json"
echo "- https://ads.thenexusengine.com/tcf-disclosure.json"
