# QUICK REFERENCE: Critical Bugs

## Top 12 Critical Fixes (Priority Order)

### 🔴 SERVICE CRASH RISK

1. **PauseAdTracker race** → Add `sync.RWMutex`
   - File: `internal/pauseads/pauseads.go:242`
   - Fix: See FIX_GUIDE_RACE_CONDITIONS.md

2. **Division by zero** → Add validation
   - File: `internal/exchange/exchange.go:777`
   - Fix: `if multiplier == 0 { return bidsByImp }`

3. **Adapter init panics** → Replace with errors
   - Files: `internal/adapters/**/init()`
   - Fix: Replace `panic()` with `log.Error()`

### 🔴 AUTHENTICATION BYPASS

4. **/metrics exposed** → Add auth middleware
   - File: `cmd/server/server.go:268`
   - Fix: Remove from bypass list

5. **/admin exposed** → Add auth middleware
   - File: `cmd/server/server.go:275-278`
   - Fix: Remove from bypass list

6. **Auth bypass** → Fix HasPrefix logic
   - File: `internal/middleware/auth.go:143`
   - Fix: Use exact path matching

### 🔴 INJECTION ATTACKS

7. **JSON injection** → Use json.NewEncoder
   - File: `internal/middleware/publisher_auth.go:204`
   - Fix: `json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})`

8. **VAST URL injection** → Use url.QueryEscape
   - File: `internal/endpoints/video_handler.go:274`
   - Fix: `url.QueryEscape(message)`

9. **Header injection** → Validate custom headers
   - File: `internal/adapters/ortb/ortb.go:418`
   - Fix: Whitelist header names

### 🔴 DATA CORRUPTION

10. **Shallow copy** → Deep copy pointers
    - File: `internal/exchange/exchange.go:1513`
    - Fix: Copy Banner/Video/Native fields

11. **EventRecorder buffer** → Copy before unlock
    - File: `pkg/idr/events.go:182`
    - Fix: See FIX_GUIDE_RACE_CONDITIONS.md

12. **Content-Length bypass** → Validate -1 case
    - File: `internal/middleware/sizelimit.go:72`
    - Fix: `if r.ContentLength < 0 || r.ContentLength > maxBodySize`

---

## High Priority Fixes (Next 10)

13. CircuitBreaker callback timeout
14. Metric cardinality explosion
15. Publisher cache race
16. GDPR PII in logs
17. TCF validation weakness
18. OpenRTB Response ID validation
19. 6 adapters hardcode bid types
20. Price adjustment bounds
21. Config validation
22. Integer overflow guards

---

## Commands

### Run Tests
```bash
# Check for races
go test -race ./...

# Run load tests
go test -v ./tests/load/...

# Check coverage
go test -cover ./...
```

### Security Scan
```bash
# Lint
golangci-lint run

# Security audit
gosec ./...

# Dependency check
go mod verify
```

### Apply Fixes
```bash
# Create branch
git checkout -b fix/critical-bugs

# Apply patches (if generated)
git apply fix-race-conditions.patch
git apply fix-resource-leaks.patch
git apply fix-validation.patch

# Test
go test -race ./...

# Commit
git commit -m "fix: address critical race conditions and security issues"
```

---

## File References

**Bug Reports:**
- `BUG_REPORT_MASTER.md` - Complete inventory (85+ bugs)

**Fix Guides:**
- `FIX_GUIDE_RACE_CONDITIONS.md` - Concurrency fixes
- `FIX_GUIDE_RESOURCE_LEAKS.md` - Memory/goroutine leaks
- `FIX_GUIDE_VALIDATION.md` - Input validation (TBD)

**Agent Transcripts:**
- `/private/tmp/claude/-Users-andrewstreets-tnevideo/tasks/*.output`

---

## Emergency Contacts

If critical issues found in production:
1. Disable affected endpoints via config
2. Roll back to previous version
3. Apply hotfixes from this guide
4. Run full test suite before redeployment
