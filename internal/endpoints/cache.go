// Package endpoints provides HTTP endpoint handlers
package endpoints

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/thenexusengine/tne_springwire/pkg/logger"
	"github.com/thenexusengine/tne_springwire/pkg/redis"
)

// maxCacheRequestSize limits cache request body reads (10MB)
const maxCacheRequestSize = 10 * 1024 * 1024

// maxCachePuts limits the number of items in a single cache request
const maxCachePuts = 100

// cacheKeyPrefix namespaces Prebid Cache entries in Redis
const cacheKeyPrefix = "prebid_cache:"

// defaultCacheTTL is the default TTL for cached items (5 minutes)
const defaultCacheTTL = 5 * time.Minute

// CacheHandler implements the Prebid Cache protocol:
//   - POST /cache: store items (VAST XML or JSON) and return UUIDs
//   - GET /cache?uuid=<uuid>: retrieve a cached item by UUID
type CacheHandler struct {
	redisClient *redis.Client
	cacheTTL    time.Duration
}

// NewCacheHandler creates a new Prebid Cache handler
func NewCacheHandler(redisClient *redis.Client) *CacheHandler {
	return &CacheHandler{
		redisClient: redisClient,
		cacheTTL:    defaultCacheTTL,
	}
}

// cachePutRequest is the incoming request format for POST /cache
type cachePutRequest struct {
	Puts []cachePutItem `json:"puts"`
}

// cachePutItem represents a single item to cache
type cachePutItem struct {
	Type  string          `json:"type"`  // "xml" or "json"
	Value json.RawMessage `json:"value"` // the content to cache
	TTL   int             `json:"ttlseconds,omitempty"`
}

// cachePutResponse is the response format for POST /cache
type cachePutResponse struct {
	Responses []cachePutResponseItem `json:"responses"`
}

// cachePutResponseItem contains the UUID assigned to a cached item
type cachePutResponseItem struct {
	UUID string `json:"uuid"`
}

// ServeHTTP routes to the appropriate handler based on HTTP method
func (h *CacheHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.handlePut(w, r)
	case http.MethodGet:
		h.handleGet(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handlePut stores items in the cache and returns UUIDs
func (h *CacheHandler) handlePut(w http.ResponseWriter, r *http.Request) {
	log := logger.Log

	body, err := io.ReadAll(io.LimitReader(r.Body, maxCacheRequestSize))
	if err != nil {
		log.Error().Err(err).Msg("Failed to read cache request body")
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req cachePutRequest
	if err := json.Unmarshal(body, &req); err != nil {
		log.Error().Err(err).Msg("Failed to parse cache request")
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(req.Puts) == 0 {
		http.Error(w, "No items to cache", http.StatusBadRequest)
		return
	}

	if len(req.Puts) > maxCachePuts {
		http.Error(w, fmt.Sprintf("Too many items (max %d)", maxCachePuts), http.StatusBadRequest)
		return
	}

	resp := cachePutResponse{
		Responses: make([]cachePutResponseItem, 0, len(req.Puts)),
	}

	ctx := r.Context()

	for _, item := range req.Puts {
		uuid, err := generateUUID()
		if err != nil {
			log.Error().Err(err).Msg("Failed to generate UUID for cache")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Determine TTL: use item-specific TTL if provided, otherwise default
		ttl := h.cacheTTL
		if item.TTL > 0 {
			ttl = time.Duration(item.TTL) * time.Second
		}

		// Build the cached value with content type metadata
		var cacheValue string
		switch item.Type {
		case "xml":
			// For XML, store the raw string value (strip JSON quotes if present)
			var xmlStr string
			if err := json.Unmarshal(item.Value, &xmlStr); err != nil {
				// If it's not a JSON string, store the raw bytes
				cacheValue = string(item.Value)
			} else {
				cacheValue = xmlStr
			}
		case "json":
			cacheValue = string(item.Value)
		default:
			cacheValue = string(item.Value)
		}

		// Store as a JSON envelope with type info so GET knows what content-type to use
		envelope := map[string]string{
			"type":  item.Type,
			"value": cacheValue,
		}
		envelopeBytes, err := json.Marshal(envelope)
		if err != nil {
			log.Error().Err(err).Msg("Failed to marshal cache envelope")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		key := cacheKeyPrefix + uuid
		if err := h.redisClient.Set(ctx, key, string(envelopeBytes), ttl); err != nil {
			log.Error().Err(err).Str("uuid", uuid).Msg("Failed to store in cache")
			http.Error(w, "Cache storage error", http.StatusInternalServerError)
			return
		}

		resp.Responses = append(resp.Responses, cachePutResponseItem{UUID: uuid})

		log.Debug().
			Str("uuid", uuid).
			Str("type", item.Type).
			Dur("ttl", ttl).
			Msg("Cached item stored")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleGet retrieves a cached item by UUID
func (h *CacheHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	log := logger.Log

	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		http.Error(w, "uuid parameter is required", http.StatusBadRequest)
		return
	}

	// Validate UUID format (hex chars only, reasonable length)
	if len(uuid) > 64 || !isHexString(uuid) {
		http.Error(w, "Invalid uuid format", http.StatusBadRequest)
		return
	}

	key := cacheKeyPrefix + uuid
	result, err := h.redisClient.Get(r.Context(), key)
	if err != nil {
		log.Error().Err(err).Str("uuid", uuid).Msg("Cache retrieval error")
		http.Error(w, "Cache error", http.StatusInternalServerError)
		return
	}

	if result == "" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	// Parse the envelope to determine content type
	var envelope map[string]string
	if err := json.Unmarshal([]byte(result), &envelope); err != nil {
		// Fallback: return raw content as application/octet-stream
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write([]byte(result))
		return
	}

	value := envelope["value"]

	switch envelope["type"] {
	case "xml":
		w.Header().Set("Content-Type", "application/xml")
	case "json":
		w.Header().Set("Content-Type", "application/json")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	w.Write([]byte(value))
}

// generateUUID creates a random hex string suitable for cache UUIDs
func generateUUID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// isHexString checks if a string contains only hexadecimal characters
func isHexString(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
