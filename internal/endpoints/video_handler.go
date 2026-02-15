package endpoints

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/thenexusengine/tne_springwire/internal/ctv"
	"github.com/thenexusengine/tne_springwire/internal/exchange"
	"github.com/thenexusengine/tne_springwire/internal/openrtb"
	"github.com/thenexusengine/tne_springwire/pkg/vast"
)

// VideoHandler handles video ad requests and returns VAST responses
type VideoHandler struct {
	exchange        *exchange.Exchange
	vastBuilder     *exchange.VASTResponseBuilder
	trackingBaseURL string
}

// NewVideoHandler creates a new video handler
func NewVideoHandler(ex *exchange.Exchange, trackingBaseURL string) *VideoHandler {
	return &VideoHandler{
		exchange:        ex,
		vastBuilder:     exchange.NewVASTResponseBuilder(trackingBaseURL),
		trackingBaseURL: trackingBaseURL,
	}
}

// HandleVASTRequest handles GET /video/vast requests
// This endpoint accepts query parameters and returns a VAST XML response
func (h *VideoHandler) HandleVASTRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse video parameters from query string
	bidReq, err := h.parseVASTRequest(r)
	if err != nil {
		log.Warn().Err(err).Msg("Invalid VAST request parameters")
		h.writeVASTError(w, "Invalid request parameters")
		return
	}

	// Detect CTV device for optimization
	if bidReq.Device != nil {
		deviceInfo := ctv.DetectDevice(bidReq.Device)
		if deviceInfo.IsCTV {
			h.applyCTVOptimizations(bidReq, deviceInfo)
		}
	}

	// Create auction request
	auctionReq := &exchange.AuctionRequest{
		BidRequest: bidReq,
		Timeout:    time.Duration(bidReq.TMax) * time.Millisecond,
	}

	// Run auction through exchange
	auctionResp, err := h.exchange.RunAuction(ctx, auctionReq)
	if err != nil {
		log.Error().Err(err).Msg("Video auction failed")
		h.writeVASTError(w, "Auction failed")
		return
	}

	// Build VAST response from auction results
	vastResp, err := h.vastBuilder.BuildVASTFromAuction(bidReq, auctionResp)
	if err != nil {
		log.Error().Err(err).Msg("Failed to build VAST response")
		h.writeVASTError(w, "Failed to build response")
		return
	}

	// Marshal and write VAST XML
	data, err := vastResp.Marshal()
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal VAST")
		h.writeVASTError(w, "Failed to serialize response")
		return
	}

	// Set headers and write response
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	// SECURITY NOTE: CORS wildcard (*) is intentional for VAST endpoints.
	// VAST/VPAID video players are typically embedded in iframes or third-party
	// contexts (e.g., video.js, JW Player, Brightcove) and require permissive CORS
	// to fetch ad responses. This is an IAB industry-standard practice for VAST
	// ad serving endpoints. The VAST response contains only ad markup, not
	// sensitive user data, so wildcard CORS does not create a security risk.
	// See: IAB VAST 4.2 spec section on "Cross-Origin Resource Sharing"
	h.setVASTCORSHeaders(w)
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	w.Write(data)

	log.Info().
		Str("request_id", bidReq.ID).
		Bool("has_ads", !vastResp.IsEmpty()).
		Msg("VAST response sent")
}

// HandleOpenRTBVideo handles POST /video/openrtb requests
// This endpoint accepts OpenRTB JSON and returns VAST XML
func (h *VideoHandler) HandleOpenRTBVideo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse OpenRTB bid request from body
	var bidReq openrtb.BidRequest
	if err := json.NewDecoder(r.Body).Decode(&bidReq); err != nil {
		log.Warn().Err(err).Msg("Invalid OpenRTB request body")
		h.writeVASTError(w, "Invalid request body")
		return
	}

	// Validate that this is a video request
	hasVideo := false
	for _, imp := range bidReq.Imp {
		if imp.Video != nil {
			hasVideo = true
			break
		}
	}
	if !hasVideo {
		h.writeVASTError(w, "No video impressions in request")
		return
	}

	// Run auction
	auctionReq := &exchange.AuctionRequest{
		BidRequest: &bidReq,
		Timeout:    time.Duration(bidReq.TMax) * time.Millisecond,
	}

	auctionResp, err := h.exchange.RunAuction(ctx, auctionReq)
	if err != nil {
		log.Error().Err(err).Msg("Video auction failed")
		h.writeVASTError(w, "Auction failed")
		return
	}

	// Build VAST response
	vastResp, err := h.vastBuilder.BuildVASTFromAuction(&bidReq, auctionResp)
	if err != nil {
		log.Error().Err(err).Msg("Failed to build VAST response")
		h.writeVASTError(w, "Failed to build response")
		return
	}

	// Marshal and write VAST XML
	data, err := vastResp.Marshal()
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal VAST")
		h.writeVASTError(w, "Failed to serialize response")
		return
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	// SECURITY NOTE: CORS wildcard intentional for VAST - see setVASTCORSHeaders
	h.setVASTCORSHeaders(w)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// setVASTCORSHeaders sets CORS headers for VAST responses.
//
// SECURITY RATIONALE: VAST endpoints intentionally use permissive CORS (Access-Control-Allow-Origin: *)
// because video players (video.js, JW Player, Brightcove, etc.) are typically embedded in third-party
// iframes and require cross-origin access to fetch ad responses. This is standard practice per IAB
// VAST specification. VAST XML contains only ad markup (media URLs, tracking pixels, etc.) and does
// not include sensitive user data, so wildcard CORS does not create a data exposure risk.
//
// This is distinct from the /openrtb2/auction endpoint which handles bid requests containing
// potentially sensitive user data and uses the configurable CORS middleware.
func (h *VideoHandler) setVASTCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")
}

// parseVASTRequest parses video parameters from query string into OpenRTB bid request
func (h *VideoHandler) parseVASTRequest(r *http.Request) (*openrtb.BidRequest, error) {
	q := r.URL.Query()

	// Required parameters
	requestID := q.Get("id")
	if requestID == "" {
		requestID = generateRequestID()
	}

	// Video dimensions (default to 1920x1080)
	width := parseInt(q.Get("w"), 1920)
	height := parseInt(q.Get("h"), 1080)

	// Duration constraints
	minDuration := parseInt(q.Get("mindur"), 5)
	maxDuration := parseInt(q.Get("maxdur"), 30)

	// Skip parameters
	skip := parseInt(q.Get("skip"), 0)
	skipAfter := parseInt(q.Get("skipafter"), 0)

	// Placement type (legacy 2.5: 1=in-stream, 3=in-article, 4=in-feed, 5=interstitial)
	placement := parseInt(q.Get("placement"), 1)

	// 2.6 plcmt: Video placement type (1=instream, 2=accompanying, 3=interstitial, 4=standalone)
	// Map legacy placement to 2.6 plcmt for backward compatibility
	plcmt := parseInt(q.Get("plcmt"), 0)
	if plcmt == 0 {
		// Auto-map from legacy placement to 2.6 plcmt
		switch placement {
		case 1: // in-stream
			plcmt = 1 // instream
		case 3, 4: // in-article, in-feed
			plcmt = 2 // accompanying content
		case 5: // interstitial
			plcmt = 3 // interstitial
		default:
			plcmt = 1 // default to instream
		}
	}

	// Playback method (1=auto-play sound on, 2=auto-play sound off, 3=click-to-play, 4=mouseover)
	playbackMethod := parseIntArray(q.Get("playbackmethod"), []int{2}) // Default: auto-play sound off

	// Protocols (comma-separated)
	protocols := parseIntArray(q.Get("protocols"), []int{2, 3, 5, 6})

	// MIME types (comma-separated)
	mimes := parseStringArray(q.Get("mimes"), []string{"video/mp4", "video/webm"})

	// Bitrate
	minBitrate := parseInt(q.Get("minbitrate"), 300)
	maxBitrate := parseInt(q.Get("maxbitrate"), 5000)

	// Floor price
	bidFloor := parseFloat(q.Get("bidfloor"), 0.0)

	// Build video object
	video := &openrtb.Video{
		Mimes:          mimes,
		MinDuration:    minDuration,
		MaxDuration:    maxDuration,
		Protocols:      protocols,
		W:              width,
		H:              height,
		Placement:      placement,           // Legacy 2.5 placement (for backward compat)
		Plcmt:          plcmt,               // 2.6 placement type
		Linearity:      1,                   // Linear/in-stream
		PlaybackMethod: playbackMethod,      // Viewability signal for DSPs
		MinBitrate:     minBitrate,
		MaxBitrate:     maxBitrate,
		API:            []int{1, 2, 5, 7},   // VPAID 1.0, VPAID 2.0, MRAID-3, OMID-1 (OMSDK)
	}

	if skip == 1 {
		skipInt := skip
		video.Skip = &skipInt
		video.SkipAfter = skipAfter
	}

	// Build impression
	secureFlag := 1
	imp := openrtb.Imp{
		ID:          "1",
		Video:       video,
		BidFloor:    bidFloor,
		BidFloorCur: "USD",
		Secure:      &secureFlag,
		TagID:       q.Get("tagid"), // Placement identifier for DSP reporting/optimization
	}

	// Build device from headers with enrichment
	device := &openrtb.Device{
		UA: r.UserAgent(),
		IP: getClientIP(r),
		W:  width,
		H:  height,
		JS: 1, // JavaScript support (assumed for web video players)
	}

	// Extract language from Accept-Language header (e.g., "en-US,en;q=0.9" -> "en")
	if acceptLang := r.Header.Get("Accept-Language"); acceptLang != "" {
		if lang := parseAcceptLanguage(acceptLang); lang != "" {
			device.Language = lang
		}
	}

	// Parse Sec-CH-UA Client Hints into SUA if available
	if sua := parseSUAFromHeaders(r); sua != nil {
		device.SUA = sua
	}

	// Build bid request
	bidReq := &openrtb.BidRequest{
		ID:   requestID,
		Imp:  []openrtb.Imp{imp},
		Device: device,
		TMax: 1000, // 1 second timeout
		Cur:  []string{"USD"},
		AT:   2, // Second-price auction
	}

	// Add site or app info if provided
	// Create Site if site_id OR domain is provided (OpenRTB allows ID to be optional)
	siteID := q.Get("site_id")
	domain := q.Get("domain")
	page := q.Get("page")

	if siteID != "" || domain != "" {
		bidReq.Site = &openrtb.Site{
			ID:     siteID,
			Domain: domain,
			Page:   page,
		}
	}

	return bidReq, nil
}

// writeVASTError writes a VAST error response
func (h *VideoHandler) writeVASTError(w http.ResponseWriter, message string) {
	// SECURITY: Escape message parameter to prevent URL injection (CVE-2026-XXXX)
	v := vast.CreateErrorVAST(fmt.Sprintf("%s/video/error?msg=%s", h.trackingBaseURL, url.QueryEscape(message)))
	data, _ := v.Marshal()

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	// SECURITY NOTE: CORS wildcard intentional for VAST error responses - see setVASTCORSHeaders
	h.setVASTCORSHeaders(w)
	w.WriteHeader(http.StatusOK) // VAST always returns 200
	w.Write(data)
}

// applyCTVOptimizations applies CTV device-specific optimizations
func (h *VideoHandler) applyCTVOptimizations(bidReq *openrtb.BidRequest, deviceInfo *ctv.DeviceInfo) {
	caps := ctv.GetCapabilities(deviceInfo.Type)

	for i := range bidReq.Imp {
		if bidReq.Imp[i].Video != nil {
			// Limit bitrate based on device capabilities
			if bidReq.Imp[i].Video.MaxBitrate > caps.MaxBitrate {
				bidReq.Imp[i].Video.MaxBitrate = caps.MaxBitrate
			}

			// Filter VPAID if not supported
			if !caps.SupportsVPAID {
				filtered := make([]int, 0)
				for _, api := range bidReq.Imp[i].Video.API {
					if api != 1 && api != 2 { // Remove VPAID 1.0 and 2.0
						filtered = append(filtered, api)
					}
				}
				bidReq.Imp[i].Video.API = filtered
			}
		}
	}
}

// Helper functions

func parseInt(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return val
}

func parseFloat(s string, defaultVal float64) float64 {
	if s == "" {
		return defaultVal
	}
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return defaultVal
	}
	return val
}

func parseIntArray(s string, defaultVal []int) []int {
	if s == "" {
		return defaultVal
	}
	var result []int
	for _, part := range parseStringArray(s, nil) {
		if val, err := strconv.Atoi(part); err == nil {
			result = append(result, val)
		}
	}
	if len(result) == 0 {
		return defaultVal
	}
	return result
}

func parseStringArray(s string, defaultVal []string) []string {
	if s == "" {
		return defaultVal
	}
	// Split by comma
	parts := []string{}
	current := ""
	for _, c := range s {
		if c == ',' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

func generateRequestID() string {
	return fmt.Sprintf("video-%d", time.Now().UnixNano())
}

// parseAcceptLanguage extracts the primary language from an Accept-Language header.
// e.g., "en-US,en;q=0.9,fr;q=0.8" -> "en"
func parseAcceptLanguage(header string) string {
	if header == "" {
		return ""
	}
	// Take the first language (highest priority)
	lang := header
	if idx := strings.IndexByte(lang, ','); idx > 0 {
		lang = lang[:idx]
	}
	// Remove quality factor
	if idx := strings.IndexByte(lang, ';'); idx > 0 {
		lang = lang[:idx]
	}
	lang = strings.TrimSpace(lang)
	// Extract the primary language subtag (e.g., "en-US" -> "en")
	if idx := strings.IndexByte(lang, '-'); idx > 0 {
		return strings.ToLower(lang[:idx])
	}
	return strings.ToLower(lang)
}

// parseSUAFromHeaders builds a Structured User Agent from User-Agent Client Hints headers.
// Sec-CH-UA: "Chromium";v="122", "Not(A:Brand";v="24", "Google Chrome";v="122"
// Sec-CH-UA-Platform: "macOS"
// Sec-CH-UA-Mobile: ?0
// Sec-CH-UA-Model: ""
func parseSUAFromHeaders(r *http.Request) *openrtb.UserAgent {
	chUA := r.Header.Get("Sec-CH-UA")
	if chUA == "" {
		return nil
	}

	sua := &openrtb.UserAgent{
		Source: 1, // Low-entropy Client Hints
	}

	// Parse Sec-CH-UA into browsers
	sua.Browsers = parseBrandVersionList(chUA)

	// Parse platform
	if platform := r.Header.Get("Sec-CH-UA-Platform"); platform != "" {
		platform = strings.Trim(platform, "\"")
		sua.Platform = &openrtb.BrandVersion{
			Brand: platform,
		}
		if pv := r.Header.Get("Sec-CH-UA-Platform-Version"); pv != "" {
			pv = strings.Trim(pv, "\"")
			sua.Platform.Version = []string{pv}
		}
	}

	// Parse mobile
	if mobile := r.Header.Get("Sec-CH-UA-Mobile"); mobile != "" {
		m := 0
		if mobile == "?1" {
			m = 1
		}
		sua.Mobile = &m
	}

	// Parse model
	if model := r.Header.Get("Sec-CH-UA-Model"); model != "" {
		model = strings.Trim(model, "\"")
		if model != "" {
			sua.Model = model
		}
	}

	// Parse architecture (high-entropy hint)
	if arch := r.Header.Get("Sec-CH-UA-Arch"); arch != "" {
		sua.Architecture = strings.Trim(arch, "\"")
		sua.Source = 2 // High-entropy Client Hints
	}

	// Parse bitness (high-entropy hint)
	if bitness := r.Header.Get("Sec-CH-UA-Bitness"); bitness != "" {
		sua.Bitness = strings.Trim(bitness, "\"")
		sua.Source = 2
	}

	return sua
}

// parseBrandVersionList parses a Sec-CH-UA header value into BrandVersion entries.
// Format: "Brand1";v="Version1", "Brand2";v="Version2"
func parseBrandVersionList(header string) []openrtb.BrandVersion {
	var brands []openrtb.BrandVersion
	for _, entry := range strings.Split(header, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		bv := openrtb.BrandVersion{}
		// Parse brand name (quoted)
		parts := strings.SplitN(entry, ";", 2)
		if len(parts) >= 1 {
			bv.Brand = strings.Trim(strings.TrimSpace(parts[0]), "\"")
		}
		// Parse version
		if len(parts) >= 2 {
			vPart := strings.TrimSpace(parts[1])
			if strings.HasPrefix(vPart, "v=") {
				ver := strings.Trim(vPart[2:], "\"")
				bv.Version = []string{ver}
			}
		}
		if bv.Brand != "" {
			brands = append(brands, bv)
		}
	}
	return brands
}
