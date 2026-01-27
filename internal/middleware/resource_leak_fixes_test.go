package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestPrivacyRequestBodySizeLimit verifies that large request bodies are rejected
func TestPrivacyRequestBodySizeLimit(t *testing.T) {
	config := DefaultPrivacyConfig()
	config.EnforceGDPR = false
	config.EnforceCOPPA = false
	config.EnforceCCPA = false

	middleware := NewPrivacyMiddleware(config)

	// Create a large request body (>1MB)
	largeBody := bytes.Repeat([]byte("x"), 2*1024*1024) // 2MB

	req := httptest.NewRequest("POST", "/openrtb2/auction", bytes.NewReader(largeBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	middleware(next).ServeHTTP(rec, req)

	// Should reject the large request
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected 413 for large request, got %d", rec.Code)
	}

	if nextCalled {
		t.Error("next handler should not have been called for oversized request")
	}
}

// TestPrivacyRequestBodySizeLimitWithinLimit verifies normal requests pass through
func TestPrivacyRequestBodySizeLimitWithinLimit(t *testing.T) {
	config := DefaultPrivacyConfig()
	config.EnforceGDPR = false
	config.EnforceCOPPA = false
	config.EnforceCCPA = false

	middleware := NewPrivacyMiddleware(config)

	// Create a small request body
	smallBody := []byte(`{"id":"test"}`)

	req := httptest.NewRequest("POST", "/openrtb2/auction", bytes.NewReader(smallBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		// Verify we can read the body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed to read body: %v", err)
		}
		if !bytes.Equal(body, smallBody) {
			t.Errorf("body mismatch: got %s, want %s", body, smallBody)
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware(next).ServeHTTP(rec, req)

	if !nextCalled {
		t.Error("next handler should have been called for normal request")
	}
}
