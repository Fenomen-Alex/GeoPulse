package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/alex/geopulse/server/internal/spatial"
)

// Geocoder is the free-text search backend used by the geocode handler.
type Geocoder interface {
	Search(query string, limit int) ([]spatial.Place, error)
}

type GeocodeResponse struct {
	Results []spatial.Place `json:"results"`
}

type GeocodeHandler struct {
	geocoder Geocoder
}

// NewGeocodeHandler builds a handler backed by the Nominatim client.
func NewGeocodeHandler() *GeocodeHandler {
	return &GeocodeHandler{geocoder: spatial.NewNominatimClient()}
}

// Handle resolves a free-text query into coordinate suggestions.
// GET /api/v1/geocode?q=berlin&limit=8
func (h *GeocodeHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("q")
	if len(query) < 3 {
		http.Error(w, `{"error":"query too short"}`, http.StatusBadRequest)
		return
	}

	limit := 8
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			limit = n
		}
	}

	results, err := h.geocoder.Search(query, limit)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Search service unavailable. Please try again later.",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GeocodeResponse{Results: results})
}
