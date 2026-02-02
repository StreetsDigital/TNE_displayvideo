# Ad Tag Server Setup Guide

## Overview

This guide explains how to wire up the direct ad tag endpoints in your TNE Catalyst server.

## Quick Integration

Add these endpoints to your server initialization in `cmd/server/server.go`:

### 1. Import Ad Tag Handlers

```go
import (
	"github.com/thenexusengine/tne_springwire/internal/endpoints"
)
```

### 2. Initialize Handlers in `initHandlers()`

Add after your existing handlers:

```go
func (s *Server) initHandlers() {
	log := logger.Log

	// ... existing handlers ...

	// Ad tag handlers
	adTagHandler := endpoints.NewAdTagHandler(s.exchange)
	adTagGenerator := endpoints.NewAdTagGeneratorHandler(s.config.HostURL)

	// ... existing routes ...

	// Ad tag endpoints
	mux.HandleFunc("/ad/js", adTagHandler.HandleJavaScriptAd)
	mux.HandleFunc("/ad/iframe", adTagHandler.HandleIframeAd)
	mux.HandleFunc("/ad/gam", adTagHandler.HandleGAMAd)
	mux.HandleFunc("/ad/track", adTagHandler.HandleAdTracking)

	// Admin endpoints
	mux.HandleFunc("/admin/adtag/generator", adTagGenerator.HandleGeneratorUI)
	mux.HandleFunc("/admin/adtag/generate", adTagGenerator.HandleGenerateTag)

	// Static assets
	mux.HandleFunc("/assets/tne-ads.js", endpoints.HandleAssets)

	log.Info().Msg("Ad tag endpoints registered")

	// ... rest of initialization ...
}
```

## Complete Implementation

Here's the complete section to add to `cmd/server/server.go`:

```go
// In initHandlers() function, add after video handlers:

	// ====================
	// Ad Tag Integration
	// ====================

	// Ad tag handler for direct publisher integration
	adTagHandler := endpoints.NewAdTagHandler(s.exchange)
	adTagGenerator := endpoints.NewAdTagGeneratorHandler(s.config.HostURL)

	log.Info().Msg("Ad tag handlers initialized")

	// ... in the routes section ...

	// Ad serving endpoints
	mux.HandleFunc("/ad/js", adTagHandler.HandleJavaScriptAd)        // JavaScript ad serving
	mux.HandleFunc("/ad/iframe", adTagHandler.HandleIframeAd)        // Iframe ad serving
	mux.HandleFunc("/ad/gam", adTagHandler.HandleGAMAd)              // GAM 3rd party script
	mux.HandleFunc("/ad/track", adTagHandler.HandleAdTracking)       // Tracking pixel

	// Admin tag generator
	mux.HandleFunc("/admin/adtag/generator", adTagGenerator.HandleGeneratorUI)
	mux.HandleFunc("/admin/adtag/generate", adTagGenerator.HandleGenerateTag)

	// Static assets
	mux.HandleFunc("/assets/", endpoints.HandleAssets)

	log.Info().Msg("Ad tag endpoints registered: /ad/*, /admin/adtag/*")
```

## Endpoints Reference

### Ad Serving Endpoints

#### GET /ad/js
**Purpose**: Serve JavaScript ad creative

**Parameters**:
- `pub` (required) - Publisher ID
- `placement` (required) - Placement ID
- `div` (required) - Container div ID
- `w` (required) - Width in pixels
- `h` (required) - Height in pixels
- `url` (optional) - Page URL
- `domain` (optional) - Site domain
- `kw` (optional) - Comma-separated keywords

**Response**: JavaScript code that renders the ad

**Example**:
```
GET /ad/js?pub=pub-123&placement=banner-1&div=ad-1&w=300&h=250
```

#### GET /ad/iframe
**Purpose**: Serve ad in iframe

**Parameters**: Same as `/ad/js`

**Response**: Complete HTML document with ad creative

**Example**:
```
GET /ad/iframe?pub=pub-123&placement=banner-1&w=728&h=90
```

#### GET /ad/gam
**Purpose**: Serve ad for GAM 3rd party script

**Parameters**: Same as `/ad/js`

**Response**: GAM-compatible JavaScript code

**Example**:
```
GET /ad/gam?pub=pub-123&placement=banner-1&div=ad-gam-1&w=300&h=250
```

#### GET /ad/track
**Purpose**: Track ad events (impressions, clicks)

**Parameters**:
- `bid` (required) - Bid ID
- `placement` (required) - Placement ID
- `event` (required) - Event type (impression, click, etc.)
- `ts` (optional) - Timestamp

**Response**: 1x1 transparent GIF

**Example**:
```
GET /ad/track?bid=bid-123&placement=banner-1&event=impression
```

### Admin Endpoints

#### GET /admin/adtag/generator
**Purpose**: UI for generating ad tags

**Response**: HTML tag generator interface

**Access**: Admin only

#### POST /admin/adtag/generate
**Purpose**: API for generating ad tags

**Request Body**:
```json
{
  "publisherId": "pub-123456",
  "placementId": "homepage-banner-1",
  "width": 300,
  "height": 250,
  "format": "async"
}
```

**Response**:
```json
{
  "html": "<!-- Generated HTML tag -->",
  "javascript": "tne.display({...});",
  "iframeUrl": "https://...",
  "gamScript": "<script>...</script>"
}
```

### Static Assets

#### GET /assets/tne-ads.js
**Purpose**: TNE Catalyst client-side SDK

**Response**: JavaScript SDK file

**Cache**: 1 hour

## Configuration

### Environment Variables

No additional environment variables required. The ad tag system uses existing configuration:

```bash
# Existing configuration
PBS_HOST_URL=https://ads.thenexusengine.com  # Used for ad tag generation
```

### CORS Configuration

Ensure CORS headers allow ad requests from publisher domains:

```go
// In middleware configuration
corsConfig := middleware.DefaultCORSConfig()
corsConfig.AllowOrigins = []string{"*"}  // or specific domains
```

## Testing

### 1. Test Ad Serving

```bash
# Test JavaScript ad endpoint
curl "http://localhost:8000/ad/js?pub=test-pub&placement=test&div=ad-1&w=300&h=250"

# Expected: JavaScript code
```

### 2. Test Iframe Ad

```bash
# Test iframe endpoint
curl "http://localhost:8000/ad/iframe?pub=test-pub&placement=test&w=728&h=90"

# Expected: HTML document
```

### 3. Test Tag Generator

```bash
# Open in browser
open http://localhost:8000/admin/adtag/generator
```

### 4. Test Client SDK

```bash
# Check SDK loads
curl http://localhost:8000/assets/tne-ads.js

# Expected: JavaScript SDK code
```

### 5. Integration Test

Create a test HTML file:

```html
<!DOCTYPE html>
<html>
<head>
  <title>Ad Tag Test</title>
</head>
<body>
  <h1>Test Ad</h1>

  <!-- Ad Container -->
  <div id="test-ad" style="width:300px;height:250px;border:1px solid #ddd;"></div>

  <!-- Ad Tag -->
  <script>
  var tne = tne || {};
  tne.cmd = tne.cmd || [];
  tne.cmd.push(function() {
    tne.setConfig({ debug: true });  // Enable debug logging
    tne.display({
      publisherId: 'test-pub',
      placementId: 'test-placement',
      divId: 'test-ad',
      size: [300, 250],
      serverUrl: 'http://localhost:8000'
    });
  });
  </script>
  <script async src="http://localhost:8000/assets/tne-ads.js"></script>
</body>
</html>
```

Open in browser and check:
1. Ad loads successfully
2. No console errors
3. Impression tracked in logs
4. Ad renders correctly

## Monitoring

### Metrics to Track

Add these Prometheus metrics (optional):

```go
// In internal/metrics/prometheus.go

adTagRequests = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "catalyst_adtag_requests_total",
        Help: "Total ad tag requests",
    },
    []string{"format", "status"},  // format: js/iframe/gam, status: success/failure
)

adTagLatency = prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name: "catalyst_adtag_latency_seconds",
        Help: "Ad tag request latency",
        Buckets: prometheus.DefBuckets,
    },
    []string{"format"},
)
```

### Log Monitoring

Watch for ad tag requests in logs:

```bash
# Watch ad tag requests
tail -f /var/log/catalyst/server.log | grep "/ad/"

# Watch tracking events
tail -f /var/log/catalyst/server.log | grep "Ad tracking event"
```

### Health Checks

Add health check for ad tag system:

```go
// In readyHandler
checks["adtag"] = map[string]interface{}{
    "status": "healthy",
    "endpoints": []string{"/ad/js", "/ad/iframe", "/ad/gam"},
}
```

## Security Considerations

### 1. Publisher Authentication

Add authentication for ad requests (optional):

```go
// In adTagHandler
if !s.authenticatePublisher(params.PublisherID, r) {
    http.Error(w, "Unauthorized", http.StatusUnauthorized)
    return
}
```

### 2. Rate Limiting

Apply rate limiting to ad endpoints:

```go
// In middleware chain
handler = s.rateLimiter.Middleware(handler)
```

### 3. Input Validation

Validate all parameters:

```go
// Already implemented in parseAdParams()
if width <= 0 || height <= 0 {
    return nil
}
```

### 4. XSS Prevention

Sanitize ad creative HTML:

```go
// Already implemented in sanitizeHTML()
html = strings.ReplaceAll(html, "<script>", "&lt;script&gt;")
```

## Performance Optimization

### 1. Caching

Cache static assets:

```go
w.Header().Set("Cache-Control", "public, max-age=3600")
```

### 2. CDN Integration

Serve assets from CDN:

```javascript
serverUrl: 'https://cdn.thenexusengine.com'
```

### 3. Compression

Enable gzip compression (already configured in middleware).

## Troubleshooting

### Issue: 404 Not Found

**Cause**: Endpoints not registered

**Fix**: Ensure handlers are registered in `initHandlers()`

### Issue: CORS Errors

**Cause**: CORS not configured for publisher domains

**Fix**: Update CORS configuration

```go
corsConfig.AllowOrigins = []string{"https://publisher.com"}
```

### Issue: Ads Not Displaying

**Cause**: No winning bids in auction

**Fix**: Check auction configuration and bidder setup

### Issue: Tracking Not Working

**Cause**: Ad blocker or incorrect tracking URL

**Fix**: Use first-party tracking domain

## Next Steps

1. ✅ Add endpoints to server
2. ✅ Test locally
3. ✅ Deploy to staging
4. ✅ Test on staging
5. ✅ Deploy to production
6. ✅ Generate production ad tags
7. ✅ Provide tags to publishers

## Support

For issues or questions:
- Email: ops@thenexusengine.io
- Docs: /docs/integrations/DIRECT_AD_TAG_INTEGRATION.md

---

**Status**: Ready for Integration
**Estimated Time**: 30 minutes
**Difficulty**: Easy
