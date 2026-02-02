# TNE Catalyst Documentation

Complete documentation for the TNE Catalyst ad exchange platform.

## Quick Links

### 📚 Guides
- [Operations Guide](./guides/OPERATIONS-GUIDE.md) - Day-to-day operations and maintenance
- [Bidder Parameters Guide](./guides/BIDDER-PARAMS-GUIDE.md) - Bidder configuration
- [Bidder Management](./guides/BIDDER-MANAGEMENT.md) - Managing bidder integrations
- [Publisher Config Guide](./guides/PUBLISHER-CONFIG-GUIDE.md) - Publisher setup
- [Publisher Management](./guides/PUBLISHER-MANAGEMENT.md) - Managing publisher accounts

### 🔌 Publisher Integrations
- [Integration Overview](./integrations/README.md) - All integration methods
  - [OpenRTB Direct](./integrations/openrtb-direct/README.md) - ✅ Production Ready
  - [Video VAST](./integrations/video-vast/README.md) - ✅ Production Ready
  - [Web via Prebid](./integrations/web-prebid/README.md) - ⚠️ Needs Examples
  - [Video via Prebid](./integrations/video-prebid/README.md) - ⚠️ Needs Examples
  - [In-App SDK](./integrations/in-app-sdk/README.md) - ❌ SDK Missing

### 🎥 Video Integration
- [Video Integration Overview](./video/VIDEO_INTEGRATION.md) - Video ad integration guide
- [Video End-to-End Complete](./video/VIDEO_E2E_COMPLETE.md) - Complete video implementation
- [Video Test Summary](./video/VIDEO_TEST_SUMMARY.md) - Video testing documentation

### 📖 API Reference
- [API Reference](./api/API-REFERENCE.md) - Complete API documentation

### 🚀 Deployment
- [Deployment Guide](./deployment/DEPLOYMENT_GUIDE.md) - Full deployment guide
- [Local Deployment](./deployment/LOCAL_DEPLOYMENT.md) - Local development setup
- [Production Readiness Report](./deployment/PRODUCTION-READINESS-REPORT.md) - Production checklist
- [Production Deployment Checklist](./deployment/PRODUCTION-DEPLOYMENT-CHECKLIST.md) - Pre-deployment verification
- [Deployment Checklist](./deployment/DEPLOYMENT-CHECKLIST.md) - General deployment steps
- [Disaster Recovery](./deployment/DISASTER-RECOVERY.md) - Disaster recovery procedures
- [Prometheus Metrics](./deployment/PROMETHEUS-METRICS.md) - Metrics and monitoring
- [Backup System Summary](./deployment/BACKUP-SYSTEM-SUMMARY.md) - Backup configuration
- [DB Health Check Summary](./deployment/DB-HEALTH-CHECK-SUMMARY.md) - Database health monitoring

### 🔒 Security
- [Security Quick Reference](./security/QUICK_REFERENCE.md) - Security guidelines
- [Bug Report Master](./security/BUG_REPORT_MASTER.md) - Security issues tracking
- [Fixes Applied](./security/FIXES_APPLIED.md) - Applied security fixes
- [Security Config Fixes](./security/SECURITY-CONFIG-FIXES.md) - Configuration security
- [Security Fix Summary](./security/SECURITY-FIX-SUMMARY.md) - Summary of fixes
- [Database Security Fixes](./security/DATABASE_SECURITY_FIXES.md) - Database security
- [Redis Password Fix](./security/REDIS-PASSWORD-FIX-SUMMARY.md) - Redis security
- [Resource Leak Fixes](./security/RESOURCE_LEAK_FIXES.md) - Memory leak fixes
- **Fix Guides:**
  - [Race Conditions](./security/guides/FIX_GUIDE_RACE_CONDITIONS.md)
  - [Resource Leaks](./security/guides/FIX_GUIDE_RESOURCE_LEAKS.md)

### 🔐 Privacy & Compliance
- [Geo & Consent Guide](./privacy/GEO-CONSENT-GUIDE.md) - GDPR/CCPA compliance
- [TCF Vendor Consent Guide](./privacy/TCF-VENDOR-CONSENT-GUIDE.md) - IAB TCF integration
- [Privacy Middleware Tests](./privacy/PRIVACY-MIDDLEWARE-TESTS-SUMMARY.md) - Test results

### ⚡ Performance
- [Performance Benchmarks](./performance/PERFORMANCE-BENCHMARKS.md) - Benchmark results
- [Performance Optimizations](./performance/PERFORMANCE_OPTIMIZATIONS.md) - Optimization guide
- [Performance Tuning](./performance/PERFORMANCE-TUNING.md) - Production tuning guide
- [Load Test Results](./performance/LOAD-TEST-RESULTS.md) - Load testing analysis

### 🧪 Testing
- [Test Coverage Status](./testing/TEST_COVERAGE_STATUS.md) - Coverage reports
- [E2E Test Report](./testing/E2E-TEST-REPORT.md) - End-to-end testing
- [Test Run Summary](./testing/TEST_RUN_SUMMARY.md) - Test execution summary
- [Security Testing](./testing/SECURITY_TESTING.md) - Security test suite
- [Video Test README](./testing/VIDEO_TEST_README.md) - Video testing documentation

### 📊 Audits
Security and code quality audits from 2026-01-26:
- [API Gatekeeper](./audits/2026-01-26-api-gatekeeper.md)
- [Concurrency Audit](./audits/2026-01-26-concurrency-audit.md)
- [Concurrency Cop](./audits/2026-01-26-concurrency-cop.md)
- [Go Guardian](./audits/2026-01-26-go-guardian.md)
- [Go Idiom Fixes](./audits/2026-01-26-go-idiom-fixes.md)
- [Privacy Compliance](./audits/2026-01-26-privacy-compliance.md)
- [Privacy Fixes](./audits/2026-01-26-privacy-fixes.md)
- [Test Tsar](./audits/2026-01-26-test-tsar.md)

### 🛠️ Development
- [Lock Ordering Fix](./development/LOCK_ORDERING_FIX.md) - Middleware lock ordering documentation
- [GeoIP Setup](./development/GEOIP_SETUP.md) - GeoIP database configuration

## Documentation Structure

```
docs/
├── README.md (this file)
├── api/                 # API documentation
├── guides/              # Operational guides and how-tos
├── integrations/        # Publisher integration methods
├── video/               # Video-specific documentation
├── deployment/          # Deployment and infrastructure
├── development/         # Development guides and fixes
├── security/            # Security documentation and fixes
├── privacy/             # Privacy and compliance
├── performance/         # Performance benchmarks and tuning
├── testing/             # Test reports and coverage
├── audits/              # Security and quality audits
└── examples/            # Example configurations and requests
```

## Getting Started

1. **For Publishers**: Start with [Integration Overview](./integrations/README.md)
2. **For Operators**: See [Operations Guide](./guides/OPERATIONS-GUIDE.md)
3. **For Developers**: Check [Security Quick Reference](./security/QUICK_REFERENCE.md)

## Additional Documentation

For deployment-specific documentation, see:
- `/tnevideo/deployment/` - Deployment configurations and scripts
- `/tnevideo/deployment/PRODUCTION-DEPLOYMENT-CHECKLIST.md` - Production deployment guide
- `/tnevideo/deployment/README.md` - Deployment documentation

---

**Last Updated:** 2026-02-02
