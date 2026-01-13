# TNE Catalyst Production Readiness Audit
**Generated**: 2026-01-13
**Target Domain**: catalyst.springwire.ai
**Status**: COMPLETE

## Audit Summary

This audit identified **21 issues** across 4 severity levels:
- CRITICAL: 4
- HIGH: 6
- MEDIUM: 7
- LOW: 4

---

## CRITICAL FINDINGS

### 1. Placeholder Passwords in Production Environment Files
**Files**:
- `/Users/andrewstreets/tne-catalyst/deployment/.env.production:32`
- `/Users/andrewstreets/tne-catalyst/deployment/.env.production:50`
- `/Users/andrewstreets/tne-catalyst/deployment/.env.staging:33`
- `/Users/andrewstreets/tne-catalyst/deployment/.env.staging:51`

**Issue**: Production and staging environment files contain placeholder passwords:
```
DB_PASSWORD=CHANGE_ME_STRONG_PASSWORD_HERE
REDIS_PASSWORD=CHANGE_ME_REDIS_PASSWORD
```

**Risk**: If deployed with these values, the database and Redis would be vulnerable.
**Remediation**: Replace with strong, randomly generated passwords before deployment.

### 2. Placeholder CORS Origins in Production Config
**Files**:
- `/Users/andrewstreets/tne-catalyst/deployment/.env.production:106`
- `/Users/andrewstreets/tne-catalyst/deployment/.env.staging:105`

**Issue**: CORS configuration contains placeholder domains:
```
CORS_ALLOWED_ORIGINS=https://yourpublisher.com,https://*.yourpublisher.com
```

**Risk**: Will reject legitimate publisher requests or may be misconfigured to allow wrong origins.
**Remediation**: Update with actual publisher domains before deployment.

### 3. Missing .env.example File Referenced in Documentation
**Files**:
- `/Users/andrewstreets/tne-catalyst/docs/DOCKER_DEPLOYMENT.md:78,234,241,250,300`

**Issue**: Documentation references `.env.example` which does not exist in the repository.
**Risk**: Deployment instructions will fail; users will be confused.
**Remediation**: Create `.env.example` file or update documentation to reference `.env.production`.

### 4. Hardcoded localhost Default for IDR Service URL
**File**: `/Users/andrewstreets/tne-catalyst/cmd/server/main.go:34`

**Issue**: Default IDR URL points to localhost:
```go
idrURL := flag.String("idr-url", getEnvOrDefault("IDR_URL", "http://localhost:5050"), "IDR service URL")
```

**File**: `/Users/andrewstreets/tne-catalyst/internal/exchange/exchange.go:119`
```go
IDRServiceURL: "http://localhost:5050",
```

**Risk**: In production, if IDR_URL env var is not set, will attempt to connect to localhost.
**Remediation**: Ensure IDR_URL is always set in production env files, or disable IDR by default.

---

## HIGH FINDINGS

### 5. Docker Image Reference Inconsistency
**File**: `/Users/andrewstreets/tne-catalyst/docs/DOCKER_DEPLOYMENT.md:488-489`

**Issue**: Documentation references `ghcr.io/streetsdigital/tne-catalyst` which may not match actual registry:
```
# Change: image: ghcr.io/streetsdigital/tne-catalyst:latest
# To: image: ghcr.io/streetsdigital/tne-catalyst:v1.0.0
```

**Risk**: Wrong registry path will cause deployment failures.
**Remediation**: Verify and update to correct container registry path.

### 6. Module Path Mismatch with Repository Name
**File**: `/Users/andrewstreets/tne-catalyst/go.mod:1`

**Issue**: Go module path is `github.com/thenexusengine/tne_springwire` but repository appears to be `tne-catalyst`.

**Risk**: Import paths in documentation and go.mod may not match actual GitHub repository.
**Remediation**: Ensure go.mod module path matches the actual GitHub repository URL.

### 7. Docker Compose Builds from Remote Git URL
**Files**:
- `/Users/andrewstreets/tne-catalyst/deployment/docker-compose.yml:7`
- `/Users/andrewstreets/tne-catalyst/deployment/docker-compose-split.yml:11,47`

**Issue**: Docker build context references remote GitHub URL:
```yaml
context: https://github.com/thenexusengine/tne_springwire.git
```

**Risk**: 
- Builds will fail if repository is private or URL is wrong
- Cannot build locally modified code
- Slower builds due to git clone

**Remediation**: Use local context (`.`) or pre-built images from container registry.

### 8. Nginx CORS Configuration Conflict
**File**: `/Users/andrewstreets/tne-catalyst/deployment/nginx.conf:93`

**Issue**: Nginx sets `Access-Control-Allow-Origin: *` unconditionally:
```nginx
add_header Access-Control-Allow-Origin "*" always;
```

This conflicts with the application's CORS middleware that has proper origin validation.

**Risk**: Bypasses application-level CORS security, allowing all origins.
**Remediation**: Remove nginx CORS headers or align with application CORS policy.

### 9. Wide-Open Admin Endpoints in Auth Bypass
**File**: `/Users/andrewstreets/tne-catalyst/internal/middleware/auth.go:46`

**Issue**: Auth middleware bypasses authentication for sensitive paths:
```go
BypassPaths: []string{"/health", "/status", "/metrics", "/info/bidders", "/cookie_sync", "/setuid", "/optout", "/openrtb2/auction", "/admin/dashboard", "/admin/metrics"},
```

**Risk**: `/admin/dashboard` and `/admin/metrics` are publicly accessible without authentication.
**Remediation**: Remove admin endpoints from bypass list or implement separate admin authentication.

### 10. IDR Enabled by Default
**File**: `/Users/andrewstreets/tne-catalyst/cmd/server/main.go:35`

**Issue**: IDR integration defaults to enabled:
```go
idrEnabled := flag.Bool("idr-enabled", getEnvBoolOrDefault("IDR_ENABLED", true), "Enable IDR integration")
```

But production env file has it disabled. If env var is not loaded, IDR attempts will be made to localhost.

**Risk**: Failed IDR connections may impact auction performance.
**Remediation**: Default to `false` and require explicit enable in production.

---

## MEDIUM FINDINGS

### 11. Example Email Addresses in Documentation
**Files**:
- `/Users/andrewstreets/tne-catalyst/docs/DOCKER_DEPLOYMENT.md:102,369,685`

**Issue**: Contains placeholder email addresses:
```
--email your-email@example.com
```

**Risk**: Users may not replace before running, causing SSL certificate issues.
**Remediation**: Use more obvious placeholder or add validation step.

### 12. Development Password in Dev Environment File
**File**: `/Users/andrewstreets/tne-catalyst/deployment/.env.dev:31`

**Issue**: Contains weak placeholder password:
```
DB_PASSWORD=dev_password_change_me
```

**Risk**: If .env.dev is accidentally used in production.
**Remediation**: Use environment variable or add checks to prevent production use.

### 13. Wildcard CORS in Development Config
**File**: `/Users/andrewstreets/tne-catalyst/deployment/.env.dev:103`

**Issue**: Development config has wildcard CORS:
```
CORS_ALLOWED_ORIGINS=*
```

**Risk**: Could be accidentally deployed to production.
**Remediation**: Add environment check or warning when wildcard is used.

### 14. Debug Endpoints Enabled in Staging
**File**: `/Users/andrewstreets/tne-catalyst/deployment/.env.staging:167-170`

**Issue**: Staging has debug/profiling endpoints enabled:
```
PPROF_ENABLED=true
DEBUG_ENDPOINTS=true
```

**Risk**: Exposes internal application state and profiling data.
**Remediation**: Document security implications; restrict access via nginx.

### 15. SSL Certificate Path Assumptions
**File**: `/Users/andrewstreets/tne-catalyst/deployment/nginx.conf:74-75`

**Issue**: Hardcoded SSL certificate paths:
```nginx
ssl_certificate /etc/nginx/ssl/fullchain.pem;
ssl_certificate_key /etc/nginx/ssl/privkey.pem;
```

**Risk**: Will fail if certificates are not in expected location.
**Remediation**: Document certificate setup or add existence check.

### 16. No Redis Password Requirement in Default Config
**Files**: 
- `/Users/andrewstreets/tne-catalyst/deployment/docker-compose.yml:45`

**Issue**: Redis runs without password requirement in docker-compose:
```yaml
command: redis-server --appendonly yes --maxmemory 1024mb --maxmemory-policy allkeys-lru
```

**Risk**: Redis accessible without authentication within container network.
**Remediation**: Add `--requirepass` with password from environment variable.

### 17. Outdated Repository URLs in Documentation
**Multiple files reference**: `https://github.com/thenexusengine/tne_springwire`

**Risk**: If repository has been renamed or moved, all documentation links are broken.
**Remediation**: Verify and update all repository references.

---

## LOW FINDINGS

### 18. Test Credentials in Test Files
**File**: `/Users/andrewstreets/tne-catalyst/internal/adapters/ortb/ortb_test.go:400,419,773`

**Issue**: Contains test credentials:
```go
config.Endpoint.AuthPassword = "pass"
config.Endpoint.AuthToken = "my-token-123"
```

**Risk**: Minimal - these are in test files only.
**Note**: Acceptable for testing purposes.

### 19. Example Domain References in Test Files
**Multiple test files contain**: `example.com`

**Risk**: None - standard test domain per RFC 2606.
**Note**: Acceptable for testing.

### 20. Version Hardcoded in Health Response
**File**: `/Users/andrewstreets/tne-catalyst/cmd/server/main.go:337`

**Issue**: Version is hardcoded:
```go
"version": "1.0.0",
```

**Risk**: Version may not reflect actual deployed version.
**Remediation**: Consider using build-time version injection.

### 21. Missing Health Check Timeout Validation
**File**: `/Users/andrewstreets/tne-catalyst/deployment/docker-compose.yml:21`

**Risk**: Health check could hang indefinitely in edge cases.
**Note**: Current timeout of 10s is reasonable.

---

## Deployment Checklist for catalyst.springwire.ai

Before deploying to production:

1. [ ] Replace `CHANGE_ME_STRONG_PASSWORD_HERE` in `.env.production` with strong DB password
2. [ ] Replace `CHANGE_ME_REDIS_PASSWORD` in `.env.production` with strong Redis password  
3. [ ] Update `CORS_ALLOWED_ORIGINS` with actual publisher domains
4. [ ] Verify repository URL matches actual GitHub location
5. [ ] Create `.env.example` file or update documentation
6. [ ] Update docker-compose to use local build context or pre-built images
7. [ ] Remove wildcard CORS from nginx.conf or align with application policy
8. [ ] Review admin endpoint access - add authentication if needed
9. [ ] Verify SSL certificates are in place at `/opt/catalyst/ssl/`
10. [ ] Add `--requirepass` to Redis configuration
11. [ ] Verify IDR_URL is set correctly or disable IDR
12. [ ] Update all `ghcr.io/streetsdigital` references to correct registry

---

## Files Audited

- Configuration: `.env.production`, `.env.staging`, `.env.dev`, `docker-compose.yml`, `nginx.conf`
- Documentation: `README.md`, `DEPLOYMENT_GUIDE.md`, `docs/DOCKER_DEPLOYMENT.md`, deployment READMEs
- Source: `cmd/server/main.go`, `internal/middleware/`, `internal/exchange/`
- Build: `Dockerfile`, `go.mod`

**Audit completed**: 2026-01-13
