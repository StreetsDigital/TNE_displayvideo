# OpenRTB Direct Integration

**Status:** ✅ Production Ready
**Timeline:** Immediate
**Difficulty:** Medium
**Best For:** DSPs, SSPs, exchanges, server-side integrations

## Overview

Direct server-to-server integration using the OpenRTB 2.5 protocol. This is the most flexible and powerful integration method, suitable for high-volume programmatic advertising.

## Quick Start (5 minutes)

### 1. Get Credentials

Contact your account manager for:
- Publisher ID: `pub-xxx`
- API Key: `your-api-key`
- Test endpoint: `https://test.tne-catalyst.com`

### 2. Send Test Request

```bash
curl -X POST https://test.tne-catalyst.com/openrtb2/auction \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "id": "test-request-1",
    "imp": [{
      "id": "1",
      "banner": {
        "w": 300,
        "h": 250
      },
      "bidfloor": 1.0
    }],
    "site": {
      "id": "pub-xxx",
      "domain": "example.com",
      "page": "https://example.com/page"
    },
    "device": {
      "ua": "Mozilla/5.0...",
      "ip": "123.45.67.89"
    }
  }'
```

### 3. Receive Response

```json
{
  "id": "test-request-1",
  "seatbid": [{
    "bid": [{
      "id": "bid-123",
      "impid": "1",
      "price": 2.50,
      "adm": "<html>...</html>",
      "crid": "creative-456",
      "w": 300,
      "h": 250
    }]
  }],
  "cur": "USD"
}
```

## Features

✅ **Full OpenRTB 2.5 Support**
- Banner, video, native, audio ad formats
- All standard request/response fields
- Custom extensions via `ext` fields

✅ **Privacy Compliance**
- GDPR (TCF v2 consent strings)
- CCPA (US Privacy strings)
- COPPA (children's content)

✅ **Advanced Targeting**
- Geo targeting (country, region, city, DMA)
- Device targeting (type, OS, browser)
- User targeting (interests, demographics)
- Content targeting (keywords, categories)

✅ **Quality & Safety**
- IVT (Invalid Traffic) detection
- Brand safety filters
- Viewability optimization
- Ad quality checks

✅ **Performance**
- Sub-100ms response times
- 10,000+ QPS capacity
- Global edge locations
- Automatic failover

## Use Cases

### 1. SSP/Exchange Integration

Perfect for supply-side platforms to access TNE demand:

```javascript
// Your SSP code
const openrtbRequest = buildOpenRTBRequest(adUnit);
const response = await fetch('https://tne-catalyst.com/openrtb2/auction', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': process.env.TNE_API_KEY
  },
  body: JSON.stringify(openrtbRequest)
});
```

### 2. DSP Integration

Access TNE supply for demand-side platforms:

```python
# Your DSP code
import requests

def get_tne_supply(bid_request):
    response = requests.post(
        'https://tne-catalyst.com/openrtb2/auction',
        headers={'X-API-Key': os.getenv('TNE_API_KEY')},
        json=bid_request
    )
    return response.json()
```

### 3. Custom Exchange

Build your own ad exchange using TNE as a demand/supply source:

```go
// Your exchange code
func callTNE(req *openrtb2.BidRequest) (*openrtb2.BidResponse, error) {
    client := &http.Client{Timeout: 100 * time.Millisecond}
    // ... make request
}
```

## Next Steps

1. **[Complete Setup Guide](./SETUP.md)** - Detailed integration instructions
2. **[API Reference](../../../API-REFERENCE.md)** - Full API documentation
3. **Test Integration** - Use test credentials in sandbox
4. **Go Live** - Request production credentials

## Example Scenarios

- [Banner Ad Request](./examples/banner.json)
- [Video Ad Request](./examples/video.json)
- [Native Ad Request](./examples/native.json)
- [Multi-Imp Request](./examples/multi-imp.json)
- [GDPR Compliance](./examples/gdpr.json)
- [CCPA Compliance](./examples/ccpa.json)

## Performance SLAs

| Metric | Target | Measured |
|--------|--------|----------|
| Response Time (P95) | < 100ms | 85ms |
| Response Time (P99) | < 200ms | 150ms |
| Availability | 99.9% | 99.95% |
| Bid Rate | > 80% | 85% |

## Support

- **Documentation**: [SETUP.md](./SETUP.md)
- **API Reference**: [API-REFERENCE.md](../../../API-REFERENCE.md)
- **Email**: integration-support@tne-catalyst.com
- **Slack**: #tne-integrations

---

**Ready to integrate?** → [Start Setup Guide](./SETUP.md)
