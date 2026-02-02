package currency

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestConverter_GetRate(t *testing.T) {
	mockRates := &CurrencyFile{
		GeneratedAt: time.Now().Format(time.RFC3339),
		DataAsOf:    time.Now().Format(time.RFC3339),
		Conversions: map[string]map[string]float64{
			"USD": {
				"EUR": 0.85,
				"GBP": 0.75,
				"JPY": 150.0,
			},
			"EUR": {
				"USD": 1.18,
				"GBP": 0.88,
			},
		},
	}

	converter := &Converter{
		rates: mockRates,
		lastFetch: time.Now(),
		staleThreshold: 24 * time.Hour,
	}

	tests := []struct {
		name    string
		from    string
		to      string
		want    float64
		wantErr bool
	}{
		{
			name: "USD to EUR",
			from: "USD",
			to:   "EUR",
			want: 0.85,
		},
		{
			name: "EUR to USD",
			from: "EUR",
			to:   "USD",
			want: 1.18,
		},
		{
			name: "same currency",
			from: "USD",
			to:   "USD",
			want: 1.0,
		},
		{
			name:    "unavailable conversion",
			from:    "USD",
			to:      "CNY",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := converter.GetRate(tt.from, tt.to)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("GetRate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConverter_Convert(t *testing.T) {
	mockRates := &CurrencyFile{
		Conversions: map[string]map[string]float64{
			"USD": {"EUR": 0.85},
		},
	}

	converter := &Converter{
		rates:          mockRates,
		lastFetch:      time.Now(),
		staleThreshold: 24 * time.Hour,
	}

	result, err := converter.Convert(100.0, "USD", "EUR")
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	expected := 85.0
	if result != expected {
		t.Errorf("Convert() = %v, want %v", result, expected)
	}
}

func TestConverter_FetchRates(t *testing.T) {
	mockResponse := CurrencyFile{
		GeneratedAt: "2026-02-02T12:00:00Z",
		DataAsOf:    "2026-02-02T00:00:00Z",
		Conversions: map[string]map[string]float64{
			"USD": {"EUR": 0.85},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	config := &Config{
		FetchURL:        server.URL,
		RefreshInterval: time.Hour,
		StaleThreshold:  24 * time.Hour,
		FetchTimeout:    5 * time.Second,
	}

	converter := NewConverter(config)
	ctx := context.Background()

	err := converter.fetchRates(ctx)
	if err != nil {
		t.Fatalf("fetchRates() error = %v", err)
	}

	if converter.rates == nil {
		t.Fatal("rates should not be nil after fetch")
	}

	if len(converter.rates.Conversions) == 0 {
		t.Fatal("conversions should not be empty")
	}
}

func TestAggregateConversions(t *testing.T) {
	// External rates
	externalRates := &Converter{
		rates: &CurrencyFile{
			Conversions: map[string]map[string]float64{
				"USD": {"EUR": 0.85},
				"EUR": {"USD": 1.18},
			},
		},
		lastFetch:      time.Now(),
		staleThreshold: 24 * time.Hour,
	}

	// Custom rates (publisher override)
	customRates := map[string]map[string]float64{
		"USD": {"EUR": 0.90}, // Publisher wants better EUR rate
	}

	t.Run("custom priority (default)", func(t *testing.T) {
		agg := NewAggregateConversions(customRates, externalRates, false)

		// Should use custom rate
		rate, err := agg.GetRate("USD", "EUR")
		if err != nil {
			t.Fatalf("GetRate() error = %v", err)
		}
		if rate != 0.90 {
			t.Errorf("expected custom rate 0.90, got %v", rate)
		}

		// Should fall back to external for EUR->USD
		rate, err = agg.GetRate("EUR", "USD")
		if err != nil {
			t.Fatalf("GetRate() error = %v", err)
		}
		if rate != 1.18 {
			t.Errorf("expected external rate 1.18, got %v", rate)
		}
	})

	t.Run("external priority", func(t *testing.T) {
		agg := NewAggregateConversions(customRates, externalRates, true)

		// Should use external rate
		rate, err := agg.GetRate("USD", "EUR")
		if err != nil {
			t.Fatalf("GetRate() error = %v", err)
		}
		if rate != 0.85 {
			t.Errorf("expected external rate 0.85, got %v", rate)
		}
	})
}

func TestConverter_Stats(t *testing.T) {
	converter := &Converter{
		rates: &CurrencyFile{
			GeneratedAt: "2026-02-02T12:00:00Z",
			DataAsOf:    "2026-02-02T00:00:00Z",
			Conversions: map[string]map[string]float64{
				"USD": {"EUR": 0.85},
				"EUR": {"USD": 1.18},
			},
		},
		lastFetch:      time.Now().Add(-1 * time.Hour),
		running:        true,
		fetchErrors:    0,
		staleThreshold: 24 * time.Hour,
	}

	stats := converter.Stats()

	if stats["running"] != true {
		t.Error("expected running to be true")
	}

	if stats["ratesLoaded"] != true {
		t.Error("expected ratesLoaded to be true")
	}

	if stats["currencies"] != 2 {
		t.Errorf("expected 2 currencies, got %v", stats["currencies"])
	}
}
