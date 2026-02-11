// Package rubicon implements the Rubicon/Magnite bidder adapter
package rubicon

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/thenexusengine/tne_springwire/internal/adapters"
	"github.com/thenexusengine/tne_springwire/internal/openrtb"
	"github.com/thenexusengine/tne_springwire/pkg/logger"
)

const (
	defaultEndpoint = "https://prebid-server.rubiconproject.com/openrtb2/auction"
)

// Adapter implements the Rubicon bidder
type Adapter struct {
	endpoint string
}

// New creates a new Rubicon adapter
func New(endpoint string) *Adapter {
	if endpoint == "" {
		endpoint = defaultEndpoint
	}
	return &Adapter{endpoint: endpoint}
}

// MakeRequests builds HTTP requests for Rubicon
func (a *Adapter) MakeRequests(request *openrtb.BidRequest, extraInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	var errors []error
	requests := make([]*adapters.RequestData, 0, len(request.Imp))

	logger.Log.Debug().
		Str("adapter", "rubicon").
		Int("impressions", len(request.Imp)).
		Str("request_id", request.ID).
		Msg("Rubicon MakeRequests called")

	// Rubicon requires one request per impression
	for _, imp := range request.Imp {
		reqCopy := *request
		reqCopy.Imp = []openrtb.Imp{imp}

		requestBody, err := json.Marshal(reqCopy)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to marshal request for imp %s: %w", imp.ID, err))
			continue
		}

		headers := http.Header{}
		headers.Set("Content-Type", "application/json;charset=utf-8")
		headers.Set("Accept", "application/json")

		requests = append(requests, &adapters.RequestData{
			Method:  "POST",
			URI:     a.endpoint,
			Body:    requestBody,
			Headers: headers,
		})

		logger.Log.Debug().
			Str("adapter", "rubicon").
			Str("imp_id", imp.ID).
			Str("endpoint", a.endpoint).
			Int("body_size", len(requestBody)).
			Msg("Rubicon request created")
	}

	logger.Log.Debug().
		Str("adapter", "rubicon").
		Int("requests_created", len(requests)).
		Int("errors", len(errors)).
		Msg("Rubicon MakeRequests completed")

	return requests, errors
}

// MakeBids parses Rubicon responses into bids
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

	// Log the raw response for debugging
	logger.Log.Debug().
		Str("adapter", "rubicon").
		Int("status_code", responseData.StatusCode).
		Int("body_size", len(responseData.Body)).
		Str("raw_response", string(responseData.Body)).
		Msg("Rubicon raw HTTP response")

	var bidResp openrtb.BidResponse
	if err := json.Unmarshal(responseData.Body, &bidResp); err != nil {
		return nil, []error{fmt.Errorf("failed to parse response: %w", err)}
	}

	logger.Log.Debug().
		Str("adapter", "rubicon").
		Str("response_id", bidResp.ID).
		Str("currency", bidResp.Cur).
		Int("seatbids", len(bidResp.SeatBid)).
		Msg("Rubicon parsed response")

	response := &adapters.BidderResponse{
		Currency:   bidResp.Cur,
		ResponseID: bidResp.ID, // P1-1: Include ResponseID for validation
		Bids:       make([]*adapters.TypedBid, 0),
	}

	// P2-3: Build impression map once for O(1) lookups instead of O(n) per bid
	impMap := adapters.BuildImpMap(request.Imp)

	for _, seatBid := range bidResp.SeatBid {
		for i := range seatBid.Bid {
			bid := &seatBid.Bid[i]
			bidType := adapters.GetBidTypeFromMap(bid, impMap)

			logger.Log.Debug().
				Str("adapter", "rubicon").
				Str("bid_id", bid.ID).
				Str("imp_id", bid.ImpID).
				Float64("price", bid.Price).
				Str("currency", bidResp.Cur).
				Str("creative_id", bid.CRID).
				Str("deal_id", bid.DealID).
				Int("width", bid.W).
				Int("height", bid.H).
				Str("bid_type", string(bidType)).
				Msg("Rubicon bid details")

			response.Bids = append(response.Bids, &adapters.TypedBid{
				Bid:     bid,
				BidType: bidType,
			})
		}
	}

	logger.Log.Debug().
		Str("adapter", "rubicon").
		Int("total_bids", len(response.Bids)).
		Msg("Rubicon MakeBids completed")

	return response, nil
}

// Info returns bidder information
func Info() adapters.BidderInfo {
	return adapters.BidderInfo{
		Enabled: true,
		Maintainer: &adapters.MaintainerInfo{
			Email: "header-bidding@rubiconproject.com",
		},
		Capabilities: &adapters.CapabilitiesInfo{
			Site: &adapters.PlatformInfo{
				MediaTypes: []adapters.BidType{
					adapters.BidTypeBanner,
					adapters.BidTypeVideo,
				},
			},
			App: &adapters.PlatformInfo{
				MediaTypes: []adapters.BidType{
					adapters.BidTypeBanner,
					adapters.BidTypeVideo,
				},
			},
		},
		GVLVendorID: 52,
		Endpoint:    defaultEndpoint,
		DemandType:  adapters.DemandTypePlatform, // Platform demand (obfuscated as "thenexusengine")
	}
}

func init() {
	if err := adapters.RegisterAdapter("rubicon", New(""), Info()); err != nil {
		logger.Log.Error().Err(err).Str("adapter", "rubicon").Msg("failed to register adapter")
	}
}
