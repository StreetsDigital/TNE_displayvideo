package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestNewBidderStore(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	defer db.Close()

	store := NewBidderStore(db)
	if store == nil {
		t.Fatal("Expected store to be created")
	}
	if store.db != db {
		t.Error("Expected store to use provided DB")
	}
}

func TestBidderStore_GetByCode_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	defer db.Close()

	store := NewBidderStore(db)
	ctx := context.Background()

	expectedBidder := createTestBidder("appnexus")
	httpHeadersJSON, _ := json.Marshal(expectedBidder.HTTPHeaders)

	rows := sqlmock.NewRows([]string{
		"id", "bidder_code", "bidder_name", "endpoint_url", "timeout_ms",
		"enabled", "status", "supports_banner", "supports_video", "supports_native", "supports_audio",
		"gvl_vendor_id", "http_headers", "description", "documentation_url", "contact_email",
		"created_at", "updated_at",
	}).AddRow(
		expectedBidder.ID,
		expectedBidder.BidderCode,
		expectedBidder.BidderName,
		expectedBidder.EndpointURL,
		expectedBidder.TimeoutMs,
		expectedBidder.Enabled,
		expectedBidder.Status,
		expectedBidder.SupportsBanner,
		expectedBidder.SupportsVideo,
		expectedBidder.SupportsNative,
		expectedBidder.SupportsAudio,
		expectedBidder.GVLVendorID,
		httpHeadersJSON,
		expectedBidder.Description,
		expectedBidder.DocumentationURL,
		expectedBidder.ContactEmail,
		expectedBidder.CreatedAt,
		expectedBidder.UpdatedAt,
	)

	mock.ExpectQuery("SELECT (.+) FROM bidders WHERE bidder_code").
		WithArgs("appnexus").
		WillReturnRows(rows)

	bidder, err := store.GetByCode(ctx, "appnexus")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if bidder == nil {
		t.Fatal("Expected bidder to be returned")
	}

	if bidder.BidderCode != "appnexus" {
		t.Errorf("Expected bidder_code 'appnexus', got '%s'", bidder.BidderCode)
	}
	if bidder.BidderName != "AppNexus" {
		t.Errorf("Expected bidder_name 'AppNexus', got '%s'", bidder.BidderName)
	}
	if bidder.EndpointURL != "https://ib.adnxs.com/openrtb2" {
		t.Errorf("Expected endpoint_url, got '%s'", bidder.EndpointURL)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestBidderStore_GetByCode_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	defer db.Close()

	store := NewBidderStore(db)
	ctx := context.Background()

	mock.ExpectQuery("SELECT (.+) FROM bidders WHERE bidder_code").
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	bidder, err := store.GetByCode(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("Expected no error for not found, got %v", err)
	}

	if bidder != nil {
		t.Error("Expected nil bidder for not found")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestBidderStore_GetByCode_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	defer db.Close()

	store := NewBidderStore(db)
	ctx := context.Background()

	expectedErr := errors.New("database connection failed")

	mock.ExpectQuery("SELECT (.+) FROM bidders WHERE bidder_code").
		WithArgs("appnexus").
		WillReturnError(expectedErr)

	bidder, err := store.GetByCode(ctx, "appnexus")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if bidder != nil {
		t.Error("Expected nil bidder on error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestBidderStore_GetByCode_InvalidJSON(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	defer db.Close()

	store := NewBidderStore(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{
		"id", "bidder_code", "bidder_name", "endpoint_url", "timeout_ms",
		"enabled", "status", "supports_banner", "supports_video", "supports_native", "supports_audio",
		"gvl_vendor_id", "http_headers", "description", "documentation_url", "contact_email",
		"created_at", "updated_at",
	}).AddRow(
		"1", "appnexus", "AppNexus", "https://example.com", 500,
		true, "active", true, true, false, false,
		nil, []byte("invalid json{"), "", "", "",
		time.Now(), time.Now(),
	)

	mock.ExpectQuery("SELECT (.+) FROM bidders WHERE bidder_code").
		WithArgs("appnexus").
		WillReturnRows(rows)

	bidder, err := store.GetByCode(ctx, "appnexus")
	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}

	if bidder != nil {
		t.Error("Expected nil bidder on JSON error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestBidderStore_ListActive_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	defer db.Close()

	store := NewBidderStore(db)
	ctx := context.Background()

	bidder1 := createTestBidder("appnexus")
	bidder2 := createTestBidder("rubicon")

	headers1, _ := json.Marshal(bidder1.HTTPHeaders)
	headers2, _ := json.Marshal(bidder2.HTTPHeaders)

	rows := sqlmock.NewRows([]string{
		"id", "bidder_code", "bidder_name", "endpoint_url", "timeout_ms",
		"enabled", "status", "supports_banner", "supports_video", "supports_native", "supports_audio",
		"gvl_vendor_id", "http_headers", "description", "documentation_url", "contact_email",
		"created_at", "updated_at",
	}).
		AddRow(
			bidder1.ID, bidder1.BidderCode, bidder1.BidderName, bidder1.EndpointURL, bidder1.TimeoutMs,
			bidder1.Enabled, bidder1.Status, bidder1.SupportsBanner, bidder1.SupportsVideo, bidder1.SupportsNative, bidder1.SupportsAudio,
			bidder1.GVLVendorID, headers1, bidder1.Description, bidder1.DocumentationURL, bidder1.ContactEmail,
			bidder1.CreatedAt, bidder1.UpdatedAt,
		).
		AddRow(
			bidder2.ID, bidder2.BidderCode, bidder2.BidderName, bidder2.EndpointURL, bidder2.TimeoutMs,
			bidder2.Enabled, bidder2.Status, bidder2.SupportsBanner, bidder2.SupportsVideo, bidder2.SupportsNative, bidder2.SupportsAudio,
			bidder2.GVLVendorID, headers2, bidder2.Description, bidder2.DocumentationURL, bidder2.ContactEmail,
			bidder2.CreatedAt, bidder2.UpdatedAt,
		)

	mock.ExpectQuery("SELECT (.+) FROM bidders WHERE enabled").
		WillReturnRows(rows)

	bidders, err := store.ListActive(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(bidders) != 2 {
		t.Fatalf("Expected 2 bidders, got %d", len(bidders))
	}

	if bidders[0].BidderCode != "appnexus" {
		t.Errorf("Expected first bidder 'appnexus', got '%s'", bidders[0].BidderCode)
	}
	if bidders[1].BidderCode != "rubicon" {
		t.Errorf("Expected second bidder 'rubicon', got '%s'", bidders[1].BidderCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestBidderStore_ListActive_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	defer db.Close()

	store := NewBidderStore(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{
		"id", "bidder_code", "bidder_name", "endpoint_url", "timeout_ms",
		"enabled", "status", "supports_banner", "supports_video", "supports_native", "supports_audio",
		"gvl_vendor_id", "http_headers", "description", "documentation_url", "contact_email",
		"created_at", "updated_at",
	})

	mock.ExpectQuery("SELECT (.+) FROM bidders WHERE enabled").
		WillReturnRows(rows)

	bidders, err := store.ListActive(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(bidders) != 0 {
		t.Errorf("Expected 0 bidders, got %d", len(bidders))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestBidderStore_ListActive_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	defer db.Close()

	store := NewBidderStore(db)
	ctx := context.Background()

	expectedErr := errors.New("database error")

	mock.ExpectQuery("SELECT (.+) FROM bidders WHERE enabled").
		WillReturnError(expectedErr)

	bidders, err := store.ListActive(ctx)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if bidders != nil {
		t.Error("Expected nil bidders on error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestBidderStore_ListActive_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	defer db.Close()

	store := NewBidderStore(db)
	ctx := context.Background()

	// Create a row with invalid type for timeout_ms (string instead of int)
	rows := sqlmock.NewRows([]string{
		"id", "bidder_code", "bidder_name", "endpoint_url", "timeout_ms",
		"enabled", "status", "supports_banner", "supports_video", "supports_native", "supports_audio",
		"gvl_vendor_id", "http_headers", "description", "documentation_url", "contact_email",
		"created_at", "updated_at",
	}).AddRow(
		"1", "appnexus", "AppNexus", "https://example.com", "invalid_int",
		true, "active", true, true, false, false,
		nil, []byte("{}"), "", "", "",
		time.Now(), time.Now(),
	)

	mock.ExpectQuery("SELECT (.+) FROM bidders WHERE enabled").
		WillReturnRows(rows)

	bidders, err := store.ListActive(ctx)
	if err == nil {
		t.Fatal("Expected scan error, got nil")
	}

	if bidders != nil {
		t.Error("Expected nil bidders on scan error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestBidderStore_GetForPublisher_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	defer db.Close()

	store := NewBidderStore(db)
	ctx := context.Background()

	bidderConfig := map[string]interface{}{"placementId": "12345"}
	bidderConfigJSON, _ := json.Marshal(bidderConfig)
	httpHeaders := map[string]interface{}{"X-API-Key": "test"}
	httpHeadersJSON, _ := json.Marshal(httpHeaders)

	rows := sqlmock.NewRows([]string{
		"id", "bidder_code", "bidder_name", "endpoint_url", "timeout_ms",
		"enabled", "status", "supports_banner", "supports_video", "supports_native", "supports_audio",
		"gvl_vendor_id", "http_headers", "description", "documentation_url", "contact_email",
		"created_at", "updated_at", "publisher_id", "publisher_name", "bidder_config",
	}).AddRow(
		"1", "appnexus", "AppNexus", "https://ib.adnxs.com/openrtb2", 500,
		true, "active", true, true, false, false,
		nil, httpHeadersJSON, "AppNexus bidder", "https://example.com", "test@example.com",
		time.Now(), time.Now(), "pub123", "Test Publisher", bidderConfigJSON,
	)

	mock.ExpectQuery("SELECT (.+) FROM bidders b CROSS JOIN publishers p").
		WithArgs("pub123").
		WillReturnRows(rows)

	publisherBidders, err := store.GetForPublisher(ctx, "pub123")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(publisherBidders) != 1 {
		t.Fatalf("Expected 1 bidder, got %d", len(publisherBidders))
	}

	pb := publisherBidders[0]
	if pb.BidderCode != "appnexus" {
		t.Errorf("Expected bidder_code 'appnexus', got '%s'", pb.BidderCode)
	}
	if pb.PublisherID != "pub123" {
		t.Errorf("Expected publisher_id 'pub123', got '%s'", pb.PublisherID)
	}
	if pb.PublisherName != "Test Publisher" {
		t.Errorf("Expected publisher_name 'Test Publisher', got '%s'", pb.PublisherName)
	}
	if pb.BidderConfig["placementId"] != "12345" {
		t.Errorf("Expected placementId '12345', got '%v'", pb.BidderConfig["placementId"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestBidderStore_GetForPublisher_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	defer db.Close()

	store := NewBidderStore(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{
		"id", "bidder_code", "bidder_name", "endpoint_url", "timeout_ms",
		"enabled", "status", "supports_banner", "supports_video", "supports_native", "supports_audio",
		"gvl_vendor_id", "http_headers", "description", "documentation_url", "contact_email",
		"created_at", "updated_at", "publisher_id", "publisher_name", "bidder_config",
	})

	mock.ExpectQuery("SELECT (.+) FROM bidders b CROSS JOIN publishers p").
		WithArgs("pub123").
		WillReturnRows(rows)

	publisherBidders, err := store.GetForPublisher(ctx, "pub123")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(publisherBidders) != 0 {
		t.Errorf("Expected 0 bidders, got %d", len(publisherBidders))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestBidderStore_GetForPublisher_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	defer db.Close()

	store := NewBidderStore(db)
	ctx := context.Background()

	expectedErr := errors.New("database error")

	mock.ExpectQuery("SELECT (.+) FROM bidders b CROSS JOIN publishers p").
		WithArgs("pub123").
		WillReturnError(expectedErr)

	publisherBidders, err := store.GetForPublisher(ctx, "pub123")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if publisherBidders != nil {
		t.Error("Expected nil bidders on error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// Helper function to create test bidders
func createTestBidder(code string) *Bidder {
	now := time.Now()
	vendorID := 32

	bidder := &Bidder{
		ID:               "1",
		BidderCode:       code,
		TimeoutMs:        500,
		Enabled:          true,
		Status:           "active",
		SupportsBanner:   true,
		SupportsVideo:    true,
		SupportsNative:   false,
		SupportsAudio:    false,
		GVLVendorID:      &vendorID,
		HTTPHeaders:      map[string]interface{}{"X-API-Key": "test"},
		Description:      "Test bidder",
		DocumentationURL: "https://example.com/docs",
		ContactEmail:     "test@example.com",
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	switch code {
	case "appnexus":
		bidder.BidderName = "AppNexus"
		bidder.EndpointURL = "https://ib.adnxs.com/openrtb2"
	case "rubicon":
		bidder.BidderName = "Rubicon Project"
		bidder.EndpointURL = "https://fastlane.rubiconproject.com/a/api/fastlane.json"
	default:
		bidder.BidderName = "Test Bidder"
		bidder.EndpointURL = "https://example.com/openrtb2"
	}

	return bidder
}
