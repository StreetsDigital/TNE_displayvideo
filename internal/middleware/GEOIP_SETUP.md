# GeoIP Setup for IVT Detection

The Invalid Traffic (IVT) detector now supports geographic IP-based filtering using MaxMind GeoIP2/GeoLite2 databases.

## Quick Start

### 1. Download GeoLite2 Database (Free)

MaxMind offers a free GeoLite2-Country database:

1. Sign up for a free account at https://www.maxmind.com/en/geolite2/signup
2. Generate a license key
3. Download the GeoLite2-Country database in MMDB format

Alternatively, use `geoipupdate` to automate updates:

```bash
# Install geoipupdate
# macOS
brew install geoipupdate

# Ubuntu/Debian
sudo apt-get install geoipupdate

# Configure with your license key
sudo vi /etc/GeoIP.conf
# Add:
# AccountID YOUR_ACCOUNT_ID
# LicenseKey YOUR_LICENSE_KEY
# EditionIDs GeoLite2-Country

# Run update
sudo geoipupdate

# Database will be at:
# /usr/share/GeoIP/GeoLite2-Country.mmdb (Linux)
# /usr/local/var/GeoIP/GeoLite2-Country.mmdb (macOS)
```

### 2. Configure Environment Variable

Set the path to the GeoIP database:

```bash
export GEOIP_DB_PATH="/usr/share/GeoIP/GeoLite2-Country.mmdb"
```

### 3. Enable Geo Checking

Enable geographic restriction checking:

```bash
export IVT_CHECK_GEO=true
```

### 4. Configure Country Restrictions (Optional)

**Whitelist approach** (only allow specific countries):

```bash
export IVT_ALLOWED_COUNTRIES="US,GB,CA,AU"
```

**Blacklist approach** (block specific countries):

```bash
export IVT_BLOCKED_COUNTRIES="CN,RU,KP"
```

**Note:** If `IVT_ALLOWED_COUNTRIES` is set, it takes precedence over `IVT_BLOCKED_COUNTRIES`.

## Configuration Options

| Environment Variable | Type | Default | Description |
|---------------------|------|---------|-------------|
| `GEOIP_DB_PATH` | string | `""` | Path to MaxMind GeoIP2/GeoLite2 database (.mmdb file) |
| `IVT_CHECK_GEO` | bool | `false` | Enable geographic IP restriction checking |
| `IVT_ALLOWED_COUNTRIES` | []string | `[]` | Whitelist of ISO country codes (comma-separated) |
| `IVT_BLOCKED_COUNTRIES` | []string | `[]` | Blacklist of ISO country codes (comma-separated) |

## Country Codes

Use ISO 3166-1 alpha-2 country codes (2-letter codes):

| Country | Code |
|---------|------|
| United States | `US` |
| United Kingdom | `GB` |
| Canada | `CA` |
| Australia | `AU` |
| Germany | `DE` |
| France | `FR` |
| China | `CN` |
| Russia | `RU` |
| India | `IN` |

Full list: https://en.wikipedia.org/wiki/ISO_3166-1_alpha-2

## Testing

Test your GeoIP configuration:

```bash
# Set environment variables
export GEOIP_DB_PATH="/usr/share/GeoIP/GeoLite2-Country.mmdb"
export IVT_CHECK_GEO=true
export IVT_BLOCKED_COUNTRIES="CN,RU"

# Run the application
go run cmd/server/main.go

# Check logs for:
# {"level":"info","path":"/usr/share/GeoIP/GeoLite2-Country.mmdb","message":"GeoIP database loaded successfully"}
```

Send a test request with a specific country IP (use a proxy or VPN):

```bash
curl -H "X-Forwarded-For: 1.2.4.8" http://localhost:8080/openrtb2/auction \
  -d '{"id":"test","imp":[{"id":"1"}]}'
```

Check the response for geo-blocking if the IP is from a restricted country.

## Troubleshooting

### Database Not Found

```
{"level":"warn","error":"open /path/to/db.mmdb: no such file or directory","message":"Failed to initialize GeoIP database"}
```

**Solution:** Verify `GEOIP_DB_PATH` points to a valid .mmdb file.

### GeoIP Lookup Fails

```
{"level":"debug","ip":"1.2.3.4","error":"...","message":"GeoIP lookup failed"}
```

**Possible causes:**
- Private IP address (10.x.x.x, 192.168.x.x, 127.0.0.1)
- Database doesn't contain the IP
- Corrupt database file

**Solution:** Use a public IP address for testing.

### Countries Not Being Blocked

1. Verify `IVT_CHECK_GEO=true` is set
2. Check `IVT_MONITORING_ENABLED=true` (required for any IVT checks)
3. Verify the IP's country code matches your blocked list
4. Check application logs for geo check results

## Performance

- **Memory:** GeoLite2-Country database is ~6MB in memory
- **Latency:** Country lookups add ~0.1-0.5ms per request
- **Caching:** Database is memory-mapped for fast lookups
- **Updates:** Refresh database weekly for accuracy

## GeoIP2 vs GeoLite2

| Feature | GeoLite2 (Free) | GeoIP2 (Paid) |
|---------|----------------|---------------|
| Country accuracy | ~99.8% | ~99.8% |
| City data | Yes | Yes (more accurate) |
| ASN data | Available | Available |
| Update frequency | Weekly | Daily |
| Commercial use | Allowed | Allowed |

For most use cases, GeoLite2-Country is sufficient. Upgrade to GeoIP2 if you need:
- Higher accuracy for city-level detection
- More frequent updates
- Additional metadata

## License

GeoLite2 databases are distributed under the Creative Commons Attribution-ShareAlike 4.0 International License.

When using GeoLite2 databases:
- Include attribution: "This product includes GeoLite2 data created by MaxMind, available from https://www.maxmind.com"
- Database updates are your responsibility

## Additional Resources

- MaxMind GeoIP2 Documentation: https://dev.maxmind.com/geoip/geoip2/downloadable/
- GeoIP2 Go Library: https://github.com/oschwald/geoip2-golang
- ISO Country Codes: https://en.wikipedia.org/wiki/ISO_3166-1_alpha-2
