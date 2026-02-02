// Package currency provides multi-currency conversion for OpenRTB auctions
// using Prebid's currency-file exchange rates
package currency

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/thenexusengine/tne_springwire/pkg/logger"
)

// CurrencyFile represents the Prebid currency file structure
type CurrencyFile struct {
	GeneratedAt string                       `json:"generatedAt"`
	DataAsOf    string                       `json:"dataAsOf"`
	Conversions map[string]map[string]float64 `json:"conversions"`
}

// Converter handles currency conversions using external and custom rates
type Converter struct {
	// Configuration
	fetchURL        string
	refreshInterval time.Duration
	staleThreshold  time.Duration
	httpClient      *http.Client

	// State
	mu           sync.RWMutex
	rates        *CurrencyFile
	lastFetch    time.Time
	fetchErrors  int
	running      bool
	stopChan     chan struct{}
}

// Config for currency converter
type Config struct {
	// URL to fetch currency rates from
	// Default: https://cdn.jsdelivr.net/gh/prebid/currency-file@1/latest.json
	FetchURL string

	// How often to refresh rates
	// Default: 30 minutes
	RefreshInterval time.Duration

	// How old rates can be before considered stale
	// Default: 24 hours
	StaleThreshold time.Duration

	// HTTP client timeout
	// Default: 10 seconds
	FetchTimeout time.Duration
}

// DefaultConfig returns recommended configuration
func DefaultConfig() *Config {
	return &Config{
		FetchURL:        "https://cdn.jsdelivr.net/gh/prebid/currency-file@1/latest.json",
		RefreshInterval: 30 * time.Minute,
		StaleThreshold:  24 * time.Hour,
		FetchTimeout:    10 * time.Second,
	}
}

// NewConverter creates a new currency converter
func NewConverter(config *Config) *Converter {
	if config == nil {
		config = DefaultConfig()
	}

	return &Converter{
		fetchURL:        config.FetchURL,
		refreshInterval: config.RefreshInterval,
		staleThreshold:  config.StaleThreshold,
		httpClient: &http.Client{
			Timeout: config.FetchTimeout,
		},
		stopChan: make(chan struct{}),
	}
}

// Start begins background rate updates
func (c *Converter) Start(ctx context.Context) error {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return fmt.Errorf("converter already running")
	}
	c.running = true
	c.mu.Unlock()

	// Initial fetch
	if err := c.fetchRates(ctx); err != nil {
		logger.Log.Warn().Err(err).Msg("initial currency fetch failed")
	}

	// Background refresh
	go c.refreshLoop(ctx)

	return nil
}

// Stop halts background updates
func (c *Converter) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		close(c.stopChan)
		c.running = false
	}
}

// refreshLoop periodically fetches new rates
func (c *Converter) refreshLoop(ctx context.Context) {
	ticker := time.NewTicker(c.refreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := c.fetchRates(ctx); err != nil {
				logger.Log.Warn().
					Err(err).
					Int("fetchErrors", c.fetchErrors).
					Msg("currency rate refresh failed")
			}
		case <-c.stopChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

// fetchRates downloads and parses the currency file
func (c *Converter) fetchRates(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.fetchURL, nil)
	if err != nil {
		c.incrementErrors()
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.incrementErrors()
		return fmt.Errorf("fetch rates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.incrementErrors()
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.incrementErrors()
		return fmt.Errorf("read body: %w", err)
	}

	var currencyFile CurrencyFile
	if err := json.Unmarshal(body, &currencyFile); err != nil {
		c.incrementErrors()
		return fmt.Errorf("parse JSON: %w", err)
	}

	// Validate rates
	if len(currencyFile.Conversions) == 0 {
		c.incrementErrors()
		return fmt.Errorf("no conversions in currency file")
	}

	c.mu.Lock()
	c.rates = &currencyFile
	c.lastFetch = time.Now()
	c.fetchErrors = 0
	c.mu.Unlock()

	logger.Log.Info().
		Int("currencies", len(currencyFile.Conversions)).
		Str("dataAsOf", currencyFile.DataAsOf).
		Msg("currency rates updated")

	return nil
}

// incrementErrors safely increments error counter
func (c *Converter) incrementErrors() {
	c.mu.Lock()
	c.fetchErrors++
	c.mu.Unlock()
}

// GetRate returns the conversion rate from one currency to another
// Returns 1.0 if currencies are the same
// Returns error if conversion not available
func (c *Converter) GetRate(from, to string) (float64, error) {
	// Same currency = 1.0 rate
	if from == to {
		return 1.0, nil
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.rates == nil {
		return 0, fmt.Errorf("no currency rates available")
	}

	// Check if rates are stale
	if time.Since(c.lastFetch) > c.staleThreshold {
		logger.Log.Warn().
			Dur("age", time.Since(c.lastFetch)).
			Dur("threshold", c.staleThreshold).
			Msg("currency rates are stale")
	}

	// Look up conversion rate
	if fromRates, ok := c.rates.Conversions[from]; ok {
		if rate, ok := fromRates[to]; ok {
			return rate, nil
		}
	}

	return 0, fmt.Errorf("no conversion available from %s to %s", from, to)
}

// Convert converts an amount from one currency to another
func (c *Converter) Convert(amount float64, from, to string) (float64, error) {
	rate, err := c.GetRate(from, to)
	if err != nil {
		return 0, err
	}
	return amount * rate, nil
}

// GetRates returns a copy of all current rates (for diagnostics)
func (c *Converter) GetRates() map[string]map[string]float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.rates == nil {
		return nil
	}

	// Deep copy
	rates := make(map[string]map[string]float64, len(c.rates.Conversions))
	for from, toRates := range c.rates.Conversions {
		rates[from] = make(map[string]float64, len(toRates))
		for to, rate := range toRates {
			rates[from][to] = rate
		}
	}

	return rates
}

// Stats returns statistics about the converter
func (c *Converter) Stats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := map[string]interface{}{
		"running":      c.running,
		"fetchErrors":  c.fetchErrors,
		"ratesLoaded":  c.rates != nil,
	}

	if c.rates != nil {
		stats["currencies"] = len(c.rates.Conversions)
		stats["dataAsOf"] = c.rates.DataAsOf
		stats["generatedAt"] = c.rates.GeneratedAt
	}

	if !c.lastFetch.IsZero() {
		stats["lastFetch"] = c.lastFetch
		stats["age"] = time.Since(c.lastFetch).String()
		stats["stale"] = time.Since(c.lastFetch) > c.staleThreshold
	}

	return stats
}

// AggregateConversions combines custom and external rates
// Custom rates take priority over external rates
type AggregateConversions struct {
	customRates   map[string]map[string]float64
	externalRates *Converter
	useExternal   bool // If true, prefer external over custom
}

// NewAggregateConversions creates a combined rate source
func NewAggregateConversions(customRates map[string]map[string]float64, externalRates *Converter, useExternal bool) *AggregateConversions {
	return &AggregateConversions{
		customRates:   customRates,
		externalRates: externalRates,
		useExternal:   useExternal,
	}
}

// GetRate returns conversion rate, checking custom then external
func (a *AggregateConversions) GetRate(from, to string) (float64, error) {
	// Same currency
	if from == to {
		return 1.0, nil
	}

	// Determine priority based on useExternal flag
	if a.useExternal {
		// Swap priority
		if a.externalRates != nil {
			rate, err := a.externalRates.GetRate(from, to)
			if err == nil {
				return rate, nil
			}
		}
		// Fall back to custom
		return a.getCustomRate(from, to)
	}

	// Custom first (default)
	rate, err := a.getCustomRate(from, to)
	if err == nil {
		return rate, nil
	}

	// Fall back to external
	if a.externalRates != nil {
		return a.externalRates.GetRate(from, to)
	}

	return 0, fmt.Errorf("no conversion available from %s to %s", from, to)
}

// getCustomRate looks up a custom rate
func (a *AggregateConversions) getCustomRate(from, to string) (float64, error) {
	if a.customRates == nil {
		return 0, fmt.Errorf("no custom rates")
	}

	if fromRates, ok := a.customRates[from]; ok {
		if rate, ok := fromRates[to]; ok {
			return rate, nil
		}
	}

	return 0, fmt.Errorf("no custom conversion from %s to %s", from, to)
}

// Convert converts an amount using aggregated rates
func (a *AggregateConversions) Convert(amount float64, from, to string) (float64, error) {
	rate, err := a.GetRate(from, to)
	if err != nil {
		return 0, err
	}
	return amount * rate, nil
}
