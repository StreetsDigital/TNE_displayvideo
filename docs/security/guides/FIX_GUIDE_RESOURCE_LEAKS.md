# FIX GUIDE: Resource Leak Issues

## Issue 1: CircuitBreaker Callback Goroutine Leak

**File:** `pkg/idr/circuitbreaker.go`
**Lines:** 178-185
**Severity:** HIGH - Goroutine leak if callback blocks

### Problem
```go
if cb.config.OnStateChange != nil {
    cb.callbackWg.Add(1)
    go func(from, to string) {
        defer cb.callbackWg.Done()
        cb.config.OnStateChange(from, to)  // Can block forever!
    }(oldState, newState)
}
```

### Fix
```go
if cb.config.OnStateChange != nil {
    cb.callbackWg.Add(1)
    go func(from, to string) {
        defer cb.callbackWg.Done()

        // Add 5-second timeout
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()

        // Run callback with timeout protection
        done := make(chan struct{})
        go func() {
            cb.config.OnStateChange(from, to)
            close(done)
        }()

        select {
        case <-done:
            // Callback completed
        case <-ctx.Done():
            // Callback timed out - goroutine will eventually return
        }
    }(oldState, newState)
}
```

**Required Import:**
```go
import (
    "context"
    "errors"
    "sync"
    "time"
)
```

---

## Issue 2: EventRecorder Shutdown Race

**File:** `pkg/idr/events.go`
**Lines:** 78-79, 92-107
**Severity:** MEDIUM - Workers may not stop cleanly

### Problem
```go
type EventRecorder struct {
    stopCh     chan struct{}  // Unbuffered!
    // ...
}

func (r *EventRecorder) flushWorker() {
    defer r.wg.Done()
    for {
        select {
        case <-r.stopCh:
            return
        case events, ok := <-r.flushQueue:
            // Race: both cases can match
            if !ok {
                return
            }
            // ...
        }
    }
}
```

### Fix Option 1: Buffered Channel (Simple)
```go
type EventRecorder struct {
    stopCh     chan struct{}  // Change in NewEventRecorder
    // ...
}

func NewEventRecorder(/* ... */) *EventRecorder {
    r := &EventRecorder{
        // ...
        flushQueue: make(chan []BidEvent, flushQueueSize),
        stopCh:     make(chan struct{}, 1),  // Buffered!
        // ...
    }
    // ...
}
```

### Fix Option 2: sync.Once (Robust)
```go
type EventRecorder struct {
    // ...
    stopCh     chan struct{}
    stopOnce   sync.Once
    // ...
}

func (r *EventRecorder) Close() error {
    var closeErr error
    r.stopOnce.Do(func() {
        // Signal workers to stop
        close(r.stopCh)

        // Flush remaining buffer
        ctx, cancel := context.WithTimeout(context.Background(), flushTimeout)
        defer cancel()
        closeErr = r.Flush(ctx)

        // Close queue and wait
        close(r.flushQueue)
        r.wg.Wait()
    })
    return closeErr
}
```

---

## Issue 3: Config Validation Missing

**File:** `cmd/server/config.go`
**Lines:** 52-84
**Severity:** MEDIUM - Bad config causes runtime failures

### Problem
```go
// No validation on parsed config
port := flag.String("port", getEnvOrDefault("PBS_PORT", "8000"), "Server port")
// Could be empty, "abc", or out of range
```

### Fix
Add validation method to ServerConfig:

```go
// Add to config.go after ServerConfig definition
func (c *ServerConfig) Validate() error {
    // Validate port
    if c.Port == "" {
        return errors.New("server port is required")
    }
    portNum, err := strconv.Atoi(c.Port)
    if err != nil {
        return fmt.Errorf("invalid port number: %w", err)
    }
    if portNum < 1 || portNum > 65535 {
        return fmt.Errorf("port must be between 1 and 65535, got %d", portNum)
    }

    // Validate IDR config
    if c.IDREnabled {
        if c.IDRUrl == "" {
            return errors.New("IDR URL is required when IDR is enabled")
        }
        if c.IDRAPIKey == "" {
            return errors.New("IDR API key is required when IDR is enabled")
        }
    }

    // Validate database config
    if c.DatabaseConfig != nil {
        if c.DatabaseConfig.Host == "" {
            return errors.New("database host is required")
        }
        if c.DatabaseConfig.User == "" {
            return errors.New("database user is required")
        }
    }

    // Validate timeout
    if c.Timeout <= 0 {
        return errors.New("timeout must be positive")
    }
    if c.Timeout > 30*time.Second {
        return errors.New("timeout cannot exceed 30 seconds")
    }

    return nil
}
```

**Call in main:**
```go
cfg := ParseConfig()
if err := cfg.Validate(); err != nil {
    log.Fatalf("Configuration validation failed: %v", err)
}
```

---

## Issue 4: Integer Overflow in Conversions

**File:** `internal/exchange/exchange.go`
**Lines:** 1950-1954
**Severity:** LOW - Unlikely but possible

### Problem
```go
case price <= 5:
    bucket = float64(int(price*100)) / 100
    // If price*100 > maxInt, overflow occurs
```

### Fix
```go
func (e *Exchange) getPriceBucket(price float64) string {
    // Input validation
    if price <= 0 {
        return "0.00"
    }
    if price > 20 {
        price = 20
    }

    const maxInt = int(^uint(0) >> 1)
    var bucket float64

    switch {
    case price <= 5:
        temp := price * 100
        if temp > float64(maxInt) {
            bucket = 5.0
        } else {
            bucket = float64(int(temp)) / 100
        }
    case price <= 10:
        temp := price * 20
        if temp > float64(maxInt) {
            bucket = 10.0
        } else {
            bucket = float64(int(temp)) / 20
        }
    case price <= 20:
        temp := price * 2
        if temp > float64(maxInt) {
            bucket = 20.0
        } else {
            bucket = float64(int(temp)) / 2
        }
    default:
        bucket = 20
    }

    return fmt.Sprintf("%.2f", bucket)
}
```

**File:** `internal/endpoints/auction.go:217`

```go
// Before
ext.ResponseTimeMillis[bidder] = int(latency.Milliseconds())

// After (with overflow check)
latencyMs := latency.Milliseconds()
if latencyMs > int64(^uint(0)>>1) {
    ext.ResponseTimeMillis[bidder] = int(^uint(0) >> 1)
} else {
    ext.ResponseTimeMillis[bidder] = int(latencyMs)
}
```

---

## Testing Checklist

- [ ] Test CircuitBreaker with slow callback (should timeout at 5s)
- [ ] Test EventRecorder shutdown under load
- [ ] Test config validation with invalid inputs
- [ ] Verify no goroutine leaks with `runtime.NumGoroutine()`
- [ ] Load test for extended periods
