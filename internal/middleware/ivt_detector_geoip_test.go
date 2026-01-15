package middleware

import (
	"net/http/httptest"
	"testing"
)

// MockGeoIP implements GeoIPLookup for testing
type MockGeoIP struct {
	countryMap map[string]string // IP -> Country code
	err        error             // Error to return
}

// NewMockGeoIP creates a mock GeoIP lookup
func NewMockGeoIP() *MockGeoIP {
	return &MockGeoIP{
		countryMap: make(map[string]string),
	}
}

// LookupCountry returns the mocked country for an IP
func (m *MockGeoIP) LookupCountry(ip string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.countryMap[ip], nil
}

// Close is a no-op for the mock
func (m *MockGeoIP) Close() error {
	return nil
}

// SetCountry sets the country for an IP (test helper)
func (m *MockGeoIP) SetCountry(ip, country string) {
	m.countryMap[ip] = country
}

// SetError sets an error to return (test helper)
func (m *MockGeoIP) SetError(err error) {
	m.err = err
}

func TestMaxMindGeoIP_NewMaxMindGeoIP_EmptyPath(t *testing.T) {
	geoip, err := NewMaxMindGeoIP("")
	if err != nil {
		t.Errorf("Expected no error for empty path, got %v", err)
	}
	if geoip != nil {
		t.Error("Expected nil GeoIP for empty path")
	}
}

func TestMaxMindGeoIP_NewMaxMindGeoIP_InvalidPath(t *testing.T) {
	geoip, err := NewMaxMindGeoIP("/nonexistent/path/database.mmdb")
	if err == nil {
		t.Error("Expected error for invalid path")
	}
	if geoip != nil {
		t.Error("Expected nil GeoIP for invalid path")
	}
}

func TestMaxMindGeoIP_LookupCountry_NilReader(t *testing.T) {
	geoip := &MaxMindGeoIP{reader: nil}
	country, err := geoip.LookupCountry("8.8.8.8")
	if err != nil {
		t.Errorf("Expected no error for nil reader, got %v", err)
	}
	if country != "" {
		t.Errorf("Expected empty country for nil reader, got %s", country)
	}
}

func TestMaxMindGeoIP_LookupCountry_InvalidIP(t *testing.T) {
	mock := NewMockGeoIP()
	country, err := mock.LookupCountry("invalid-ip")
	if err != nil {
		t.Errorf("Expected no error for invalid IP, got %v", err)
	}
	if country != "" {
		t.Errorf("Expected empty country for invalid IP, got %s", country)
	}
}

func TestMaxMindGeoIP_Close_NilReader(t *testing.T) {
	geoip := &MaxMindGeoIP{reader: nil}
	err := geoip.Close()
	if err != nil {
		t.Errorf("Expected no error closing nil reader, got %v", err)
	}
}

func TestCheckGeoWithConfig_Disabled(t *testing.T) {
	config := &IVTConfig{
		CheckGeo: false,
	}

	mock := NewMockGeoIP()
	mock.SetCountry("1.2.3.4", "CN")

	detector := &IVTDetector{
		config:  config,
		geoip:   mock,
		metrics: &IVTMetrics{},
	}

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	result := &IVTResult{}

	detector.checkGeoWithConfig(req, result, config)

	if len(result.Signals) != 0 {
		t.Error("Expected no signals when geo check is disabled")
	}
}

func TestCheckGeoWithConfig_NoGeoIP(t *testing.T) {
	config := &IVTConfig{
		CheckGeo: true,
	}

	detector := &IVTDetector{
		config:  config,
		geoip:   nil, // No GeoIP available
		metrics: &IVTMetrics{},
	}

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	result := &IVTResult{}

	detector.checkGeoWithConfig(req, result, config)

	if len(result.Signals) != 0 {
		t.Error("Expected no signals when GeoIP is not available")
	}
}

func TestCheckGeoWithConfig_NoIP(t *testing.T) {
	config := &IVTConfig{
		CheckGeo: true,
	}

	mock := NewMockGeoIP()
	detector := &IVTDetector{
		config:  config,
		geoip:   mock,
		metrics: &IVTMetrics{},
	}

	req := httptest.NewRequest("GET", "/test", nil)
	// No X-Forwarded-For or RemoteAddr
	result := &IVTResult{}

	detector.checkGeoWithConfig(req, result, config)

	if len(result.Signals) != 0 {
		t.Error("Expected no signals when no IP available")
	}
}

func TestCheckGeoWithConfig_NoCountryFound(t *testing.T) {
	config := &IVTConfig{
		CheckGeo: true,
	}

	mock := NewMockGeoIP()
	// Don't set any country for the IP

	detector := &IVTDetector{
		config:  config,
		geoip:   mock,
		metrics: &IVTMetrics{},
	}

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	result := &IVTResult{}

	detector.checkGeoWithConfig(req, result, config)

	if len(result.Signals) != 0 {
		t.Error("Expected no signals when no country found")
	}
}

func TestCheckGeoWithConfig_AllowedCountries_Pass(t *testing.T) {
	config := &IVTConfig{
		CheckGeo:         true,
		AllowedCountries: []string{"US", "GB", "CA"},
	}

	mock := NewMockGeoIP()
	mock.SetCountry("1.2.3.4", "US")

	detector := &IVTDetector{
		config:  config,
		geoip:   mock,
		metrics: &IVTMetrics{},
	}

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	result := &IVTResult{}

	detector.checkGeoWithConfig(req, result, config)

	if len(result.Signals) != 0 {
		t.Error("Expected no signals for allowed country")
	}
}

func TestCheckGeoWithConfig_AllowedCountries_Block(t *testing.T) {
	config := &IVTConfig{
		CheckGeo:         true,
		AllowedCountries: []string{"US", "GB", "CA"},
	}

	mock := NewMockGeoIP()
	mock.SetCountry("1.2.3.4", "CN")

	detector := &IVTDetector{
		config:  config,
		geoip:   mock,
		metrics: &IVTMetrics{},
	}

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	result := &IVTResult{}

	detector.checkGeoWithConfig(req, result, config)

	if len(result.Signals) != 1 {
		t.Fatalf("Expected 1 signal for disallowed country, got %d", len(result.Signals))
	}

	signal := result.Signals[0]
	if signal.Type != "geo_restricted" {
		t.Errorf("Expected signal type 'geo_restricted', got '%s'", signal.Type)
	}
	if signal.Severity != "high" {
		t.Errorf("Expected severity 'high', got '%s'", signal.Severity)
	}
	if signal.Description != "country CN not in allowed list" {
		t.Errorf("Unexpected description: %s", signal.Description)
	}

	// Check metrics
	if detector.metrics.GeoMismatches != 1 {
		t.Errorf("Expected GeoMismatches=1, got %d", detector.metrics.GeoMismatches)
	}
}

func TestCheckGeoWithConfig_BlockedCountries_Pass(t *testing.T) {
	config := &IVTConfig{
		CheckGeo:         true,
		BlockedCountries: []string{"CN", "RU", "KP"},
	}

	mock := NewMockGeoIP()
	mock.SetCountry("1.2.3.4", "US")

	detector := &IVTDetector{
		config:  config,
		geoip:   mock,
		metrics: &IVTMetrics{},
	}

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	result := &IVTResult{}

	detector.checkGeoWithConfig(req, result, config)

	if len(result.Signals) != 0 {
		t.Error("Expected no signals for non-blocked country")
	}
}

func TestCheckGeoWithConfig_BlockedCountries_Block(t *testing.T) {
	config := &IVTConfig{
		CheckGeo:         true,
		BlockedCountries: []string{"CN", "RU", "KP"},
	}

	mock := NewMockGeoIP()
	mock.SetCountry("1.2.3.4", "CN")

	detector := &IVTDetector{
		config:  config,
		geoip:   mock,
		metrics: &IVTMetrics{},
	}

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	result := &IVTResult{}

	detector.checkGeoWithConfig(req, result, config)

	if len(result.Signals) != 1 {
		t.Fatalf("Expected 1 signal for blocked country, got %d", len(result.Signals))
	}

	signal := result.Signals[0]
	if signal.Type != "geo_blocked" {
		t.Errorf("Expected signal type 'geo_blocked', got '%s'", signal.Type)
	}
	if signal.Severity != "high" {
		t.Errorf("Expected severity 'high', got '%s'", signal.Severity)
	}
	if signal.Description != "country CN is blocked" {
		t.Errorf("Unexpected description: %s", signal.Description)
	}

	// Check metrics
	if detector.metrics.GeoMismatches != 1 {
		t.Errorf("Expected GeoMismatches=1, got %d", detector.metrics.GeoMismatches)
	}
}

func TestCheckGeoWithConfig_AllowedTakesPrecedence(t *testing.T) {
	config := &IVTConfig{
		CheckGeo:         true,
		AllowedCountries: []string{"US", "GB"},
		BlockedCountries: []string{"CN", "RU"},
	}

	mock := NewMockGeoIP()
	mock.SetCountry("1.2.3.4", "CN")

	detector := &IVTDetector{
		config:  config,
		geoip:   mock,
		metrics: &IVTMetrics{},
	}

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	result := &IVTResult{}

	detector.checkGeoWithConfig(req, result, config)

	// Should trigger geo_restricted, not geo_blocked
	if len(result.Signals) != 1 {
		t.Fatalf("Expected 1 signal, got %d", len(result.Signals))
	}

	signal := result.Signals[0]
	if signal.Type != "geo_restricted" {
		t.Errorf("Expected 'geo_restricted' when allowed list is present, got '%s'", signal.Type)
	}
}

func TestCheckGeoWithConfig_MultipleCountries(t *testing.T) {
	testCases := []struct {
		name         string
		allowList    []string
		blockList    []string
		countryCode  string
		expectSignal bool
		signalType   string
	}{
		{
			name:         "Allowed US",
			allowList:    []string{"US", "GB", "CA"},
			blockList:    []string{},
			countryCode:  "US",
			expectSignal: false,
		},
		{
			name:         "Allowed GB",
			allowList:    []string{"US", "GB", "CA"},
			blockList:    []string{},
			countryCode:  "GB",
			expectSignal: false,
		},
		{
			name:         "Disallowed CN",
			allowList:    []string{"US", "GB", "CA"},
			blockList:    []string{},
			countryCode:  "CN",
			expectSignal: true,
			signalType:   "geo_restricted",
		},
		{
			name:         "Blocked RU",
			allowList:    []string{},
			blockList:    []string{"CN", "RU", "KP"},
			countryCode:  "RU",
			expectSignal: true,
			signalType:   "geo_blocked",
		},
		{
			name:         "Non-blocked DE",
			allowList:    []string{},
			blockList:    []string{"CN", "RU", "KP"},
			countryCode:  "DE",
			expectSignal: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &IVTConfig{
				CheckGeo:         true,
				AllowedCountries: tc.allowList,
				BlockedCountries: tc.blockList,
			}

			mock := NewMockGeoIP()
			mock.SetCountry("1.2.3.4", tc.countryCode)

			detector := &IVTDetector{
				config:  config,
				geoip:   mock,
				metrics: &IVTMetrics{},
			}

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("X-Forwarded-For", "1.2.3.4")
			result := &IVTResult{}

			detector.checkGeoWithConfig(req, result, config)

			if tc.expectSignal {
				if len(result.Signals) != 1 {
					t.Fatalf("Expected 1 signal, got %d", len(result.Signals))
				}
				if result.Signals[0].Type != tc.signalType {
					t.Errorf("Expected signal type '%s', got '%s'", tc.signalType, result.Signals[0].Type)
				}
			} else {
				if len(result.Signals) != 0 {
					t.Errorf("Expected no signals, got %d", len(result.Signals))
				}
			}
		})
	}
}

func TestContains(t *testing.T) {
	testCases := []struct {
		name     string
		slice    []string
		value    string
		expected bool
	}{
		{"Found first", []string{"US", "GB", "CA"}, "US", true},
		{"Found middle", []string{"US", "GB", "CA"}, "GB", true},
		{"Found last", []string{"US", "GB", "CA"}, "CA", true},
		{"Not found", []string{"US", "GB", "CA"}, "CN", false},
		{"Empty slice", []string{}, "US", false},
		{"Empty value", []string{"US", "GB"}, "", false},
		{"Case sensitive", []string{"US", "GB"}, "us", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := contains(tc.slice, tc.value)
			if result != tc.expected {
				t.Errorf("contains(%v, %q) = %v, expected %v", tc.slice, tc.value, result, tc.expected)
			}
		})
	}
}

func TestIVTDetector_Close(t *testing.T) {
	t.Run("WithGeoIP", func(t *testing.T) {
		mock := NewMockGeoIP()
		detector := &IVTDetector{
			geoip: mock,
		}

		err := detector.Close()
		if err != nil {
			t.Errorf("Expected no error closing detector, got %v", err)
		}
	})

	t.Run("WithoutGeoIP", func(t *testing.T) {
		detector := &IVTDetector{
			geoip: nil,
		}

		err := detector.Close()
		if err != nil {
			t.Errorf("Expected no error closing detector without GeoIP, got %v", err)
		}
	})
}

func TestNewIVTDetector_WithGeoIPPath(t *testing.T) {
	// Test with invalid path (should fail gracefully)
	config := &IVTConfig{
		GeoIPDBPath: "/nonexistent/path/database.mmdb",
	}

	detector := NewIVTDetector(config)
	if detector == nil {
		t.Fatal("Expected detector to be created even with invalid GeoIP path")
	}

	// GeoIP should be nil since the path is invalid
	if detector.geoip != nil {
		t.Error("Expected GeoIP to be nil for invalid path")
	}

	// Cleanup
	if err := detector.Close(); err != nil {
		t.Errorf("Error closing detector: %v", err)
	}
}

func TestNewIVTDetector_WithoutGeoIPPath(t *testing.T) {
	config := &IVTConfig{
		GeoIPDBPath: "", // No path
	}

	detector := NewIVTDetector(config)
	if detector == nil {
		t.Fatal("Expected detector to be created")
	}

	if detector.geoip != nil {
		t.Error("Expected GeoIP to be nil when no path provided")
	}

	// Cleanup
	if err := detector.Close(); err != nil {
		t.Errorf("Error closing detector: %v", err)
	}
}

func TestCheckGeoWithConfig_XForwardedFor(t *testing.T) {
	config := &IVTConfig{
		CheckGeo:         true,
		BlockedCountries: []string{"CN"},
	}

	mock := NewMockGeoIP()
	mock.SetCountry("203.0.113.1", "CN")

	detector := &IVTDetector{
		config:  config,
		geoip:   mock,
		metrics: &IVTMetrics{},
	}

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.1")
	result := &IVTResult{}

	detector.checkGeoWithConfig(req, result, config)

	// Should use first IP from X-Forwarded-For
	if len(result.Signals) != 1 {
		t.Fatalf("Expected 1 signal for blocked country, got %d", len(result.Signals))
	}

	if result.Signals[0].Type != "geo_blocked" {
		t.Errorf("Expected geo_blocked signal, got %s", result.Signals[0].Type)
	}
}
