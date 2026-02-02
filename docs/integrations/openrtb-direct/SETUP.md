# OpenRTB Direct Integration - Setup Guide

## Prerequisites

- [ ] Publisher account and credentials
- [ ] Server with HTTPS support
- [ ] Ability to make HTTP POST requests
- [ ] Understanding of OpenRTB 2.5 protocol

## Step 1: Account Setup

### 1.1 Create Publisher Account

Contact your TNE Catalyst account manager:
- Email: sales@tne-catalyst.com
- Provide: Company name, domain(s), expected volume

### 1.2 Receive Credentials

You'll receive:
```
Publisher ID: pub-123456789
API Key: tne_live_abcdef123456
Test API Key: tne_test_abcdef123456
```

### 1.3 Access Admin Dashboard

Login at: https://admin.tne-catalyst.com
- Configure bidder preferences
- Set floor prices
- View analytics

## Step 2: Technical Integration

### 2.1 Configure Endpoint

**Production:**
```
POST https://api.tne-catalyst.com/openrtb2/auction
```

**Test/Sandbox:**
```
POST https://test.tne-catalyst.com/openrtb2/auction
```

### 2.2 Authentication

All requests must include:

```http
POST /openrtb2/auction HTTP/1.1
Host: api.tne-catalyst.com
Content-Type: application/json
X-API-Key: tne_live_abcdef123456
X-Publisher-ID: pub-123456789
```

### 2.3 Minimum Viable Request

```json
{
  "id": "unique-auction-id",
  "imp": [{
    "id": "1",
    "banner": {
      "w": 300,
      "h": 250
    }
  }],
  "site": {
    "id": "pub-123456789",
    "domain": "yoursite.com"
  }
}
```

### 2.4 Handle Response

**Success (200 OK):**
```json
{
  "id": "unique-auction-id",
  "seatbid": [{
    "bid": [{
      "id": "bid-123",
      "impid": "1",
      "price": 2.50,
      "adm": "<html>ad markup</html>"
    }]
  }]
}
```

**No Bid (204 No Content):**
```
Empty body - no ads available
```

**Error (400/500):**
```json
{
  "error": "Invalid request",
  "details": ["Field 'imp' is required"]
}
```

## Step 3: Privacy Compliance

### 3.1 GDPR Implementation

For EU traffic, include TCF consent string:

```json
{
  "id": "auction-id",
  "user": {
    "ext": {
      "consent": "CPXxRfAPXxRfAAfKABENDXCgAAAAAAAAAAAAAAAAAAAA"
    }
  },
  "regs": {
    "ext": {
      "gdpr": 1
    }
  }
}
```

### 3.2 CCPA Implementation

For US traffic, include privacy string:

```json
{
  "id": "auction-id",
  "regs": {
    "ext": {
      "us_privacy": "1YNN"
    }
  }
}
```

### 3.3 COPPA Implementation

For children's content:

```json
{
  "id": "auction-id",
  "regs": {
    "coppa": 1
  }
}
```

## Step 4: Advanced Configuration

### 4.1 Multiple Ad Formats

Request multiple impressions:

```json
{
  "id": "auction-id",
  "imp": [
    {
      "id": "1",
      "banner": {"w": 300, "h": 250}
    },
    {
      "id": "2",
      "video": {
        "w": 1920,
        "h": 1080,
        "minduration": 5,
        "maxduration": 30
      }
    }
  ]
}
```

### 4.2 Floor Prices

Set minimum CPM:

```json
{
  "imp": [{
    "id": "1",
    "banner": {"w": 300, "h": 250},
    "bidfloor": 1.50,
    "bidfloorcur": "USD"
  }]
}
```

### 4.3 First Party Data

Pass custom targeting:

```json
{
  "site": {
    "ext": {
      "data": {
        "category": "sports",
        "keywords": ["football", "nfl"]
      }
    }
  }
}
```

### 4.4 User ID Sync

Enable cookie sync for better targeting:

```javascript
// Include in your page
fetch('https://api.tne-catalyst.com/cookie_sync')
  .then(r => r.json())
  .then(data => {
    data.sync_urls.forEach(url => {
      // Fire sync pixels
      new Image().src = url;
    });
  });
```

## Step 5: Testing

### 5.1 Test with cURL

```bash
curl -X POST https://test.tne-catalyst.com/openrtb2/auction \
  -H "Content-Type: application/json" \
  -H "X-API-Key: tne_test_abcdef123456" \
  -d @test-request.json
```

### 5.2 Validate Responses

Check for:
- [ ] Response time < 100ms
- [ ] Valid bid prices
- [ ] Correct ad markup
- [ ] Proper currency
- [ ] Privacy compliance

### 5.3 Test Edge Cases

- No bid scenarios
- Invalid requests
- Timeout handling
- GDPR consent variations
- Different ad formats

## Step 6: Monitoring

### 6.1 Health Check

```bash
curl https://api.tne-catalyst.com/health
```

Response:
```json
{
  "status": "ok",
  "timestamp": "2026-02-02T12:00:00Z"
}
```

### 6.2 Metrics Endpoint

Track performance (requires auth):
```bash
curl -H "X-API-Key: your-key" \
  https://api.tne-catalyst.com/metrics
```

### 6.3 Set Up Alerts

Monitor:
- Response time degradation
- Bid rate drops
- Error rate increases
- Timeout spikes

## Step 7: Go Live

### 7.1 Pre-Launch Checklist

- [ ] Successfully tested in sandbox
- [ ] Privacy compliance verified
- [ ] Floor prices configured
- [ ] Monitoring setup
- [ ] Error handling implemented
- [ ] Timeout configured (100-200ms)

### 7.2 Switch to Production

1. Update endpoint URL
2. Update API key to production
3. Start with 10% traffic
4. Monitor for 24 hours
5. Gradually increase to 100%

### 7.3 Post-Launch

- Monitor dashboard metrics
- Review first day performance
- Optimize floor prices
- Enable additional bidders

## Troubleshooting

### Common Issues

**Issue: 401 Unauthorized**
```
Solution: Check API key and Publisher ID headers
```

**Issue: 400 Bad Request**
```
Solution: Validate OpenRTB request format
```

**Issue: 204 No Bids**
```
Solution: Lower floor prices, check targeting
```

**Issue: Timeouts**
```
Solution: Set client timeout to 200ms
```

## Code Examples

- [Node.js Integration](./examples/nodejs-integration.js)
- [Python Integration](./examples/python-integration.py)
- [Go Integration](./examples/go-integration.go)
- [Java Integration](./examples/java-integration.java)

## Support Resources

- **API Reference**: [API-REFERENCE.md](../../../API-REFERENCE.md)
- **Troubleshooting**: [TROUBLESHOOTING.md](./TROUBLESHOOTING.md)
- **Best Practices**: [BEST-PRACTICES.md](./BEST-PRACTICES.md)
- **Email**: integration-support@tne-catalyst.com

---

**Next:** Review [Best Practices](./BEST-PRACTICES.md) for optimization tips
