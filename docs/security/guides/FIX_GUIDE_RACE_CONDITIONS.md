# FIX GUIDE: Race Condition Issues

## Issue 1: PauseAdTracker Concurrent Map Access

**File:** `internal/pauseads/pauseads.go`
**Lines:** 242-246
**Severity:** CRITICAL - Guaranteed panic under load

### Problem
```go
// BROKEN - No mutex protection
type PauseAdTracker struct {
    impressions map[string][]time.Time
}

func (t *PauseAdTracker) RecordImpression(sessionID string) {
    t.impressions[sessionID] = append(t.impressions[sessionID], time.Now())
    t.cleanupOldImpressions(sessionID)
}
```

### Fix
```go
// FIXED - Add mutex protection
type PauseAdTracker struct {
    mu          sync.RWMutex
    impressions map[string][]time.Time
}

func (t *PauseAdTracker) CanShowAd(sessionID string, cap *FrequencyCap) bool {
    if cap == nil {
        return true
    }

    now := time.Now()
    cutoff := now.Add(-time.Duration(cap.TimeWindowSeconds) * time.Second)

    t.mu.RLock()
    impressions, ok := t.impressions[sessionID]
    t.mu.RUnlock()

    if !ok {
        return true
    }

    count := 0
    for _, imp := range impressions {
        if imp.After(cutoff) {
            count++
        }
    }

    return count < cap.MaxImpressions
}

func (t *PauseAdTracker) RecordImpression(sessionID string) {
    t.mu.Lock()
    defer t.mu.Unlock()

    t.impressions[sessionID] = append(t.impressions[sessionID], time.Now())
    t.cleanupOldImpressionsLocked(sessionID)
}

func (t *PauseAdTracker) cleanupOldImpressionsLocked(sessionID string) {
    cutoff := time.Now().Add(-24 * time.Hour)
    impressions := t.impressions[sessionID]

    var cleaned []time.Time
    for _, imp := range impressions {
        if imp.After(cutoff) {
            cleaned = append(cleaned, imp)
        }
    }

    if len(cleaned) > 0 {
        t.impressions[sessionID] = cleaned
    } else {
        delete(t.impressions, sessionID)
    }
}
```

---

## Issue 2: EventRecorder Buffer Race

**File:** `pkg/idr/events.go`
**Lines:** 182-189
**Severity:** CRITICAL - Memory corruption

### Problem
```go
r.mu.Lock()
r.buffer = append(r.buffer, event)
shouldFlush := len(r.buffer) >= r.bufferSize
var eventsToFlush []BidEvent
if shouldFlush {
    eventsToFlush = r.buffer  // References same underlying array!
    r.buffer = make([]BidEvent, 0, r.bufferSize)
}
r.mu.Unlock()

// eventsToFlush still references old buffer memory
```

### Fix
```go
r.mu.Lock()
r.buffer = append(r.buffer, event)
shouldFlush := len(r.buffer) >= r.bufferSize
var eventsToFlush []BidEvent
if shouldFlush {
    // Copy buffer before unlock
    eventsToFlush = make([]BidEvent, len(r.buffer))
    copy(eventsToFlush, r.buffer)
    r.buffer = make([]BidEvent, 0, r.bufferSize)
}
r.mu.Unlock()
```

---

## Issue 3: Publisher Cache Race

**File:** `internal/middleware/publisher_auth.go`
**Lines:** 100-105
**Severity:** HIGH - Data races

### Problem
```go
type PublisherAuth struct {
    // ...
    rateLimits   map[string]*rateLimitEntry      // No mutex!
    publisherCache   map[string]*publisherCacheEntry  // No mutex!
}
```

### Fix
Maps already have separate mutexes declared (lines 101, 105) but they're not consistently used. Ensure all access uses locks:

```go
// Already correct in struct:
rateLimits   map[string]*rateLimitEntry
rateLimitsMu sync.RWMutex

publisherCache   map[string]*publisherCacheEntry
publisherCacheMu sync.RWMutex

// Ensure all map access uses appropriate locks
func (p *PublisherAuth) checkRateLimit(publisherID string) bool {
    p.rateLimitsMu.Lock()
    defer p.rateLimitsMu.Unlock()

    // ... access rateLimits map safely
}

func (p *PublisherAuth) getCachedPublisher(publisherID string) string {
    p.publisherCacheMu.RLock()
    defer p.publisherCacheMu.RUnlock()

    // ... access publisherCache map safely
}
```

The code mostly has this correct, but verify all map access paths use the mutexes.

---

## Testing Checklist

After applying fixes:

- [ ] Run `go test -race ./...` to detect remaining races
- [ ] Load test with `go test -race -run=TestConcurrentPauseAd`
- [ ] Verify no panics under concurrent load
- [ ] Check performance impact of locks (should be minimal)
- [ ] Review all mutex lock/unlock pairs for deadlock potential
