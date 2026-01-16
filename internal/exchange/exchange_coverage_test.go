package exchange

import (
	"context"
	"testing"

	"github.com/thenexusengine/tne_springwire/internal/adapters"
	"github.com/thenexusengine/tne_springwire/internal/openrtb"
	"github.com/thenexusengine/tne_springwire/pkg/idr"
)

// Mock publisher for testing bid multiplier extraction
type mockPublisherWithMultiplier struct {
	PublisherID   string
	BidMultiplier float64
}

func (m *mockPublisherWithMultiplier) GetPublisherID() string {
	return m.PublisherID
}

func (m *mockPublisherWithMultiplier) GetBidMultiplier() float64 {
	return m.BidMultiplier
}

func (m *mockPublisherWithMultiplier) GetAllowedDomains() string {
	return "example.com"
}

// TestGetDemandType_NotFound tests demand type for unknown bidders
func TestGetDemandType_NotFound(t *testing.T) {
	registry := adapters.NewRegistry()
	exchange := New(registry, nil)

	// Unknown bidder should default to platform
	demandType := exchange.getDemandType("unknown-bidder", nil)
	if demandType != adapters.DemandTypePlatform {
		t.Errorf("Expected DemandTypePlatform for unknown bidder, got %v", demandType)
	}
}

// TestBuildImpFloorMap_NoPublisher tests floor map building without publisher
func TestBuildImpFloorMap_NoPublisher(t *testing.T) {
	registry := adapters.NewRegistry()
	exchange := New(registry, nil)

	req := &openrtb.BidRequest{
		ID: "test-request",
		Imp: []openrtb.Imp{
			{ID: "imp1", BidFloor: 1.0},
			{ID: "imp2", BidFloor: 2.5},
			{ID: "imp3", BidFloor: 0.0},
		},
	}

	ctx := context.Background()
	floorMap := exchange.buildImpFloorMap(ctx, req)

	if floorMap["imp1"] != 1.0 {
		t.Errorf("Expected floor 1.0 for imp1, got %f", floorMap["imp1"])
	}
	if floorMap["imp2"] != 2.5 {
		t.Errorf("Expected floor 2.5 for imp2, got %f", floorMap["imp2"])
	}
	if floorMap["imp3"] != 0.0 {
		t.Errorf("Expected floor 0.0 for imp3, got %f", floorMap["imp3"])
	}
}

// TestExtractBidMultiplier_Interface tests multiplier extraction via interface
func TestExtractBidMultiplier_Interface(t *testing.T) {
	pub := &mockPublisherWithMultiplier{
		BidMultiplier: 1.05,
	}

	multiplier, ok := extractBidMultiplier(pub)
	if !ok {
		t.Error("Expected to extract bid multiplier")
	}
	if multiplier != 1.05 {
		t.Errorf("Expected 1.05, got %f", multiplier)
	}
}

// TestExtractBidMultiplier_NotFound tests multiplier extraction when not present
func TestExtractBidMultiplier_NotFound(t *testing.T) {
	type noBidMultiplier struct {
		SomeField string
	}
	obj := &noBidMultiplier{SomeField: "value"}

	_, ok := extractBidMultiplier(obj)
	if ok {
		t.Error("Expected not to extract bid multiplier from object without field")
	}
}

// TestExtractPublisherID_Interface tests publisher ID extraction
func TestExtractPublisherID_Interface(t *testing.T) {
	pub := &mockPublisherWithMultiplier{
		PublisherID: "pub-123",
	}

	id, ok := extractPublisherID(pub)
	if !ok {
		t.Error("Expected to extract publisher ID")
	}
	if id != "pub-123" {
		t.Errorf("Expected 'pub-123', got '%s'", id)
	}
}

// TestExtractPublisherID_EmptyID tests publisher ID extraction with empty ID
func TestExtractPublisherID_EmptyID(t *testing.T) {
	pub := &mockPublisherWithMultiplier{
		PublisherID: "",
	}

	_, ok := extractPublisherID(pub)
	if ok {
		t.Error("Expected not to extract empty publisher ID")
	}
}

// TestExtractPublisherID_NotFound tests publisher ID extraction when not present
func TestExtractPublisherID_NotFound(t *testing.T) {
	type noPublisherID struct {
		SomeField string
	}
	obj := &noPublisherID{SomeField: "value"}

	_, ok := extractPublisherID(obj)
	if ok {
		t.Error("Expected not to extract publisher ID from object without field")
	}
}

// TestBuildBidExtension_PlatformDemand tests bid extension for platform demand
func TestBuildBidExtension_PlatformDemand(t *testing.T) {
	registry := adapters.NewRegistry()
	exchange := New(registry, nil)

	vb := ValidatedBid{
		Bid: &adapters.TypedBid{
			Bid: &openrtb.Bid{
				ID:     "bid1",
				ImpID:  "imp1",
				Price:  2.5,
				W:      300,
				H:      250,
				DealID: "deal123",
			},
			BidType: adapters.BidTypeBanner,
		},
		BidderCode: "appnexus",
		DemandType: adapters.DemandTypePlatform,
	}

	ext := exchange.buildBidExtension(vb)

	if ext.Prebid == nil {
		t.Fatal("Expected non-nil Prebid extension")
	}

	// Should use "thenexusengine" for platform demand
	if ext.Prebid.Targeting["hb_bidder"] != "thenexusengine" {
		t.Errorf("Expected hb_bidder 'thenexusengine', got '%s'", ext.Prebid.Targeting["hb_bidder"])
	}

	// Should include deal ID
	if ext.Prebid.Targeting["hb_deal"] != "deal123" {
		t.Errorf("Expected hb_deal 'deal123', got '%s'", ext.Prebid.Targeting["hb_deal"])
	}

	// Should include size
	if ext.Prebid.Targeting["hb_size"] != "300x250" {
		t.Errorf("Expected hb_size '300x250', got '%s'", ext.Prebid.Targeting["hb_size"])
	}
}

// TestBuildBidExtension_PublisherDemand tests bid extension for publisher demand
func TestBuildBidExtension_PublisherDemand(t *testing.T) {
	registry := adapters.NewRegistry()
	exchange := New(registry, nil)

	vb := ValidatedBid{
		Bid: &adapters.TypedBid{
			Bid: &openrtb.Bid{
				ID:    "bid1",
				ImpID: "imp1",
				Price: 2.5,
				W:     728,
				H:     90,
			},
			BidType: adapters.BidTypeBanner,
		},
		BidderCode: "rubicon",
		DemandType: adapters.DemandTypePublisher,
	}

	ext := exchange.buildBidExtension(vb)

	if ext.Prebid == nil {
		t.Fatal("Expected non-nil Prebid extension")
	}

	// Should use original bidder code for publisher demand
	if ext.Prebid.Targeting["hb_bidder"] != "rubicon" {
		t.Errorf("Expected hb_bidder 'rubicon', got '%s'", ext.Prebid.Targeting["hb_bidder"])
	}

	// Should include size
	if ext.Prebid.Targeting["hb_size"] != "728x90" {
		t.Errorf("Expected hb_size '728x90', got '%s'", ext.Prebid.Targeting["hb_size"])
	}
}

// TestBuildBidExtension_VideoType tests bid extension for video
func TestBuildBidExtension_VideoType(t *testing.T) {
	registry := adapters.NewRegistry()
	exchange := New(registry, nil)

	vb := ValidatedBid{
		Bid: &adapters.TypedBid{
			Bid: &openrtb.Bid{
				ID:    "bid1",
				ImpID: "imp1",
				Price: 5.0,
				W:     640,
				H:     480,
			},
			BidType: adapters.BidTypeVideo,
		},
		BidderCode: "appnexus",
		DemandType: adapters.DemandTypePlatform,
	}

	ext := exchange.buildBidExtension(vb)

	if ext.Prebid == nil {
		t.Fatal("Expected non-nil Prebid extension")
	}

	if ext.Prebid.Type != "video" {
		t.Errorf("Expected type 'video', got '%s'", ext.Prebid.Type)
	}

	if ext.Prebid.Meta.MediaType != "video" {
		t.Errorf("Expected media_type 'video', got '%s'", ext.Prebid.Meta.MediaType)
	}
}

// TestSetMetrics tests setting metrics recorder
func TestSetMetrics(t *testing.T) {
	registry := adapters.NewRegistry()
	exchange := New(registry, nil)

	// Create mock metrics recorder
	metrics := &mockMetricsRecorder{}

	exchange.SetMetrics(metrics)

	// Verify it was set by checking it's not nil (we can't access private field directly)
	if exchange.metrics == nil {
		t.Error("Expected metrics to be set")
	}
}

// TestClose_WithEventRecorder tests Close with event recorder
func TestClose_WithEventRecorder(t *testing.T) {
	registry := adapters.NewRegistry()

	// Create real event recorder with empty URL (will work for Close)
	eventRecorder := idr.NewEventRecorder("", 10)

	exchange := New(registry, nil)
	exchange.eventRecorder = eventRecorder

	err := exchange.Close()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// TestClose_NoEventRecorder tests Close without event recorder
func TestClose_NoEventRecorder(t *testing.T) {
	registry := adapters.NewRegistry()
	exchange := New(registry, nil)
	exchange.eventRecorder = nil

	err := exchange.Close()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// TestBuildImpFloorMap_MultipleImpressions tests floor map with multiple impressions
func TestBuildImpFloorMap_MultipleImpressions(t *testing.T) {
	registry := adapters.NewRegistry()
	exchange := New(registry, nil)

	req := &openrtb.BidRequest{
		ID: "test-request",
		Imp: []openrtb.Imp{
			{ID: "imp1", BidFloor: 1.0},
			{ID: "imp2", BidFloor: 2.5},
			{ID: "imp3", BidFloor: 0.5},
		},
	}

	ctx := context.Background()
	floorMap := exchange.buildImpFloorMap(ctx, req)

	// Should have all impression IDs in the map
	if len(floorMap) != 3 {
		t.Errorf("Expected 3 floors in map, got %d", len(floorMap))
	}

	// Should preserve original floors when no multiplier
	if floorMap["imp1"] != 1.0 {
		t.Errorf("Expected floor 1.0 for imp1, got %f", floorMap["imp1"])
	}
	if floorMap["imp2"] != 2.5 {
		t.Errorf("Expected floor 2.5 for imp2, got %f", floorMap["imp2"])
	}
	if floorMap["imp3"] != 0.5 {
		t.Errorf("Expected floor 0.5 for imp3, got %f", floorMap["imp3"])
	}
}

// Mock implementations for testing

type mockMetricsRecorder struct{}

func (m *mockMetricsRecorder) RecordMargin(publisher, bidder, mediaType string, originalPrice, adjustedPrice, platformCut float64) {
}
func (m *mockMetricsRecorder) RecordFloorAdjustment(publisher string) {}
