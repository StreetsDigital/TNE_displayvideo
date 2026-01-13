package unruly

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/thenexusengine/tne_springwire/internal/adapters"
	"github.com/thenexusengine/tne_springwire/internal/openrtb"
)

func TestNew(t *testing.T) {
	// Test with default endpoint
	adapter := New("")
	if adapter.SimpleAdapter.Endpoint != defaultEndpoint {
		t.Errorf("Expected default endpoint %s, got %s", defaultEndpoint, adapter.SimpleAdapter.Endpoint)
	}

	// Test with custom endpoint
	customEndpoint := "https://custom.endpoint.com"
	adapter = New(customEndpoint)
	if adapter.SimpleAdapter.Endpoint != customEndpoint {
		t.Errorf("Expected custom endpoint %s, got %s", customEndpoint, adapter.SimpleAdapter.Endpoint)
	}
}

func TestMakeRequests(t *testing.T) {
	adapter := New("")

	request := &openrtb.BidRequest{
		ID: "test-request-1",
		Imp: []openrtb.Imp{
			{
				ID: "imp-1",
				Video: &openrtb.Video{
					W:     640,
					H:     480,
					Mimes: []string{"video/mp4"},
				},
			},
		},
		Site: &openrtb.Site{
			Domain: "example.com",
			Page:   "https://example.com/page",
		},
	}

	requests, errs := adapter.MakeRequests(request, nil)

	if len(errs) > 0 {
		t.Fatalf("Unexpected errors: %v", errs)
	}

	if len(requests) != 1 {
		t.Fatalf("Expected 1 request, got %d", len(requests))
	}

	req := requests[0]
	if req.Method != "POST" {
		t.Errorf("Expected POST method, got %s", req.Method)
	}

	if req.URI != defaultEndpoint {
		t.Errorf("Expected URI %s, got %s", defaultEndpoint, req.URI)
	}

	// Verify request body is valid JSON
	var parsed openrtb.BidRequest
	if err := json.Unmarshal(req.Body, &parsed); err != nil {
		t.Errorf("Request body is not valid JSON: %v", err)
	}

	if parsed.ID != "test-request-1" {
		t.Errorf("Expected request ID 'test-request-1', got '%s'", parsed.ID)
	}
}

func TestMakeBids_VideoSuccess(t *testing.T) {
	adapter := New("")

	request := &openrtb.BidRequest{
		ID: "test-request-1",
		Imp: []openrtb.Imp{
			{
				ID: "imp-1",
				Video: &openrtb.Video{
					W:     640,
					H:     480,
					Mimes: []string{"video/mp4"},
				},
			},
		},
	}

	responseBody := `{
		"id": "response-1",
		"cur": "USD",
		"seatbid": [{
			"bid": [{
				"id": "bid-1",
				"impid": "imp-1",
				"price": 2.50,
				"adm": "<VAST></VAST>",
				"w": 640,
				"h": 480
			}]
		}]
	}`

	response := &adapters.ResponseData{
		StatusCode: http.StatusOK,
		Body:       []byte(responseBody),
	}

	bidderResponse, errs := adapter.MakeBids(request, response)

	if len(errs) > 0 {
		t.Fatalf("Unexpected errors: %v", errs)
	}

	if bidderResponse == nil {
		t.Fatal("Expected bidder response, got nil")
	}

	if bidderResponse.Currency != "USD" {
		t.Errorf("Expected currency USD, got %s", bidderResponse.Currency)
	}

	if len(bidderResponse.Bids) != 1 {
		t.Fatalf("Expected 1 bid, got %d", len(bidderResponse.Bids))
	}

	bid := bidderResponse.Bids[0]
	if bid.Bid.Price != 2.50 {
		t.Errorf("Expected bid price 2.50, got %f", bid.Bid.Price)
	}

	if bid.BidType != adapters.BidTypeVideo {
		t.Errorf("Expected video bid type, got %v", bid.BidType)
	}
}

func TestMakeBids_NoContent(t *testing.T) {
	adapter := New("")

	response := &adapters.ResponseData{
		StatusCode: http.StatusNoContent,
		Body:       nil,
	}

	bidderResponse, errs := adapter.MakeBids(&openrtb.BidRequest{}, response)

	if len(errs) > 0 {
		t.Fatalf("Unexpected errors: %v", errs)
	}

	if bidderResponse != nil {
		t.Error("Expected nil response for NoContent")
	}
}

func TestMakeBids_BadRequest(t *testing.T) {
	adapter := New("")

	response := &adapters.ResponseData{
		StatusCode: http.StatusBadRequest,
		Body:       []byte("invalid request"),
	}

	_, errs := adapter.MakeBids(&openrtb.BidRequest{}, response)

	if len(errs) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(errs))
	}
}

func TestMakeBids_InvalidJSON(t *testing.T) {
	adapter := New("")

	response := &adapters.ResponseData{
		StatusCode: http.StatusOK,
		Body:       []byte("not json"),
	}

	_, errs := adapter.MakeBids(&openrtb.BidRequest{}, response)

	if len(errs) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(errs))
	}
}

func TestInfo(t *testing.T) {
	info := Info()

	if !info.Enabled {
		t.Error("Expected adapter to be enabled")
	}

	if info.GVLVendorID != gvlVendorID {
		t.Errorf("Expected GVL vendor ID %d, got %d", gvlVendorID, info.GVLVendorID)
	}

	if info.Endpoint != defaultEndpoint {
		t.Errorf("Expected endpoint %s, got %s", defaultEndpoint, info.Endpoint)
	}

	if info.Capabilities == nil {
		t.Fatal("Expected capabilities to be set")
	}

	if info.Capabilities.Site == nil {
		t.Fatal("Expected site capabilities")
	}

	// Check video support
	hasVideo := false
	for _, mt := range info.Capabilities.Site.MediaTypes {
		if mt == adapters.BidTypeVideo {
			hasVideo = true
			break
		}
	}
	if !hasVideo {
		t.Error("Expected video support in site capabilities")
	}
}
