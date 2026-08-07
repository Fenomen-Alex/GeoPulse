package spatial

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Place is a geocoding result: a human-readable label plus WGS84 coordinates.
type Place struct {
	Lat   float64 `json:"lat"`
	Lng   float64 `json:"lng"`
	Label string  `json:"label"`
}

// Geocoder resolves free-text queries (cities, streets, addresses) into
// coordinates. Implemented by NominatimClient; stubbed in tests.
type Geocoder interface {
	Search(query string, limit int) ([]Place, error)
}

// NominatimClient proxies geocoding requests to the free OpenStreetMap
// Nominatim API. No API key is required, but the service requires a valid
// User-Agent and discourages heavy client-side use, so all traffic flows
// through the backend.
type NominatimClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewNominatimClient builds a geocoder backed by the public Nominatim API.
func NewNominatimClient() *NominatimClient {
	return &NominatimClient{
		baseURL:    "https://nominatim.openstreetmap.org",
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Search geocodes a free-text query and returns up to limit results.
func (c *NominatimClient) Search(query string, limit int) ([]Place, error) {
	if limit <= 0 || limit > 20 {
		limit = 10
	}

	params := url.Values{}
	params.Set("q", query)
	params.Set("format", "jsonv2")
	params.Set("addressdetails", "1")
	params.Set("limit", strconv.Itoa(limit))
	params.Set("accept-language", "en")

	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/search?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("create geocode request: %w", err)
	}
	req.Header.Set("User-Agent", "GeoPulse/1.0 (spatial workbench; admin@geopulse.io)")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("geocode request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("geocode API returned status %d", resp.StatusCode)
	}

	var raw []nominatimResult
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode geocode response: %w", err)
	}

	places := make([]Place, 0, len(raw))
	for _, r := range raw {
		lat, err1 := strconv.ParseFloat(r.Lat, 64)
		lng, err2 := strconv.ParseFloat(r.Lon, 64)
		if err1 != nil || err2 != nil {
			continue
		}
		places = append(places, Place{Lat: lat, Lng: lng, Label: r.DisplayName})
	}
	return places, nil
}

// --- Nominatim JSON response shape ---

type nominatimResult struct {
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	DisplayName string `json:"display_name"`
}
