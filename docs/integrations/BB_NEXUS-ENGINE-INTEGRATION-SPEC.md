# The Nexus Engine - MAI Publisher Integration Specification

**Version:** 1.0
**Date:** February 4, 2026
**Audience:** Nexus Engine Server-Side Platform Team
**Purpose:** Define exact integration requirements for Catalyst bidder with MAI Publisher

---

## Executive Summary

This document specifies how **The Nexus Engine's Catalyst server-side platform** must integrate with the MAI Publisher client-side JavaScript ad stack. The integration follows a client-server model where:

- **Client:** MAI Publisher JavaScript (browser-side)
- **Server:** Nexus Engine Catalyst platform (server-side bidding)
- **Protocol:** JavaScript SDK bridge + HTTP/HTTPS bid requests

The integration must coordinate with two existing bidders (Prebid.js and Amazon UAM) and respect a **2800ms timeout** for all bidding operations.

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Client-Side SDK Requirements](#client-side-sdk-requirements)
3. [Server-Side API Requirements](#server-side-api-requirements)
4. [Bid Request Format](#bid-request-format)
5. [Bid Response Format](#bid-response-format)
6. [Initialization Flow](#initialization-flow)
7. [Coordination Protocol](#coordination-protocol)
8. [Timeout Handling](#timeout-handling)
9. [GDPR Consent](#gdpr-consent)
10. [Error Handling](#error-handling)
11. [Testing Requirements](#testing-requirements)
12. [Performance SLA](#performance-sla)

---

## 1. Architecture Overview

### System Context

```
┌─────────────────────────────────────────────────────────────┐
│                    Browser (User's Device)                   │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐ │
│  │          MAI Publisher JavaScript Stack                 │ │
│  │                                                         │ │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────────────────┐ │ │
│  │  │ Prebid.js│  │ Amazon   │  │ Catalyst SDK         │ │ │
│  │  │          │  │ UAM      │  │ (Nexus Engine)       │ │ │
│  │  └────┬─────┘  └────┬─────┘  └──────────┬───────────┘ │ │
│  │       │             │                    │             │ │
│  │       └─────────────┴────────────────────┘             │ │
│  │                     │                                  │ │
│  │              ┌──────▼──────┐                           │ │
│  │              │requestManager│                          │ │
│  │              │(Coordinator) │                          │ │
│  │              └──────┬──────┘                           │ │
│  │                     │                                  │ │
│  │              ┌──────▼──────┐                           │ │
│  │              │  Google GPT │                           │ │
│  │              │  Ad Server  │                           │ │
│  │              └─────────────┘                           │ │
│  └──────────────────────┬──────────────────────────────────┘ │
│                         │                                    │
│                         │ HTTPS Bid Request                  │
└─────────────────────────┼────────────────────────────────────┘
                          │
                          ▼
        ┌─────────────────────────────────────────┐
        │   The Nexus Engine (Server-Side)        │
        │                                         │
        │  ┌───────────────────────────────────┐ │
        │  │   Catalyst Bidding Platform       │ │
        │  │                                   │ │
        │  │  • Receive bid requests           │ │
        │  │  • Run server-side auctions       │ │
        │  │  • Apply business rules           │ │
        │  │  • Return winning bids            │ │
        │  │  • Track metrics                  │ │
        │  └───────────────────────────────────┘ │
        │                                         │
        │  ┌───────────────────────────────────┐ │
        │  │   Demand Partners Integration     │ │
        │  │   (Your SSP/Exchange Connections) │ │
        │  └───────────────────────────────────┘ │
        └─────────────────────────────────────────┘
```

### Integration Points

| Layer | Component | Owner | Responsibility |
|-------|-----------|-------|----------------|
| **Client** | MAI Publisher JS | MAI/Mediavine | Coordinate bidders, manage ad slots |
| **Client** | Catalyst SDK | **Nexus Engine** | Load SDK, send requests, set targeting |
| **Server** | Catalyst Platform | **Nexus Engine** | Process bids, run auctions, return results |
| **Client** | Google GPT | Google | Render ads based on targeting |

---

## 2. Client-Side SDK Requirements

### 2.1 SDK Loading

**You must provide:** A JavaScript SDK that can be loaded dynamically.

**URL Pattern Expected:**
```javascript
// MAI Publisher will inject your SDK like this:
const script = document.createElement('script');
script.src = 'https://your-cdn.nexusengine.com/catalyst-sdk.js';
script.async = true;
document.head.appendChild(script);
```

**SDK File Requirements:**
- ✅ Must be served over HTTPS
- ✅ Must support async loading
- ✅ Must be cacheable (set appropriate Cache-Control headers)
- ✅ Should be gzipped/compressed
- ✅ Recommended size: < 50KB (gzipped)
- ✅ Must not conflict with existing global variables

### 2.2 Global Object

**You must expose:** A global `window.catalyst` object.

**Required Interface:**
```javascript
window.catalyst = {
  // Initialize the SDK with configuration
  init: function(config) {
    // config = {
    //   accountId: "publisher-account-id",
    //   timeout: 2800,
    //   debug: true/false
    // }
  },

  // Request bids for ad slots
  requestBids: function(requestConfig, callback) {
    // requestConfig = {
    //   accountId: "publisher-account-id",
    //   timeout: 2800,
    //   slots: [...]
    // }
    // callback = function(bids) { ... }
  },

  // Set targeting on GPT slots (optional - MAI can handle this)
  setTargeting: function(bids) {
    // Apply bids to GPT slots
  },

  // Get SDK version (for debugging)
  version: "1.0.0"
};
```

### 2.3 Initialization Sequence

**Expected Call Pattern:**

```javascript
// Step 1: MAI Publisher loads your SDK
<script src="https://your-cdn.nexusengine.com/catalyst-sdk.js" async></script>

// Step 2: MAI Publisher waits for SDK to load
// (Your SDK should set window.catalyst when ready)

// Step 3: MAI Publisher calls init()
window.catalyst.init({
  accountId: "mai-publisher-12345",
  timeout: 2800,
  debug: true
});

// Step 4: MAI Publisher calls requestBids()
window.catalyst.requestBids({
  accountId: "mai-publisher-12345",
  timeout: 2800,
  slots: [
    {
      divId: "mai-ad-leaderboard",
      sizes: [[728, 90], [970, 250]],
      adUnitPath: "/123456/homepage/leaderboard"
    }
  ]
}, function(bids) {
  // Step 5: MAI Publisher receives bids
  console.log('Catalyst returned bids:', bids);

  // Step 6: MAI Publisher sets targeting and calls GPT
});
```

### 2.4 SDK Initialization Timing

**Critical Requirements:**

1. **Before `init()` is called:**
   - ✅ `window.catalyst` object must exist
   - ✅ All methods must be defined (can be no-ops initially)

2. **After `init()` is called:**
   - ✅ SDK must be ready to receive `requestBids()` calls
   - ✅ Configuration must be stored for use in bid requests

3. **Timeout:**
   - ✅ SDK must be fully loaded and initialized within 500ms
   - ✅ If initialization takes longer, use a loading queue pattern

**Recommended Loading Pattern:**

```javascript
// Your SDK should do this:
(function() {
  // Create stub immediately
  window.catalyst = window.catalyst || {
    init: function(config) {
      window.catalyst._config = config;
      window.catalyst._queue = window.catalyst._queue || [];
    },
    requestBids: function(config, callback) {
      window.catalyst._queue = window.catalyst._queue || [];
      window.catalyst._queue.push({ config, callback });
    }
  };

  // When SDK fully loads, replace stubs with real implementation
  function sdkReady() {
    const config = window.catalyst._config;
    const queue = window.catalyst._queue || [];

    // Replace with real implementation
    window.catalyst = new CatalystSDK(config);

    // Process queued requests
    queue.forEach(({ config, callback }) => {
      window.catalyst.requestBids(config, callback);
    });
  }

  // Load SDK asynchronously
  // ... your loading code ...
  // When done: sdkReady();
})();
```

---

## 3. Server-Side API Requirements

### 3.1 Bid Request Endpoint

**You must provide:** An HTTPS endpoint to receive bid requests.

**Endpoint URL:** (You specify this)
```
https://api.nexusengine.com/v1/bid
```

**Method:** POST

**Headers:**
```
Content-Type: application/json
Accept: application/json
```

**Authentication:** (Your choice - options below)

1. **API Key in Header:**
   ```
   X-Catalyst-API-Key: your-api-key
   ```

2. **Account ID in Payload:**
   ```json
   {
     "accountId": "mai-publisher-12345",
     ...
   }
   ```

3. **HMAC Signature:**
   ```
   X-Catalyst-Signature: sha256=...
   X-Catalyst-Timestamp: 1234567890
   ```

### 3.2 Request Timeout

**SLA:** Your API must respond within **2500ms** (leaves 300ms buffer from 2800ms total)

**If you exceed 2500ms:**
- MAI Publisher will cancel the request
- Your bid will not be considered
- Page will continue loading without Catalyst bids

### 3.3 CORS Headers

**Required Response Headers:**
```
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: POST, OPTIONS
Access-Control-Allow-Headers: Content-Type, X-Catalyst-API-Key
Access-Control-Max-Age: 86400
```

**Preflight Request:**
Your endpoint must handle OPTIONS requests for CORS preflight.

---

## 4. Bid Request Format

### 4.1 Request Payload

**The SDK will send:**

```json
{
  "accountId": "mai-publisher-12345",
  "timeout": 2800,
  "slots": [
    {
      "divId": "mai-ad-leaderboard",
      "sizes": [[728, 90], [970, 250], [970, 90]],
      "adUnitPath": "/123456/homepage/leaderboard",
      "position": "atf",
      "enabled_bidders": ["prebid", "amazon", "catalyst"]
    },
    {
      "divId": "mai-ad-rectangle-1",
      "sizes": [[300, 250], [300, 600]],
      "adUnitPath": "/123456/homepage/sidebar",
      "position": "atf",
      "enabled_bidders": ["prebid", "amazon", "catalyst"]
    }
  ],
  "page": {
    "url": "https://example.com/article/12345",
    "domain": "example.com",
    "keywords": ["sports", "football"],
    "categories": ["IAB17", "IAB17-2"]
  },
  "user": {
    "consentGiven": true,
    "gdprApplies": true,
    "uspConsent": "1YNN"
  },
  "device": {
    "width": 1920,
    "height": 1080,
    "deviceType": "desktop",
    "userAgent": "Mozilla/5.0..."
  }
}
```

### 4.2 Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `accountId` | string | ✅ Yes | Publisher's Catalyst account identifier |
| `timeout` | number | ✅ Yes | Timeout in milliseconds (always 2800) |
| `slots` | array | ✅ Yes | Ad slots requesting bids |
| `slots[].divId` | string | ✅ Yes | Unique div ID for the ad slot |
| `slots[].sizes` | array | ✅ Yes | Array of [width, height] pairs |
| `slots[].adUnitPath` | string | ✅ Yes | GPT ad unit path |
| `slots[].position` | string | ❌ No | Position hint: "atf", "btf", "sticky" |
| `slots[].enabled_bidders` | array | ❌ No | Which bidders are enabled for this slot |
| `page.url` | string | ✅ Yes | Current page URL |
| `page.domain` | string | ✅ Yes | Domain name |
| `page.keywords` | array | ❌ No | Content keywords |
| `page.categories` | array | ❌ No | IAB categories |
| `user.consentGiven` | boolean | ✅ Yes | GDPR consent status |
| `user.gdprApplies` | boolean | ✅ Yes | Whether GDPR applies |
| `user.uspConsent` | string | ❌ No | CCPA/USP consent string |
| `device.*` | object | ❌ No | Device information |

### 4.3 Request Validation

**You must validate:**

1. ✅ `accountId` is valid and active
2. ✅ `slots` array is not empty
3. ✅ Each slot has valid `divId`, `sizes`, and `adUnitPath`
4. ✅ `user.consentGiven` is true (if GDPR applies)

**If validation fails:**
- Return HTTP 400 with error message
- SDK should handle gracefully (no bids returned)

---

## 5. Bid Response Format

### 5.1 Response Payload

**You must return:**

```json
{
  "bids": [
    {
      "divId": "mai-ad-leaderboard",
      "cpm": 2.50,
      "currency": "USD",
      "width": 728,
      "height": 90,
      "adId": "catalyst-bid-abc123",
      "creativeId": "creative-xyz789",
      "dealId": "deal-premium-001",
      "meta": {
        "advertiserDomains": ["example-advertiser.com"],
        "networkId": "12345",
        "networkName": "Example Network"
      }
    },
    {
      "divId": "mai-ad-rectangle-1",
      "cpm": 1.75,
      "currency": "USD",
      "width": 300,
      "height": 250,
      "adId": "catalyst-bid-def456",
      "creativeId": "creative-uvw012"
    }
  ],
  "responseTime": 1247
}
```

### 5.2 Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `bids` | array | ✅ Yes | Array of bid objects (can be empty) |
| `bids[].divId` | string | ✅ Yes | Must match request `divId` |
| `bids[].cpm` | number | ✅ Yes | Bid price in CPM (USD cents) |
| `bids[].currency` | string | ✅ Yes | Currency code (default: "USD") |
| `bids[].width` | number | ✅ Yes | Creative width |
| `bids[].height` | number | ✅ Yes | Creative height |
| `bids[].adId` | string | ✅ Yes | Unique identifier for this bid |
| `bids[].creativeId` | string | ✅ Yes | Creative identifier |
| `bids[].dealId` | string | ❌ No | Private marketplace deal ID |
| `bids[].meta` | object | ❌ No | Additional metadata |
| `bids[].meta.advertiserDomains` | array | ❌ No | Advertiser domains (for transparency) |
| `responseTime` | number | ❌ No | Server processing time in ms |

### 5.3 Response Validation

**MAI Publisher will validate:**

1. ✅ `bids` is an array (can be empty)
2. ✅ Each bid has required fields
3. ✅ `divId` matches a requested slot
4. ✅ `cpm` is a positive number
5. ✅ `width` and `height` match one of the requested sizes

**If validation fails:**
- Invalid bids are discarded
- Valid bids are still used

### 5.4 No Bids Response

**If you have no bids:**

```json
{
  "bids": [],
  "responseTime": 123
}
```

**Do NOT return:**
- HTTP 204 No Content
- HTTP 404 Not Found
- Empty response body

**Always return HTTP 200 with valid JSON.**

---

## 6. Initialization Flow

### 6.1 Full Initialization Sequence

```
User Loads Page
       ↓
┌──────────────────────────────────────────┐
│  MAI Publisher: Page Load Event          │
│  - DOM Ready                              │
│  - LCP event fired                        │
└──────────────┬───────────────────────────┘
               ↓
┌──────────────────────────────────────────┐
│  MAI Publisher: CMP Initialization       │
│  - Load Sourcepoint CMP                   │
│  - Wait for consent signal                │
│  - Set window.maiPubAdsVars.hasConsent    │
└──────────────┬───────────────────────────┘
               ↓
┌──────────────────────────────────────────┐
│  MAI Publisher: Bidder Initialization    │
│  - Initialize Prebid.js (parallel)        │
│  - Initialize Amazon UAM (parallel)       │
│  - Initialize Catalyst (parallel) ←──────┐
└──────────────┬───────────────────────────┘│
               ↓                             │
┌──────────────────────────────────────────┐│
│  Catalyst SDK: Load & Init               ││
│  1. Inject <script> tag                  ││
│  2. Download catalyst-sdk.js             ││
│  3. window.catalyst object created       ││
│  4. Call catalyst.init(config)           ││
└──────────────┬───────────────────────────┘│
               ↓                             │
┌──────────────────────────────────────────┐│
│  MAI Publisher: Check All Systems Ready  ││
│  - window.initStatus.prebid = true       ││
│  - window.initStatus.amazon = true       ││
│  - window.initStatus.catalyst = true ←───┘
└──────────────┬───────────────────────────┘
               ↓
┌──────────────────────────────────────────┐
│  MAI Publisher: Request Bids             │
│  - Build slot list                        │
│  - Filter by enabled_bidders              │
│  - Call each bidder requestBids()         │
│    • Prebid.requestBids()                 │
│    • Amazon apstag.fetchBids()            │
│    • Catalyst.requestBids() ←────────────┐
└──────────────┬───────────────────────────┘│
               ↓                             │
┌──────────────────────────────────────────┐│
│  Catalyst SDK: Send Bid Request          ││
│  1. Build request payload                ││
│  2. POST to Nexus Engine API             ││
│  3. Wait for response (max 2500ms)       ││
│  4. Parse bid response                   ││
│  5. Call callback(bids) ←────────────────┘
└──────────────┬───────────────────────────┘
               ↓
┌──────────────────────────────────────────┐
│  MAI Publisher: Bidder Coordination      │
│  - window.requestManager.prebidReady     │
│  - window.requestManager.amazonReady     │
│  - window.requestManager.catalystReady   │
│  - All ready? → sendAdserverRequest()    │
└──────────────┬───────────────────────────┘
               ↓
┌──────────────────────────────────────────┐
│  MAI Publisher: Set GPT Targeting        │
│  - Merge bids from all bidders            │
│  - Set hb_bidder, hb_pb, hb_size, etc.   │
│  - Winner: highest CPM across bidders     │
└──────────────┬───────────────────────────┘
               ↓
┌──────────────────────────────────────────┐
│  MAI Publisher: Call GPT                 │
│  - googletag.pubads().refresh(slots)      │
└──────────────┬───────────────────────────┘
               ↓
       Ads Render on Page
```

### 6.2 Critical Timing Requirements

| Event | Max Time | Requirement |
|-------|----------|-------------|
| SDK Download | 500ms | CDN must be fast |
| SDK Initialization | 100ms | `init()` must be quick |
| Bid Request | 2500ms | API must respond fast |
| Total (init → bids) | 2800ms | Hard timeout |

---

## 7. Coordination Protocol

### 7.1 Request Manager

MAI Publisher uses a **requestManager** to coordinate all three bidders.

**State Machine:**

```javascript
window.requestManager = {
  prebidReady: false,
  amazonReady: false,
  catalystReady: false,
  adserverRequestSent: false
};

// Called when Catalyst finishes (success or timeout)
function biddersReady(bidder) {
  window.requestManager[bidder + 'Ready'] = true;

  checkAllBiddersReady();
}

function checkAllBiddersReady() {
  const { prebidReady, amazonReady, catalystReady } = window.requestManager;

  if (prebidReady && amazonReady && catalystReady) {
    if (!window.requestManager.adserverRequestSent) {
      window.requestManager.adserverRequestSent = true;
      sendAdserverRequest(); // Calls GPT
    }
  }
}
```

### 7.2 Catalyst's Responsibility

**Your SDK must:**

1. ✅ Call `biddersReady('catalyst')` when bids are ready
2. ✅ Call `biddersReady('catalyst')` on timeout
3. ✅ Call `biddersReady('catalyst')` on error
4. ✅ **Never block** - always signal completion

**Example Implementation:**

```javascript
window.catalyst.requestBids = function(config, callback) {
  const timeoutMs = config.timeout || 2800;
  let completed = false;

  // Set timeout
  const timeoutId = setTimeout(() => {
    if (!completed) {
      completed = true;
      console.warn('Catalyst: Timeout reached');
      callback([]);
      window.biddersReady('catalyst');
    }
  }, timeoutMs);

  // Make API request
  fetch('https://api.nexusengine.com/v1/bid', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(config)
  })
  .then(response => response.json())
  .then(data => {
    if (!completed) {
      completed = true;
      clearTimeout(timeoutId);
      callback(data.bids || []);
      window.biddersReady('catalyst');
    }
  })
  .catch(error => {
    if (!completed) {
      completed = true;
      clearTimeout(timeoutId);
      console.error('Catalyst: Request failed', error);
      callback([]);
      window.biddersReady('catalyst');
    }
  });
};
```

---

## 8. Timeout Handling

### 8.1 Timeout Architecture

**Two-Level Timeout:**

1. **Per-Bidder Timeout:** 2800ms (enforced by MAI Publisher)
2. **Overall Timeout:** 2800ms (enforced by requestManager)

**If Catalyst exceeds 2800ms:**
- MAI Publisher cancels the request
- `biddersReady('catalyst')` is called automatically
- Page continues without Catalyst bids

### 8.2 Recommended Server-Side Timeout

**Your API should timeout at 2500ms** (internal timeout)

This leaves 300ms buffer for:
- Network latency
- DNS resolution
- TLS handshake
- Response parsing

**Implementation:**

```javascript
// Server-side (pseudocode)
async function handleBidRequest(request) {
  const internalTimeout = 2500; // 300ms buffer
  const startTime = Date.now();

  try {
    // Run auction with timeout
    const bids = await runAuctionWithTimeout(request, internalTimeout);

    return {
      bids: bids,
      responseTime: Date.now() - startTime
    };
  } catch (timeoutError) {
    // Return empty bids on timeout
    return {
      bids: [],
      responseTime: Date.now() - startTime
    };
  }
}
```

### 8.3 Timeout Monitoring

**You should track:**
- % of requests that timeout
- Average response time
- P50, P95, P99 latencies
- Slow requests by publisher

**Alert if:**
- Timeout rate > 5%
- P95 latency > 2000ms
- P99 latency > 2500ms

---

## 9. GDPR Consent

### 9.1 Consent Check

**Before sending bid requests, MAI Publisher checks:**

```javascript
const hasConsent = window.maiPubAdsVars.hasConsent;
const gdprApplies = window.maiPubAdsVars.gdpr;

if (gdprApplies && !hasConsent) {
  // Skip Catalyst entirely
  console.log('Catalyst: Skipped due to no consent');
  window.biddersReady('catalyst');
  return;
}

// Proceed with bid request
window.catalyst.requestBids(...);
```

### 9.2 Consent Fields in Request

**You will receive:**

```json
{
  "user": {
    "consentGiven": true,
    "gdprApplies": true,
    "uspConsent": "1YNN"
  }
}
```

### 9.3 Your Responsibilities

**You must:**

1. ✅ Check `user.consentGiven` before using personal data
2. ✅ Respect GDPR if `user.gdprApplies = true`
3. ✅ Respect CCPA if `user.uspConsent` indicates opt-out
4. ✅ **Do not use personal data if consent is false**

**Personal Data includes:**
- User IDs
- Cookie values
- Device fingerprints
- Location data (beyond country/state)

**Allowed without consent:**
- Contextual targeting (page keywords, categories)
- Geographic targeting (country/state level)
- Device type (desktop, mobile, tablet)

---

## 10. Error Handling

### 10.1 Error Scenarios

**Your SDK must handle:**

| Error | Scenario | Expected Behavior |
|-------|----------|-------------------|
| Network Error | API unreachable | Return empty bids, call ready callback |
| HTTP 4xx | Invalid request | Log error, return empty bids |
| HTTP 5xx | Server error | Log error, return empty bids |
| Timeout | No response in 2800ms | Cancel request, return empty bids |
| Invalid JSON | Malformed response | Log error, return empty bids |
| Invalid Bids | Missing required fields | Filter invalid, return valid bids |

### 10.2 Error Logging

**Your SDK should log to console (when debug=true):**

```javascript
console.log('[Catalyst] Requesting bids for 3 slots');
console.log('[Catalyst] Response time: 1247ms');
console.log('[Catalyst] Returned 2 bids');

// On error:
console.error('[Catalyst] Request failed:', error.message);
console.error('[Catalyst] Timeout after 2800ms');
```

**Do NOT:**
- Throw unhandled exceptions
- Block page rendering
- Spam console with excessive logs

### 10.3 Graceful Degradation

**If Catalyst fails:**
- Page must continue loading normally
- Other bidders (Prebid, Amazon) must still work
- Ads must still render (with other bids)
- No JavaScript errors in console

---

## 11. Testing Requirements

### 11.1 Mock SDK for Testing

**You should provide a mock SDK for testing:**

```javascript
// catalyst-mock-sdk.js
window.catalyst = {
  init: function(config) {
    console.log('[Catalyst Mock] Initialized with config:', config);
  },

  requestBids: function(config, callback) {
    console.log('[Catalyst Mock] Requesting bids for', config.slots.length, 'slots');

    // Simulate network delay
    setTimeout(() => {
      const mockBids = config.slots.map(slot => ({
        divId: slot.divId,
        cpm: Math.random() * 5,
        currency: 'USD',
        width: slot.sizes[0][0],
        height: slot.sizes[0][1],
        adId: 'mock-' + Math.random(),
        creativeId: 'mock-creative'
      }));

      callback(mockBids);
      if (window.biddersReady) {
        window.biddersReady('catalyst');
      }
    }, 200);
  },

  version: '1.0.0-mock'
};
```

### 11.2 Test Accounts

**You should provide:**

1. **Test Account ID:** For integration testing
2. **Staging API URL:** For pre-production testing
3. **Test Credentials:** API keys or auth tokens

**Example:**
```
Test Account ID: test-mai-publisher-001
Staging API: https://staging-api.nexusengine.com/v1/bid
API Key: test_sk_abc123xyz789
```

### 11.3 Test Scenarios

**We will test:**

| Scenario | Test Case | Expected Result |
|----------|-----------|-----------------|
| Happy Path | Normal bid request | Return valid bids |
| No Bids | No demand available | Return empty bids array |
| Timeout | Slow API response | SDK times out at 2800ms |
| Error | API returns 500 | SDK returns empty bids |
| Invalid Request | Missing account ID | API returns 400 error |
| No Consent | GDPR applies, no consent | No bid request sent |

---

## 12. Performance SLA

### 12.1 Service Level Agreement

| Metric | Target | Monitoring |
|--------|--------|------------|
| **Availability** | 99.9% uptime | API health checks every 60s |
| **Response Time** | P95 < 2000ms | Track all requests |
| **Timeout Rate** | < 5% | Requests exceeding 2500ms |
| **Error Rate** | < 1% | HTTP 5xx responses |
| **CDN Performance** | SDK loads in < 500ms | CDN monitoring |

### 12.2 Monitoring Requirements

**You should monitor:**

1. **API Latency:**
   - P50, P95, P99 response times
   - By publisher account
   - By geographic region

2. **Error Rates:**
   - HTTP 4xx (client errors)
   - HTTP 5xx (server errors)
   - Timeout rate

3. **Bid Quality:**
   - % of requests with bids
   - Average CPM
   - Win rate (when competing with Prebid/Amazon)

4. **SDK Performance:**
   - Load time
   - Initialization time
   - JavaScript errors

### 12.3 Alerting Thresholds

**Alert when:**

- ❌ Availability < 99.5% (over 5 minutes)
- ❌ P95 latency > 2500ms (over 5 minutes)
- ❌ Error rate > 5% (over 5 minutes)
- ❌ Timeout rate > 10% (over 5 minutes)
- ❌ SDK load time > 1000ms (over 5 minutes)

---

## Appendix A: Complete Integration Example

### Client-Side Integration

```javascript
// MAI Publisher implementation (already done)

// Step 1: Inject Catalyst SDK
function loadCatalystSDK() {
  if (window.maiPubAdsVars.catalyst !== true) return;

  const script = document.createElement('script');
  script.src = 'https://cdn.nexusengine.com/catalyst-sdk.js';
  script.async = true;
  script.onload = () => {
    console.log('[MAI] Catalyst SDK loaded');
    initializeCatalyst();
  };
  script.onerror = () => {
    console.error('[MAI] Catalyst SDK failed to load');
    window.biddersReady('catalyst');
  };
  document.head.appendChild(script);
}

// Step 2: Initialize Catalyst
function initializeCatalyst() {
  if (typeof window.catalyst === 'undefined') {
    console.error('[MAI] Catalyst SDK not found');
    window.biddersReady('catalyst');
    return;
  }

  window.catalyst.init({
    accountId: window.maiPubAdsVars.catalystConfig.accountId,
    timeout: 2800,
    debug: window.maiPubAdsVars.debug
  });

  window.initStatus.catalyst = true;
  console.log('[MAI] Catalyst initialized');
}

// Step 3: Request bids
function requestCatalystBids(slots) {
  if (!window.catalyst) {
    window.biddersReady('catalyst');
    return;
  }

  // Filter slots that have Catalyst enabled
  const catalystSlots = slots.filter(slot =>
    slot.enabled_bidders.includes('catalyst')
  );

  if (catalystSlots.length === 0) {
    console.log('[MAI] No slots enabled for Catalyst');
    window.biddersReady('catalyst');
    return;
  }

  console.log('[MAI] Requesting Catalyst bids for', catalystSlots.length, 'slots');

  window.catalyst.requestBids({
    accountId: window.maiPubAdsVars.catalystConfig.accountId,
    timeout: 2800,
    slots: catalystSlots,
    page: {
      url: window.location.href,
      domain: window.location.hostname,
      keywords: window.maiPubAdsVars.keywords || [],
      categories: window.maiPubAdsVars.categories || []
    },
    user: {
      consentGiven: window.maiPubAdsVars.hasConsent,
      gdprApplies: window.maiPubAdsVars.gdpr
    },
    device: {
      width: window.screen.width,
      height: window.screen.height,
      deviceType: getDeviceType()
    }
  }, function(bids) {
    console.log('[MAI] Catalyst returned', bids.length, 'bids');

    // Store bids
    window.catalystBids = bids;

    // Signal ready
    window.biddersReady('catalyst');
  });
}
```

### Server-Side Integration (Your Implementation)

```javascript
// Nexus Engine API endpoint (example)
app.post('/v1/bid', async (req, res) => {
  const startTime = Date.now();

  try {
    // Validate request
    const { accountId, slots, user } = req.body;

    if (!accountId || !slots || slots.length === 0) {
      return res.status(400).json({ error: 'Invalid request' });
    }

    // Check consent
    if (user.gdprApplies && !user.consentGiven) {
      return res.status(200).json({ bids: [] });
    }

    // Run auction with timeout
    const bids = await runAuction(req.body, 2500);

    // Return response
    res.json({
      bids: bids,
      responseTime: Date.now() - startTime
    });

  } catch (error) {
    console.error('Auction error:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
});

async function runAuction(request, timeoutMs) {
  return new Promise((resolve, reject) => {
    const timeout = setTimeout(() => {
      resolve([]); // Return empty bids on timeout
    }, timeoutMs);

    // Run your bidding logic
    yourBiddingLogic(request)
      .then(bids => {
        clearTimeout(timeout);
        resolve(bids);
      })
      .catch(error => {
        clearTimeout(timeout);
        reject(error);
      });
  });
}
```

---

## Appendix B: GPT Targeting Format

### How MAI Publisher Sets Targeting

```javascript
// After receiving bids from all bidders
function setGPTTargeting(prebidBids, amazonBids, catalystBids) {
  // Merge all bids
  const allBids = [
    ...prebidBids.map(b => ({ ...b, bidder: 'prebid' })),
    ...amazonBids.map(b => ({ ...b, bidder: 'amazon' })),
    ...catalystBids.map(b => ({ ...b, bidder: 'catalyst' }))
  ];

  // For each ad slot
  slots.forEach(slot => {
    // Get bids for this slot
    const slotBids = allBids.filter(b => b.divId === slot.divId);

    // Find highest CPM
    const winningBid = slotBids.reduce((max, bid) =>
      bid.cpm > max.cpm ? bid : max
    , { cpm: 0 });

    if (winningBid.cpm > 0) {
      // Set targeting on GPT slot
      const gptSlot = googletag.pubads().getSlots()
        .find(s => s.getSlotElementId() === slot.divId);

      if (gptSlot) {
        gptSlot.setTargeting('hb_bidder', winningBid.bidder);
        gptSlot.setTargeting('hb_pb', winningBid.cpm.toFixed(2));
        gptSlot.setTargeting('hb_size', `${winningBid.width}x${winningBid.height}`);
        gptSlot.setTargeting('hb_adid', winningBid.adId);
      }
    }
  });
}
```

---

## Appendix C: Contact Information

### Integration Support

**For technical questions:**
- Email: integration@nexusengine.com (example)
- Slack: #catalyst-integration (example)
- Documentation: https://docs.nexusengine.com (example)

**For account setup:**
- Email: accounts@nexusengine.com (example)
- Provide: Publisher domain, expected traffic, test credentials

**For performance issues:**
- Email: support@nexusengine.com (example)
- Include: Account ID, timestamp, request/response logs

---

## Appendix D: Checklist

### Pre-Launch Checklist

Before going live, verify:

**SDK:**
- [ ] SDK loads in < 500ms
- [ ] `window.catalyst` object exists
- [ ] `init()` and `requestBids()` methods work
- [ ] Timeout handling works (test with slow API)
- [ ] Error handling works (test with API down)
- [ ] Console logging works (when debug=true)
- [ ] No JavaScript errors in console

**API:**
- [ ] Endpoint responds in < 2500ms (P95)
- [ ] Returns valid JSON structure
- [ ] CORS headers set correctly
- [ ] Authentication works
- [ ] Validation errors return 400
- [ ] Server errors return 500 (not crash)
- [ ] Empty bids return { bids: [] }

**Integration:**
- [ ] Catalyst coordinates with Prebid & Amazon
- [ ] `biddersReady('catalyst')` called on completion
- [ ] GPT targeting set correctly
- [ ] Ads render with Catalyst bids
- [ ] Page loads normally if Catalyst fails
- [ ] GDPR consent respected
- [ ] No bids sent without consent

**Monitoring:**
- [ ] API latency tracked
- [ ] Error rate tracked
- [ ] Timeout rate tracked
- [ ] SDK load time tracked
- [ ] Alerts configured

**Documentation:**
- [x] API endpoint URL provided
- [x] SDK CDN URL provided
- [x] Test account credentials provided
- [x] Integration guide published

---

## Deployment Information

### Production Endpoints

**Catalyst SDK (JavaScript)**
- URL: `https://cdn.thenexusengine.com/assets/catalyst-sdk.js`
- CDN: CloudFront
- Cache: 1 hour (max-age=3600)
- Size: < 50KB gzipped
- Integrity: Available via SRI (see below)

**Catalyst API (Bid Endpoint)**
- URL: `https://ads.thenexusengine.com/v1/bid`
- Method: POST
- Content-Type: application/json
- CORS: Enabled for all origins
- Rate Limit: 1000 req/min per IP

### Staging Endpoints

**Staging SDK**
- URL: `https://staging-cdn.thenexusengine.com/assets/catalyst-sdk.js`

**Staging API**
- URL: `https://staging-ads.thenexusengine.com/v1/bid`

### Test Account Credentials

**Staging Account**
- Account ID: `mai-staging-test`
- Description: For MAI Publisher staging integration testing
- Rate Limit: Unlimited

**Production Account**
- Account ID: `mai-publisher-12345`
- Description: Production MAI Publisher account
- Rate Limit: Standard limits apply

### SDK Integration Code

```html
<!-- Add to MAI Publisher <head> section -->
<script src="https://cdn.thenexusengine.com/assets/catalyst-sdk.js" async></script>

<!-- With Subresource Integrity (SRI) -->
<script
  src="https://cdn.thenexusengine.com/assets/catalyst-sdk.js"
  integrity="sha384-[HASH]"
  crossorigin="anonymous"
  async>
</script>

<!-- Initialize when ready -->
<script>
window.catalyst = window.catalyst || {};
catalyst.cmd = catalyst.cmd || [];

catalyst.cmd.push(function() {
  catalyst.init({
    accountId: 'mai-publisher-12345',
    timeout: 2800,
    debug: false
  });
});
</script>
```

### Health Check Endpoints

**Liveness Check**
- URL: `https://ads.thenexusengine.com/health`
- Response: `{"status": "healthy", "timestamp": "...", "version": "1.0.0"}`

**Readiness Check**
- URL: `https://ads.thenexusengine.com/health/ready`
- Response: `{"ready": true, "checks": {...}}`

### Monitoring

**Prometheus Metrics**
- URL: `https://ads.thenexusengine.com/metrics`
- Format: Prometheus exposition format
- Key Metrics:
  - `catalyst_bid_requests_total` - Total bid requests
  - `catalyst_bid_latency_seconds` - Bid request latency histogram
  - `catalyst_bid_timeouts_total` - Total timeouts
  - `catalyst_bid_errors_total` - Total errors

**Grafana Dashboards**
- URL: `https://grafana.thenexusengine.com/d/catalyst`
- Dashboards:
  - Catalyst Overview
  - SLA Compliance
  - Per-Account Metrics

### Support

**Technical Support**
- Email: tech-support@thenexusengine.com
- Slack: #catalyst-integration
- On-Call: pagerduty@thenexusengine.com

**Documentation**
- Deployment Guide: [CATALYST_DEPLOYMENT_GUIDE.md](./CATALYST_DEPLOYMENT_GUIDE.md)
- API Reference: [CATALYST_API_REFERENCE.md](./CATALYST_API_REFERENCE.md)
- Status Page: `https://status.thenexusengine.com`

---

## Document Version History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-02-04 | AI Assistant | Initial specification based on MAI Publisher integration |
| 1.1 | 2026-02-04 | AI Assistant | Added deployment information and production URLs |

---

## Implementation Status

✅ **Phase 1: Bid Endpoint** - Complete
- MAI-compatible bid request handler
- JSON request/response format
- 2500ms server timeout
- OpenRTB conversion

✅ **Phase 2: JavaScript SDK** - Complete
- `window.catalyst` namespace
- `init()` and `requestBids()` methods
- POST /v1/bid integration
- `biddersReady('catalyst')` callback
- < 50KB gzipped

✅ **Phase 3: Server Integration** - Complete
- `/v1/bid` endpoint registered
- `/assets/catalyst-sdk.js` endpoint registered
- CORS configuration

✅ **Phase 4: Timeout Configuration** - Complete
- 2500ms auction timeout
- Per-request timeout support

✅ **Phase 5: Testing** - Complete
- Unit tests (Go)
- Integration tests (Go)
- Browser test page (HTML)

✅ **Phase 6: Documentation** - Complete
- Deployment guide
- Integration specification updated
- Test credentials provided

**Status:** Ready for staging deployment

---

**End of Specification**

This document defines all requirements for The Nexus Engine Catalyst platform to integrate with MAI Publisher. All technical requirements are based on actual exploration and implementation of the MAI Publisher codebase.
