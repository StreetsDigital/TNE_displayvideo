# In-App SDK Integration

**Status:** ❌ SDK Not Built - Backend Ready
**Timeline:** 4-6 weeks to complete
**Difficulty:** Hard
**Best For:** Mobile app developers (iOS, Android, React Native)

## Overview

Native SDK for integrating TNE Catalyst ads directly into mobile applications. Provides a simple API for requesting and displaying ads without needing to understand OpenRTB protocol details.

## Current Status

✅ **Backend Complete:**
- OpenRTB 2.5 auction endpoint
- Full ad format support (banner, video, native)
- Privacy compliance (GDPR/CCPA/COPPA)
- Mobile device detection
- App-specific targeting

❌ **SDK Not Built:**
- No iOS SDK
- No Android SDK
- No React Native SDK
- No Flutter SDK
- No Unity SDK

## What It Will Look Like (When Complete)

### iOS (Swift)

```swift
import TNECatalystSDK

// Initialize SDK
TNECatalyst.shared.configure(
    publisherId: "pub-mobile-123",
    apiKey: "your-api-key"
)

// Request banner ad
let bannerView = TNEBannerAdView(frame: CGRect(x: 0, y: 0, width: 320, height: 50))
bannerView.delegate = self
bannerView.loadAd()

// Request interstitial ad
let interstitial = TNEInterstitialAd()
interstitial.delegate = self
interstitial.load()
interstitial.present(from: self)

// Request rewarded video
let rewardedAd = TNERewardedAd()
rewardedAd.delegate = self
rewardedAd.load()
rewardedAd.present(from: self)
```

### Android (Kotlin)

```kotlin
import com.tne.catalyst.sdk.*

// Initialize SDK
TNECatalyst.initialize(
    context = this,
    publisherId = "pub-mobile-123",
    apiKey = "your-api-key"
)

// Request banner ad
val bannerView = TNEBannerAdView(this)
bannerView.adUnitId = "banner-123"
bannerView.setAdSize(AdSize.BANNER_320x50)
bannerView.adListener = object : AdListener {
    override fun onAdLoaded() { }
    override fun onAdFailedToLoad(error: LoadAdError) { }
}
bannerView.loadAd()

// Request interstitial ad
val interstitial = TNEInterstitialAd(this)
interstitial.adUnitId = "interstitial-123"
interstitial.load()

// Request rewarded video
val rewardedAd = TNERewardedAd(this)
rewardedAd.adUnitId = "rewarded-123"
rewardedAd.load()
```

### React Native

```javascript
import { TNECatalyst, BannerAd, InterstitialAd, RewardedAd } from 'tne-catalyst-sdk';

// Initialize
TNECatalyst.initialize({
  publisherId: 'pub-mobile-123',
  apiKey: 'your-api-key'
});

// Banner component
<BannerAd
  adUnitId="banner-123"
  size="320x50"
  onAdLoaded={() => console.log('Ad loaded')}
  onAdFailedToLoad={(error) => console.log('Failed:', error)}
/>

// Interstitial
await InterstitialAd.load('interstitial-123');
await InterstitialAd.show();

// Rewarded Video
await RewardedAd.load('rewarded-123');
await RewardedAd.show();
```

## Planned Features

### Ad Formats
- Banner ads (multiple sizes)
- Interstitial ads (full-screen)
- Rewarded video ads
- Native ads (customizable)
- Video ads (in-stream)

### Privacy & Compliance
- GDPR consent management
- CCPA opt-out
- COPPA compliance
- ATT (App Tracking Transparency) support
- IDFA handling

### Targeting
- Geo targeting
- Device targeting
- App category targeting
- Custom user targeting
- First-party data

### Advanced Features
- Mediation support
- Ad caching
- Preloading
- Auto-refresh
- Frequency capping
- A/B testing

### Analytics
- Impression tracking
- Click-through tracking
- Viewability
- Fill rate
- Revenue reporting

## Use Cases

### 1. Free Apps with Ads
Monetize free mobile apps with banner, interstitial, and video ads.

### 2. Gaming Apps
Reward users with in-game currency for watching rewarded video ads.

### 3. News/Content Apps
Display native ads that blend with content.

### 4. Utility Apps
Show banner ads without disrupting user experience.

## What Needs to Be Built

See [WORK_REQUIRED.md](./WORK_REQUIRED.md) for complete roadmap:

### Core SDKs (Required)
1. **iOS SDK** (Swift)
2. **Android SDK** (Kotlin)
3. **React Native SDK** (JavaScript/Native)

### Additional SDKs (Optional)
4. Flutter SDK (Dart)
5. Unity SDK (C#)
6. Xamarin SDK (C#)

### Supporting Infrastructure
7. SDK documentation
8. Sample apps
9. Integration guides
10. Testing tools

## Estimated Timeline

| SDK | Effort | Priority |
|-----|--------|----------|
| iOS (Swift) | 3-4 weeks | Critical |
| Android (Kotlin) | 3-4 weeks | Critical |
| React Native | 2 weeks | High |
| Documentation | 2 weeks | Critical |
| Sample Apps | 1 week | High |
| Flutter | 2-3 weeks | Medium |
| Unity | 2-3 weeks | Low |
| **Total (Core)** | **10-12 weeks** | - |

## Technical Requirements

### SDK Architecture
- Native iOS/Android implementation
- OpenRTB protocol handling
- Ad rendering engine
- Event tracking
- Error handling
- Caching layer
- Privacy management

### Distribution
- CocoaPods (iOS)
- Swift Package Manager (iOS)
- Maven Central (Android)
- NPM (React Native)
- Flutter pub (Flutter)
- Unity Asset Store (Unity)

### Minimum Requirements
- **iOS:** iOS 12.0+, Swift 5.0+
- **Android:** Android 5.0+ (API 21+), Kotlin 1.5+
- **React Native:** RN 0.63+

## Backend Endpoints (Already Available)

The SDK will use these existing endpoints:

```
POST /openrtb2/auction
GET /cookie_sync
GET /setuid
GET /optout
GET /health
```

All endpoints support mobile-specific parameters.

## Next Steps

1. **Review**: [WORK_REQUIRED.md](./WORK_REQUIRED.md) - See complete development plan
2. **Contact**: Email sdk-development@tne-catalyst.com for discussions
3. **Partnership**: Inquire about white-label SDK opportunities

## Temporary Workaround

Until SDK is built, mobile apps can:

1. **Use WebView** - Embed web ads via WebView
2. **Direct OpenRTB** - Call `/openrtb2/auction` directly
3. **Third-party Mediation** - Use existing mediation platforms

Example direct integration:

```swift
// iOS - Direct OpenRTB call
func requestAd() {
    let url = URL(string: "https://api.tne-catalyst.com/openrtb2/auction")!
    var request = URLRequest(url: url)
    request.httpMethod = "POST"
    request.setValue("application/json", forHTTPHeaderField: "Content-Type")
    request.setValue(apiKey, forHTTPHeaderField: "X-API-Key")

    let openRTBRequest = createOpenRTBRequest()
    request.httpBody = try? JSONEncoder().encode(openRTBRequest)

    URLSession.shared.dataTask(with: request) { data, response, error in
        // Handle response
    }.resume()
}
```

## Support

- **Development Status**: [WORK_REQUIRED.md](./WORK_REQUIRED.md)
- **Email**: sdk-development@tne-catalyst.com
- **Notify Me**: Request notification when SDK is available

---

**Want to help build the SDK?** → Email sdk-development@tne-catalyst.com

**Need mobile monetization now?** → Use [OpenRTB Direct](../openrtb-direct/) integration
