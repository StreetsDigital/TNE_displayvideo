package endpoints

import (
	"strings"
)

// GetCookieDomain extracts the proper cookie domain from a request
// For "ads.thenexusengine.com" -> ".thenexusengine.com" (wildcard for subdomains)
// For "localhost:8000" -> "localhost" (local development)
func GetCookieDomain(host string) string {
	// Strip port
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}

	// Handle localhost and IP addresses
	if host == "localhost" ||
		strings.Contains(host, "127.0.0.1") ||
		strings.Contains(host, "::1") ||
		strings.Count(host, ".") == 0 {
		return host
	}

	parts := strings.Split(host, ".")

	// Special handling for multi-level TLDs (.co.uk, .com.au, etc.)
	if len(parts) >= 3 {
		secondLevel := parts[len(parts)-2]
		// Known second-level TLD patterns
		if secondLevel == "co" || secondLevel == "com" ||
			secondLevel == "gov" || secondLevel == "org" ||
			secondLevel == "ac" || secondLevel == "edu" {
			// Return last 3 parts: .example.co.uk
			return "." + strings.Join(parts[len(parts)-3:], ".")
		}
	}

	// Default: return last 2 parts with leading dot
	// ads.thenexusengine.com -> .thenexusengine.com
	if len(parts) >= 2 {
		return "." + strings.Join(parts[len(parts)-2:], ".")
	}

	return host
}
