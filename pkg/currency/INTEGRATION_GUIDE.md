# Currency Conversion Integration Guide

Complete guide for integrating multi-currency support into the TNE Catalyst auction exchange.

## Status: ✅ Module Created, Integration Needed

The currency conversion module is complete and ready to integrate into the exchange auction flow.

## Files Created

1. ✅ `pkg/currency/converter.go` - Currency converter implementation
2. ✅ `pkg/currency/converter_test.go` - Unit tests
3. ✅ `pkg/currency/README.md` - Documentation
4. ✅ `internal/exchange/currency.go` - Exchange helper functions

## Files Modified

1. ✅ `internal/exchange/exchange.go` - Added currency converter to Exchange struct and Config

## Integration Steps

### Step 1: Initialize Currency Converter (Server Startup)

In `cmd/server/server.go`, add currency converter initialization:

```go
import "github.com/thenexusengine/tne_springwire/pkg/currency"

// In your server initialization code:
func setupExchange() *exchange.Exchange {
    // Initialize currency converter
    currencyConverter := currency.NewConverter(currency.DefaultConfig())

    // Start background rate updates
    ctx := context.Background()
    if err := currencyConverter.Start(ctx); err != nil {
        log.Fatal("Failed to start currency converter:", err)
    }

    // Create exchange config with currency support
    exchangeConfig := &exchange.Config{
        DefaultTimeout:    1000 * time.Millisecond,
        CurrencyConverter: currencyConverter,  // ← Add this
        DefaultCurrency:   "USD",
        // ... other config
    }

    ex := exchange.New(registry, exchangeConfig)

    return ex
}
```

### Step 2: Add Currency Conversion to RunAuction

In `internal/exchange/exchange.go`, in the `RunAuction` function, add currency conversion after collecting bid responses:

**Location:** After line ~1400 (after bidder responses are collected, before building final response)

```go
// Extract target currency from request
targetCurrency := e.extractTargetCurrency(req.BidRequest)

// Extract custom rates if provided
customRates, useExternalRates := extractCustomRates(req.BidRequest)

// Convert all bidder responses to target currency
for bidderCode, result := range response.BidderResults {
    if result.BidResponse != nil {
        err := e.convertBidderResponse(
            result.BidResponse,
            bidderCode,
            targetCurrency,
            customRates,
            useExternalRates,
        )

        if err != nil {
            // Log conversion error
            logger.Warn("currency conversion failed",
                "bidder", bidderCode,
                "error", err,
            )

            // Mark bidder result as error
            result.Error = fmt.Sprintf("currency conversion failed: %v", err)
            result.BidResponse = nil // Reject bids
        }
    }
}

// Set currency in final response
if response.BidResponse != nil {
    response.BidResponse.Cur = targetCurrency
}
```

### Step 3: Update Metrics (Optional but Recommended)

Add currency conversion metrics to track conversions and failures:

```go
// In internal/metrics/prometheus.go, add new metrics:

currencyConversions = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "catalyst_currency_conversions_total",
        Help: "Total number of currency conversions performed",
    },
    []string{"from", "to", "status"}, // status: success/failure
)

// In convertBidCurrency() function:
if err == nil {
    e.metrics.RecordCurrencyConversion(bidCurrency, targetCurrency, "success")
} else {
    e.metrics.RecordCurrencyConversion(bidCurrency, targetCurrency, "failure")
}
```

### Step 4: Add Configuration (Environment Variables)

In `cmd/server/config.go`, add currency-related configuration:

```go
type Config struct {
    // ... existing fields

    // Currency conversion
    CurrencyEnabled         bool          `env:"CURRENCY_ENABLED" envDefault:"true"`
    CurrencyDefaultCurrency string        `env:"CURRENCY_DEFAULT" envDefault:"USD"`
    CurrencyFetchURL        string        `env:"CURRENCY_FETCH_URL" envDefault:"https://cdn.jsdelivr.net/gh/prebid/currency-file@1/latest.json"`
    CurrencyRefreshInterval time.Duration `env:"CURRENCY_REFRESH_INTERVAL" envDefault:"30m"`
    CurrencyStaleThreshold  time.Duration `env:"CURRENCY_STALE_THRESHOLD" envDefault:"24h"`
}
```

In `.env.production`, add:

```bash
# Currency Conversion
CURRENCY_ENABLED=true
CURRENCY_DEFAULT=USD
CURRENCY_FETCH_URL=https://cdn.jsdelivr.net/gh/prebid/currency-file@1/latest.json
CURRENCY_REFRESH_INTERVAL=30m
CURRENCY_STALE_THRESHOLD=24h
```

### Step 5: Add Health Check Endpoint

Add a currency health check endpoint to monitor rate freshness:

```go
// In cmd/server/server.go or relevant handler file:

func (s *Server) currencyHealthHandler(w http.ResponseWriter, r *http.Request) {
    if s.currencyConverter == nil {
        http.Error(w, "Currency converter not initialized", http.StatusServiceUnavailable)
        return
    }

    stats := s.currencyConverter.Stats()

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(stats)
}

// Register route:
mux.HandleFunc("/currency/stats", s.currencyHealthHandler)
```

## Testing the Integration

### 1. Unit Tests

Run existing tests:

```bash
go test ./pkg/currency/...
go test ./internal/exchange/...
```

### 2. Integration Test

Create a test auction request with multiple currencies:

```bash
curl -X POST https://ads.thenexusengine.com/openrtb2/auction \
  -H "Content-Type: application/json" \
  -d '{
    "id": "test-multi-currency",
    "cur": ["USD"],
    "imp": [{
      "id": "1",
      "banner": {"w": 300, "h": 250}
    }],
    "site": {
      "domain": "example.com"
    },
    "ext": {
      "prebid": {
        "currency": {
          "rates": {
            "USD": {"EUR": 0.92}
          }
        }
      }
    }
  }'
```

Expected: All bid responses should be in USD.

### 3. Test Custom Rates

Test that custom rates override external rates:

```json
{
  "cur": ["EUR"],
  "ext": {
    "prebid": {
      "currency": {
        "rates": {
          "USD": {"EUR": 0.95}
        },
        "usepbsrates": false
      }
    }
  }
}
```

### 4. Test Unsupported Currency

Test behavior when requesting unsupported currency:

```json
{
  "cur": ["XYZ"]  // Invalid currency code
}
```

Expected: Should fall back to USD or return error.

## Verification Checklist

After integration, verify:

- [ ] Currency converter starts successfully at server startup
- [ ] Rates are fetched from Prebid CDN within 30 seconds
- [ ] `/currency/stats` endpoint shows healthy status
- [ ] Multi-currency auction requests work correctly
- [ ] Bid prices are converted to request currency
- [ ] Custom rates override external rates when provided
- [ ] Same-currency requests don't trigger conversion
- [ ] Conversion failures are logged appropriately
- [ ] Metrics track currency conversions
- [ ] Stale rates generate warnings in logs

## Monitoring

### Key Metrics to Track

1. **Currency Conversions**: Count of successful/failed conversions by currency pair
2. **Rate Freshness**: Age of currency rates data
3. **Fetch Errors**: Consecutive failures fetching rates from CDN
4. **Conversion Errors**: Bids rejected due to currency conversion failures

### Grafana Dashboard Queries

```promql
# Conversion rate by currency pair
rate(catalyst_currency_conversions_total{status="success"}[5m])

# Conversion failure rate
rate(catalyst_currency_conversions_total{status="failure"}[5m])

# Rate staleness (in hours)
(time() - catalyst_currency_last_fetch_timestamp) / 3600
```

## Troubleshooting

### Rates Not Loading

**Symptom**: `/currency/stats` shows `ratesLoaded: false`

**Solutions**:
1. Check network connectivity to CDN
2. Verify `CURRENCY_FETCH_URL` is accessible
3. Check firewall rules for outbound HTTPS
4. Review logs for fetch errors

### High Conversion Failures

**Symptom**: Many bids rejected due to currency conversion errors

**Solutions**:
1. Check if bidders are returning unsupported currencies
2. Verify rates data includes required currency pairs
3. Consider adding custom rate overrides for missing pairs
4. Check if currency codes are valid ISO 4217

### Stale Rates

**Symptom**: Warnings about stale currency rates

**Solutions**:
1. Check if background refresh is running
2. Verify CDN is accessible and responding
3. Consider reducing `CURRENCY_STALE_THRESHOLD`
4. Check for network/proxy issues

## Performance Impact

Expected performance impact with currency conversion:

- **Memory**: +50KB for rate storage
- **CPU**: Negligible (<0.1ms per bid conversion)
- **Latency**: +0.5-1ms per auction (for conversions)
- **Network**: 1 HTTP request every 30 minutes (rate refresh)

## Rollback Plan

If issues arise, disable currency conversion:

```bash
# In .env.production:
CURRENCY_ENABLED=false
```

Or set `CurrencyConverter: nil` in exchange config.

## Future Enhancements

Potential improvements:

1. **Publisher-specific default currencies**: Allow publishers to set preferred currency
2. **Bidder currency preferences**: Track which currencies each bidder supports
3. **Currency caching**: Cache conversion results for auction duration
4. **Rate source fallback**: Use multiple rate sources for redundancy
5. **Historical rate tracking**: Store rate history for analysis
6. **Currency arbitrage detection**: Alert on suspicious rate differences

## References

- [Prebid Currency Conversion](https://docs.prebid.org/prebid-server/features/pbs-currency.html)
- [OpenRTB 2.5 Specification](https://www.iab.com/wp-content/uploads/2016/03/OpenRTB-API-Specification-Version-2-5-FINAL.pdf)
- [ISO 4217 Currency Codes](https://www.iso.org/iso-4217-currency-codes.html)
- [TNE Catalyst Currency Module README](./README.md)

---

**Integration Status**: Ready for implementation
**Estimated Integration Time**: 2-4 hours
**Risk Level**: Low (isolated feature, easy to disable)
**Testing Required**: Unit + integration tests

**Questions?** Contact ops@thenexusengine.io
