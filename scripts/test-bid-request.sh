#!/bin/bash

# Test Catalyst bid request with all ad units from mapping

echo "=== Catalyst Bid Request Test ==="
echo ""

# Test localhost or production
HOST="${1:-https://ads.thenexusengine.com}"

echo "Testing against: $HOST"
echo ""

# Test 1: Single ad unit (leaderboard)
echo "Test 1: Single ad unit (leaderboard)"
echo "---"
curl -s -X POST "$HOST/v1/bid" \
  -H "Content-Type: application/json" \
  -d '{
    "accountId": "icisic-media",
    "timeout": 2800,
    "slots": [{
      "divId": "test-leaderboard",
      "sizes": [[728, 90], [970, 90]],
      "adUnitPath": "totalprosports.com/leaderboard"
    }]
  }' | jq '.'
echo ""

# Test 2: Multiple ad units
echo "Test 2: Multiple ad units (leaderboard + rectangle)"
echo "---"
curl -s -X POST "$HOST/v1/bid" \
  -H "Content-Type: application/json" \
  -d '{
    "accountId": "icisic-media",
    "timeout": 2800,
    "slots": [
      {
        "divId": "test-leaderboard",
        "sizes": [[728, 90], [970, 90]],
        "adUnitPath": "totalprosports.com/leaderboard"
      },
      {
        "divId": "test-rectangle",
        "sizes": [[300, 250]],
        "adUnitPath": "totalprosports.com/rectangle-medium"
      }
    ],
    "page": {
      "url": "https://totalprosports.com/test",
      "domain": "totalprosports.com",
      "keywords": ["sports", "news"],
      "categories": ["IAB17"]
    }
  }' | jq '.'
echo ""

# Test 3: Unknown ad unit (should warn in logs but not fail)
echo "Test 3: Unknown ad unit (should return empty bids)"
echo "---"
curl -s -X POST "$HOST/v1/bid" \
  -H "Content-Type: application/json" \
  -d '{
    "accountId": "icisic-media",
    "timeout": 2800,
    "slots": [{
      "divId": "test-unknown",
      "sizes": [[728, 90]],
      "adUnitPath": "unknown.com/test"
    }]
  }' | jq '.'
echo ""

echo "=== Test Complete ==="
echo ""
echo "Expected results:"
echo "- Test 1 & 2: Should return bid response (may have 0 bids initially)"
echo "- Test 3: Should return empty bids with warning in logs"
echo ""
echo "Check server logs for bidder parameter injection:"
echo "  ssh user@ads.thenexusengine.com 'sudo journalctl -u catalyst -n 100'"
