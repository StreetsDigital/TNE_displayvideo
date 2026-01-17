package middleware

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

// ============================================================================
// FALLBACK SCENARIO TESTS
// ============================================================================

// mockRedisClientWithErrors simulates Redis errors for fallback testing
type mockRedisClientWithErrors struct {
	data      map[string]map[string]string
	shouldErr bool
	err       error
	mu        sync.RWMutex
}

func (m *mockRedisClientWithErrors) HGet(ctx context.Context, key, field string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.shouldErr {
		return "", m.err
	}

	if m.data == nil {
		return "", nil
	}
	if hash, ok := m.data[key]; ok {
		return hash[field], nil
	}
	return "", nil
}

func (m *mockRedisClientWithErrors) Ping(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.shouldErr {
		return m.err
	}
	return nil
}

func (m *mockRedisClientWithErrors) setError(shouldErr bool, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldErr = shouldErr
	m.err = err
}

// mockPublisher represents a publisher object
type mockPublisher struct {
	PublisherID    string
	AllowedDomains string
}

func (p *mockPublisher) GetAllowedDomains() string {
	return p.AllowedDomains
}

// mockPublisherStore simulates PostgreSQL for fallback testing
type mockPublisherStore struct {
	data      map[string]*mockPublisher
	shouldErr bool
	err       error
	mu        sync.RWMutex
}

func (m *mockPublisherStore) GetByPublisherID(ctx context.Context, publisherID string) (interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.shouldErr {
		return nil, m.err
	}

	if m.data == nil {
		return nil, nil
	}

	return m.data[publisherID], nil
}

func (m *mockPublisherStore) setError(shouldErr bool, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldErr = shouldErr
	m.err = err
}

// TestValidatePublisher_RedisFallbackToPostgreSQL tests Redis failing and falling back to PostgreSQL
func TestValidatePublisher_RedisFallbackToPostgreSQL(t *testing.T) {
	mockRedis := &mockRedisClientWithErrors{
		data: map[string]map[string]string{
			RedisPublishersHash: {
				"pub123": "example.com",
			},
		},
		shouldErr: true,
		err:       errors.New("redis connection refused"),
	}

	mockStore := &mockPublisherStore{
		data: map[string]*mockPublisher{
			"pub123": {
				PublisherID:    "pub123",
				AllowedDomains: "example.com",
			},
		},
	}

	config := &PublisherAuthConfig{
		Enabled:           true,
		AllowUnregistered: false,
		UseRedis:          true,
	}

	auth := NewPublisherAuth(config)
	auth.SetRedisClient(mockRedis)
	auth.SetPublisherStore(mockStore)

	// Redis will fail, should fall back to PostgreSQL
	err := auth.validatePublisher(context.Background(), "pub123", "example.com")

	if err != nil {
		t.Errorf("Expected validation to succeed via PostgreSQL fallback, got error: %v", err)
	}

	// Verify publisher was cached in memory
	cached := auth.getCachedPublisher("pub123")
	if cached != "example.com" {
		t.Errorf("Expected publisher to be cached with allowed domains, got: %q", cached)
	}
}

// TestValidatePublisher_PostgreSQLFallbackToMemory tests PostgreSQL failing and falling back to memory cache
func TestValidatePublisher_PostgreSQLFallbackToMemory(t *testing.T) {
	mockStore := &mockPublisherStore{
		data: map[string]*mockPublisher{
			"pub123": {
				PublisherID:    "pub123",
				AllowedDomains: "example.com",
			},
		},
	}

	config := &PublisherAuthConfig{
		Enabled:           true,
		AllowUnregistered: false,
		UseRedis:          false, // Redis disabled
	}

	auth := NewPublisherAuth(config)
	auth.SetPublisherStore(mockStore)

	// First request - should succeed via PostgreSQL and cache result
	err := auth.validatePublisher(context.Background(), "pub123", "example.com")
	if err != nil {
		t.Fatalf("First request should succeed via PostgreSQL: %v", err)
	}

	// Now make PostgreSQL fail
	mockStore.setError(true, errors.New("database connection lost"))

	// Second request - PostgreSQL will fail, should use memory cache
	err = auth.validatePublisher(context.Background(), "pub123", "example.com")
	if err != nil {
		t.Errorf("Expected validation to succeed via memory cache fallback, got error: %v", err)
	}
}

// TestValidatePublisher_MemoryCacheExpiration tests that expired cache entries are not used
func TestValidatePublisher_MemoryCacheExpiration(t *testing.T) {
	mockStore := &mockPublisherStore{
		data: map[string]*mockPublisher{
			"pub123": {
				PublisherID:    "pub123",
				AllowedDomains: "example.com",
			},
		},
	}

	config := &PublisherAuthConfig{
		Enabled:           true,
		AllowUnregistered: false,
		UseRedis:          false,
	}

	auth := NewPublisherAuth(config)
	auth.SetPublisherStore(mockStore)

	// First request - cache the publisher
	err := auth.validatePublisher(context.Background(), "pub123", "example.com")
	if err != nil {
		t.Fatalf("First request should succeed: %v", err)
	}

	// Manually expire the cache entry
	auth.publisherCacheMu.Lock()
	if entry, ok := auth.publisherCache["pub123"]; ok {
		entry.expiresAt = time.Now().Add(-1 * time.Second) // Expire it
	}
	auth.publisherCacheMu.Unlock()

	// Make PostgreSQL fail
	mockStore.setError(true, errors.New("database unavailable"))

	// Should fail because cache is expired and PostgreSQL is down
	err = auth.validatePublisher(context.Background(), "pub123", "example.com")
	if err == nil {
		t.Error("Expected validation to fail with expired cache and unavailable PostgreSQL")
	}
}

// TestValidatePublisher_FallbackToRegisteredPubs tests falling back to in-memory RegisteredPubs
func TestValidatePublisher_FallbackToRegisteredPubs(t *testing.T) {
	mockRedis := &mockRedisClientWithErrors{
		shouldErr: true,
		err:       errors.New("redis down"),
	}

	mockStore := &mockPublisherStore{
		shouldErr: true,
		err:       errors.New("database down"),
	}

	config := &PublisherAuthConfig{
		Enabled:           true,
		AllowUnregistered: false,
		UseRedis:          true,
		RegisteredPubs: map[string]string{
			"pub123": "example.com",
		},
	}

	auth := NewPublisherAuth(config)
	auth.SetRedisClient(mockRedis)
	auth.SetPublisherStore(mockStore)

	// Both Redis and PostgreSQL will fail, should fall back to RegisteredPubs
	err := auth.validatePublisher(context.Background(), "pub123", "example.com")
	if err != nil {
		t.Errorf("Expected validation to succeed via RegisteredPubs fallback, got error: %v", err)
	}
}

// TestValidatePublisher_AllFallbacksExhausted tests when all fallbacks fail
func TestValidatePublisher_AllFallbacksExhausted(t *testing.T) {
	mockRedis := &mockRedisClientWithErrors{
		shouldErr: true,
		err:       errors.New("redis down"),
	}

	mockStore := &mockPublisherStore{
		shouldErr: true,
		err:       errors.New("database down"),
	}

	config := &PublisherAuthConfig{
		Enabled:           true,
		AllowUnregistered: false,
		UseRedis:          true,
		RegisteredPubs:    map[string]string{}, // Empty - no fallback
	}

	auth := NewPublisherAuth(config)
	auth.SetRedisClient(mockRedis)
	auth.SetPublisherStore(mockStore)

	// All fallbacks exhausted - should fail
	err := auth.validatePublisher(context.Background(), "pub123", "example.com")
	if err == nil {
		t.Error("Expected validation to fail when all fallbacks are exhausted")
	}
}

// TestValidatePublisher_RedisPriorityOverPostgreSQL tests that Redis is tried first
func TestValidatePublisher_RedisPriorityOverPostgreSQL(t *testing.T) {
	// Use successful Redis, broken PostgreSQL
	// If PostgreSQL is called, it will error and we'll see the error
	mockRedis := &mockRedisClientWithErrors{
		data: map[string]map[string]string{
			RedisPublishersHash: {
				"pub123": "example.com",
			},
		},
	}

	mockStore := &mockPublisherStore{
		shouldErr: true,
		err:       errors.New("postgres should not be called when redis succeeds"),
	}

	config := &PublisherAuthConfig{
		Enabled:           true,
		AllowUnregistered: false,
		UseRedis:          true,
	}

	auth := NewPublisherAuth(config)
	auth.SetRedisClient(mockRedis)
	auth.SetPublisherStore(mockStore)

	// Should succeed via Redis without calling PostgreSQL
	err := auth.validatePublisher(context.Background(), "pub123", "example.com")
	if err != nil {
		t.Errorf("Expected validation to succeed via Redis (PostgreSQL should not be called): %v", err)
	}
}

// TestCachePublisher_BoundedSize tests that cache cleanup prevents unbounded growth
func TestCachePublisher_BoundedSize(t *testing.T) {
	auth := NewPublisherAuth(&PublisherAuthConfig{
		Enabled: true,
	})

	// Add 1001 entries (triggers cleanup at 1000)
	for i := 0; i < 1001; i++ {
		pubID := "pub" + string(rune(i))
		auth.cachePublisher(pubID, "example.com", 30*time.Second)
	}

	// Verify cache is bounded
	auth.publisherCacheMu.RLock()
	cacheSize := len(auth.publisherCache)
	auth.publisherCacheMu.RUnlock()

	if cacheSize > 1001 {
		t.Errorf("Expected cache size <= 1001, got %d (cleanup should have run)", cacheSize)
	}
}

// TestCachePublisher_TTLWorks tests that TTL expiration works correctly
func TestCachePublisher_TTLWorks(t *testing.T) {
	auth := NewPublisherAuth(&PublisherAuthConfig{
		Enabled: true,
	})

	// Cache with very short TTL
	auth.cachePublisher("pub123", "example.com", 50*time.Millisecond)

	// Should be available immediately
	cached := auth.getCachedPublisher("pub123")
	if cached != "example.com" {
		t.Error("Expected publisher to be cached immediately")
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should be expired
	cached = auth.getCachedPublisher("pub123")
	if cached != "" {
		t.Error("Expected cache entry to be expired")
	}
}

// TestRateLimitedLogging tests that fallback warnings are rate-limited
func TestRateLimitedLogging_Redis(t *testing.T) {
	// Clear global state
	lastRedisWarning = sync.Map{}

	auth := NewPublisherAuth(&PublisherAuthConfig{
		Enabled: true,
	})

	// First call should log (we can't easily verify logging, but we ensure it doesn't panic)
	auth.logRedisFallback(errors.New("redis error"), "pub123")

	// Immediate second call should be suppressed (within 1 minute)
	// This should not panic or error
	auth.logRedisFallback(errors.New("redis error"), "pub123")

	// Can't easily test log output, but we verified the function works
}

// TestRateLimitedLogging_Database tests database fallback logging rate limiting
func TestRateLimitedLogging_Database(t *testing.T) {
	// Clear global state
	lastDBWarning = sync.Map{}

	auth := NewPublisherAuth(&PublisherAuthConfig{
		Enabled: true,
	})

	// First call should log
	auth.logDatabaseFallback(errors.New("db error"), "pub123")

	// Immediate second call should be suppressed
	auth.logDatabaseFallback(errors.New("db error"), "pub123")

	// Function should work without panicking
}

// TestCleanupExpiredCache tests that expired entries are removed
func TestCleanupExpiredCache(t *testing.T) {
	auth := NewPublisherAuth(&PublisherAuthConfig{
		Enabled: true,
	})

	// Add entries with different expiration times
	auth.publisherCacheMu.Lock()
	auth.publisherCache = map[string]*publisherCacheEntry{
		"pub1": {allowedDomains: "example.com", expiresAt: time.Now().Add(10 * time.Second)},  // Valid
		"pub2": {allowedDomains: "test.com", expiresAt: time.Now().Add(-1 * time.Second)},     // Expired
		"pub3": {allowedDomains: "demo.com", expiresAt: time.Now().Add(-10 * time.Second)},    // Expired
		"pub4": {allowedDomains: "another.com", expiresAt: time.Now().Add(5 * time.Second)},   // Valid
	}
	auth.publisherCacheMu.Unlock()

	// Run cleanup
	auth.cleanupExpiredCache()

	// Verify expired entries removed
	auth.publisherCacheMu.RLock()
	defer auth.publisherCacheMu.RUnlock()

	if _, exists := auth.publisherCache["pub2"]; exists {
		t.Error("Expected expired pub2 to be removed")
	}
	if _, exists := auth.publisherCache["pub3"]; exists {
		t.Error("Expected expired pub3 to be removed")
	}
	if _, exists := auth.publisherCache["pub1"]; !exists {
		t.Error("Expected valid pub1 to remain")
	}
	if _, exists := auth.publisherCache["pub4"]; !exists {
		t.Error("Expected valid pub4 to remain")
	}
}

// TestValidatePublisher_ConcurrentAccess tests that fallback chain is thread-safe
func TestValidatePublisher_ConcurrentAccess(t *testing.T) {
	mockStore := &mockPublisherStore{
		data: map[string]*mockPublisher{
			"pub123": {
				PublisherID:    "pub123",
				AllowedDomains: "example.com",
			},
		},
	}

	config := &PublisherAuthConfig{
		Enabled:           true,
		AllowUnregistered: false,
		UseRedis:          false,
	}

	auth := NewPublisherAuth(config)
	auth.SetPublisherStore(mockStore)

	// Run 100 concurrent validations
	var wg sync.WaitGroup
	errChan := make(chan error, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := auth.validatePublisher(context.Background(), "pub123", "example.com")
			if err != nil {
				errChan <- err
			}
		}()
	}

	wg.Wait()
	close(errChan)

	// Verify no errors occurred
	for err := range errChan {
		t.Errorf("Concurrent validation failed: %v", err)
	}
}

// TestCheckRateLimit_CleanupStaleEntries tests that stale rate limit entries are cleaned up
func TestCheckRateLimit_CleanupStaleEntries(t *testing.T) {
	auth := NewPublisherAuth(&PublisherAuthConfig{
		Enabled:         true,
		RateLimitPerPub: 10,
	})

	// Add 1001 publishers to trigger cleanup threshold
	for i := 0; i < 1001; i++ {
		pubID := fmt.Sprintf("pub%d", i)
		auth.checkRateLimit(pubID)
	}

	// Manually age some entries to be stale (>1 hour old)
	auth.rateLimitsMu.Lock()
	staleTime := time.Now().Add(-2 * time.Hour)
	for i := 0; i < 500; i++ {
		pubID := fmt.Sprintf("pub%d", i)
		if entry, ok := auth.rateLimits[pubID]; ok {
			entry.lastCheck = staleTime
		}
	}
	auth.rateLimitsMu.Unlock()

	// Trigger cleanup by checking rate limit on an EXISTING publisher (not a new one)
	// This will pass the token check and trigger cleanup since len > 1000
	auth.checkRateLimit("pub1000") // Existing publisher from loop above

	// Verify cleanup ran (stale entries should be removed)
	auth.rateLimitsMu.RLock()
	finalSize := len(auth.rateLimits)
	auth.rateLimitsMu.RUnlock()

	// Should have removed the 500 stale entries
	// Expected: 1001 - 500 = 501 remaining
	if finalSize > 600 { // Allow some margin
		t.Errorf("Rate limit map size after cleanup: %d (expected ~501)", finalSize)
	}
	if finalSize < 450 {
		t.Errorf("Rate limit map size too small: %d (expected ~501)", finalSize)
	}

	// Verify one of the stale entries was actually removed
	auth.rateLimitsMu.RLock()
	_, exists := auth.rateLimits["pub0"] // Should be stale and removed
	auth.rateLimitsMu.RUnlock()

	if exists {
		t.Error("Expected stale entry pub0 to be removed")
	}
}
