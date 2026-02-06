#!/bin/bash
# Test script for TCF Device Storage Disclosure endpoint

set -e

HOST="${1:-https://ads.thenexusengine.com}"

echo "Testing TCF Disclosure endpoints on $HOST"
echo "=========================================="
echo ""

# Test 1: Standard .well-known path
echo "Test 1: Standard .well-known path"
echo "URL: $HOST/.well-known/tcf-disclosure.json"
RESPONSE=$(curl -s -w "\n%{http_code}" "$HOST/.well-known/tcf-disclosure.json")
HTTP_CODE=$(echo "$RESPONSE" | tail -1)
BODY=$(echo "$RESPONSE" | head -n -1)

if [ "$HTTP_CODE" = "200" ]; then
    echo "✅ Status: $HTTP_CODE (OK)"
    DISCLOSURE_COUNT=$(echo "$BODY" | jq '.disclosures | length')
    DOMAIN_COUNT=$(echo "$BODY" | jq '.domains | length')
    echo "✅ Disclosures: $DISCLOSURE_COUNT"
    echo "✅ Domains: $DOMAIN_COUNT"
else
    echo "❌ Status: $HTTP_CODE (Expected 200)"
    exit 1
fi
echo ""

# Test 2: Alternative root path
echo "Test 2: Alternative root path"
echo "URL: $HOST/tcf-disclosure.json"
RESPONSE=$(curl -s -w "\n%{http_code}" "$HOST/tcf-disclosure.json")
HTTP_CODE=$(echo "$RESPONSE" | tail -1)

if [ "$HTTP_CODE" = "200" ]; then
    echo "✅ Status: $HTTP_CODE (OK)"
else
    echo "❌ Status: $HTTP_CODE (Expected 200)"
    exit 1
fi
echo ""

# Test 3: CORS headers
echo "Test 3: CORS headers"
CORS_HEADER=$(curl -s -I "$HOST/tcf-disclosure.json" | grep -i "access-control-allow-origin" || echo "")
if [ -n "$CORS_HEADER" ]; then
    echo "✅ CORS header present: $CORS_HEADER"
else
    echo "❌ CORS header missing"
    exit 1
fi
echo ""

# Test 4: Content-Type header
echo "Test 4: Content-Type header"
CONTENT_TYPE=$(curl -s -I "$HOST/tcf-disclosure.json" | grep -i "content-type" || echo "")
if echo "$CONTENT_TYPE" | grep -q "application/json"; then
    echo "✅ Content-Type correct: $CONTENT_TYPE"
else
    echo "❌ Content-Type incorrect: $CONTENT_TYPE"
    exit 1
fi
echo ""

# Test 5: Cache-Control header
echo "Test 5: Cache-Control header"
CACHE_CONTROL=$(curl -s -I "$HOST/tcf-disclosure.json" | grep -i "cache-control" || echo "")
if echo "$CACHE_CONTROL" | grep -q "max-age"; then
    echo "✅ Cache-Control present: $CACHE_CONTROL"
else
    echo "⚠️  Cache-Control missing or incorrect: $CACHE_CONTROL"
fi
echo ""

# Test 6: Validate JSON structure
echo "Test 6: Validate JSON structure"
BODY=$(curl -s "$HOST/tcf-disclosure.json")

# Check required fields
HAS_DISCLOSURES=$(echo "$BODY" | jq 'has("disclosures")')
HAS_DOMAINS=$(echo "$BODY" | jq 'has("domains")')

if [ "$HAS_DISCLOSURES" = "true" ] && [ "$HAS_DOMAINS" = "true" ]; then
    echo "✅ JSON structure valid"
else
    echo "❌ JSON structure invalid (missing required fields)"
    exit 1
fi
echo ""

# Test 7: Check for key bidders
echo "Test 7: Check for key bidders"
KARGO=$(echo "$BODY" | jq '.disclosures[] | select(.identifier == "kuid") | .identifier')
RUBICON=$(echo "$BODY" | jq '.disclosures[] | select(.identifier == "rubiconproject_uid") | .identifier')
PUBMATIC=$(echo "$BODY" | jq '.disclosures[] | select(.identifier == "KRTBCOOKIE_*") | .identifier')

if [ -n "$KARGO" ]; then
    echo "✅ Kargo declared (kuid)"
else
    echo "❌ Kargo missing"
fi

if [ -n "$RUBICON" ]; then
    echo "✅ Rubicon declared (rubiconproject_uid)"
else
    echo "❌ Rubicon missing"
fi

if [ -n "$PUBMATIC" ]; then
    echo "✅ PubMatic declared (KRTBCOOKIE_*)"
else
    echo "❌ PubMatic missing"
fi
echo ""

echo "=========================================="
echo "✅ All TCF disclosure tests passed!"
echo ""
echo "Next steps:"
echo "1. Update publisher Sourcepoint CMP configuration"
echo "2. Reference: $HOST/tcf-disclosure.json"
echo "3. Validate at: https://iabeurope.eu/vendorjson"
