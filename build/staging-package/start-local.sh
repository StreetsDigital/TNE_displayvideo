#!/bin/bash
# Quick start script for local staging testing

set -euo pipefail

echo "================================================"
echo "  Catalyst Staging Server - Local Test"
echo "================================================"
echo ""

# Load environment
if [ -f .env ]; then
    echo "✓ Loading .env configuration"
    set -a
    source .env
    set +a
else
    echo "✗ .env file not found!"
    exit 1
fi

# Override for local testing
export PBS_PORT=8000
export PBS_HOST_URL="http://localhost:8000"

echo "✓ Configuration loaded"
echo "  Server URL: $PBS_HOST_URL"
echo "  Port: $PBS_PORT"
echo ""

# Check if port is available
if lsof -Pi :8000 -sTCP:LISTEN -t >/dev/null 2>&1; then
    echo "✗ Port 8000 is already in use!"
    echo "  Kill the process or choose a different port"
    exit 1
fi

echo "✓ Port 8000 is available"
echo ""

# Start server
echo "Starting Catalyst server..."
echo "Press Ctrl+C to stop"
echo ""
echo "Test endpoints:"
echo "  Health: http://localhost:8000/health"
echo "  SDK: http://localhost:8000/assets/catalyst-sdk.js"
echo "  Bid API: curl -X POST http://localhost:8000/v1/bid -H 'Content-Type: application/json' -d '{\"accountId\":\"test\",\"slots\":[{\"divId\":\"test\",\"sizes\":[[728,90]]}]}'"
echo ""

exec ./catalyst-server
