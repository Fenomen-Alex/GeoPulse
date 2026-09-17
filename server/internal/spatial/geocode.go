package spatial

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
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

// CachingGeocoder wraps a Geocoder with a small in-memory LRU result cache
// keyed by normalized query (+limit). Street addresses and city coordinates
// are stable over days, so repeated search-bar keystrokes and duplicate
// queries never touch the upstream service again.
type CachingGeocoder struct {
	inner Geocoder
	ttl   time.Duration
	max   int

	mu    sync.Mutex
	cache map[string]cacheEntry
}

type cacheEntry struct {
	places    []Place
	expiresAt time.Time
}

// NewCachingGeocoder wraps inner with a cache holding up to max queries for ttl.
func NewCachingGeocoder(inner Geocoder, max int, ttl time.Duration) *CachingGeocoder {
	return &CachingGeocoder{
		inner: inner,
		ttl:   ttl,
		max:   max,
		cache: make(map[string]cacheEntry),
	}
}

func (c *CachingGeocoder) Search(query string, limit int) ([]Place, error) {
	if limit <= 0 || limit > 20 {
		limit = 10
	}
	key := normalizeQuery(query) + "|" + strconv.Itoa(limit)
	now := time.Now()

	c.mu.Lock()
	if e, ok := c.cache[key]; ok {
		if now.Before(e.expiresAt) {
			places := e.places
			c.mu.Unlock()
			return places, nil
		}
		delete(c.cache, key)
	}
	c.mu.Unlock()

	places, err := c.inner.Search(query, limit)

	c.mu.Lock()
	defer c.mu.Unlock()
	if err == nil {
		c.cache[key] = cacheEntry{places: places, expiresAt: now.Add(c.ttl)}
		if len(c.cache) > c.max {
			c.evictExpired(now)
		}
		if len(c.cache) > c.max {
			c.evictOldest()
		}
	}
	return places, err
}

// evictExpired drops stale entries (also amortizes cleanup so a full scan is
// rare). Callers must hold c.mu.
func (c *CachingGeocoder) evictExpired(now time.Time) {
	for k, e := range c.cache {
		if !now.Before(e.expiresAt) {
			delete(c.cache, k)
		}
	}
}

// evictOldest drops the single least-recently-used entry to make room.
// Callers must hold c.mu.
func (c *CachingGeocoder) evictOldest() {
	var oldestKey string
	var oldestExp time.Time
	for k, e := range c.cache {
		if oldestKey == "" || e.expiresAt.Before(oldestExp) {
			oldestKey = k
			oldestExp = e.expiresAt
		}
	}
	if oldestKey != "" {
		delete(c.cache, oldestKey)
	}
}

var _ Geocoder = (*CachingGeocoder)(nil)
