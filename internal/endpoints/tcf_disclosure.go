package endpoints

import (
	"net/http"
)

// HandleTCFDisclosure serves the IAB TCF v2 device storage disclosure JSON file
// This file declares all programmatic bidders and their cookie/storage usage
// for GDPR compliance and proper consent management via CMPs like Sourcepoint.
//
// The disclosure file is required for:
// - TCF v2 compliance (mandatory by Feb 28, 2026)
// - Publisher CMP configuration (references this URL)
// - Programmatic bidder cookie syncing
// - User consent collection for vendors
//
// Standards:
// - IAB Transparency & Consent Framework v2
// - Device Storage Disclosure v1.1
// - https://github.com/InteractiveAdvertisingBureau/GDPR-Transparency-and-Consent-Framework
func HandleTCFDisclosure(w http.ResponseWriter, r *http.Request) {
	// Set proper headers for TCF compliance
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Cache-Control", "public, max-age=86400") // 24 hour cache

	// Handle OPTIONS preflight for CORS
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Serve the static JSON file
	http.ServeFile(w, r, "assets/tcf-disclosure.json")
}
