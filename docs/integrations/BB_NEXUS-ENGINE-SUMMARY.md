# The Nexus Engine Integration - Quick Summary

## What They Need to Build

### 1. Client-Side JavaScript SDK

**Requirements:**
- Lightweight SDK (< 50KB gzipped)
- Hosted on fast CDN (< 500ms load time)
- Exposes `window.catalyst` global object
- Two main methods: `init()` and `requestBids()`

**Example:**
```javascript
window.catalyst = {
  init: function(config) { ... },
  requestBids: function(config, callback) { ... }
};
```

### 2. Server-Side Bid API

**Requirements:**
- HTTPS POST endpoint (e.g., `https://api.nexusengine.com/v1/bid`)
- Responds in < 2500ms (P95 latency)
- Accepts JSON bid requests
- Returns JSON bid responses
- Handles CORS properly

**Example:**
```json
POST /v1/bid
{
  "accountId": "mai-publisher-12345",
  "slots": [...]
}

Response:
{
  "bids": [
    {
      "divId": "mai-ad-leaderboard",
      "cpm": 2.50,
      "width": 728,
      "height": 90
    }
  ]
}
```

## Key Integration Points

### 1. Initialization (500ms)
```
MAI loads SDK → SDK creates window.catalyst → MAI calls init()
```

### 2. Bid Request (2800ms max)
```
MAI calls requestBids() → SDK POSTs to API → API returns bids → SDK calls callback
```

### 3. Coordination
```
Catalyst signals ready → MAI waits for all bidders → MAI sets GPT targeting
```

## Critical Requirements

### Performance SLA
- ✅ SDK loads in < 500ms
- ✅ API responds in < 2500ms (P95)
- ✅ 99.9% uptime
- ✅ < 5% timeout rate
- ✅ < 1% error rate

### Protocol Requirements
- ✅ Always call `biddersReady('catalyst')` when done
- ✅ Never block page load
- ✅ Handle errors gracefully (return empty bids)
- ✅ Respect GDPR consent
- ✅ Support CORS

### Data Format
- ✅ Request includes: slots, accountId, page context, user consent
- ✅ Response includes: bids with divId, cpm, width, height
- ✅ All prices in USD
- ✅ Currency code must be "USD"

## What MAI Publisher Does

### MAI's Responsibilities
1. ✅ Load Catalyst SDK via `<script>` tag
2. ✅ Call `catalyst.init(config)` with account ID
3. ✅ Call `catalyst.requestBids()` with slot data
4. ✅ Wait for callback or 2800ms timeout
5. ✅ Merge Catalyst bids with Prebid/Amazon
6. ✅ Set GPT targeting
7. ✅ Render ads

### MAI Provides to Catalyst
- Account ID
- Ad slot details (divId, sizes, position)
- Page context (URL, keywords, categories)
- User consent status
- Device information

## Example Flow

```
1. Page loads → MAI initializes
2. MAI loads: Prebid, Amazon, Catalyst (parallel)
3. All ready → MAI calls requestBids() on all three
4. Catalyst SDK → POST to Nexus Engine API
5. Nexus Engine → runs auction → returns bids
6. Catalyst SDK → calls callback(bids)
7. MAI merges bids → finds highest CPM
8. MAI sets GPT targeting → calls refresh()
9. Ads render
```

## Testing Approach

We built a comprehensive test suite with:
- ✅ 77 automated tests (Playwright)
- ✅ Tests all integration points
- ✅ Validates timing, coordination, errors
- ✅ Mock SDK provided for testing

## Documents Available

1. **NEXUS-ENGINE-INTEGRATION-SPEC.md** (81 pages)
   - Complete technical specification
   - API contracts
   - Request/response formats
   - Performance requirements
   - Testing requirements

2. **TEST-PLAN-SUMMARY.md**
   - Test infrastructure overview
   - How to run tests
   - Expected results

3. **TESTING.md**
   - Quick start guide
   - Test execution commands

## Next Steps for Nexus Engine

### Phase 1: SDK Development
1. Build JavaScript SDK with `init()` and `requestBids()`
2. Host on CDN
3. Implement timeout handling
4. Add error handling
5. Test with mock API

### Phase 2: API Development
1. Build bid request endpoint
2. Implement auction logic
3. Add timeout handling (2500ms internal)
4. Set up CORS
5. Test with mock requests

### Phase 3: Integration Testing
1. Provide test account credentials
2. Run our Playwright test suite
3. Fix any issues found
4. Optimize performance

### Phase 4: Production Launch
1. Monitor performance metrics
2. Track timeout rates
3. Optimize based on data
4. Scale as needed

## Contact

For questions about this integration:
- Review: `NEXUS-ENGINE-INTEGRATION-SPEC.md`
- Test with: `./run-all-tests.sh`
- Ask: Integration team

---

**Status:** Specification complete, test suite ready, waiting for Nexus Engine implementation.
