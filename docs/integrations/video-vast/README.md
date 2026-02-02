# Video VAST Integration

**Status:** ✅ Production Ready
**Timeline:** Immediate
**Difficulty:** Easy
**Best For:** Video publishers, CTV/OTT platforms, video players

## Overview

Direct VAST tag integration for video advertising. Generate VAST XML responses from simple HTTP requests. Supports VAST 2.0-4.0, inline and wrapper formats, and is optimized for CTV/OTT platforms.

## Quick Start (2 minutes)

### Method 1: Query Parameters (Simplest)

```html
<!-- In your video player -->
<video>
  <source src="content.mp4" />
</video>

<script>
// Initialize IMA SDK or your video ad framework
const vastUrl = 'https://api.tne-catalyst.com/video/vast?' +
  'w=1920&h=1080&' +
  'mindur=5&maxdur=30&' +
  'mimes=video/mp4&' +
  'pub_id=pub-123456&' +
  'bidfloor=3.0';

videoPlayer.loadAd(vastUrl);
</script>
```

### Method 2: POST with OpenRTB (Advanced)

```javascript
// For more control, use OpenRTB format
const response = await fetch('https://api.tne-catalyst.com/video/openrtb', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': 'your-api-key'
  },
  body: JSON.stringify({
    id: 'request-123',
    imp: [{
      id: '1',
      video: {
        w: 1920,
        h: 1080,
        minduration: 5,
        maxduration: 30,
        mimes: ['video/mp4', 'video/webm'],
        protocols: [2, 3, 5, 6],
        placement: 1
      },
      bidfloor: 3.0
    }],
    site: {
      id: 'pub-123456',
      domain: 'yoursite.com'
    }
  })
});

const vastXml = await response.text();
videoPlayer.loadAdFromXml(vastXml);
```

### Receive VAST Response

```xml
<?xml version="1.0" encoding="UTF-8"?>
<VAST version="4.0">
  <Ad id="ad-123">
    <InLine>
      <AdSystem>TNE Catalyst</AdSystem>
      <AdTitle>Your Video Ad</AdTitle>
      <Impression>https://track.tne-catalyst.com/imp?id=123</Impression>
      <Creatives>
        <Creative>
          <Linear>
            <Duration>00:00:15</Duration>
            <TrackingEvents>
              <Tracking event="start">https://track.tne-catalyst.com/start</Tracking>
              <Tracking event="firstQuartile">https://track.tne-catalyst.com/q25</Tracking>
              <Tracking event="midpoint">https://track.tne-catalyst.com/q50</Tracking>
              <Tracking event="thirdQuartile">https://track.tne-catalyst.com/q75</Tracking>
              <Tracking event="complete">https://track.tne-catalyst.com/complete</Tracking>
            </TrackingEvents>
            <MediaFiles>
              <MediaFile delivery="progressive" type="video/mp4" width="1920" height="1080">
                https://cdn.example.com/video.mp4
              </MediaFile>
            </MediaFiles>
          </Linear>
        </Creative>
      </Creatives>
    </InLine>
  </Ad>
</VAST>
```

## Features

✅ **VAST Support**
- VAST 2.0, 3.0, 4.0 formats
- Inline VAST (direct creative)
- Wrapper VAST (mediation)
- Multi-level wrapper unwrapping

✅ **Video Formats**
- In-stream (pre-roll, mid-roll, post-roll)
- Out-stream (in-feed, in-article)
- Rewarded video
- Interstitial video

✅ **CTV/OTT Optimization**
- 10+ platform detection (Roku, Fire TV, Apple TV, etc.)
- 4K video support
- Bitrate selection
- VPAID filtering for TV platforms
- HLS/DASH protocol support

✅ **Tracking Events**
- Impression tracking
- Video events (start, quartiles, complete)
- User interactions (click, mute, pause, fullscreen)
- Error tracking
- Custom event parameters

✅ **Advanced Features**
- Skippable ads (configurable skip offset)
- Companion ads (banner, HTML, iframe)
- VPAID/MRAID support
- OMID viewability
- Multiple media files (bitrate selection)

## Supported Video Players

Works with all major video players:

- **Google IMA SDK** ✅
- **Video.js** ✅
- **JW Player** ✅
- **Brightcove** ✅
- **Kaltura** ✅
- **THEOplayer** ✅
- **Shaka Player** ✅
- **Custom Players** ✅

## Endpoints

### 1. GET /video/vast

Simple query parameter interface.

**URL:** `https://api.tne-catalyst.com/video/vast`

**Parameters:**

| Parameter | Required | Description | Example |
|-----------|----------|-------------|---------|
| `pub_id` | Yes | Publisher ID | `pub-123456` |
| `w` | Yes | Video width | `1920` |
| `h` | Yes | Video height | `1080` |
| `mindur` | Yes | Min duration (seconds) | `5` |
| `maxdur` | Yes | Max duration (seconds) | `30` |
| `mimes` | Yes | Supported MIME types | `video/mp4,video/webm` |
| `bidfloor` | No | Minimum CPM | `3.0` |
| `placement` | No | Placement type (1-5) | `1` |
| `protocols` | No | Supported protocols | `2,3,5,6` |
| `session_id` | No | User session ID | `session-abc` |

**Example:**
```
GET /video/vast?pub_id=pub-123&w=1920&h=1080&mindur=5&maxdur=30&mimes=video/mp4&bidfloor=3.0
```

### 2. POST /video/openrtb

Full OpenRTB request interface.

**URL:** `https://api.tne-catalyst.com/video/openrtb`

**Body:** OpenRTB 2.5 BidRequest with video object

**Example:** See [Setup Guide](./SETUP.md)

### 3. GET /video/event/{type}

Track video events.

**URL:** `https://api.tne-catalyst.com/video/event/{type}`

**Event Types:**
- `impression` - Ad impression
- `start` - Video started
- `firstQuartile` - 25% complete
- `midpoint` - 50% complete
- `thirdQuartile` - 75% complete
- `complete` - 100% complete
- `click` - User clicked ad
- `pause` - User paused
- `mute` - User muted
- `error` - Error occurred

## CTV/OTT Support

Automatically detects and optimizes for:

| Platform | Detection | Optimization |
|----------|-----------|--------------|
| Roku | ✅ User-Agent | VPAID filter, bitrate limit |
| Amazon Fire TV | ✅ User-Agent | VPAID filter, 4K support |
| Apple TV | ✅ User-Agent | HLS preferred, VPAID filter |
| Android TV | ✅ User-Agent | Standard optimization |
| Samsung Tizen | ✅ User-Agent | VPAID filter |
| LG webOS | ✅ User-Agent | VPAID filter |
| Chromecast | ✅ User-Agent | Standard optimization |
| Xbox | ✅ User-Agent | VPAID filter |
| PlayStation | ✅ User-Agent | VPAID filter |
| Vizio SmartCast | ✅ User-Agent | VPAID filter |

## Use Cases

### 1. Website Video Player

```html
<!DOCTYPE html>
<html>
<head>
  <script src="//imasdk.googleapis.com/js/sdkloader/ima3.js"></script>
</head>
<body>
  <video id="video" width="640" height="360"></video>
  <div id="ad-container"></div>

  <script>
    const adDisplayContainer = new google.ima.AdDisplayContainer(
      document.getElementById('ad-container'),
      document.getElementById('video')
    );

    const adsLoader = new google.ima.AdsLoader(adDisplayContainer);

    const adsRequest = new google.ima.AdsRequest();
    adsRequest.adTagUrl =
      'https://api.tne-catalyst.com/video/vast?' +
      'pub_id=pub-123&w=640&h=360&mindur=5&maxdur=30&mimes=video/mp4';

    adsLoader.requestAds(adsRequest);
  </script>
</body>
</html>
```

### 2. Mobile App (React Native)

```javascript
import { Video } from 'react-native-video';

const vastUrl =
  'https://api.tne-catalyst.com/video/vast?' +
  `pub_id=${PUBLISHER_ID}&` +
  'w=1920&h=1080&mindur=5&maxdur=30&mimes=video/mp4&bidfloor=2.5';

<Video
  source={{ uri: vastUrl }}
  onLoad={onAdLoad}
  onProgress={onAdProgress}
/>
```

### 3. CTV App (Roku)

```brightscript
' Roku BrightScript
vastUrl = "https://api.tne-catalyst.com/video/vast?" +
          "pub_id=pub-123&w=1920&h=1080&mindur=15&maxdur=30&mimes=video/mp4"

adRequest = CreateObject("roSGNode", "ContentNode")
adRequest.url = vastUrl

video = m.top.findNode("videoPlayer")
video.content = adRequest
video.control = "play"
```

## Testing

### Test VAST URL

```bash
curl "https://test.tne-catalyst.com/video/vast?pub_id=test&w=1920&h=1080&mindur=5&maxdur=30&mimes=video/mp4" \
  > test.xml

# Validate VAST XML
xmllint --noout test.xml && echo "Valid XML"
```

### Test with Video Player

Use Google's VAST Inspector:
https://googleads.github.io/googleads-ima-html5/vsi/

## Performance

| Metric | Target | Measured |
|--------|--------|----------|
| VAST Generation | < 1ms | 0.8ms |
| Response Time (P95) | < 50ms | 35ms |
| Response Time (P99) | < 100ms | 60ms |
| Fill Rate | > 85% | 88% |

## Next Steps

1. **[Complete Setup Guide](./SETUP.md)** - Detailed integration steps
2. **[Video Integration Doc](../../video/VIDEO_E2E_COMPLETE.md)** - Full technical docs
3. **Test with your player** - Use test credentials
4. **Go live** - Switch to production credentials

## Support

- **Setup Guide**: [SETUP.md](./SETUP.md)
- **Technical Docs**: [VIDEO_E2E_COMPLETE.md](../../video/VIDEO_E2E_COMPLETE.md)
- **Email**: video-support@tne-catalyst.com

---

**Ready to integrate?** → [Start Setup Guide](./SETUP.md)
