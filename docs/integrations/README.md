# TNE Catalyst Publisher Integrations

This directory contains comprehensive integration guides for publishers to connect with the TNE Catalyst ad exchange platform.

## Available Integration Methods

| Method | Status | Difficulty | Timeline | Best For |
|--------|--------|-----------|----------|----------|
| [OpenRTB Direct](#openrtb-direct) | ✅ Production Ready | Medium | Immediate | DSPs, SSPs, Direct integration |
| [Video VAST](#video-vast) | ✅ Production Ready | Easy | Immediate | Video publishers, CTV/OTT platforms |
| [Web via Prebid](#web-prebid) | ⚠️ Needs Examples | Easy | 1-2 weeks | Display publishers using Prebid.js |
| [Video via Prebid](#video-prebid) | ⚠️ Needs Examples | Medium | 1-2 weeks | Video publishers using Prebid.js |
| [In-App SDK](#in-app-sdk) | ❌ SDK Missing | Hard | 4-6 weeks | Mobile app developers |

## Quick Navigation

### Production Ready (Use Today)

1. **[OpenRTB Direct](./openrtb-direct/)** - Server-to-server integration
   - Complete API documentation
   - Full OpenRTB 2.5 support
   - GDPR/CCPA/COPPA compliant
   - [Get Started →](./openrtb-direct/README.md)

2. **[Video VAST](./video-vast/)** - Direct VAST tag integration
   - 3 endpoints (GET/POST)
   - VAST 2.0-4.0 support
   - CTV/OTT optimized
   - [Get Started →](./video-vast/README.md)

### Needs Documentation (Coming Soon)

3. **[Web via Prebid](./web-prebid/)** - Prebid.js for display ads
   - Backend ready
   - Needs client examples
   - [View Status →](./web-prebid/WORK_REQUIRED.md)

4. **[Video via Prebid](./video-prebid/)** - Prebid.js for video ads
   - Backend ready
   - Needs client examples
   - [View Status →](./video-prebid/WORK_REQUIRED.md)

### Under Development

5. **[In-App SDK](./in-app-sdk/)** - JavaScript/TypeScript SDK
   - Backend ready
   - SDK needs development
   - [View Roadmap →](./in-app-sdk/WORK_REQUIRED.md)

## Integration Decision Guide

### I'm a publisher with...

**Video inventory (CTV, OTT, in-stream)**
→ Use [Video VAST](./video-vast/) for direct integration
→ Use [Video via Prebid](./video-prebid/) if you already use Prebid.js

**Display inventory (banner, native)**
→ Use [Web via Prebid](./web-prebid/) if you already use Prebid.js
→ Use [OpenRTB Direct](./openrtb-direct/) for server-side integration

**Mobile app**
→ Use [In-App SDK](./in-app-sdk/) (coming soon)
→ Use [OpenRTB Direct](./openrtb-direct/) for server-side integration today

**Custom exchange/SSP**
→ Use [OpenRTB Direct](./openrtb-direct/)

## Support Matrix

| Feature | OpenRTB | VAST | Web Prebid | Video Prebid | In-App SDK |
|---------|---------|------|------------|--------------|------------|
| Display Ads | ✅ | ❌ | ✅ | ❌ | 🚧 |
| Video Ads | ✅ | ✅ | ❌ | ✅ | 🚧 |
| Native Ads | ✅ | ❌ | ✅ | ❌ | 🚧 |
| CTV/OTT | ✅ | ✅ | ❌ | ✅ | ❌ |
| GDPR Support | ✅ | ✅ | ✅ | ✅ | 🚧 |
| CCPA Support | ✅ | ✅ | ✅ | ✅ | 🚧 |
| Server-Side | ✅ | ✅ | ❌ | ❌ | ❌ |
| Client-Side | ❌ | ✅ | ✅ | ✅ | ✅ |

**Legend:** ✅ Supported | ❌ Not Supported | 🚧 In Development

## Getting Started

### Step 1: Choose Your Integration

Review the [Integration Decision Guide](#integration-decision-guide) above.

### Step 2: Get Credentials

Contact your TNE Catalyst account manager to receive:
- Publisher ID
- API Key
- Access to admin dashboard

### Step 3: Follow Integration Guide

Each integration method has its own directory with:
- `README.md` - Onboarding guide and quick start
- `SETUP.md` - Detailed setup instructions
- `WORK_REQUIRED.md` - Current status and remaining work

### Step 4: Test Integration

Use the test endpoints and credentials provided in each guide.

### Step 5: Go Live

Follow the go-live checklist in your integration guide.

## Common Requirements

All integration methods require:

1. **Publisher Account**
   - Contact sales@tne-catalyst.com for onboarding
   - Receive Publisher ID and API credentials

2. **Privacy Compliance**
   - Implement GDPR consent management (EU traffic)
   - Implement CCPA opt-out (US traffic)
   - Handle COPPA compliance (children's content)

3. **Technical Requirements**
   - HTTPS endpoints (TLS 1.2+)
   - Valid domain ownership
   - Ability to set cookies (for user sync)

4. **Content Compliance**
   - Adherence to content policies
   - Brand safety requirements
   - Ad quality standards

## Technical Support

- **Documentation**: This directory
- **API Reference**: `/tnevideo/API-REFERENCE.md`
- **Video Integration**: `/tnevideo/docs/video/VIDEO_E2E_COMPLETE.md`
- **Email**: support@tne-catalyst.com
- **Status Page**: https://status.tne-catalyst.com

## Additional Resources

- [Publisher Configuration Guide](../../PUBLISHER-CONFIG-GUIDE.md)
- [Geo & Consent Guide](../../GEO-CONSENT-GUIDE.md)
- [Bidder Parameters Guide](../../BIDDER-PARAMS-GUIDE.md)
- [Security Documentation](../security/)

## Changelog

- **2026-02-02**: Created integrations directory structure
- **2026-01-26**: Completed video VAST integration (150+ tests)
- **2026-01-26**: Applied critical security fixes

---

**Need Help?** Start with the integration method that matches your use case, or contact your account manager for personalized guidance.
