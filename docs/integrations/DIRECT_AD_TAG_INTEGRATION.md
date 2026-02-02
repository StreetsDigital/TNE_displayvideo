# Direct Ad Tag Integration Guide

## Overview

TNE Catalyst provides direct ad tag integration, allowing publishers to embed ads of any size directly into their web pages. The ads connect to your server and can integrate seamlessly with Google Ad Manager (GAM) via 3rd party script creative type.

## Quick Start

### 1. Generate Your Ad Tag

Visit the tag generator:
```
https://ads.thenexusengine.com/admin/adtag/generator
```

**Fill in:**
- Publisher ID (e.g., `pub-123456`)
- Placement ID (e.g., `homepage-banner-1`)
- Ad Size (e.g., 300x250)
- Format (Async JavaScript, GAM, Iframe, or Sync)

### 2. Copy & Paste

Copy the generated code and paste it into your HTML where you want the ad to appear:

```html
<!-- TNE Catalyst Ad Tag - 300x250 -->
<div id="tne-ad-homepage-banner-1" style="width:300px;height:250px;"></div>
<script>
(function() {
  var tne = tne || {};
  tne.cmd = tne.cmd || [];
  tne.cmd.push(function() {
    tne.display({
      publisherId: 'pub-123456',
      placementId: 'homepage-banner-1',
      divId: 'tne-ad-homepage-banner-1',
      size: [300, 250],
      serverUrl: 'https://ads.thenexusengine.com'
    });
  });
})();
</script>
<script async src="https://ads.thenexusengine.com/assets/tne-ads.js"></script>
```

### 3. Test

Open your page in a browser. The ad should load automatically!

## Integration Methods

### Method 1: Async JavaScript (Recommended)

**Best for:** Modern websites, SPAs, responsive sites

**Advantages:**
- ✅ Non-blocking page load
- ✅ Fully responsive
- ✅ Auto-refresh support
- ✅ Advanced targeting

**Code:**
```html
<div id="tne-ad-slot-1" style="width:728px;height:90px;"></div>
<script>
var tne = tne || {};
tne.cmd = tne.cmd || [];
tne.cmd.push(function() {
  tne.display({
    publisherId: 'pub-123456',
    placementId: 'leaderboard-top',
    divId: 'tne-ad-slot-1',
    size: [728, 90],
    serverUrl: 'https://ads.thenexusengine.com'
  });
});
</script>
<script async src="https://ads.thenexusengine.com/assets/tne-ads.js"></script>
```

### Method 2: GAM 3rd Party Script

**Best for:** Google Ad Manager integration

**Advantages:**
- ✅ Seamless GAM integration
- ✅ Works in GAM creative templates
- ✅ Full tracking support
- ✅ Automated rendering

**GAM Setup:**

1. In GAM, create a new **3rd Party Tag** creative
2. Paste this code in the "Creative Code" field:

```html
<script>
(function() {
  var tneConfig = {
    publisherId: 'pub-123456',
    placementId: '%%PATTERN:url%%',
    width: %%WIDTH%%,
    height: %%HEIGHT%%,
    serverUrl: 'https://ads.thenexusengine.com',
    pageUrl: '%%PATTERN:url%%',
    domain: '%%PATTERN:url%%'
  };

  var container = document.createElement('div');
  container.id = 'tne-gam-' + tneConfig.placementId;
  container.style.width = tneConfig.width + 'px';
  container.style.height = tneConfig.height + 'px';
  document.write(container.outerHTML);

  var script = document.createElement('script');
  script.src = tneConfig.serverUrl + '/ad/gam?' +
    'pub=' + encodeURIComponent(tneConfig.publisherId) +
    '&placement=' + encodeURIComponent(tneConfig.placementId) +
    '&w=' + tneConfig.width +
    '&h=' + tneConfig.height +
    '&div=' + encodeURIComponent(container.id) +
    '&url=' + encodeURIComponent(tneConfig.pageUrl) +
    '&domain=' + encodeURIComponent(tneConfig.domain);
  script.async = true;
  document.body.appendChild(script);
})();
</script>
```

**GAM Macros:**
- `%%WIDTH%%` - Ad width
- `%%HEIGHT%%` - Ad height
- `%%PATTERN:url%%` - Page URL
- Custom macros can be passed as query parameters

### Method 3: Iframe

**Best for:** Maximum isolation, third-party sites

**Advantages:**
- ✅ Complete security isolation
- ✅ No JavaScript conflicts
- ✅ Simple implementation
- ✅ Works everywhere

**Code:**
```html
<iframe src="https://ads.thenexusengine.com/ad/iframe?pub=pub-123456&placement=sidebar-1&w=300&h=600"
        width="300"
        height="600"
        frameborder="0"
        scrolling="no"
        loading="lazy">
</iframe>
```

### Method 4: Sync JavaScript

**Best for:** Legacy sites, simple implementation

**Advantages:**
- ✅ Simple one-line code
- ✅ Works on old browsers
- ✅ Guaranteed placement

**Code:**
```html
<div id="tne-ad-footer" style="width:970px;height:250px;"></div>
<script src="https://ads.thenexusengine.com/ad/js?pub=pub-123456&placement=footer-1&div=tne-ad-footer&w=970&h=250"></script>
```

## Common Ad Sizes

### Desktop

| Size | Name | Use Case |
|------|------|----------|
| 728x90 | Leaderboard | Header/footer banner |
| 970x250 | Billboard | Premium header placement |
| 300x250 | Medium Rectangle | Sidebar, in-content |
| 300x600 | Half Page | Sidebar |
| 160x600 | Wide Skyscraper | Sidebar |
| 970x90 | Super Leaderboard | Header |

### Mobile

| Size | Name | Use Case |
|------|------|----------|
| 320x50 | Mobile Banner | Top/bottom |
| 320x100 | Large Mobile Banner | Top/bottom |
| 300x250 | Mobile Rectangle | In-content |
| 336x280 | Large Rectangle | In-content |

### Video

| Size | Name | Use Case |
|------|------|----------|
| 1920x1080 | Full HD | Video player |
| 640x480 | Standard | Video player |
| 480x360 | Small | Video player |

## Advanced Configuration

### Auto-Refresh Ads

Refresh ads automatically every N seconds:

```javascript
tne.display({
  publisherId: 'pub-123456',
  placementId: 'refreshing-ad',
  divId: 'tne-ad-1',
  size: [300, 250],
  serverUrl: 'https://ads.thenexusengine.com',
  refreshRate: 30  // Refresh every 30 seconds
});
```

### Keyword Targeting

Pass keywords for better targeting:

```javascript
tne.display({
  publisherId: 'pub-123456',
  placementId: 'article-ad',
  divId: 'tne-ad-1',
  size: [728, 90],
  serverUrl: 'https://ads.thenexusengine.com',
  keywords: ['technology', 'smartphones', 'reviews']
});
```

### Custom Data

Pass custom key-value pairs:

```javascript
tne.display({
  publisherId: 'pub-123456',
  placementId: 'custom-ad',
  divId: 'tne-ad-1',
  size: [300, 600],
  serverUrl: 'https://ads.thenexusengine.com',
  customData: {
    section: 'sports',
    author: 'john-doe',
    premium: 'true'
  }
});
```

### Multiple Ads

Place multiple ads on the same page:

```html
<!-- Header Ad -->
<div id="tne-ad-header" style="width:728px;height:90px;"></div>

<!-- Sidebar Ad -->
<div id="tne-ad-sidebar" style="width:300px;height:250px;"></div>

<!-- Footer Ad -->
<div id="tne-ad-footer" style="width:970px;height:250px;"></div>

<script>
var tne = tne || {};
tne.cmd = tne.cmd || [];

tne.cmd.push(function() {
  // Header
  tne.display({
    publisherId: 'pub-123456',
    placementId: 'header',
    divId: 'tne-ad-header',
    size: [728, 90],
    serverUrl: 'https://ads.thenexusengine.com'
  });

  // Sidebar
  tne.display({
    publisherId: 'pub-123456',
    placementId: 'sidebar',
    divId: 'tne-ad-sidebar',
    size: [300, 250],
    serverUrl: 'https://ads.thenexusengine.com'
  });

  // Footer
  tne.display({
    publisherId: 'pub-123456',
    placementId: 'footer',
    divId: 'tne-ad-footer',
    size: [970, 250],
    serverUrl: 'https://ads.thenexusengine.com'
  });
});
</script>
<script async src="https://ads.thenexusengine.com/assets/tne-ads.js"></script>
```

## API Reference

### tne.display(options)

Display an ad in a container.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| publisherId | string | Yes | Publisher account ID |
| placementId | string | Yes | Ad placement/slot ID |
| divId | string | Yes | Container div ID |
| size | array | Yes | [width, height] in pixels |
| serverUrl | string | No | Ad server URL (default: current domain) |
| pageUrl | string | No | Page URL (default: window.location.href) |
| domain | string | No | Site domain (default: window.location.hostname) |
| keywords | array | No | Targeting keywords |
| customData | object | No | Custom key-value pairs |
| refreshRate | number | No | Auto-refresh interval in seconds (0 = disabled) |

**Example:**
```javascript
tne.display({
  publisherId: 'pub-123456',
  placementId: 'homepage-banner',
  divId: 'ad-container',
  size: [300, 250],
  serverUrl: 'https://ads.thenexusengine.com',
  keywords: ['tech', 'news'],
  refreshRate: 60
});
```

### tne.refreshAd(divId)

Manually refresh an ad slot.

**Parameters:**
- `divId` (string) - Container div ID

**Example:**
```javascript
tne.refreshAd('tne-ad-1');
```

### tne.destroySlot(divId)

Remove an ad slot.

**Parameters:**
- `divId` (string) - Container div ID

**Example:**
```javascript
tne.destroySlot('tne-ad-1');
```

### tne.setConfig(config)

Update global configuration.

**Parameters:**
- `config` (object) - Configuration object

**Example:**
```javascript
tne.setConfig({
  serverUrl: 'https://ads.thenexusengine.com',
  debug: true,
  timeout: 3000
});
```

## Tracking & Analytics

### Automatic Tracking

The SDK automatically tracks:
- **Impressions** - When ad is displayed
- **Clicks** - When user clicks on ad
- **Viewability** - When ad is in viewport

### Custom Tracking

Track custom events:

```javascript
// Track impression
tne.trackImpression('bid-id-123', 'placement-id');

// Track click
tne.trackClick('bid-id-123', 'placement-id');
```

### Access Tracking Data

Track data is available at:
```
GET /admin/metrics?placement=<placement-id>&date=<date>
```

## Troubleshooting

### Ad Not Showing

**1. Check container exists:**
```javascript
var container = document.getElementById('tne-ad-1');
console.log('Container:', container); // Should not be null
```

**2. Enable debug mode:**
```javascript
tne.setConfig({ debug: true });
// Check browser console for logs
```

**3. Verify publisher ID:**
```javascript
// Check if publisher ID is correct
console.log('Publisher ID:', 'pub-123456');
```

**4. Check network requests:**
- Open browser DevTools → Network tab
- Look for requests to `/ad/js` or `/ad/iframe`
- Check for errors (404, 500, etc.)

### Ad Size Issues

**Container too small:**
```html
<!-- Make sure container matches ad size -->
<div id="tne-ad-1" style="width:300px;height:250px;"></div>
<!-- Not: width:100px (too small!) -->
```

**Responsive sizing:**
```html
<div id="tne-ad-1" style="max-width:300px;height:250px;"></div>
```

### Script Loading Issues

**Async loading problems:**
```html
<!-- Ensure SDK loads before calling tne.display() -->
<script async src="https://ads.thenexusengine.com/assets/tne-ads.js"></script>

<!-- Use command queue -->
<script>
var tne = tne || {};
tne.cmd = tne.cmd || [];
tne.cmd.push(function() {
  // This will execute after SDK loads
  tne.display({...});
});
</script>
```

### CORS Issues

If you see CORS errors:

1. **Check server configuration** - Ensure CORS headers are set
2. **Use iframe method** - Avoids CORS completely
3. **Contact support** - May need to whitelist your domain

## Performance Optimization

### 1. Lazy Loading

Load ads only when visible:

```html
<div id="tne-ad-1"
     style="width:300px;height:250px;"
     data-lazy-ad="true"></div>

<script>
// Load ad when in viewport
var observer = new IntersectionObserver(function(entries) {
  entries.forEach(function(entry) {
    if (entry.isIntersecting) {
      tne.display({
        publisherId: 'pub-123456',
        placementId: 'lazy-ad',
        divId: entry.target.id,
        size: [300, 250]
      });
      observer.unobserve(entry.target);
    }
  });
});

observer.observe(document.getElementById('tne-ad-1'));
</script>
```

### 2. Async Loading

Always use async script loading:

```html
<!-- Good -->
<script async src="https://ads.thenexusengine.com/assets/tne-ads.js"></script>

<!-- Bad (blocks page load) -->
<script src="https://ads.thenexusengine.com/assets/tne-ads.js"></script>
```

### 3. Minimize Refreshes

Don't refresh too frequently:

```javascript
// Good: 30-60 seconds
refreshRate: 30

// Bad: Too frequent
refreshRate: 5
```

## Security Best Practices

### 1. Content Security Policy

Add to your CSP headers:

```
Content-Security-Policy:
  script-src 'self' https://ads.thenexusengine.com;
  img-src 'self' https://ads.thenexusengine.com;
  frame-src 'self' https://ads.thenexusengine.com;
```

### 2. Sandboxed Iframes

Use sandbox attribute:

```html
<iframe src="..."
        sandbox="allow-scripts allow-same-origin allow-popups"
        ...>
</iframe>
```

### 3. HTTPS Only

Always use HTTPS URLs:

```javascript
// Good
serverUrl: 'https://ads.thenexusengine.com'

// Bad
serverUrl: 'http://ads.thenexusengine.com'
```

## Support

### Resources
- **Tag Generator**: https://ads.thenexusengine.com/admin/adtag/generator
- **Documentation**: https://ads.thenexusengine.com/docs
- **API Reference**: https://ads.thenexusengine.com/docs/api

### Contact
- **Email**: ops@thenexusengine.io
- **Support**: https://support.thenexusengine.com

## Examples

See complete examples in:
- `docs/examples/adtag-async.html`
- `docs/examples/adtag-gam.html`
- `docs/examples/adtag-iframe.html`
- `docs/examples/adtag-responsive.html`

---

**Version**: 1.0.0
**Last Updated**: 2026-02-02
**Maintainer**: TNE Catalyst Team
