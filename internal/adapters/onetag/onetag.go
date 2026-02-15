// Package onetag implements the OneTag bidder adapter
package onetag

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/thenexusengine/tne_springwire/internal/adapters"
	"github.com/thenexusengine/tne_springwire/internal/openrtb"
	"github.com/thenexusengine/tne_springwire/pkg/logger"
)

const (
	defaultEndpoint = "https://prebid-server.onetag-sys.com/prebid-server"
)

// Adapter implements the OneTag bidder
type Adapter struct {
	endpoint string
}

// OneTagParams represents the OneTag-specific bidder parameters
type OneTagParams struct {
	PublisherID string `json:"publisherId"`
	PubID       string `json:"pubId"` // Alternative parameter name
}

// New creates a new OneTag adapter
func New(endpoint string) *Adapter {
	if endpoint == "" {
		endpoint = defaultEndpoint
	}
	return &Adapter{endpoint: endpoint}
}

// MakeRequests builds HTTP requests for OneTag
func (a *Adapter) MakeRequests(request *openrtb.BidRequest, extraInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	var errors []error

	// Extract publisher ID from the first impression's ext
	publisherID, err := a.getPublisherID(request)
	if err != nil {
		return nil, []error{err}
	}

	// Build endpoint URL with publisher ID
	endpoint := fmt.Sprintf("%s/%s", strings.TrimSuffix(a.endpoint, "/"), publisherID)

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, []error{fmt.Errorf("failed to marshal request: %w", err)}
	}

	headers := adapters.MakeOpenRTBHeaders()

	return []*adapters.RequestData{
		{
			Method:  "POST",
			URI:     endpoint,
			Body:    requestBody,
			Headers: headers,
		},
	}, errors
}

// getPublisherID extracts the publisher ID from impression extensions
func (a *Adapter) getPublisherID(request *openrtb.BidRequest) (string, error) {
	if len(request.Imp) == 0 {
		return "", fmt.Errorf("no impressions in request")
	}

	// Extract publisher ID from first impression
	var params OneTagParams
	if err := json.Unmarshal(request.Imp[0].Ext, &params); err != nil {
		return "", fmt.Errorf("failed to parse onetag params: %w", err)
	}

	// Support both publisherId and pubId parameter names
	publisherID := params.PublisherID
	if publisherID == "" {
		publisherID = params.PubID
	}

	if publisherID == "" {
		return "", fmt.Errorf("publisherId is required")
	}

	// Validate that all impressions use the same publisher ID
	for i := 1; i < len(request.Imp); i++ {
		var p OneTagParams
		if err := json.Unmarshal(request.Imp[i].Ext, &p); err != nil {
			return "", fmt.Errorf("failed to parse onetag params for imp %d: %w", i, err)
		}

		impPubID := p.PublisherID
		if impPubID == "" {
			impPubID = p.PubID
		}

		if impPubID != publisherID {
			return "", fmt.Errorf("all impressions must use the same publisherId")
		}
	}

	return publisherID, nil
}

// MakeBids parses OneTag responses into bids
func (a *Adapter) MakeBids(request *openrtb.BidRequest, responseData *adapters.ResponseData) (*adapters.BidderResponse, []error) {
	if responseData.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	if responseData.StatusCode == http.StatusBadRequest {
		return nil, []error{fmt.Errorf("bad request: %s", string(responseData.Body))}
	}

	if responseData.StatusCode != http.StatusOK {
		return nil, []error{fmt.Errorf("unexpected status: %d", responseData.StatusCode)}
	}

	var bidResp openrtb.BidResponse
	if err := json.Unmarshal(responseData.Body, &bidResp); err != nil {
		return nil, []error{fmt.Errorf("failed to parse response: %w", err)}
	}

	response := &adapters.BidderResponse{
		Currency:   bidResp.Cur,
		ResponseID: bidResp.ID,
		Bids:       make([]*adapters.TypedBid, 0),
	}

	// Build impression map once for O(1) lookups instead of O(n) per bid
	impMap := adapters.BuildImpMap(request.Imp)

	for _, seatBid := range bidResp.SeatBid {
		for i := range seatBid.Bid {
			bid := &seatBid.Bid[i]
			bidType := adapters.GetBidTypeFromMap(bid, impMap)

			response.Bids = append(response.Bids, &adapters.TypedBid{
				Bid:     bid,
				BidType: bidType,
			})
		}
	}

	return response, nil
}

// Info returns bidder information
func Info() adapters.BidderInfo {
	return adapters.BidderInfo{
		Enabled: true,
		Maintainer: &adapters.MaintainerInfo{
			Email: "devops@onetag.com",
		},
		Capabilities: &adapters.CapabilitiesInfo{
			Site: &adapters.PlatformInfo{
				MediaTypes: []adapters.BidType{
					adapters.BidTypeBanner,
					adapters.BidTypeVideo,
					adapters.BidTypeNative,
				},
			},
			App: &adapters.PlatformInfo{
				MediaTypes: []adapters.BidType{
					adapters.BidTypeBanner,
					adapters.BidTypeVideo,
					adapters.BidTypeNative,
				},
			},
		},
		GVLVendorID: 241,
		Endpoint:    defaultEndpoint,
		DemandType:  adapters.DemandTypePublisher, // Publisher demand (shown transparently)
	}
}

func init() {
	if err := adapters.RegisterAdapter("onetag", New(""), Info()); err != nil {
		logger.Log.Error().Err(err).Str("adapter", "onetag").Msg("failed to register adapter")
	}
}
