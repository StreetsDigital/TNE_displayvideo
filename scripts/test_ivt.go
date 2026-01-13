// Test IVT detection with various request scenarios
package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http/httptest"
	"os"

	"github.com/thenexusengine/tne_springwire/internal/middleware"
)

// Color codes for terminal output
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
)

type TestScenario struct {
	Name        string
	UserAgent   string
	Referer     string
	Domain      string
	PublisherID string
	Description string
	ExpectScore int
	ExpectValid bool
}

func main() {
	fmt.Printf("%s=== IVT Detection Test Suite ===%s\n\n", colorCyan, colorReset)

	scenarios := []TestScenario{
		{
			Name:        "Clean Traffic",
			UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			Referer:     "https://example.com/article",
			Domain:      "example.com",
			PublisherID: "pub-example",
			Description: "Normal browser request from legitimate publisher",
			ExpectScore: 0,
			ExpectValid: true,
		},
		{
			Name:        "Bot User Agent",
			UserAgent:   "Googlebot/2.1 (+http://www.google.com/bot.html)",
			Referer:     "https://example.com",
			Domain:      "example.com",
			PublisherID: "pub-example",
			Description: "Search engine bot with valid referer",
			ExpectScore: 50,
			ExpectValid: true, // Not blocked (score < 70)
		},
		{
			Name:        "Scraper",
			UserAgent:   "curl/7.68.0",
			Referer:     "https://example.com",
			Domain:      "example.com",
			PublisherID: "pub-example",
			Description: "Command-line scraper tool",
			ExpectScore: 50,
			ExpectValid: true, // Flagged but not blocked (score < 70)
		},
		{
			Name:        "Domain Mismatch",
			UserAgent:   "Mozilla/5.0 (normal browser)",
			Referer:     "https://malicious.com/page",
			Domain:      "example.com",
			PublisherID: "pub-example",
			Description: "Normal UA but referer doesn't match domain",
			ExpectScore: 50,
			ExpectValid: true, // Flagged but not blocked (score < 70)
		},
		{
			Name:        "Bot + Domain Mismatch",
			UserAgent:   "curl/7.68.0",
			Referer:     "https://malicious.com",
			Domain:      "example.com",
			PublisherID: "pub-example",
			Description: "Scraper with mismatched domain (high risk)",
			ExpectScore: 100,
			ExpectValid: false, // BLOCKED (score >= 70)
		},
		{
			Name:        "Headless Browser",
			UserAgent:   "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) HeadlessChrome/91.0.4472.124 Safari/537.36",
			Referer:     "https://example.com",
			Domain:      "example.com",
			PublisherID: "pub-example",
			Description: "Headless Chrome (automation/scraping tool)",
			ExpectScore: 50,
			ExpectValid: true, // Flagged but not blocked
		},
		{
			Name:        "Python Scraper + Mismatch",
			UserAgent:   "python-requests/2.25.1",
			Referer:     "https://competitor.com",
			Domain:      "example.com",
			PublisherID: "pub-example",
			Description: "Python scraper with wrong domain (fraud)",
			ExpectScore: 100,
			ExpectValid: false, // BLOCKED
		},
		{
			Name:        "Empty User Agent",
			UserAgent:   "",
			Referer:     "https://example.com",
			Domain:      "example.com",
			PublisherID: "pub-example",
			Description: "No user agent provided (suspicious)",
			ExpectScore: 50,
			ExpectValid: true, // Flagged but not blocked
		},
		{
			Name:        "Subdomain Valid",
			UserAgent:   "Mozilla/5.0 (normal browser)",
			Referer:     "https://www.example.com/page",
			Domain:      "example.com",
			PublisherID: "pub-example",
			Description: "Valid subdomain referer",
			ExpectScore: 0,
			ExpectValid: true,
		},
		{
			Name:        "Selenium WebDriver",
			UserAgent:   "Mozilla/5.0 (selenium; like Gecko) Chrome/91.0",
			Referer:     "https://malicious.com",
			Domain:      "example.com",
			PublisherID: "pub-example",
			Description: "Selenium automation with wrong domain",
			ExpectScore: 100,
			ExpectValid: false, // BLOCKED
		},
	}

	// Test in monitoring mode first
	fmt.Printf("%s--- Testing in MONITORING MODE (flags but doesn't block) ---%s\n\n", colorYellow, colorReset)
	runTests(scenarios, false)

	fmt.Printf("\n%s--- Testing in BLOCKING MODE (blocks score >= 70) ---%s\n\n", colorRed, colorReset)
	runTests(scenarios, true)

	// Show metrics summary
	showMetricsSummary()
}

func runTests(scenarios []TestScenario, blockMode bool) {
	config := middleware.DefaultIVTConfig()
	config.BlockingEnabled = blockMode
	detector := middleware.NewIVTDetector(config)

	passCount := 0
	flaggedCount := 0
	blockedCount := 0

	for i, scenario := range scenarios {
		fmt.Printf("%s[%d/%d] %s%s\n", colorBlue, i+1, len(scenarios), scenario.Name, colorReset)
		fmt.Printf("  %s\n", scenario.Description)

		// Create test request
		req := httptest.NewRequest("POST", "/openrtb2/auction", bytes.NewReader([]byte(`{
			"id": "test-request",
			"imp": [{"id": "1", "banner": {"format": [{"w": 300, "h": 250}]}}],
			"site": {"domain": "`+scenario.Domain+`", "publisher": {"id": "`+scenario.PublisherID+`"}}
		}`)))
		req.Header.Set("User-Agent", scenario.UserAgent)
		req.Header.Set("Referer", scenario.Referer)

		// Run IVT detection
		result := detector.Validate(context.Background(), req, scenario.PublisherID, scenario.Domain)

		// Display results
		fmt.Printf("  UA: %s\n", truncate(scenario.UserAgent, 60))
		fmt.Printf("  Referer: %s\n", scenario.Referer)
		fmt.Printf("  Domain: %s\n", scenario.Domain)

		// Show score with color coding
		scoreColor := colorGreen
		if result.Score >= 70 {
			scoreColor = colorRed
		} else if result.Score >= 35 {
			scoreColor = colorYellow
		}
		fmt.Printf("  %sScore: %d/100%s", scoreColor, result.Score, colorReset)
		if result.Score >= 70 {
			fmt.Printf(" %s(THRESHOLD EXCEEDED)%s", colorRed, colorReset)
		}
		fmt.Printf("\n")

		// Show signals
		if len(result.Signals) > 0 {
			fmt.Printf("  Signals: ")
			for j, signal := range result.Signals {
				if j > 0 {
					fmt.Printf(", ")
				}
				sigColor := colorYellow
				if signal.Severity == "high" {
					sigColor = colorRed
				}
				fmt.Printf("%s%s (%s)%s", sigColor, signal.Type, signal.Severity, colorReset)
			}
			fmt.Printf("\n")
		}

		// Show decision
		if result.ShouldBlock {
			fmt.Printf("  Decision: %s❌ BLOCKED%s (Reason: %s)\n", colorRed, colorReset, result.BlockReason)
			blockedCount++
		} else if !result.IsValid {
			fmt.Printf("  Decision: %s⚠️  FLAGGED%s (monitoring only)\n", colorYellow, colorReset)
			flaggedCount++
		} else {
			fmt.Printf("  Decision: %s✅ ALLOWED%s\n", colorGreen, colorReset)
			passCount++
		}

		// Check if result matches expectation
		scoreMatch := result.Score == scenario.ExpectScore
		validMatch := result.IsValid == scenario.ExpectValid

		if scoreMatch && validMatch {
			fmt.Printf("  %s✓ Test PASSED%s\n", colorGreen, colorReset)
		} else {
			fmt.Printf("  %s✗ Test FAILED%s (expected score=%d valid=%v, got score=%d valid=%v)\n",
				colorRed, colorReset, scenario.ExpectScore, scenario.ExpectValid, result.Score, result.IsValid)
		}

		fmt.Println()
	}

	// Summary
	total := len(scenarios)
	fmt.Printf("%sSummary:%s\n", colorCyan, colorReset)
	fmt.Printf("  Total Requests: %d\n", total)
	fmt.Printf("  %s✅ Allowed: %d%s\n", colorGreen, passCount, colorReset)
	fmt.Printf("  %s⚠️  Flagged: %d%s\n", colorYellow, flaggedCount, colorReset)
	if blockMode {
		fmt.Printf("  %s❌ Blocked: %d%s\n", colorRed, blockedCount, colorReset)
	}
}

func showMetricsSummary() {
	fmt.Printf("\n%s=== Production Deployment Recommendations ===%s\n\n", colorCyan, colorReset)

	fmt.Printf("%sPhase 1: Monitoring (Week 1-2)%s\n", colorGreen, colorReset)
	fmt.Printf("  IVT_MONITORING_ENABLED=true\n")
	fmt.Printf("  IVT_BLOCKING_ENABLED=false\n")
	fmt.Printf("  → Collect metrics, identify false positives\n\n")

	fmt.Printf("%sPhase 2: Selective Blocking (Week 3-4)%s\n", colorYellow, colorReset)
	fmt.Printf("  IVT_MONITORING_ENABLED=true\n")
	fmt.Printf("  IVT_BLOCKING_ENABLED=true\n")
	fmt.Printf("  IVT_CHECK_UA=true\n")
	fmt.Printf("  IVT_CHECK_REFERER=false  # Start with UA only\n")
	fmt.Printf("  → Monitor impact, tune thresholds\n\n")

	fmt.Printf("%sPhase 3: Full Protection (Production)%s\n", colorRed, colorReset)
	fmt.Printf("  IVT_MONITORING_ENABLED=true\n")
	fmt.Printf("  IVT_BLOCKING_ENABLED=true\n")
	fmt.Printf("  IVT_CHECK_UA=true\n")
	fmt.Printf("  IVT_CHECK_REFERER=true\n")
	fmt.Printf("  IVT_ALLOWED_COUNTRIES=\"US,GB,CA,AU,NZ\"  # Optional\n")
	fmt.Printf("  → Full fraud protection active\n\n")

	fmt.Printf("%sMonitoring:%s\n", colorCyan, colorReset)
	fmt.Printf("  - Check logs: grep 'IVT detected'\n")
	fmt.Printf("  - Review metrics: GetIVTMetrics()\n")
	fmt.Printf("  - Watch headers: X-IVT-Score, X-IVT-Signals\n")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func init() {
	// Disable JSON output for cleaner test display
	os.Setenv("LOG_LEVEL", "error")
}
