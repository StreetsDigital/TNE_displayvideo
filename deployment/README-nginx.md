# nginx.conf - Reverse Proxy Configuration

## Purpose
Nginx acts as the reverse proxy and SSL termination point for catalyst.springwire.ai. It handles all incoming traffic and forwards requests to the Catalyst container.

## What This File Does

```
Internet Traffic
      ↓
catalyst.springwire.ai:443 (HTTPS)
      ↓
   NGINX (this config)
      ↓
catalyst:8000 (internal Docker network)
```

## Key Configuration Decisions

### 1. Domain Name
```nginx
server_name catalyst.springwire.ai;
```
**Decision**: Hardcoded to catalyst.springwire.ai
**Why**: Single-server deployment, specific to this integration
**To Change**: Edit this line if domain changes

### 2. SSL Certificates
```nginx
ssl_certificate /etc/nginx/ssl/fullchain.pem;
ssl_certificate_key /etc/nginx/ssl/privkey.pem;
```
**Decision**: Expects certs in `./ssl/` directory
**Why**: Your colleague handles SSL certificate setup externally
**Required Files**:
- `/opt/catalyst/ssl/fullchain.pem` (certificate)
- `/opt/catalyst/ssl/privkey.pem` (private key)

### 3. Rate Limiting
```nginx
limit_req_zone $binary_remote_addr zone=general:10m rate=100r/s;
limit_req_zone $binary_remote_addr zone=auction:10m rate=50r/s;
```
**Decision**:
- General endpoints: 100 requests/second per IP
- Auction endpoint: 50 requests/second per IP (stricter)
**Why**: Prevent abuse and DoS attacks on auction endpoint
**To Adjust**: Increase rates if legitimate traffic is being blocked

### 4. Auction Endpoint Timeouts
```nginx
location /openrtb2/auction {
    proxy_connect_timeout 5s;
    proxy_send_timeout 10s;
    proxy_read_timeout 10s;
    proxy_buffering off;
}
```
**Decision**: Short timeouts (5-10 seconds)
**Why**: Ad auctions need fast responses (<1 second typical)
**Buffering Off**: Real-time auction responses can't wait
**To Change**: Only increase if you see timeout errors in logs

### 5. Security Headers
```nginx
add_header X-Frame-Options "SAMEORIGIN" always;
add_header X-Content-Type-Options "nosniff" always;
add_header X-XSS-Protection "1; mode=block" always;
add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
```
**Decision**: Modern security headers enabled
**Why**: Protect against common web attacks (XSS, clickjacking, MIME sniffing)
**HSTS**: Forces HTTPS for 1 year after first visit

### 6. HTTP to HTTPS Redirect
```nginx
server {
    listen 80;
    location / {
        return 301 https://$server_name$request_uri;
    }
}
```
**Decision**: All HTTP traffic redirected to HTTPS
**Why**: Enforce encrypted connections
**Exception**: Health check allowed on HTTP for monitoring

### 7. CORS Headers
```nginx
add_header Access-Control-Allow-Origin "*" always;
add_header Access-Control-Allow-Methods "GET, POST, OPTIONS" always;
```
**Decision**: Allow all origins (wildcard)
**Why**: Catalyst handles CORS in application code, this is fallback
**To Restrict**: Change `"*"` to specific domains if needed

### 8. Connection Limits
```nginx
limit_conn addr 10;
```
**Decision**: Max 10 concurrent connections per IP
**Why**: Prevent single client from exhausting server resources
**To Adjust**: Increase if publisher has many concurrent requests

### 9. Request Body Size
```nginx
client_max_body_size 1M;
```
**Decision**: Max 1MB request body
**Why**: OpenRTB bid requests are typically <50KB
**To Change**: Increase only if you see "413 Entity Too Large" errors

### 10. Worker Connections
```nginx
worker_connections 2048;
```
**Decision**: 2048 simultaneous connections
**Why**: Handles ~2000 QPS comfortably
**To Scale**: Increase for higher traffic (4096, 8192, etc.)

## Traffic Flow

### Normal Auction Request
```
Publisher Website
    ↓ (HTTPS POST)
https://catalyst.springwire.ai/openrtb2/auction
    ↓ (Nginx receives)
Rate limit check (50/s per IP) ✓
    ↓
SSL termination
    ↓
Proxy to http://catalyst:8000/openrtb2/auction
    ↓
Catalyst processes auction
    ↓
Response back through Nginx
    ↓
Publisher Website
```

### Health Check Request
```
Monitoring Tool
    ↓
https://catalyst.springwire.ai/health
    ↓ (Nginx receives)
No rate limit (health checks exempt)
    ↓
Proxy to http://catalyst:8000/health
    ↓
Response: {"status":"healthy"}
```

## Logging

### Access Logs
```
Location: /opt/catalyst/nginx-logs/access.log
Format: Includes request timing (rt, uct, uht, urt)
```

**What's logged:**
- Client IP
- Request method and URL
- Response status code
- Request time (total time to serve request)
- Upstream connect time (time to connect to Catalyst)
- Upstream response time (time for Catalyst to respond)

**Example log line:**
```
203.0.113.42 - - [13/Jan/2025:10:30:15 +0000] "POST /openrtb2/auction HTTP/2.0" 200 1024
"https://publisher.com" "Mozilla/5.0..." rt=0.045 uct="0.001" uht="0.002" urt="0.042"
```

### Error Logs
```
Location: /opt/catalyst/nginx-logs/error.log
Level: warn (warnings and errors only)
```

## Common Adjustments

### Increase Rate Limits
If you're blocking legitimate traffic:
```nginx
# Change this:
limit_req_zone $binary_remote_addr zone=auction:10m rate=50r/s;

# To this:
limit_req_zone $binary_remote_addr zone=auction:10m rate=200r/s;
```

### Restrict CORS to Specific Domains
If you want to lock down CORS:
```nginx
# Remove wildcard
# add_header Access-Control-Allow-Origin "*" always;

# Add specific domains
add_header Access-Control-Allow-Origin "https://yourpublisher.com" always;
```

### Increase Timeouts
If seeing timeout errors:
```nginx
# In /openrtb2/auction location
proxy_connect_timeout 10s;  # was 5s
proxy_read_timeout 15s;     # was 10s
```

### Add IP Whitelist
To only allow specific IPs:
```nginx
location /openrtb2/auction {
    allow 203.0.113.0/24;  # Allow this subnet
    allow 198.51.100.42;   # Allow specific IP
    deny all;              # Deny everyone else

    # ... rest of config
}
```

## Troubleshooting

### Problem: 502 Bad Gateway
**Cause**: Nginx can't reach Catalyst container
**Check**:
```bash
docker compose ps catalyst  # Is it running?
docker compose logs catalyst  # Any errors?
```

### Problem: 413 Request Entity Too Large
**Cause**: Request body exceeds 1MB limit
**Fix**: Increase `client_max_body_size`

### Problem: 429 Too Many Requests
**Cause**: Rate limit exceeded
**Check logs**: Which IP is hitting limits?
**Fix**: Increase rate limit or investigate if it's an attack

### Problem: SSL certificate errors
**Cause**: Certificates not found or invalid
**Check**:
```bash
ls -la /opt/catalyst/ssl/
# Should see fullchain.pem and privkey.pem
```

### Problem: Slow response times
**Check logs**: Look at `urt` (upstream response time) values
- If high: Problem is in Catalyst
- If low: Problem is in network/Nginx

## Testing

### Test HTTP to HTTPS redirect
```bash
curl -I http://catalyst.springwire.ai/health
# Should see: HTTP/1.1 301 Moved Permanently
# Location: https://catalyst.springwire.ai/health
```

### Test SSL configuration
```bash
curl -I https://catalyst.springwire.ai/health
# Should see: HTTP/2 200
```

### Test rate limiting
```bash
# Send 100 requests quickly
for i in {1..100}; do curl https://catalyst.springwire.ai/health; done
# After ~50 requests, should see 429 errors
```

### Test proxy to Catalyst
```bash
curl -X POST https://catalyst.springwire.ai/openrtb2/auction \
  -H "Content-Type: application/json" \
  -d '{"id":"test","imp":[{"id":"1","banner":{"w":300,"h":250}}]}'
```

## Security Considerations

### What's Protected
- ✅ SSL/TLS encryption (TLS 1.2 and 1.3 only)
- ✅ Rate limiting per IP
- ✅ Connection limits per IP
- ✅ Security headers (XSS, clickjacking protection)
- ✅ HSTS (force HTTPS)
- ✅ Hidden file access blocked (/.git, /.env, etc.)
- ✅ Server version hidden (server_tokens off)

### What's NOT Protected
- ❌ DDoS attacks from distributed sources (need Cloudflare for this)
- ❌ Application-level attacks (handled by Catalyst)
- ❌ Brute force on specific endpoints (add fail2ban if needed)

## Performance

### Expected Capacity
With current settings:
- **Throughput**: ~2000 requests/second
- **Concurrent connections**: 2048
- **Rate limit**: 50 auction requests/second per IP

### To Scale Higher
1. Increase `worker_connections` to 4096 or 8192
2. Increase rate limits
3. Add more Catalyst instances and use Nginx load balancing

## File Location on Server

```
/opt/catalyst/nginx.conf
```

This file is mounted into the Nginx container at:
```
/etc/nginx/nginx.conf
```

## Reloading Configuration

After making changes:
```bash
# Test configuration
docker compose exec nginx nginx -t

# Reload (no downtime)
docker compose exec nginx nginx -s reload

# Or restart container
docker compose restart nginx
```

## Related Files
- `docker-compose.yml` - Defines how Nginx container runs
- `ssl/fullchain.pem` - SSL certificate (your colleague provides)
- `ssl/privkey.pem` - SSL private key (your colleague provides)
- `.env` - Environment variables (doesn't affect Nginx directly)

---

**Last Updated**: 2025-01-13
**Deployment**: catalyst.springwire.ai
**Maintainer**: The Nexus Engine / Springwire
