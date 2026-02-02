package endpoints

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockRedisClient implements a minimal in-memory store for testing
type mockRedisClient struct {
	store map[string]string
}

func (m *mockRedisClient) get(key string) string {
	return m.store[key]
}

func TestCacheHandler_Put_ValidRequest(t *testing.T) {
	handler := &CacheHandler{
		redisClient: nil, // We'll test the HTTP layer; Redis is tested separately
		cacheTTL:    defaultCacheTTL,
	}

	// We can't easily test with real Redis in a unit test, so we test
	// the request parsing and validation logic via the HTTP handler.
	// Full integration with Redis would be an integration test.

	// Test: empty puts array returns 400
	body := `{"puts":[]}`
	req := httptest.NewRequest(http.MethodPost, "/cache", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	handler.handlePut(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for empty puts, got %d", w.Code)
	}
}

func TestCacheHandler_Put_InvalidJSON(t *testing.T) {
	handler := &CacheHandler{
		redisClient: nil,
		cacheTTL:    defaultCacheTTL,
	}

	req := httptest.NewRequest(http.MethodPost, "/cache", bytes.NewBufferString("not json"))
	w := httptest.NewRecorder()
	handler.handlePut(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for invalid JSON, got %d", w.Code)
	}
}

func TestCacheHandler_Put_TooManyItems(t *testing.T) {
	handler := &CacheHandler{
		redisClient: nil,
		cacheTTL:    defaultCacheTTL,
	}

	// Create request with 101 items (over the max of 100)
	puts := make([]cachePutItem, 101)
	for i := range puts {
		puts[i] = cachePutItem{Type: "json", Value: json.RawMessage(`"test"`)}
	}
	body, _ := json.Marshal(cachePutRequest{Puts: puts})

	req := httptest.NewRequest(http.MethodPost, "/cache", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	handler.handlePut(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for too many items, got %d", w.Code)
	}
}

func TestCacheHandler_Get_MissingUUID(t *testing.T) {
	handler := &CacheHandler{
		redisClient: nil,
		cacheTTL:    defaultCacheTTL,
	}

	req := httptest.NewRequest(http.MethodGet, "/cache", nil)
	w := httptest.NewRecorder()
	handler.handleGet(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for missing uuid, got %d", w.Code)
	}
}

func TestCacheHandler_Get_InvalidUUID(t *testing.T) {
	handler := &CacheHandler{
		redisClient: nil,
		cacheTTL:    defaultCacheTTL,
	}

	// Test with non-hex characters
	req := httptest.NewRequest(http.MethodGet, "/cache?uuid=not-hex-!!!", nil)
	w := httptest.NewRecorder()
	handler.handleGet(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for invalid uuid, got %d", w.Code)
	}
}

func TestCacheHandler_Get_TooLongUUID(t *testing.T) {
	handler := &CacheHandler{
		redisClient: nil,
		cacheTTL:    defaultCacheTTL,
	}

	// UUID longer than 64 characters
	longUUID := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	req := httptest.NewRequest(http.MethodGet, "/cache?uuid="+longUUID, nil)
	w := httptest.NewRecorder()
	handler.handleGet(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for too-long uuid, got %d", w.Code)
	}
}

func TestCacheHandler_MethodNotAllowed(t *testing.T) {
	handler := &CacheHandler{
		redisClient: nil,
		cacheTTL:    defaultCacheTTL,
	}

	req := httptest.NewRequest(http.MethodDelete, "/cache", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected 405 for DELETE method, got %d", w.Code)
	}
}

func TestCacheHandler_Routes(t *testing.T) {
	handler := &CacheHandler{
		redisClient: nil,
		cacheTTL:    defaultCacheTTL,
	}

	// POST should go to handlePut (will fail on empty body but not 405)
	req := httptest.NewRequest(http.MethodPost, "/cache", bytes.NewBufferString("{}"))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code == http.StatusMethodNotAllowed {
		t.Error("POST should not return 405")
	}

	// GET should go to handleGet (will fail on missing uuid but not 405)
	req = httptest.NewRequest(http.MethodGet, "/cache", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code == http.StatusMethodNotAllowed {
		t.Error("GET should not return 405")
	}
}

func TestGenerateUUID(t *testing.T) {
	uuid1, err := generateUUID()
	if err != nil {
		t.Fatalf("generateUUID() failed: %v", err)
	}
	if len(uuid1) != 32 {
		t.Errorf("Expected UUID length 32, got %d", len(uuid1))
	}

	uuid2, err := generateUUID()
	if err != nil {
		t.Fatalf("generateUUID() failed: %v", err)
	}

	if uuid1 == uuid2 {
		t.Error("Two generated UUIDs should not be identical")
	}
}

func TestIsHexString(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"abcdef0123456789", true},
		{"ABCDEF", true},
		{"abc123", true},
		{"", true}, // empty string is trivially hex
		{"xyz", false},
		{"abc-def", false},
		{"abc def", false},
		{"abc!@#", false},
	}

	for _, tt := range tests {
		result := isHexString(tt.input)
		if result != tt.expected {
			t.Errorf("isHexString(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}
