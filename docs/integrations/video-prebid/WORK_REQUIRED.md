# Video via Prebid Integration - Work Required

**Current Status:** ⚠️ **NEEDS CLIENT DOCUMENTATION**

**Backend:** ✅ 100% Complete
**Frontend:** ❌ 0% Complete (needs Prebid-specific docs and examples)

## What Exists (Backend)

| Component | Status | Details |
|-----------|--------|---------|
| Video Endpoints | ✅ Complete | GET /video/vast, POST /video/openrtb |
| VAST Generation | ✅ Complete | VAST 2.0-4.0 support |
| OpenRTB Video | ✅ Complete | Full video object support |
| Event Tracking | ✅ Complete | 8 tracked events |
| CTV Optimization | ✅ Complete | 10+ platforms |
| Privacy | ✅ Complete | GDPR/CCPA/COPPA |

**Backend Files:**
- ✅ `internal/endpoints/video_handler.go` - Video endpoints
- ✅ `pkg/vast/` - Complete VAST library
- ✅ `internal/exchange/vast_response.go` - VAST builder
- ✅ `internal/endpoints/video_events.go` - Event tracking
- ✅ `internal/ctv/` - CTV detection and optimization

**See Also:**
- [Video VAST Integration](../video-vast/WORK_REQUIRED.md) - Backend implementation details

## What's Missing (Prebid Integration)

### 1. Prebid Video Configuration Guide

**Priority:** HIGH
**Effort:** 2 days
**Owner:** Documentation Team

**File to Create:**
- `docs/integrations/video-prebid/SETUP.md`

**Required Content:**

#### Basic Prebid Video Setup
```javascript
// Example of what needs to be documented

var videoAdUnit = {
  code: 'video-ad-unit',
  mediaTypes: {
    video: {
      // Video parameters
      playerSize: [640, 480],
      context: 'instream',  // or 'outstream'
      mimes: ['video/mp4', 'video/webm'],
      protocols: [2, 3, 5, 6],  // VAST protocols
      minduration: 5,
      maxduration: 30,
      api: [2],  // VPAID 2.0
      placement: 1,  // In-stream
      playbackmethod: [1, 2],  // Auto-play, click-to-play
      // Skip settings
      skip: 1,
      skipafter: 5,
      // Bidfloor
      bidfloor: 3.50,
      bidfloorcur: 'USD'
    }
  },
  bids: [{
    bidder: 'tne-catalyst',  // or use generic adapter
    params: {
      publisherId: 'pub-video-123',
      placementId: 'homepage-video',
      // Pass-through to OpenRTB video object
      video: {
        // Additional video params
      }
    }
  }]
};
```

#### Configuration Sections Needed:
1. Prerequisites (Prebid.js setup)
2. Add TNE Catalyst video bidder
3. Configure video ad units
4. Player integration (IMA, Video.js, JW)
5. VAST URL handling
6. Event tracking setup
7. Privacy compliance
8. CTV/OTT configuration
9. Testing
10. Troubleshooting

### 2. Video Player Integration Examples

**Priority:** HIGH
**Effort:** 2 days
**Owner:** Engineering Team

**Files to Create:**

#### Google IMA SDK Example
`docs/integrations/video-prebid/examples/ima-integration.html`

```html
<!-- Example structure -->
<!DOCTYPE html>
<html>
<head>
  <script src="prebid.js"></script>
  <script src="//imasdk.googleapis.com/js/sdkloader/ima3.js"></script>
</head>
<body>
  <video id="video"></video>
  <div id="ad-container"></div>

  <script>
    // 1. Configure Prebid video ad unit with TNE Catalyst
    // 2. Request bids
    // 3. On bids back, get VAST URL or XML
    // 4. Pass to IMA SDK
    // 5. Handle video events
  </script>
</body>
</html>
```

#### Video.js Example
`docs/integrations/video-prebid/examples/videojs-integration.html`

```html
<!-- Video.js + Prebid + TNE Catalyst -->
```

#### JW Player Example
`docs/integrations/video-prebid/examples/jwplayer-integration.html`

```html
<!-- JW Player + Prebid + TNE Catalyst -->
```

#### Out-stream Example
`docs/integrations/video-prebid/examples/outstream-integration.html`

```html
<!-- Out-stream video in article -->
```

### 3. CTV/OTT Platform Examples

**Priority:** HIGH
**Effort:** 2 days
**Owner:** Engineering Team

**Files to Create:**

#### Roku Integration
`docs/integrations/video-prebid/examples/ctv/roku-integration.brs`

```brightscript
' BrightScript example
' Using TNE Catalyst with Prebid on Roku
```

#### Fire TV Integration
`docs/integrations/video-prebid/examples/ctv/firetv-integration.java`

```java
// Android/Fire TV integration
// Using Prebid SDK for Fire TV
```

#### Apple TV Integration
`docs/integrations/video-prebid/examples/ctv/appletv-integration.swift`

```swift
// Swift/tvOS integration
// Using Prebid SDK for Apple TV
```

### 4. Video Parameters Documentation

**Priority:** HIGH
**Effort:** 1 day
**Owner:** Documentation Team

**File to Create:**
`docs/integrations/video-prebid/PARAMETERS.md`

**Content:**

Document all video-specific parameters:

```javascript
{
  bidder: 'tne-catalyst',
  params: {
    // Required
    publisherId: 'pub-video-123',
    placementId: 'homepage-video',

    // Optional - Video specific
    video: {
      minduration: 5,      // Min video duration (seconds)
      maxduration: 30,     // Max video duration (seconds)
      skip: 1,             // 1 = skippable, 0 = not skippable
      skipafter: 5,        // Skip button appears after N seconds
      protocols: [2,3,5,6],// VAST protocols
      mimes: ['video/mp4'],// Supported MIME types
      api: [2],            // VPAID 2.0
      placement: 1,        // 1=in-stream, 2=in-banner, etc.
      playbackmethod: [1,2],// 1=auto, 2=click-to-play
      linearity: 1,        // 1=linear, 2=non-linear
    },

    // Optional - Targeting
    bidfloor: 3.50,        // CPM floor
    keywords: ['sports'],   // Content keywords
    firstPartyData: {}     // First-party data
  }
}
```

**Map to OpenRTB:**
Show how Prebid params map to OpenRTB 2.5 video object fields.

### 5. VAST Response Handling

**Priority:** HIGH
**Effort:** 1 day
**Owner:** Documentation Team

**File to Add Section To:**
`docs/integrations/video-prebid/SETUP.md`

**Content:**

Document how to handle VAST responses from TNE Catalyst:

```javascript
// Option 1: VAST URL in bid response
pbjs.requestBids({
  bidsBackHandler: function() {
    var vastUrl = pbjs.getAdserverTargetingForAdUnitCode('video-unit');
    player.loadAd(vastUrl);
  }
});

// Option 2: VAST XML in bid response
pbjs.requestBids({
  bidsBackHandler: function() {
    var vastXml = pbjs.getBidResponsesForAdUnitCode('video-unit');
    player.loadAdFromXml(vastXml);
  }
});

// Option 3: Use with Google Ad Manager
var vastUrl = pbjs.adServers.dfp.buildVideoUrl({
  adUnit: videoAdUnit,
  params: {
    iu: '/12345/video',
    cust_params: pbjs.adServers.dfp.getAdserverTargeting()
  }
});
```

### 6. Privacy Compliance for Video

**Priority:** MEDIUM
**Effort:** 1 day
**Owner:** Documentation Team

**File to Create:**
`docs/integrations/video-prebid/PRIVACY.md`

**Content:**

#### GDPR for Video
```javascript
// Pass TCF consent string
pbjs.setConfig({
  consentManagement: {
    gdpr: {
      cmpApi: 'iab',
      timeout: 3000,
      defaultGdprScope: true
    }
  }
});

// TNE Catalyst will automatically receive consent string
```

#### CCPA for Video
```javascript
// Pass US Privacy string
pbjs.setConfig({
  consentManagement: {
    usp: {
      cmpApi: 'iab',
      timeout: 100
    }
  }
});
```

### 7. Testing & Debugging Guide

**Priority:** MEDIUM
**Effort:** 1 day
**Owner:** Documentation Team

**File to Create:**
`docs/integrations/video-prebid/TESTING.md`

**Content:**

#### Enable Prebid Debug Mode
```javascript
pbjs.setConfig({ debug: true });
localStorage.setItem('pbjs_debug', 'true');
```

#### Test VAST Response
```javascript
// Check bid response
pbjs.getBidResponses();

// Validate VAST XML
// Use Google VAST Inspector: https://googleads.github.io/googleads-ima-html5/vsi/
```

#### Test Events
```javascript
// Monitor tracking pixels
// Check network tab for event URLs
// Verify in TNE Catalyst dashboard
```

#### Common Issues
- VAST parsing errors
- Video player compatibility
- VPAID issues on CTV
- Cookie sync for video
- Floor price too high

### 8. Performance Optimization

**Priority:** LOW
**Effort:** 1 day
**Owner:** Documentation Team

**File to Create:**
`docs/integrations/video-prebid/OPTIMIZATION.md`

**Topics:**
- Prebid video timeout settings
- Parallel vs sequential loading
- Video caching strategies
- Preloading techniques
- Bandwidth optimization
- CTV-specific optimizations

### 9. Troubleshooting Guide

**Priority:** MEDIUM
**Effort:** 1 day
**Owner:** Support Team

**File to Create:**
`docs/integrations/video-prebid/TROUBLESHOOTING.md`

**Common Issues:**

1. **No video bids**
   - Check video parameters
   - Verify floor prices
   - Check MIME type support
   - Validate video dimensions

2. **VAST errors**
   - Validate VAST XML
   - Check media file URLs
   - Verify VAST version compatibility
   - Check wrapper unwrapping

3. **Video won't play**
   - Check VPAID compatibility
   - Verify player configuration
   - Check media file format
   - Test on different browsers

4. **Events not tracking**
   - Check tracking URL format
   - Verify no ad blockers
   - Check CORS settings
   - Test URLs manually

5. **CTV-specific issues**
   - VPAID not supported on Roku/Fire TV
   - Check platform detection
   - Verify HLS support
   - Check 4K bitrates

## Implementation Plan

### Phase 1: Core Documentation (Week 1)

**Goal:** Enable basic video integration

- [ ] Create SETUP.md with step-by-step guide
- [ ] Create PARAMETERS.md with all video params
- [ ] Create basic IMA SDK example
- [ ] Create basic Video.js example
- [ ] Test end-to-end with real player

**Deliverables:**
- Publishers can add video ads via Prebid
- Basic examples work
- Parameters documented

### Phase 2: Advanced Examples (Week 2)

**Goal:** Cover all video scenarios

- [ ] Create JW Player example
- [ ] Create out-stream example
- [ ] Create PRIVACY.md guide
- [ ] Create TESTING.md guide
- [ ] Create TROUBLESHOOTING.md

**Deliverables:**
- Multiple player examples
- Privacy compliance docs
- Testing support

### Phase 3: CTV/OTT (Optional, Week 3)

**Goal:** Support CTV platforms

- [ ] Create Roku example
- [ ] Create Fire TV example
- [ ] Create Apple TV example
- [ ] Document CTV-specific considerations
- [ ] Test on real devices

**Deliverables:**
- CTV integration examples
- Platform-specific docs

## Dependencies

**Requires (from Web Prebid):**
- [ ] Prebid bidder adapter configuration (Option A: generic, Option B: custom)
- [ ] Basic Prebid setup docs

**Builds On:**
- ✅ Video VAST integration (complete)
- ✅ OpenRTB video support (complete)

## Testing Checklist

Before marking complete:

- [ ] Test with Google IMA SDK
- [ ] Test with Video.js
- [ ] Test with JW Player
- [ ] Test in-stream video
- [ ] Test out-stream video
- [ ] Test skippable ads
- [ ] Test VPAID creative
- [ ] Test companion ads
- [ ] Test on desktop browsers
- [ ] Test on mobile browsers
- [ ] Test on Roku (if CTV phase included)
- [ ] Test on Fire TV (if CTV phase included)
- [ ] Test GDPR consent
- [ ] Test CCPA opt-out
- [ ] Verify event tracking
- [ ] Check performance (< 100ms)

## Resources Needed

**Engineering:**
- 1 developer x 3 days (examples)
- 1 developer x 2 days (testing)

**Documentation:**
- 1 technical writer x 4 days

**QA:**
- 1 QA engineer x 2 days
- 1 CTV device lab (if Phase 3)

**Total Effort:** 2-3 weeks (without CTV), 3-4 weeks (with CTV)

## Success Criteria

**Complete When:**
- [ ] Documentation published
- [ ] 3+ player examples work
- [ ] Privacy compliance documented
- [ ] Testing guide available
- [ ] 3+ beta publishers successfully integrated
- [ ] No major issues reported

## Next Actions

1. **Assign owner** for Phase 1
2. **Schedule kickoff** meeting
3. **Set target date** (recommend 2-3 weeks)
4. **Identify beta publishers** with video inventory
5. **Coordinate with Web Prebid** work (shared adapter)

---

**Last Updated:** 2026-02-02
**Target Completion:** 2026-02-23 (3 weeks)
**Status:** Ready to start
**Depends On:** Web Prebid bidder adapter (can proceed in parallel)
