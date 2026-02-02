# Video via Prebid Integration

**Status:** ⚠️ Backend Ready - Needs Client Examples
**Timeline:** 1-2 weeks to complete
**Difficulty:** Medium
**Best For:** Video publishers already using Prebid.js for video monetization

## Overview

Integrate TNE Catalyst video demand into your existing Prebid.js video setup. Perfect for publishers who already use Prebid.js for video header bidding and want to add TNE Catalyst as an additional demand source.

## Current Status

✅ **Backend Complete:**
- Video VAST endpoints (/video/vast, /video/openrtb)
- Full VAST 2.0-4.0 support
- OpenRTB video auction
- CTV/OTT optimization
- Event tracking
- Privacy compliance

⚠️ **Needs Documentation:**
- Prebid.js video configuration for TNE
- Video-specific bidder params
- Player integration examples
- CTV/OTT specific examples

## Quick Overview (How It Will Work)

```javascript
// This is what the final integration will look like:

var videoAdUnit = {
  code: 'video-div',
  mediaTypes: {
    video: {
      playerSize: [640, 480],
      context: 'instream',
      mimes: ['video/mp4', 'video/webm'],
      protocols: [2, 3, 5, 6],
      minduration: 5,
      maxduration: 30,
      api: [2],  // VPAID 2.0
      placement: 1  // In-stream
    }
  },
  bids: [{
    bidder: 'tne-catalyst',
    params: {
      publisherId: 'pub-video-123',
      placement: 'homepage-video',
      bidfloor: 3.50
    }
  }]
};

pbjs.requestBids({
  adUnits: [videoAdUnit],
  bidsBackHandler: function(bids) {
    // Winning bid will contain VAST XML or URL
    var vastUrl = pbjs.adServers.dfp.buildVideoUrl({
      adUnit: videoAdUnit,
      params: {
        // GAM params
      }
    });
    player.loadAd(vastUrl);
  }
});
```

## Features (When Complete)

### Video Formats
- ✅ In-stream (pre-roll, mid-roll, post-roll)
- ✅ Out-stream (in-feed, in-article)
- ✅ Rewarded video
- ✅ Interstitial video

### VAST Support
- ✅ VAST 2.0, 3.0, 4.0
- ✅ Inline VAST (direct creative)
- ✅ Wrapper VAST (mediation)
- ✅ Multiple media files (bitrate selection)
- ✅ Companion ads

### CTV/OTT
- ✅ Platform detection (Roku, Fire TV, Apple TV, etc.)
- ✅ VPAID filtering for TV platforms
- ✅ 4K video support
- ✅ HLS/DASH protocols
- ✅ Bitrate optimization

### Privacy & Compliance
- ✅ GDPR consent (TCF v2)
- ✅ CCPA (US Privacy)
- ✅ COPPA compliance

### Advanced Features
- ✅ Skippable ads
- ✅ VPAID/MRAID support
- ✅ OMID viewability
- ✅ Quartile tracking
- ✅ Completion tracking

## Use Cases

### 1. Video Header Bidding
Add TNE Catalyst to existing Prebid.js video setup.

### 2. CTV/OTT Monetization
Monetize connected TV and OTT platforms via Prebid.

### 3. Multi-Format Video
Mix in-stream and out-stream video ads.

## What's Needed

See [WORK_REQUIRED.md](./WORK_REQUIRED.md) for complete list:

1. **Prebid Video Configuration Guide**
2. **Video Player Integration Examples**
3. **CTV/OTT Platform Examples**
4. **Video-Specific Parameters Documentation**
5. **Testing Guide**

## Backend Endpoints (Already Live)

**Simple VAST (Query Params):**
```
GET https://api.tne-catalyst.com/video/vast?pub_id=...&w=1920&h=1080
```

**Advanced VAST (OpenRTB):**
```
POST https://api.tne-catalyst.com/video/openrtb
Content-Type: application/json
```

**Event Tracking:**
```
GET https://api.tne-catalyst.com/video/event/{type}
```

## Estimated Timeline

| Task | Effort | Priority |
|------|--------|----------|
| Create video config guide | 2 days | High |
| Write player integration examples | 2 days | High |
| Create CTV examples | 2 days | High |
| Document video parameters | 1 day | High |
| Create testing guide | 1 day | Medium |
| **Total** | **1-2 weeks** | - |

## Supported Video Players

Examples will be provided for:
- Google IMA SDK ✅
- Video.js ✅
- JW Player ✅
- Brightcove ✅
- Custom players ✅

## Next Steps

1. **Review**: [WORK_REQUIRED.md](./WORK_REQUIRED.md) - See what needs to be built
2. **Contact**: Email video-integration@tne-catalyst.com to express interest
3. **Beta**: Sign up for beta access when ready

## Temporary Workaround

Until Prebid integration is complete, you can:

1. Use [Video VAST](../video-vast/) for direct integration
2. Build custom Prebid adapter following [OpenRTB Direct](../openrtb-direct/)

## Support

- **Status Updates**: [WORK_REQUIRED.md](./WORK_REQUIRED.md)
- **Email**: video-prebid@tne-catalyst.com
- **Notify Me**: Request notification when ready

---

**Interested in beta testing?** → Email video-prebid@tne-catalyst.com
