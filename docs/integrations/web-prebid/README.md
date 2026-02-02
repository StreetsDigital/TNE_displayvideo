# Web via Prebid Integration

**Status:** ⚠️ Backend Ready - Needs Client Examples
**Timeline:** 1-2 weeks to complete
**Difficulty:** Easy
**Best For:** Display publishers already using Prebid.js

## Overview

Integrate TNE Catalyst as a demand partner in your existing Prebid.js setup. Perfect for publishers who already use Prebid.js for header bidding and want to add TNE Catalyst demand.

## Current Status

✅ **Backend Complete:**
- OpenRTB 2.5 auction endpoint
- Publisher authentication
- Privacy compliance (GDPR/CCPA)
- Cookie sync endpoints
- Rate limiting

⚠️ **Needs Documentation:**
- Prebid.js configuration examples
- Bidder adapter setup guide
- Publisher integration steps

## Quick Overview (How It Will Work)

```javascript
// This is what the final integration will look like:

var adUnits = [{
  code: 'div-gpt-ad-1234',
  mediaTypes: {
    banner: {
      sizes: [[300, 250], [728, 90]]
    }
  },
  bids: [{
    bidder: 'tne-catalyst',  // TNE Catalyst bidder
    params: {
      publisherId: 'pub-123456',
      placement: 'homepage-banner',
      bidfloor: 1.50
    }
  }]
}];

pbjs.requestBids({
  adUnits: adUnits,
  bidsBackHandler: function(bids) {
    // TNE Catalyst bids will be included
  }
});
```

## Features (When Complete)

### Supported Ad Formats
- ✅ Banner (all standard sizes)
- ✅ Native (in-feed, content recommendation)
- ✅ Multi-size banners
- ✅ Lazy loading
- ✅ Refresh

### Privacy & Compliance
- ✅ GDPR consent (TCF v2)
- ✅ CCPA (US Privacy)
- ✅ COPPA compliance
- ✅ Automatic geo-detection

### Advanced Features
- ✅ First-party data passing
- ✅ Custom targeting
- ✅ Floor prices
- ✅ Currency conversion
- ✅ User ID sync

## Use Cases

### 1. Header Bidding
Add TNE Catalyst to your existing header bidding stack.

### 2. Multi-Format Monetization
Monetize banner and native inventory.

### 3. International Publishers
GDPR/CCPA compliant with automatic geo-detection.

## What's Needed

See [WORK_REQUIRED.md](./WORK_REQUIRED.md) for complete list:

1. **Prebid.js Bidder Adapter** (or configuration for existing OpenRTB adapter)
2. **Publisher Integration Guide**
3. **Configuration Examples**
4. **Test Credentials**
5. **Troubleshooting Guide**

## Backend Endpoints (Already Live)

**Auction Endpoint:**
```
POST https://api.tne-catalyst.com/openrtb2/auction
```

**Cookie Sync:**
```
GET https://api.tne-catalyst.com/cookie_sync
GET https://api.tne-catalyst.com/setuid?bidder={bidder}&uid={uid}
```

**User Opt-Out:**
```
GET https://api.tne-catalyst.com/optout
```

## Estimated Timeline

| Task | Effort | Priority |
|------|--------|----------|
| Create bidder adapter config | 2-3 days | High |
| Write integration guide | 2 days | High |
| Create code examples | 2 days | High |
| Test with real Prebid.js | 2 days | High |
| Create troubleshooting guide | 1 day | Medium |
| **Total** | **1-2 weeks** | - |

## Next Steps

1. **Review**: [WORK_REQUIRED.md](./WORK_REQUIRED.md) - See what needs to be built
2. **Contact**: Email integration@tne-catalyst.com to express interest
3. **Beta**: Sign up for beta access when ready

## Temporary Workaround

Until Prebid integration is complete, you can:

1. Use [OpenRTB Direct](../openrtb-direct/) for server-side integration
2. Use [Video VAST](../video-vast/) for video inventory

## Support

- **Status Updates**: [WORK_REQUIRED.md](./WORK_REQUIRED.md)
- **Email**: prebid-integration@tne-catalyst.com
- **Notify Me**: Request notification when ready

---

**Interested in beta testing?** → Email prebid-integration@tne-catalyst.com
