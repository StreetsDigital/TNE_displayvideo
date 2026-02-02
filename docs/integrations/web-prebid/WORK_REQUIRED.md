# Web via Prebid Integration - Work Required

**Current Status:** ⚠️ **NEEDS CLIENT DOCUMENTATION**

**Backend:** ✅ 100% Complete
**Frontend:** ❌ 0% Complete (needs examples and docs)

## What Exists (Backend)

| Component | Status | Details |
|-----------|--------|---------|
| OpenRTB Endpoint | ✅ Complete | `/openrtb2/auction` |
| Banner Support | ✅ Complete | All standard sizes |
| Native Support | ✅ Complete | IAB native spec |
| Privacy (GDPR) | ✅ Complete | TCF v2 consent |
| Privacy (CCPA) | ✅ Complete | US Privacy string |
| Cookie Sync | ✅ Complete | `/cookie_sync`, `/setuid` |
| Authentication | ✅ Complete | API key based |
| Rate Limiting | ✅ Complete | Per-publisher |

**Backend Files:**
- ✅ `internal/endpoints/auction.go` - Auction handler
- ✅ `internal/endpoints/cookie_sync.go` - User sync
- ✅ `internal/openrtb/request.go` - Request models
- ✅ `internal/openrtb/response.go` - Response models
- ✅ `internal/middleware/auth.go` - Authentication
- ✅ `internal/middleware/privacy.go` - Privacy compliance

## What's Missing (Client-Side)

### 1. Prebid.js Bidder Adapter Configuration

**Priority:** HIGH
**Effort:** 2-3 days
**Owner:** Engineering Team

**Task:** Create or document Prebid.js bidder adapter for TNE Catalyst

**Two Options:**

#### Option A: Use Existing OpenRTB Generic Adapter
Prebid.js has a generic OpenRTB adapter. We just need to document configuration:

**File to Create:**
- `docs/integrations/web-prebid/SETUP.md`

**Content Needed:**
```javascript
// Example configuration
pbjs.bidderSettings = {
  'generic': {
    endpoint: 'https://api.tne-catalyst.com/openrtb2/auction',
    syncUrl: 'https://api.tne-catalyst.com/cookie_sync'
  }
};

var adUnits = [{
  bids: [{
    bidder: 'generic',
    params: {
      endpoint: 'https://api.tne-catalyst.com/openrtb2/auction',
      publisherId: 'pub-123456',
      placementId: 'homepage-banner'
    }
  }]
}];
```

#### Option B: Create Custom TNE Catalyst Adapter (Better)
Create dedicated bidder adapter in Prebid.js codebase.

**Files to Create:**
- `prebid.js/modules/tneCatalystBidAdapter.js` - Bidder adapter
- `prebid.js/modules/tneCatalystBidAdapter.md` - Documentation
- `prebid.js/test/spec/modules/tneCatalystBidAdapter_spec.js` - Tests

**Process:**
1. Fork Prebid.js repo
2. Create adapter following Prebid guidelines
3. Write tests (required)
4. Submit PR to Prebid.js
5. Wait for review and merge (~2-4 weeks)

**Recommended:** Start with Option A (generic adapter) for immediate use, then pursue Option B for better DX.

### 2. Publisher Integration Guide

**Priority:** HIGH
**Effort:** 2 days
**Owner:** Documentation Team

**File to Create:**
- `docs/integrations/web-prebid/SETUP.md`

**Required Sections:**
1. Prerequisites (Prebid.js setup)
2. Add TNE Catalyst as bidder
3. Configure ad units
4. Configure user sync
5. Set floor prices
6. Privacy compliance (GDPR/CCPA)
7. Testing steps
8. Go-live checklist
9. Troubleshooting

**Template Outline:**
```markdown
# Web Prebid Integration - Setup Guide

## Prerequisites
- Prebid.js 7.0+ installed
- Publisher ID from TNE Catalyst
- Basic Prebid.js knowledge

## Step 1: Add TNE Catalyst Bidder

## Step 2: Configure Ad Units

## Step 3: Enable Cookie Sync

## Step 4: Privacy Compliance

## Step 5: Testing

## Step 6: Go Live
```

### 3. Code Examples

**Priority:** HIGH
**Effort:** 2 days
**Owner:** Engineering Team

**Files to Create:**

#### Basic Examples
- `docs/integrations/web-prebid/examples/basic-banner.html`
  - Single 300x250 banner
  - Simple Prebid setup
  - TNE Catalyst only

- `docs/integrations/web-prebid/examples/multi-size-banner.html`
  - Responsive ad unit
  - Multiple sizes [[300,250], [728,90]]
  - TNE + other bidders

- `docs/integrations/web-prebid/examples/native-ad.html`
  - Native ad unit
  - Image + title + description
  - Styling examples

#### Advanced Examples
- `docs/integrations/web-prebid/examples/multi-format.html`
  - Banner + Native on same page
  - Multiple ad units
  - Different floor prices

- `docs/integrations/web-prebid/examples/video-banner.html`
  - Both video and banner
  - Different configurations per format

- `docs/integrations/web-prebid/examples/lazy-loading.html`
  - Lazy load ad units
  - Viewport detection
  - Performance optimization

- `docs/integrations/web-prebid/examples/refresh.html`
  - Auto-refresh ads
  - User activity detection
  - Frequency capping

#### Privacy Examples
- `docs/integrations/web-prebid/examples/gdpr-consent.html`
  - TCF v2 integration
  - CMP (Consent Management Platform)
  - Consent string passing

- `docs/integrations/web-prebid/examples/ccpa-optout.html`
  - US Privacy string
  - Opt-out link
  - State detection

### 4. Test Credentials

**Priority:** MEDIUM
**Effort:** 1 day
**Owner:** Operations Team

**Task:** Create test publisher accounts

**Files to Update:**
- `docs/integrations/web-prebid/SETUP.md` (add test credentials section)

**Test Credentials Needed:**
```
Test Publisher ID: pub-test-prebid-001
Test API Key: tne_test_prebid_xxx
Test Endpoint: https://test.tne-catalyst.com/openrtb2/auction
```

**Test Ad Units:**
- 300x250 banner (always returns bid)
- 728x90 banner (always returns bid)
- Native ad (always returns bid)
- Test creative URLs

### 5. Prebid Adapter Parameters Documentation

**Priority:** HIGH
**Effort:** 1 day
**Owner:** Documentation Team

**File to Create:**
- `docs/integrations/web-prebid/PARAMETERS.md`

**Content:**
Document all supported bidder parameters:

```javascript
{
  bidder: 'tne-catalyst',
  params: {
    publisherId: 'pub-123456',    // Required
    placementId: 'homepage-top',   // Required
    bidfloor: 1.50,                // Optional, CPM floor
    currency: 'USD',               // Optional, default USD
    keywords: ['sports', 'news'],  // Optional, targeting
    firstPartyData: {              // Optional, FPD
      category: 'sports',
      section: 'nfl'
    }
  }
}
```

### 6. Troubleshooting Guide

**Priority:** MEDIUM
**Effort:** 1 day
**Owner:** Support Team

**File to Create:**
- `docs/integrations/web-prebid/TROUBLESHOOTING.md`

**Common Issues to Document:**

1. **No bids returned**
   - Check publisher ID
   - Verify floor prices
   - Check ad unit configuration
   - Test endpoint directly

2. **Slow response times**
   - Check timeout configuration
   - Verify network connectivity
   - Check Prebid debug logs

3. **Cookie sync not working**
   - Check CORS settings
   - Verify cookie sync URL
   - Check browser cookie settings

4. **Privacy consent issues**
   - Verify TCF string format
   - Check consent purposes
   - Test with/without consent

5. **Ads not rendering**
   - Check creative rendering
   - Verify Google Ad Manager setup
   - Check browser console errors

### 7. Performance Optimization Guide

**Priority:** LOW
**Effort:** 1 day
**Owner:** Documentation Team

**File to Create:**
- `docs/integrations/web-prebid/OPTIMIZATION.md`

**Topics:**
- Timeout configuration
- Lazy loading best practices
- Refresh strategies
- Floor price optimization
- Reducing page weight

### 8. Migration Guide

**Priority:** LOW
**Effort:** 1 day
**Owner:** Documentation Team

**File to Create:**
- `docs/integrations/web-prebid/MIGRATION.md`

**Content:**
Help publishers migrate from other SSPs to TNE Catalyst:
- Side-by-side comparison
- Feature mapping
- Configuration conversion
- Testing in parallel
- Gradual rollout

## Implementation Plan

### Phase 1: Minimum Viable Documentation (Week 1)

**Goal:** Enable first publishers to integrate

- [ ] Create SETUP.md with generic OpenRTB adapter config
- [ ] Create basic banner example (basic-banner.html)
- [ ] Create PARAMETERS.md with all params
- [ ] Create test credentials
- [ ] Test integration end-to-end

**Deliverables:**
- Publishers can integrate using generic adapter
- Basic example works
- Test credentials available

### Phase 2: Complete Examples (Week 2)

**Goal:** Cover all common use cases

- [ ] Create multi-size banner example
- [ ] Create native ad example
- [ ] Create GDPR consent example
- [ ] Create CCPA opt-out example
- [ ] Create TROUBLESHOOTING.md

**Deliverables:**
- Full example library
- Troubleshooting support
- Privacy compliance examples

### Phase 3: Custom Adapter (Optional, 4-6 weeks)

**Goal:** Better developer experience

- [ ] Fork Prebid.js
- [ ] Create tneCatalystBidAdapter.js
- [ ] Write comprehensive tests
- [ ] Submit PR to Prebid.js
- [ ] Address review comments
- [ ] Merge into Prebid.js

**Deliverables:**
- Native Prebid.js adapter
- Listed on Prebid.org
- Better publisher DX

## Testing Checklist

Before marking as complete:

- [ ] Test basic banner integration
- [ ] Test multi-size banners
- [ ] Test native ads
- [ ] Test with multiple bidders
- [ ] Test GDPR consent flow
- [ ] Test CCPA opt-out
- [ ] Test cookie sync
- [ ] Test on desktop browsers (Chrome, Firefox, Safari)
- [ ] Test on mobile browsers (iOS Safari, Android Chrome)
- [ ] Test with ad blockers (should gracefully fail)
- [ ] Verify performance (< 100ms bid response)
- [ ] Check for console errors
- [ ] Validate OpenRTB requests/responses

## Resources Needed

**Engineering:**
- 1 developer x 2-3 days (Phase 1)
- 1 developer x 2-3 days (Phase 2)
- 1 developer x 2-4 weeks (Phase 3, optional)

**Documentation:**
- 1 technical writer x 3-4 days

**Operations:**
- 1 ops engineer x 1 day (test credentials)

**QA:**
- 1 QA engineer x 2 days (testing)

**Total Effort:** 2-3 weeks (without custom adapter), 6-8 weeks (with custom adapter)

## Success Criteria

**Phase 1 Complete When:**
- [ ] Publisher can add TNE Catalyst to Prebid.js
- [ ] Basic example works end-to-end
- [ ] Documentation published
- [ ] Test credentials available

**Phase 2 Complete When:**
- [ ] All examples work
- [ ] Troubleshooting guide published
- [ ] 5+ beta publishers successfully integrated
- [ ] No major issues reported

**Phase 3 Complete When:**
- [ ] Custom adapter merged into Prebid.js
- [ ] Listed on Prebid.org
- [ ] Appears in Prebid.js release notes

## Dependencies

**Blocked By:**
- None - backend is complete

**Blocks:**
- Publisher onboarding for Prebid users
- Prebid marketplace listings
- Integration partnerships

## Next Actions

1. **Assign owner** for Phase 1 work
2. **Create GitHub project** to track tasks
3. **Schedule kickoff** meeting
4. **Set target completion date** (recommend 2 weeks)
5. **Identify beta publishers** for testing

---

**Last Updated:** 2026-02-02
**Target Completion:** 2026-02-16 (2 weeks)
**Status:** Ready to start
