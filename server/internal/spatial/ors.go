package spatial

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Point is a WGS84 geographic coordinate.
type Point struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// Step is a single turn-by-turn navigation instruction along a route.
type Step struct {
	Instruction string    `json:"instruction"`
	Name        string    `json:"name"`
	Distance    float64   `json:"distance"`
	Duration    float64   `json:"duration"`
	Type        int       `json:"type"`
	Location    []float64 `json:"location"`
}

// Route is the result of a street-navigation directions query.
type Route struct {
	DistanceM   float64     `json:"distance_m"`
	DurationS   float64     `json:"duration_s"`
	Coordinates [][]float64 `json:"coordinates"`
	Steps       []Step      `json:"steps"`
}

var profileByMode = map[string]string{
	"walk":  "foot-walking",
	"bike":  "cycling-regular",
	"drive": "driving-car",
}

// ORSClient proxies street-navigation requests to the OpenRouteService
// Directions API. Upstream credentials never leave the backend.
type ORSClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	cache      *routeCache
}

// NewORSClient builds an ORS client. baseURL defaults to the hosted API.
func NewORSClient(apiKey string) *ORSClient {
	return &ORSClient{
		apiKey:     apiKey,
		baseURL:    "https://api.openrouteservice.org",
		httpClient: &http.Client{Timeout: 15 * time.Second},
		cache:      newRouteCache(),
	}
}

// Directions computes a street route from start to end for the given mode.
func (c *ORSClient) Directions(start, end Point, mode string) (*Route, error) {
	profile, ok := profileByMode[mode]
	if !ok {
		profile = profileByMode["walk"]
	}

	cacheKey := fmt.Sprintf("%s|%.6f,%.6f|%.6f,%.6f", profile, start.Lat, start.Lng, end.Lat, end.Lng)
	if hit, ok := c.cache.get(cacheKey); ok {
		return hit, nil
	}

	body := map[string]interface{}{
		"coordinates":  [][]float64{{start.Lng, start.Lat}, {end.Lng, end.Lat}},
		"instructions": true,
		"maneuvers":    true,
		"language":     "en",
	}
	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal directions body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/v2/directions/"+profile+"/geojson", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("create directions request: %w", err)
	}
	req.Header.Set("Authorization", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/geo+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("directions request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("directions API returned status %d", resp.StatusCode)
	}

	var fc orsFeatureCollection
	if err := json.NewDecoder(resp.Body).Decode(&fc); err != nil {
		return nil, fmt.Errorf("decode directions response: %w", err)
	}

	if len(fc.Features) == 0 {
		return nil, fmt.Errorf("directions response contains no route features")
	}

	feature := fc.Features[0]

	route := &Route{
		DistanceM:   feature.Properties.Summary.Distance,
		DurationS:   feature.Properties.Summary.Duration,
		Coordinates: feature.Geometry.Coordinates,
	}

	for _, seg := range feature.Properties.Segments {
		for _, s := range seg.Steps {
			route.Steps = append(route.Steps, Step{
				Instruction: s.Instruction,
				Name:        s.Name,
				Distance:    s.Distance,
				Duration:    s.Duration,
				Type:        s.Type,
				Location:    s.Maneuver.Location,
			})
		}
	}

	c.cache.set(cacheKey, route)
	return route, nil
}

// --- OpenRouteService GeoJSON response shape ---

type orsFeatureCollection struct {
	Type     string       `json:"type"`
	Features []orsFeature `json:"features"`
}

type orsFeature struct {
	Type       string        `json:"type"`
	Geometry   orsGeometry   `json:"geometry"`
	Properties orsProperties `json:"properties"`
}

type orsGeometry struct {
	Type        string      `json:"type"`
	Coordinates [][]float64 `json:"coordinates"`
}

type orsProperties struct {
	Summary  orsSummary   `json:"summary"`
	Segments []orsSegment `json:"segments"`
}

type orsSummary struct {
	Distance float64 `json:"distance"`
	Duration float64 `json:"duration"`
}

type orsSegment struct {
	Steps []orsStep `json:"steps"`
}

type orsStep struct {
	Distance    float64     `json:"distance"`
	Duration    float64     `json:"duration"`
	Type        int         `json:"type"`
	Instruction string      `json:"instruction"`
	Name        string      `json:"name"`
	Maneuver    orsManeuver `json:"maneuver"`
}

type orsManeuver struct {
	Location []float64 `json:"location"`
}

// --- small in-memory TTL cache ---

type routeCache struct {
	mu    sync.Mutex
	items map[string]routeCacheEntry
}

type routeCacheEntry struct {
	route   *Route
	expires time.Time
}

func newRouteCache() *routeCache {
	return &routeCache{items: make(map[string]routeCacheEntry)}
}

func (c *routeCache) get(key string) (*Route, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.items[key]
	if !ok {
		return nil, false
	}
	if time.Now().After(entry.expires) {
		delete(c.items, key)
		return nil, false
	}
	return entry.route, true
}

func (c *routeCache) set(key string, r *Route) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.items) >= 500 {
		now := time.Now()
		for k, e := range c.items {
			if now.After(e.expires) {
				delete(c.items, k)
			}
		}
	}
	c.items[key] = routeCacheEntry{route: r, expires: time.Now().Add(24 * time.Hour)}
}
