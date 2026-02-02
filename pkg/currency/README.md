# Currency Conversion Module

Multi-currency support for OpenRTB auctions using Prebid's currency-file exchange rates.

## Features

- ✅ Automatic rate updates from Prebid currency CDN
- ✅ Support for 32+ currencies (USD, EUR, GBP, JPY, etc.)
- ✅ Custom rate overrides per request
- ✅ Stale rate detection
- ✅ Thread-safe operation
- ✅ Background refresh with configurable intervals

## Quick Start

### 1. Initialize the Converter

```go
import "github.com/thenexusengine/tne_springwire/pkg/currency"

// Use default configuration
converter := currency.NewConverter(currency.DefaultConfig())

// Start background updates
ctx := context.Background()
if err := converter.Start(ctx); err != nil {
    log.Fatal(err)
}
defer converter.Stop()
```

### 2. Convert Bid Prices

```go
// Convert EUR bid to USD
usdPrice, err := converter.Convert(1.00, "EUR", "USD")
if err != nil {
    // Handle conversion error
}
// usdPrice ≈ 1.18
```

### 3. Use with Custom Rates (from OpenRTB request)

```go
// Extract custom rates from request.ext.prebid.currency
customRates := map[string]map[string]float64{
    "USD": {"EUR": 0.92}, // Publisher override
}

// Create aggregate conversions
agg := currency.NewAggregateConversions(
    customRates,     // Custom rates (priority)
    converter,       // External rates (fallback)
    false,           // Use custom rates first
)

// Convert with priority logic
rate, err := agg.GetRate("USD", "EUR")
// Uses custom rate: 0.92
```

## Configuration

### Default Configuration

```go
config := currency.DefaultConfig()
// FetchURL: https://cdn.jsdelivr.net/gh/prebid/currency-file@1/latest.json
// RefreshInterval: 30 minutes
// StaleThreshold: 24 hours
// FetchTimeout: 10 seconds
```

### Custom Configuration

```go
config := &currency.Config{
    FetchURL:        "https://custom-cdn.com/rates.json",
    RefreshInterval: 15 * time.Minute,  // Refresh every 15 min
    StaleThreshold:  12 * time.Hour,    // Stale after 12 hours
    FetchTimeout:    5 * time.Second,   // HTTP timeout
}

converter := currency.NewConverter(config)
```

## Supported Currencies

The Prebid currency file supports **32 currencies**:

**Major:**
- USD, EUR, GBP, JPY, CHF, AUD, CAD, NZD

**European:**
- CZK, DKK, HUF, PLN, RON, SEK, NOK, ISK

**Asian:**
- CNY, HKD, IDR, INR, KRW, MYR, PHP, SGD, THB

**Others:**
- TRY, BRL, MXN, ILS, ZAR

All currencies use **ISO 4217 three-letter codes**.

## OpenRTB Integration

### Request Currency

```json
{
  "cur": ["USD"],  // Preferred currency
  "ext": {
    "prebid": {
      "currency": {
        "rates": {
          "USD": { "EUR": 0.92 }  // Custom override
        },
        "usepbsrates": false  // false = prefer custom rates
      }
    }
  }
}
```

### Response Currency

```json
{
  "cur": "USD",  // All bid prices in USD
  "seatbid": [...]
}
```

## Error Handling

### Conversion Not Available

```go
rate, err := converter.GetRate("USD", "XYZ")
if err != nil {
    // Error: "no conversion available from USD to XYZ"
    // Reject bid or use fallback logic
}
```

### Stale Rates

```go
stats := converter.Stats()
if stats["stale"] == true {
    // Rates are older than StaleThreshold
    // Log warning, continue using stale rates
}
```

### Network Failures

The converter gracefully handles network failures:
- Continues using last successfully fetched rates
- Logs warnings for fetch errors
- Tracks consecutive failures in stats

## Monitoring

### Check Stats

```go
stats := converter.Stats()
fmt.Printf("Stats: %+v\n", stats)
// Output:
// {
//   "running": true,
//   "fetchErrors": 0,
//   "ratesLoaded": true,
//   "currencies": 32,
//   "dataAsOf": "2026-02-02T00:00:00.000Z",
//   "lastFetch": "2026-02-02T12:30:00Z",
//   "age": "15m0s",
//   "stale": false
// }
```

### Metrics to Track

- **fetchErrors** - Consecutive fetch failures
- **age** - Time since last successful fetch
- **stale** - Whether rates exceed StaleThreshold
- **currencies** - Number of base currencies loaded

## Best Practices

### 1. Initialize at Startup

```go
// In main() or init()
converter := currency.NewConverter(currency.DefaultConfig())
if err := converter.Start(context.Background()); err != nil {
    log.Fatal("Failed to start currency converter:", err)
}
```

### 2. Pass to Exchange

```go
exchange := exchange.New(&exchange.Config{
    CurrencyConverter: converter,
    // ... other config
})
```

### 3. Handle Same-Currency Pass-Through

```go
// No conversion needed if same currency
if bidCurrency == requestCurrency {
    // Use bid price as-is
} else {
    // Convert to request currency
    convertedPrice, err := converter.Convert(
        bidPrice,
        bidCurrency,
        requestCurrency,
    )
}
```

### 4. Reject Bids on Conversion Failure

```go
convertedPrice, err := converter.Convert(bidPrice, bidCur, reqCur)
if err != nil {
    // Cannot convert - reject bid
    return nil, fmt.Errorf("currency conversion failed: %w", err)
}
```

### 5. Log Conversion Stats

```go
// Log successful conversions for analysis
logger.Info("converted bid",
    "from", bidCurrency,
    "to", requestCurrency,
    "originalPrice", bidPrice,
    "convertedPrice", convertedPrice,
    "rate", rate,
)
```

## Testing

### Unit Tests

```bash
go test ./pkg/currency/...
```

### Integration Testing

```go
// Test with mock HTTP server
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    json.NewEncoder(w).Encode(currency.CurrencyFile{
        Conversions: map[string]map[string]float64{
            "USD": {"EUR": 0.85},
        },
    })
}))

config := &currency.Config{FetchURL: server.URL}
converter := currency.NewConverter(config)
```

## Performance

- **Memory**: ~50KB for 32 currencies with 1000 conversion pairs
- **Lookup**: O(1) constant time for rate lookups
- **Thread-safe**: RWMutex allows concurrent reads

## Data Source

**Prebid Currency File:**
- Primary CDN: `https://cdn.jsdelivr.net/gh/prebid/currency-file@1/latest.json`
- Fallback: `http://currency.prebid.org/latest.json`
- Updated daily from European Central Bank
- Published automatically via GitHub Actions

## Troubleshooting

### Rates Not Loading

```go
stats := converter.Stats()
if stats["ratesLoaded"] == false {
    // Check network connectivity
    // Check fetchErrors count
    // Verify FetchURL is accessible
}
```

### High Fetch Errors

```go
if stats["fetchErrors"].(int) > 5 {
    // Network issues or CDN down
    // Consider using fallback URL
    // Check firewall/proxy settings
}
```

### Unexpected Rates

```go
// Get all rates for inspection
rates := converter.GetRates()
fmt.Printf("USD->EUR: %v\n", rates["USD"]["EUR"])
```

## References

- [Prebid Currency Conversion](https://docs.prebid.org/prebid-server/features/pbs-currency.html)
- [Prebid Currency File](https://github.com/prebid/currency-file)
- [OpenRTB 2.5 Spec](https://www.iab.com/wp-content/uploads/2016/03/OpenRTB-API-Specification-Version-2-5-FINAL.pdf) (Section 3.2.4)
- [ISO 4217 Currency Codes](https://www.iso.org/iso-4217-currency-codes.html)

---

**Version**: 1.0.0
**Last Updated**: 2026-02-02
**Maintainer**: TNE Catalyst Team
