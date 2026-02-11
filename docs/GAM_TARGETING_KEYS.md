# Catalyst GAM Targeting Keys Reference

## Overview

Catalyst sets specific targeting keys in Google Ad Manager (GAM) to identify server-side bids and enable reporting. All Catalyst keys use the `_catalyst` suffix to avoid conflicts with client-side Prebid.js bidders.

## Complete Key List

### Core Bidding Keys

| Key | Description | Example Value | Always Set |
|-----|-------------|---------------|------------|
| `hb_pb_catalyst` | Catalyst bid price (CPM) | `"2.50"` | ✅ Yes |
| `hb_size_catalyst` | Ad size (WxH) | `"300x250"` | ✅ Yes |
| `hb_adid_catalyst` | Unique ad/bid ID | `"abc123"` | ✅ Yes |
| `hb_creative_catalyst` | Creative ID | `"creative-456"` | ✅ Yes |
| `hb_bidder_catalyst` | Demand partner name | `"thenexusengine"` | ✅ Yes |
| `hb_partner` | Demand partner (alias) | `"thenexusengine"` | ✅ Yes |

### Standard Prebid-Compatible Keys

| Key | Description | Example Value | Always Set |
|-----|-------------|---------------|------------|
| `hb_source_catalyst` | Bid source type | `"s2s"` (server-to-server) | ✅ Yes |
| `hb_format_catalyst` | Ad format | `"banner"` | ✅ Yes |
| `hb_deal_catalyst` | Deal ID for PMP deals | `"DEAL-12345"` | ❌ Only if deal |
| `hb_adomain_catalyst` | Advertiser domain | `"example.com"` | ❌ When available |

## Key Purposes

### Revenue Reporting
- **`hb_pb_catalyst`** - Use for CPM-based reporting and revenue calculations
- **`hb_deal_catalyst`** - Track PMP deal performance separately

### Inventory Management
- **`hb_size_catalyst`** - Filter by ad size (e.g., only 300x250)
- **`hb_format_catalyst`** - Currently always "banner"

### Demand Partner Analysis
- **`hb_bidder_catalyst`** or **`hb_partner`** - Identify which SSP/exchange filled
- **`hb_adomain_catalyst`** - Block or allow specific advertisers

### Technical Tracking
- **`hb_source_catalyst`** - Always "s2s" (server-side)
- **`hb_adid_catalyst`** - Unique identifier for debugging

## GAM Setup Examples

### Line Item Targeting

**Target Catalyst bids over $2.00:**
```
hb_pb_catalyst >= 2.00
AND hb_bidder_catalyst = thenexusengine
```

**Target specific deal IDs:**
```
hb_deal_catalyst = DEAL-12345
```

**Target specific ad sizes:**
```
hb_size_catalyst = 300x250
OR hb_size_catalyst = 728x90
```

### Price Granularity

Catalyst sends exact CPM values. You can create line items with:
- **Dense:** $0.01 increments (e.g., 0.01, 0.02, 0.03...)
- **Standard:** $0.10 increments (e.g., 0.10, 0.20, 0.30...)
- **Custom:** Match your Prebid.js granularity

**Example Line Item Setup:**
```
Priority: 12 (Header Bidding)
Type: Price Priority
Rate: $2.50
Targeting: hb_pb_catalyst = 2.50
```

### Reporting Dimensions

Use these keys in GAM reports:

**Revenue by Demand Partner:**
- Dimension: `hb_bidder_catalyst` or `hb_partner`
- Metric: Revenue, Impressions, eCPM

**Revenue by Deal:**
- Dimension: `hb_deal_catalyst`
- Metric: Revenue, Fill Rate

**Revenue by Size:**
- Dimension: `hb_size_catalyst`
- Metric: Revenue, Impressions

## Comparison: Catalyst vs Prebid Keys

### No Overlap (No Conflicts!)

| Prebid Sets | Catalyst Sets | Purpose |
|-------------|---------------|---------|
| `hb_bidder` | `hb_bidder_catalyst` | Bidder identification |
| `hb_pb` | `hb_pb_catalyst` | Bid price |
| `hb_size` | `hb_size_catalyst` | Ad size |
| `hb_adid` | `hb_adid_catalyst` | Ad ID |
| `hb_source` | `hb_source_catalyst` | Bid source |
| `hb_format` | `hb_format_catalyst` | Ad format |
| `hb_deal` | `hb_deal_catalyst` | Deal ID |
| `hb_adomain` | `hb_adomain_catalyst` | Advertiser domain |

**Key Insight:** Prebid and Catalyst use separate namespaces, so both can set targeting simultaneously without conflicts.

## Integration Pattern

When using Catalyst + Prebid.js together:

1. **Both fetch bids in parallel** (no slowdown)
2. **Both set targeting keys** (separate namespaces)
3. **GAM receives all keys** (Prebid keys + Catalyst keys)
4. **Line items can target either or both**

**Example: Combined Targeting**
```
(hb_pb >= 1.00 AND hb_bidder = rubicon)  // Client-side Prebid
OR
(hb_pb_catalyst >= 1.00 AND hb_bidder_catalyst = thenexusengine)  // Server-side Catalyst
```

## Troubleshooting

### Keys Not Appearing in GAM

**Check browser console:**
```javascript
googletag.pubads().getSlots().forEach(slot => {
    console.log('Slot:', slot.getAdUnitPath());
    slot.getTargetingKeys().forEach(key => {
        if (key.includes('catalyst')) {
            console.log(`  ${key}:`, slot.getTargeting(key));
        }
    });
});
```

**Expected output:**
```
hb_pb_catalyst: ["2.50"]
hb_size_catalyst: ["300x250"]
hb_bidder_catalyst: ["thenexusengine"]
hb_source_catalyst: ["s2s"]
hb_format_catalyst: ["banner"]
hb_partner: ["thenexusengine"]
```

### Empty Bracket Values

If you see `hb_pb_catalyst: []`, it means:
- Catalyst didn't return a bid for that slot
- OR the bid timed out
- OR Catalyst SDK wasn't initialized

### Keys Overwritten

If Catalyst keys disappear after Prebid runs:
- ✅ **Fixed in v1.0.0+** - Catalyst now uses unique `_catalyst` suffix
- Old versions set `hb_bidder` directly (conflicted with Prebid)

## Version History

### v1.0.0 (Current)
- ✅ Added `_catalyst` suffix to all keys
- ✅ Added `hb_source_catalyst` (s2s)
- ✅ Added `hb_format_catalyst` (banner)
- ✅ Added `hb_deal_catalyst` (PMP support)
- ✅ Added `hb_adomain_catalyst` (advertiser domain)
- ✅ Removed conflicting keys (`hb_bidder`, `hb_pb`, `hb_size`)

### v0.x (Legacy)
- ❌ Set `hb_bidder`, `hb_pb`, `hb_size` directly
- ❌ Conflicted with Prebid.js
- ❌ Required sequential timing coordination

## Support

For questions about GAM setup or targeting keys:
- Check Catalyst SDK logs: `window.catalyst._config.debug = true`
- Review bid responses: `[Catalyst] Set slot targeting for...`
- Contact: support@thenexusengine.com
