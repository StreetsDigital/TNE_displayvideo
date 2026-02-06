# TCF Device Storage Disclosure - Deployment Guide

## Overview

This document describes the implementation and deployment of the IAB TCF v2 Device Storage Disclosure for ads.thenexusengine.com, enabling GDPR-compliant cookie syncing for all 26 programmatic bidders.

## What Was Implemented

### 1. TCF Disclosure JSON File
**File:** `assets/tcf-disclosure.json`

- **29 cookie disclosures:**
  - 26 programmatic bidder cookies (Kargo, Rubicon, PubMatic, etc.)
  - 3 internal Nexus Engine cookies (_nxs_uid, _nxs_session, _nxs_consent)
- **28 domain declarations:** All bidder sync domains properly documented
- **IAB TCF v2 compliant:** Follows Device Storage Disclosure v1.1 specification

### 2. HTTP Handler
**File:** `internal/endpoints/tcf_disclosure.go`

Features:
- Serves static JSON file with proper headers
- CORS enabled for CMP access (`Access-Control-Allow-Origin: *`)
- 24-hour cache headers (`Cache-Control: public, max-age=86400`)
- OPTIONS preflight handling
- Proper content type (`application/json; charset=utf-8`)

### 3. Server Routes
**File:** `cmd/server/server.go`

Two endpoints registered:
- `/.well-known/tcf-disclosure.json` (IAB standard path)
- `/tcf-disclosure.json` (convenience path)

### 4. Test Script
**File:** `scripts/test-tcf-disclosure.sh`

Validates:
- HTTP 200 responses
- CORS headers
- Content-Type headers
- Cache-Control headers
- JSON structure
- Key bidder declarations

## Deployment Instructions

### Step 1: Upload Files to Server

```bash
# SSH to server
ssh ec2-user@18.209.163.224 -i ~/.ssh/lightsail-catalyst.pem

# Navigate to catalyst directory
cd ~/catalyst

# Exit for now - we'll upload from local
exit
```

### Step 2: Upload from Local Machine

```bash
# Upload TCF disclosure JSON
scp -i ~/.ssh/lightsail-catalyst.pem \
  assets/tcf-disclosure.json \
  ec2-user@18.209.163.224:~/catalyst/assets/

# Upload TCF handler
scp -i ~/.ssh/lightsail-catalyst.pem \
  internal/endpoints/tcf_disclosure.go \
  ec2-user@18.209.163.224:~/catalyst/internal/endpoints/

# Upload updated server.go
scp -i ~/.ssh/lightsail-catalyst.pem \
  cmd/server/server.go \
  ec2-user@18.209.163.224:~/catalyst/cmd/server/

# Upload test script
scp -i ~/.ssh/lightsail-catalyst.pem \
  scripts/test-tcf-disclosure.sh \
  ec2-user@18.209.163.224:~/catalyst/scripts/
```

### Step 3: Rebuild and Restart on Server

```bash
# SSH back to server
ssh ec2-user@18.209.163.224 -i ~/.ssh/lightsail-catalyst.pem

# Navigate to catalyst directory
cd ~/catalyst

# Build the new binary
make build

# Restart the service with docker-compose
docker-compose restart catalyst

# Check logs for successful startup
docker-compose logs -f catalyst | head -50
```

### Step 4: Verify Deployment

```bash
# Run test script
chmod +x scripts/test-tcf-disclosure.sh
./scripts/test-tcf-disclosure.sh https://ads.thenexusengine.com
```

Expected output:
```
Testing TCF Disclosure endpoints on https://ads.thenexusengine.com
==========================================================

Test 1: Standard .well-known path
✅ Status: 200 (OK)
✅ Disclosures: 29
✅ Domains: 28

Test 2: Alternative root path
✅ Status: 200 (OK)

Test 3: CORS headers
✅ CORS header present: access-control-allow-origin: *

Test 4: Content-Type header
✅ Content-Type correct: content-type: application/json; charset=utf-8

Test 5: Cache-Control header
✅ Cache-Control present: cache-control: public, max-age=86400

Test 6: Validate JSON structure
✅ JSON structure valid

Test 7: Check for key bidders
✅ Kargo declared (kuid)
✅ Rubicon declared (rubiconproject_uid)
✅ PubMatic declared (KRTBCOOKIE_*)

==========================================================
✅ All TCF disclosure tests passed!
```

### Step 5: Manual Verification

```bash
# Test directly with curl
curl -I https://ads.thenexusengine.com/tcf-disclosure.json

# View JSON content
curl https://ads.thenexusengine.com/tcf-disclosure.json | jq .

# Count disclosures
curl -s https://ads.thenexusengine.com/tcf-disclosure.json | jq '.disclosures | length'
# Should return: 29

# Count domains
curl -s https://ads.thenexusengine.com/tcf-disclosure.json | jq '.domains | length'
# Should return: 28
```

## IAB Validation

After deployment, validate the disclosure file using the IAB tool:

1. Visit: https://iabeurope.eu/vendorjson
2. Enter URL: `https://ads.thenexusengine.com/tcf-disclosure.json`
3. Click "Validate"
4. Ensure all fields pass TCF v2 schema validation

## Publisher Integration

### Updating Sourcepoint CMP Configuration

Publishers need to update their Sourcepoint CMP configuration to reference the new disclosure file:

```javascript
// In Sourcepoint privacy manager configuration
{
  "accountId": 1234,
  "propertyId": 5678,
  "deviceStorageDisclosureUrl": "https://ads.thenexusengine.com/tcf-disclosure.json",
  // ... other config
}
```

### Testing User Consent Flow

After updating CMP configuration:

1. Clear browser cookies
2. Visit publisher site
3. CMP should show consent prompt
4. Verify all 26 bidders appear in vendor list
5. Accept consent
6. Verify cookie sync pixels fire from bidders
7. Check `/cookie_sync` endpoint for successful syncs

## Bidders Declared

The following 26 programmatic bidders are declared in the TCF disclosure:

| Bidder | GVL ID | Cookie Name | Domain |
|--------|--------|-------------|--------|
| Kargo | 972 | kuid | *.krxd.net |
| Rubicon/Magnite | 52 | rubiconproject_uid | *.rubiconproject.com |
| PubMatic | 76 | KRTBCOOKIE_* | *.pubmatic.com |
| Sovrn | 13 | ljt_reader | *.lijit.com |
| TripleLift | 28 | tluid | *.3lift.com |
| AppNexus/Xandr | 32 | uuid2 | *.adnxs.com |
| Index Exchange | 10 | CMID | *.casalemedia.com |
| OpenX | 69 | i | *.openx.net |
| Criteo | 91 | cto_bundle | *.criteo.com |
| 33Across | 58 | 33x_ps | *.33across.com |
| Aniview | 780 | aniview_uid | *.aniview.com |
| Adform | 50 | uid | *.adform.net |
| Beachfront | 335 | beach_uid | *.beachfront.com |
| Conversant | 24 | cnvr_uid | *.conversantmedia.com |
| GumGum | 61 | gguid | *.gumgum.com |
| Improve Digital | 253 | id_sync | *.360yield.com |
| Media.net | 142 | media_net_uid | *.media.net |
| OMS | 883 | oms_uid | *.adserver.com |
| OneTag | 241 | onetag_sync | *.onetag.com |
| Outbrain | 164 | ob_uid | *.outbrain.com |
| Sharethrough | 80 | stx_uid | *.sharethrough.com |
| Smart Ad Server | 45 | sas_uid | *.smartadserver.com |
| SpotX | 165 | spotx_uid | *.spotxchange.com |
| Taboola | 42 | t_gid | *.taboola.com |
| Teads | 132 | _tfpvi | *.teads.tv |
| Unruly | 36 | unruly_uid | *.video.unrulymedia.com |

Plus 3 internal Nexus Engine cookies:
- `_nxs_uid` - User identifier
- `_nxs_session` - Session cookie
- `_nxs_consent` - Consent preferences

## TCF Purpose IDs

The disclosure uses these IAB TCF purpose IDs:

- **Purpose 1:** Store and/or access information on a device
- **Purpose 2:** Select basic ads
- **Purpose 3:** Create a personalised ads profile
- **Purpose 4:** Select personalised ads
- **Purpose 5:** Create a personalised content profile
- **Purpose 7:** Measure ad performance

## Expected Impact

After successful deployment and publisher CMP configuration:

✅ **Compliance:**
- TCF v2 compliant (meets Feb 28, 2026 deadline)
- Proper vendor declarations for CMPs
- Legal basis for cookie syncing

✅ **Technical:**
- CMPs can fetch vendor list
- Programmatic bidders can write cookies
- Cookie sync success rate increases
- User ID recognition improves

✅ **Revenue:**
- Higher bid CPMs from improved user recognition
- More bidders able to participate in auctions
- Better match rates across publisher inventory

## Monitoring

After deployment, monitor:

1. **Cookie sync endpoint:** `/cookie_sync` success rate
2. **Bidder participation:** Number of bidders responding to auctions
3. **CPM improvements:** Average bid prices pre/post deployment
4. **User sync rates:** Percentage of users with synced IDs

## Troubleshooting

### Issue: 404 Not Found

```bash
# Check if file exists on server
ssh ec2-user@18.209.163.224 -i ~/Downloads/lightsail-catalyst.pem
ls -la ~/catalyst/assets/tcf-disclosure.json

# If missing, re-upload
exit
scp -i ~/Downloads/lightsail-catalyst.pem assets/tcf-disclosure.json ec2-user@18.209.163.224:~/catalyst/assets/
```

### Issue: CORS Headers Missing

Check nginx configuration for proxy_pass headers. The Go handler sets CORS headers, but nginx may override them.

### Issue: IAB Validation Fails

- Ensure JSON syntax is valid: `cat assets/tcf-disclosure.json | jq .`
- Check all required fields are present
- Verify maxAgeSeconds values are integers, not strings
- Confirm purposes are arrays of integers

## References

- [IAB TCF v2 Framework](https://github.com/InteractiveAdvertisingBureau/GDPR-Transparency-and-Consent-Framework)
- [Device Storage Disclosure Spec](https://github.com/InteractiveAdvertisingBureau/GDPR-Transparency-and-Consent-Framework/blob/master/TCFv2/IAB%20Tech%20Lab%20-%20Device%20Storage%20Disclosure.md)
- [TCF Validation Tool](https://iabeurope.eu/vendorjson)
- [Sourcepoint CMP Docs](https://docs.sourcepoint.com/)

## Files Modified

1. **Created:** `assets/tcf-disclosure.json` - TCF disclosure JSON with 29 cookie declarations
2. **Created:** `internal/endpoints/tcf_disclosure.go` - HTTP handler for serving disclosure
3. **Modified:** `cmd/server/server.go` - Added route registration for TCF endpoints
4. **Created:** `scripts/test-tcf-disclosure.sh` - Test script for validation
5. **Created:** `docs/TCF_DISCLOSURE_DEPLOYMENT.md` - This deployment guide
