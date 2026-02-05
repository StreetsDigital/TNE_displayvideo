# Catalyst SDK User Sync - Publisher Guide

## What is User Sync?

User sync (also called cookie sync) allows ad bidders to recognize your users, which enables them to:
- Bid more accurately based on user data
- Offer higher CPMs (typically 10-30% higher)
- Improve fill rates
- Provide better ad relevance

The Catalyst SDK now **automatically syncs users** with all configured bidders when your page loads.

## Quick Start

### Basic Setup (Recommended)

User sync is **enabled by default**. Just initialize the SDK normally:

```javascript
catalyst.init({
  accountId: 'your-account-id'
});
```

That's it! The SDK will automatically:
- Wait 1 second after initialization
- Sync with 5 bidders: Kargo, Rubicon, Pubmatic, Sovrn, Triplelift
- Respect user privacy (GDPR/CCPA)
- Fire up to 5 sync pixels/iframes

### See It in Action

Enable debug mode to watch the sync process:

```javascript
catalyst.init({
  accountId: 'your-account-id',
  debug: true
});
```

Open browser console to see:
```
[Catalyst] Catalyst SDK initialized with accountId: your-account-id
[Catalyst] Starting user sync for bidders: kargo,rubicon,pubmatic,sovrn,triplelift
[Catalyst] Fired iframe sync for kargo
[Catalyst] Fired pixel sync for rubicon
[Catalyst] Fired 5 user syncs
```

## Configuration Options

### Disable User Sync

If you don't want automatic user sync:

```javascript
catalyst.init({
  accountId: 'your-account-id',
  userSync: false  // Disable entirely
});
```

### Customize Bidders

Sync with specific bidders only:

```javascript
catalyst.init({
  accountId: 'your-account-id',
  userSync: {
    bidders: ['kargo', 'rubicon']  // Only these two
  }
});
```

Available bidders:
- `kargo`
- `rubicon`
- `pubmatic`
- `sovrn`
- `triplelift`

### Adjust Sync Timing

Change when sync occurs:

```javascript
catalyst.init({
  accountId: 'your-account-id',
  userSync: {
    syncDelay: 2000  // Wait 2 seconds after init (default: 1000)
  }
});
```

### Limit Number of Syncs

Reduce the number of syncs per page:

```javascript
catalyst.init({
  accountId: 'your-account-id',
  userSync: {
    maxSyncs: 3  // Only sync with 3 bidders (default: 5)
  }
});
```

### Disable Sync Types

Control which sync methods are used:

```javascript
catalyst.init({
  accountId: 'your-account-id',
  userSync: {
    iframeEnabled: true,   // Allow iframe syncs (default: true)
    pixelEnabled: false    // Disable pixel syncs (default: true)
  }
});
```

### Advanced Configuration

Combine multiple options:

```javascript
catalyst.init({
  accountId: 'your-account-id',
  debug: true,
  userSync: {
    enabled: true,
    bidders: ['kargo', 'rubicon', 'pubmatic'],
    syncDelay: 1500,
    iframeEnabled: true,
    pixelEnabled: true,
    maxSyncs: 3
  }
});
```

## Privacy & Compliance

### GDPR (European Users)

The SDK **automatically checks** for GDPR consent before syncing:

```javascript
// SDK checks for Consent Management Platform (CMP)
if (window.__tcfapi) {
  // Checks if user has given consent
  // Only syncs if consent granted
}
```

**You don't need to do anything** - the SDK handles it automatically.

### CCPA (US Users)

The SDK **automatically checks** for CCPA opt-out:

```javascript
// SDK checks for US Privacy API
if (window.__uspapi) {
  // Checks if user has opted out
  // Blocks sync if user opted out
}
```

**You don't need to do anything** - the SDK handles it automatically.

### No Consent Framework

If you don't use a consent management platform:
- User sync will proceed normally
- This is appropriate for non-regulated regions
- Or if you handle consent differently

## Performance Impact

### Page Load

User sync **does not block** your page or ad requests:
- Sync happens **after** SDK initialization
- Runs **asynchronously** with 1-second delay (configurable)
- Bid requests proceed immediately without waiting for sync

### Network Traffic

User sync adds minimal network overhead:
- 1 POST request to `/cookie_sync` (~1 KB)
- 5 sync pixels/iframes (~5 requests total)
- All async and non-blocking
- Only happens once per user session

### When Sync Happens

User sync fires:
- ✅ Once per page load (on initialization)
- ❌ Not on every bid request
- ❌ Not on user interaction
- ❌ Not repeatedly

## Verification

### Check if Sync is Working

1. **Enable debug mode**:
   ```javascript
   catalyst.init({ accountId: 'xxx', debug: true });
   ```

2. **Open browser DevTools console**

3. **Look for these messages**:
   ```
   [Catalyst] Starting user sync for bidders: [...]
   [Catalyst] Fired iframe sync for kargo
   [Catalyst] Fired pixel sync for rubicon
   [Catalyst] Fired 5 user syncs
   ```

4. **Check Network tab**:
   - Look for POST to `/cookie_sync`
   - Should return 200 OK
   - Should see subsequent requests to bidder URLs

5. **Check Elements tab**:
   - Search for `iframe[data-bidder]`
   - Search for `img[data-bidder]`
   - Should see hidden elements for each synced bidder

### Test Page

Use the provided test page to verify sync functionality:

```
https://ads.thenexusengine.com/assets/test-user-sync.html
```

Features:
- Real-time console output
- Visual display of sync elements
- Manual sync trigger button
- Configuration display

## Troubleshooting

### "User sync disabled" in console

**Cause**: User sync is disabled in configuration

**Solution**:
```javascript
catalyst.init({
  accountId: 'your-account-id',
  userSync: true  // or just omit (enabled by default)
});
```

### "User sync blocked by privacy settings"

**Cause**: User has not consented or has opted out (GDPR/CCPA)

**This is correct behavior** - respecting user privacy. No action needed.

### "User sync already performed"

**Cause**: Sync only happens once per page load

**This is correct behavior** - prevents duplicate syncs. Refresh page to sync again.

### No sync messages in console

**Check**:
1. Is debug mode enabled? `debug: true`
2. Is user sync enabled? Check config
3. Check JavaScript console for errors
4. Verify SDK loaded correctly

### Sync request fails (Network error)

**Check**:
1. Is `/cookie_sync` endpoint available?
2. Check Network tab for HTTP errors
3. Verify server is running
4. Check for CORS issues

## Expected Results

### Immediate (Day 1)
- User sync fires successfully
- Sync pixels/iframes created in DOM
- No errors in console
- No page performance degradation

### Short-term (Week 1)
- Bidders recognize more users
- Increased bid participation
- Higher fill rates

### Medium-term (Month 1)
- **10-30% CPM increase** for synced users
- Better bid competition
- Improved overall revenue

## Best Practices

### ✅ Recommended

1. **Keep default settings** (works for most publishers)
2. **Enable debug mode** during initial testing
3. **Monitor CPMs** after implementation
4. **Use test page** to verify setup

### ⚠️ Consider

1. **Adjust syncDelay** if you have slow page loads
2. **Reduce maxSyncs** if you have bandwidth concerns
3. **Customize bidders** if you only work with specific SSPs

### ❌ Avoid

1. **Don't disable user sync** unless required (loses revenue)
2. **Don't set syncDelay too low** (< 500ms) - may impact performance
3. **Don't set maxSyncs too high** (> 10) - diminishing returns

## Integration Examples

### WordPress

Add to your theme's header or footer:

```html
<script src="https://ads.thenexusengine.com/assets/catalyst-sdk.js"></script>
<script>
  catalyst.init({
    accountId: 'your-account-id',
    debug: false
  });
</script>
```

### React

```jsx
import { useEffect } from 'react';

function App() {
  useEffect(() => {
    // Load Catalyst SDK
    const script = document.createElement('script');
    script.src = 'https://ads.thenexusengine.com/assets/catalyst-sdk.js';
    script.async = true;
    script.onload = () => {
      window.catalyst.init({
        accountId: 'your-account-id'
      });
    };
    document.body.appendChild(script);
  }, []);

  return <div>Your App</div>;
}
```

### Next.js

```jsx
// pages/_app.js
import Script from 'next/script';

function MyApp({ Component, pageProps }) {
  return (
    <>
      <Script
        src="https://ads.thenexusengine.com/assets/catalyst-sdk.js"
        strategy="afterInteractive"
        onLoad={() => {
          window.catalyst.init({
            accountId: 'your-account-id'
          });
        }}
      />
      <Component {...pageProps} />
    </>
  );
}
```

### Google Tag Manager

1. Create new **Custom HTML** tag
2. Add this code:
   ```html
   <script src="https://ads.thenexusengine.com/assets/catalyst-sdk.js"></script>
   <script>
     catalyst.init({
       accountId: 'your-account-id'
     });
   </script>
   ```
3. Set trigger to **All Pages**
4. Publish container

## Support

If you encounter issues:

1. **Check console** with debug mode enabled
2. **Use test page** to verify setup
3. **Review Network tab** for failed requests
4. **Contact support** with:
   - Your account ID
   - Console logs (with debug enabled)
   - Network request details
   - Browser and OS version

## FAQ

### Q: Will this slow down my site?

**A:** No. User sync is asynchronous, delayed by 1 second, and doesn't block ad requests or page load.

### Q: How often does sync happen?

**A:** Once per page load, only if the user hasn't already been synced in the current session.

### Q: Do I need consent from users?

**A:** The SDK automatically checks for GDPR/CCPA consent and respects user choices. If you use a Consent Management Platform (CMP), no additional work needed.

### Q: Can I see which users are synced?

**A:** User sync happens client-side. Bidders receive the sync and match users on their end. You'll see the impact in higher CPMs and better fill rates.

### Q: What if a bidder sync fails?

**A:** The SDK handles errors gracefully. Failed syncs don't affect other syncs or bid requests. Enable debug mode to see which syncs succeeded.

### Q: Can I sync with my own bidders?

**A:** Yes! Contact support to add your bidders to the server configuration. Then add them to your SDK config:
```javascript
userSync: { bidders: ['your-bidder', 'kargo', ...] }
```

### Q: Does this work with all browsers?

**A:** Yes. Works with all modern browsers (Chrome, Firefox, Safari, Edge). Uses standard JavaScript APIs.

### Q: What's the difference between iframe and pixel sync?

**A:**
- **Iframe sync**: Loads bidder page in hidden iframe (more reliable, supports cookies)
- **Pixel sync**: Loads 1x1 transparent image (lightweight, simpler)

Both work well. The SDK uses whichever type each bidder requires.

## Summary

**Default behavior** (recommended for most publishers):
```javascript
catalyst.init({
  accountId: 'your-account-id'
});
```

This gives you:
- ✅ Automatic user sync with 5 major bidders
- ✅ Privacy-compliant (GDPR/CCPA)
- ✅ Non-blocking and performant
- ✅ Once per session
- ✅ 10-30% CPM increase expected

**That's all you need!** The SDK handles everything else automatically.
