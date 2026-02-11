package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/thenexusengine/tne_springwire/internal/exchange"
	"github.com/thenexusengine/tne_springwire/internal/openrtb"
	"github.com/thenexusengine/tne_springwire/pkg/logger"
)

// BidderMapping represents the full bidder parameter mapping configuration
type BidderMapping struct {
	Publisher struct {
		PublisherID    string   `json:"publisherId"`
		Domain         string   `json:"domain"`
		DefaultBidders []string `json:"defaultBidders"`
	} `json:"publisher"`
	AdUnits map[string]AdUnitConfig `json:"adUnits"`
}

// AdUnitConfig contains bidder-specific parameters for an ad unit
type AdUnitConfig struct {
	Rubicon    *RubiconParams    `json:"rubicon,omitempty"`
	Kargo      *KargoParams      `json:"kargo,omitempty"`
	Sovrn      *SovrnParams      `json:"sovrn,omitempty"`
	OMS        *OMSParams        `json:"oms,omitempty"`
	Aniview    *AniviewParams    `json:"aniview,omitempty"`
	Pubmatic   *PubmaticParams   `json:"pubmatic,omitempty"`
	Triplelift *TripleliftParams `json:"triplelift,omitempty"`
}

// RubiconParams are Rubicon/Magnite adapter parameters
type RubiconParams struct {
	AccountID        int  `json:"accountId"`
	SiteID           int  `json:"siteId"`
	ZoneID           int  `json:"zoneId"`
	BidOnMultiFormat bool `json:"bidonmultiformat"`
}

// KargoParams are Kargo adapter parameters
type KargoParams struct {
	PlacementID string `json:"placementId"`
}

// SovrnParams are Sovrn adapter parameters
type SovrnParams struct {
	TagID int `json:"tagid"`
}

// OMSParams are OMS (Onemobile) adapter parameters
type OMSParams struct {
	PublisherID int `json:"publisherId"`
}

// AniviewParams are Aniview adapter parameters
type AniviewParams struct {
	PublisherID string `json:"publisherId"`
	ChannelID   string `json:"channelId"`
}

// PubmaticParams are Pubmatic adapter parameters
type PubmaticParams struct {
	PublisherID int `json:"publisherId"`
	AdSlot      int `json:"adSlot"`
}

// TripleliftParams are Triplelift adapter parameters
type TripleliftParams struct {
	InventoryCode string `json:"inventoryCode"`
}

// CatalystBidHandler handles MAI Publisher-compatible bid requests
type CatalystBidHandler struct {
	exchange *exchange.Exchange
	mapping  *BidderMapping
}

// LoadBidderMapping loads bidder parameter mapping from JSON file
func LoadBidderMapping(path string) (*BidderMapping, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read mapping file: %w", err)
	}

	var mapping BidderMapping
	if err := json.Unmarshal(data, &mapping); err != nil {
		return nil, fmt.Errorf("failed to parse mapping JSON: %w", err)
	}

	logger.Log.Info().
		Int("ad_units", len(mapping.AdUnits)).
		Str("publisher", mapping.Publisher.PublisherID).
		Msg("Loaded bidder mapping")

	return &mapping, nil
}

// NewCatalystBidHandler creates a new Catalyst bid handler
func NewCatalystBidHandler(ex *exchange.Exchange, mapping *BidderMapping) *CatalystBidHandler {
	return &CatalystBidHandler{
		exchange: ex,
		mapping:  mapping,
	}
}

// MAIBidRequest represents the MAI Publisher bid request format
type MAIBidRequest struct {
	AccountID string       `json:"accountId"`
	Timeout   int          `json:"timeout"` // Client-side timeout in ms
	Slots     []MAISlot    `json:"slots"`
	Page      *MAIPage     `json:"page,omitempty"`
	User      *MAIUser     `json:"user,omitempty"`
	Device    *MAIDevice   `json:"device,omitempty"`
}

// MAISlot represents an ad slot
type MAISlot struct {
	DivID          string      `json:"divId"`
	Sizes          [][]int     `json:"sizes"`
	AdUnitPath     string      `json:"adUnitPath,omitempty"`
	Position       string      `json:"position,omitempty"`
	EnabledBidders []string    `json:"enabled_bidders,omitempty"`
}

// MAIPage represents page context
type MAIPage struct {
	URL        string   `json:"url,omitempty"`
	Domain     string   `json:"domain,omitempty"`
	Keywords   []string `json:"keywords,omitempty"`
	Categories []string `json:"categories,omitempty"`
}

// MAIUser represents user/privacy info
type MAIUser struct {
	ConsentGiven bool   `json:"consentGiven,omitempty"`
	GDPRApplies  bool   `json:"gdprApplies,omitempty"`
	USPConsent   string `json:"uspConsent,omitempty"`
}

// MAIDevice represents device info
type MAIDevice struct {
	Width      int    `json:"width,omitempty"`
	Height     int    `json:"height,omitempty"`
	DeviceType string `json:"deviceType,omitempty"`
	UserAgent  string `json:"userAgent,omitempty"`
}

// MAIBidResponse represents the MAI Publisher bid response format
type MAIBidResponse struct {
	Bids         []MAIBid `json:"bids"`
	ResponseTime int      `json:"responseTime"` // In milliseconds
}

// MAIBid represents a single bid
type MAIBid struct {
	DivID      string      `json:"divId"`
	CPM        float64     `json:"cpm"`
	Currency   string      `json:"currency"`
	Width      int         `json:"width"`
	Height     int         `json:"height"`
	AdID       string      `json:"adId"`
	CreativeID string      `json:"creativeId"`
	DealID     string      `json:"dealId,omitempty"`
	Meta       *MAIBidMeta `json:"meta,omitempty"`
}

// MAIBidMeta represents bid metadata
type MAIBidMeta struct {
	AdvertiserDomains []string `json:"advertiserDomains,omitempty"`
	NetworkID         string   `json:"networkId,omitempty"`
	NetworkName       string   `json:"networkName,omitempty"`
}

// HandleBidRequest handles POST /v1/bid requests
func (h *CatalystBidHandler) HandleBidRequest(w http.ResponseWriter, r *http.Request) {
	log := logger.Log
	startTime := time.Now()

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only accept POST
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse MAI bid request
	var maiBidReq MAIBidRequest
	if err := json.NewDecoder(r.Body).Decode(&maiBidReq); err != nil {
		log.Error().Err(err).Msg("Failed to parse MAI bid request")
		h.writeErrorResponse(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := h.validateMAIBidRequest(&maiBidReq); err != nil {
		log.Error().Err(err).Msg("Invalid MAI bid request")
		h.writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Convert to OpenRTB
	ortbReq, impToSlot, err := h.convertToOpenRTB(r, &maiBidReq)
	if err != nil {
		log.Error().Err(err).Msg("Failed to convert to OpenRTB")
		h.writeErrorResponse(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// Run auction with 2500ms timeout (MAI Publisher requirement)
	ctx, cancel := context.WithTimeout(r.Context(), 2500*time.Millisecond)
	defer cancel()

	auctionReq := &exchange.AuctionRequest{
		BidRequest: ortbReq,
		Timeout:    2500 * time.Millisecond,
	}

	auctionResp, err := h.exchange.RunAuction(ctx, auctionReq)
	if err != nil {
		log.Error().Err(err).Msg("Auction failed")
		// Return empty bids on error (MAI Publisher requirement)
		h.writeMAIResponse(w, &MAIBidResponse{
			Bids:         []MAIBid{},
			ResponseTime: int(time.Since(startTime).Milliseconds()),
		})
		return
	}

	// Convert OpenRTB response to MAI format
	maiResp := h.convertToMAIResponse(auctionResp, impToSlot)
	maiResp.ResponseTime = int(time.Since(startTime).Milliseconds())

	// Log the full response payload for debugging
	if respJSON, err := json.Marshal(maiResp); err == nil {
		log.Debug().
			Str("full_response", string(respJSON)).
			Int("bids", len(maiResp.Bids)).
			Int("response_time_ms", maiResp.ResponseTime).
			Msg("Catalyst full response payload")
	}

	// Write response
	h.writeMAIResponse(w, maiResp)

	log.Info().
		Str("account_id", maiBidReq.AccountID).
		Int("slots", len(maiBidReq.Slots)).
		Int("bids", len(maiResp.Bids)).
		Int("response_time_ms", maiResp.ResponseTime).
		Msg("Catalyst bid request completed")
}

// validateMAIBidRequest validates the MAI bid request
func (h *CatalystBidHandler) validateMAIBidRequest(req *MAIBidRequest) error {
	if req.AccountID == "" {
		return fmt.Errorf("accountId is required")
	}
	if len(req.Slots) == 0 {
		return fmt.Errorf("at least one slot is required")
	}
	for i, slot := range req.Slots {
		if slot.DivID == "" {
			return fmt.Errorf("slot[%d].divId is required", i)
		}
		if len(slot.Sizes) == 0 {
			return fmt.Errorf("slot[%d].sizes is required", i)
		}
		for j, size := range slot.Sizes {
			if len(size) != 2 || size[0] <= 0 || size[1] <= 0 {
				return fmt.Errorf("slot[%d].sizes[%d] must be [width, height] with positive values", i, j)
			}
		}
	}
	return nil
}

// convertToOpenRTB converts MAI bid request to OpenRTB format
func (h *CatalystBidHandler) convertToOpenRTB(r *http.Request, maiBid *MAIBidRequest) (*openrtb.BidRequest, map[string]string, error) {
	// Generate request ID
	requestID := fmt.Sprintf("catalyst-%d", time.Now().UnixNano())

	// Build impressions and track mapping (impID -> divID)
	imps := make([]openrtb.Imp, 0, len(maiBid.Slots))
	impToSlot := make(map[string]string) // Maps impression ID to slot divID

	for i, slot := range maiBid.Slots {
		impID := fmt.Sprintf("%d", i+1)
		impToSlot[impID] = slot.DivID

		// Convert sizes to format array
		formats := make([]openrtb.Format, len(slot.Sizes))
		for j, size := range slot.Sizes {
			formats[j] = openrtb.Format{
				W: size[0],
				H: size[1],
			}
		}

		imp := openrtb.Imp{
			ID: impID,
			Banner: &openrtb.Banner{
				W:      slot.Sizes[0][0], // Use first size as primary
				H:      slot.Sizes[0][1],
				Format: formats,
			},
			TagID: slot.AdUnitPath,
		}

		// Look up bidder parameters from mapping
		if slot.AdUnitPath != "" && h.mapping != nil {
			if adUnitConfig, ok := h.mapping.AdUnits[slot.AdUnitPath]; ok {
				logger.Log.Debug().
					Str("ad_unit", slot.AdUnitPath).
					Msg("Found mapping for ad unit")

				// Build imp.ext with all configured bidders
				impExt := make(map[string]interface{})

				// Rubicon/Magnite
				if adUnitConfig.Rubicon != nil {
					impExt["rubicon"] = map[string]interface{}{
						"accountId":        adUnitConfig.Rubicon.AccountID,
						"siteId":           adUnitConfig.Rubicon.SiteID,
						"zoneId":           adUnitConfig.Rubicon.ZoneID,
						"bidonmultiformat": adUnitConfig.Rubicon.BidOnMultiFormat,
					}
				}

				// Kargo
				if adUnitConfig.Kargo != nil {
					impExt["kargo"] = map[string]interface{}{
						"placementId": adUnitConfig.Kargo.PlacementID,
					}
				}

				// Sovrn
				if adUnitConfig.Sovrn != nil {
					impExt["sovrn"] = map[string]interface{}{
						"tagid": adUnitConfig.Sovrn.TagID,
					}
				}

				// OMS (Onemobile) - uses "onetag" in OpenRTB
				if adUnitConfig.OMS != nil {
					impExt["onetag"] = map[string]interface{}{
						"publisherId": adUnitConfig.OMS.PublisherID,
					}
				}

				// Aniview
				if adUnitConfig.Aniview != nil {
					impExt["aniview"] = map[string]interface{}{
						"publisherId": adUnitConfig.Aniview.PublisherID,
						"channelId":   adUnitConfig.Aniview.ChannelID,
					}
				}

				// Pubmatic
				if adUnitConfig.Pubmatic != nil {
					impExt["pubmatic"] = map[string]interface{}{
						"publisherId": adUnitConfig.Pubmatic.PublisherID,
						"adSlot":      adUnitConfig.Pubmatic.AdSlot,
					}
				}

				// Triplelift
				if adUnitConfig.Triplelift != nil {
					impExt["triplelift"] = map[string]interface{}{
						"inventoryCode": adUnitConfig.Triplelift.InventoryCode,
					}
				}

				// Marshal and attach to impression
				if len(impExt) > 0 {
					extJSON, err := json.Marshal(impExt)
					if err == nil {
						imp.Ext = extJSON
						logger.Log.Info().
							Str("ad_unit", slot.AdUnitPath).
							Int("bidders", len(impExt)).
							Msg("Injected bidder parameters")
					} else {
						logger.Log.Error().
							Err(err).
							Str("ad_unit", slot.AdUnitPath).
							Msg("Failed to marshal bidder parameters")
					}
				}
			} else {
				logger.Log.Warn().
					Str("ad_unit", slot.AdUnitPath).
					Msg("No mapping found for ad unit")
			}
		}

		imps = append(imps, imp)
	}

	// Build site
	site := &openrtb.Site{
		ID: maiBid.AccountID,
	}

	if maiBid.Page != nil {
		site.Domain = maiBid.Page.Domain
		if maiBid.Page.URL != "" {
			site.Page = maiBid.Page.URL
		}
		if maiBid.Page.Domain == "" && maiBid.Page.URL != "" {
			// Extract domain from URL if not provided
			site.Domain = extractDomain(maiBid.Page.URL)
		}
		if len(maiBid.Page.Keywords) > 0 {
			site.Keywords = strings.Join(maiBid.Page.Keywords, ",")
		}
		if len(maiBid.Page.Categories) > 0 {
			site.Cat = maiBid.Page.Categories
		}
	}

	// Build device
	device := &openrtb.Device{
		UA: r.Header.Get("User-Agent"),
		IP: getClientIP(r),
	}

	if maiBid.Device != nil {
		if maiBid.Device.UserAgent != "" {
			device.UA = maiBid.Device.UserAgent
		}
		if maiBid.Device.Width > 0 && maiBid.Device.Height > 0 {
			device.W = maiBid.Device.Width
			device.H = maiBid.Device.Height
		}
		// Map device type to OpenRTB device type
		switch strings.ToLower(maiBid.Device.DeviceType) {
		case "mobile", "phone":
			device.DeviceType = 1 // Mobile/Tablet
		case "tablet":
			device.DeviceType = 5 // Tablet
		case "desktop", "pc":
			device.DeviceType = 2 // Personal Computer
		case "tv", "ctv", "connected_tv":
			device.DeviceType = 3 // Connected TV
		}
	}

	// Build user and regulations
	var user *openrtb.User
	var regs *openrtb.Regs

	if maiBid.User != nil {
		user = &openrtb.User{}
		if maiBid.User.ConsentGiven {
			user.Consent = "1"
		} else {
			user.Consent = "0"
		}

		regs = &openrtb.Regs{}
		if maiBid.User.GDPRApplies {
			gdpr := 1
			regs.GDPR = &gdpr
		}
		if maiBid.User.USPConsent != "" {
			regs.USPrivacy = maiBid.User.USPConsent
		}
	}

	// Build OpenRTB request
	ortbReq := &openrtb.BidRequest{
		ID:     requestID,
		Imp:    imps,
		Site:   site,
		Device: device,
		User:   user,
		Regs:   regs,
		Cur:    []string{"USD"},
		TMax:   2500, // 2500ms internal timeout
	}

	return ortbReq, impToSlot, nil
}

// convertToMAIResponse converts OpenRTB response to MAI format
func (h *CatalystBidHandler) convertToMAIResponse(auctionResp *exchange.AuctionResponse, impToSlot map[string]string) *MAIBidResponse {
	maiResp := &MAIBidResponse{
		Bids: []MAIBid{},
	}

	if auctionResp == nil || auctionResp.BidResponse == nil {
		return maiResp
	}

	// Extract all bids from all seats
	for _, seatBid := range auctionResp.BidResponse.SeatBid {
		for _, bid := range seatBid.Bid {
			// Map impression ID back to divID
			divID, ok := impToSlot[bid.ImpID]
			if !ok {
				continue // Skip if we can't map back to slot
			}

			maiBid := MAIBid{
				DivID:      divID,
				CPM:        bid.Price,
				Currency:   auctionResp.BidResponse.Cur,
				Width:      bid.W,
				Height:     bid.H,
				AdID:       bid.ID,
				CreativeID: bid.CRID,
				DealID:     bid.DealID,
			}

			// Set default currency if not specified
			if maiBid.Currency == "" {
				maiBid.Currency = "USD"
			}

			// Build metadata
			if len(bid.ADomain) > 0 || bid.CID != "" {
				maiBid.Meta = &MAIBidMeta{
					AdvertiserDomains: bid.ADomain,
					NetworkID:         bid.CID,
					NetworkName:       seatBid.Seat,
				}
			}

			maiResp.Bids = append(maiResp.Bids, maiBid)
		}
	}

	return maiResp
}

// writeMAIResponse writes MAI-formatted JSON response
func (h *CatalystBidHandler) writeMAIResponse(w http.ResponseWriter, resp *MAIBidResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Log.Error().Err(err).Msg("Failed to encode MAI response")
	}
}

// writeErrorResponse writes error response
func (h *CatalystBidHandler) writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
