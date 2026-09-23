package spatial

import (
	"encoding/json"
	"math"
)

// ModeSpeeds maps travel modes to their assumed average speed in km/h. These
// are synthetic estimates used to derive isochrone radii; the actual road
// network is approximated by a circle until a live POI/routing provider is
// wired in.
var ModeSpeeds = map[string]float64{
	"walk":  5.0,
	"bike":  15.0,
	"drive": 40.0,
}

// ValidModes is the set of accepted travel modes.
var ValidModes = map[string]bool{
	"walk":  true,
	"bike":  true,
	"drive": true,
}

// BandColors is the ordered palette used to color isochrone bands from the
// outermost to the innermost ring.
var BandColors = []struct {
	Fill        string
	Stroke      string
	FillOpacity float64
}{
	{"#10b981", "#34d399", 0.40},
	{"#f59e0b", "#fbbf24", 0.35},
	{"#ef4444", "#f87171", 0.25},
	{"#8b5cf6", "#a78bfa", 0.20},
	{"#3b82f6", "#60a5fa", 0.15},
}

// BandMinutesFor returns the set of discrete isochrone bands for a given
// maximum travel time in minutes. Bands are capped at 60 minutes.
func BandMinutesFor(maxMinutes int) []int {
	allBands := []int{5, 15, 30, 45, 60}
	result := make([]int, 0)
	for _, b := range allBands {
		if b <= maxMinutes {
			result = append(result, b)
		}
	}
	if len(result) == 0 {
		result = append(result, maxMinutes)
	}
	return result
}

// IsochroneBand is one reachability ring for a given travel time.
type IsochroneBand struct {
	Minutes     int             `json:"minutes"`
	Area        float64         `json:"area"`
	FillColor   string          `json:"fillColor"`
	StrokeColor string          `json:"strokeColor"`
	FillOpacity float64         `json:"fillOpacity"`
	GeoJSON     json.RawMessage `json:"geojson"`
}

// AnalysisResult is the full output of a reachability analysis.
type AnalysisResult struct {
	Lat            float64         `json:"lat"`
	Lng            float64         `json:"lng"`
	Mode           string          `json:"mode"`
	Minutes        int             `json:"minutes"`
	TotalArea      float64         `json:"totalArea"`
	PoiCount       int             `json:"poiCount"`
	Score          int             `json:"score"`
	PoiBreakdown   map[string]int  `json:"poiBreakdown,omitempty"`
	Bands          []IsochroneBand `json:"bands"`
}

// RunAnalysis computes a reachability analysis for the given parameters. It is
// pure (no I/O) so it can be called from handlers and tests alike.
func RunAnalysis(lat, lng float64, mode string, minutes int, poiCategories []string) *AnalysisResult {
	if minutes <= 0 || minutes > 60 {
		minutes = 15
	}
	if !ValidModes[mode] {
		mode = "walk"
	}
	speed := ModeSpeeds[mode]

	bandMinutes := BandMinutesFor(minutes)
	bands := make([]IsochroneBand, 0, len(bandMinutes))

	var totalArea float64
	for i, mins := range bandMinutes {
		radiusKm := speed * (float64(mins) / 60.0)
		area := math.Pi * radiusKm * radiusKm
		totalArea = area // outermost band area is the total reachable area

		colorIdx := i
		if colorIdx >= len(BandColors) {
			colorIdx = len(BandColors) - 1
		}

		polygon := generateCircleGeoJSON(lat, lng, radiusKm, 64)
		polyBytes, _ := json.Marshal(polygon)

		bands = append(bands, IsochroneBand{
			Minutes:     mins,
			Area:        math.Round(area*100) / 100,
			FillColor:   BandColors[colorIdx].Fill,
			StrokeColor: BandColors[colorIdx].Stroke,
			FillOpacity: BandColors[colorIdx].FillOpacity,
			GeoJSON:     polyBytes,
		})
	}

	// Simulated POI count and score.
	poiCount := int(totalArea * 12)
	score := min(100, int(totalArea*8)+10)

	// Distribute POIs across categories for richer analytics.
	poiBreakdown := map[string]int{}
	if len(poiCategories) == 0 {
		poiCategories = []string{"amenity", "food", "transit", "shopping", "education", "health"}
	}
	remaining := poiCount
	for i, cat := range poiCategories {
		var share int
		if i == len(poiCategories)-1 {
			share = remaining
		} else {
			share = int(float64(poiCount) * (0.35 / float64(len(poiCategories))))
			if cat == "transit" {
				share = int(float64(poiCount) * 0.25)
			}
			if cat == "health" {
				share = int(float64(poiCount) * 0.15)
			}
			remaining -= share
		}
		if share < 0 {
			share = 0
		}
		poiBreakdown[cat] = share
	}

	return &AnalysisResult{
		Lat:           lat,
		Lng:           lng,
		Mode:          mode,
		Minutes:       minutes,
		TotalArea:     math.Round(totalArea*10) / 10,
		PoiCount:      poiCount,
		Score:         score,
		PoiBreakdown:  poiBreakdown,
		Bands:         bands,
	}
}

// generateCircleGeoJSON builds an approximate circle polygon around a point.
func generateCircleGeoJSON(lat, lng, radiusKm float64, numPoints int) map[string]interface{} {
	coords := make([][]float64, 0, numPoints+1)
	for i := 0; i <= numPoints; i++ {
		angle := 2 * math.Pi * float64(i) / float64(numPoints)
		dLat := (radiusKm / 111.32) * math.Cos(angle)
		dLng := (radiusKm / (111.32 * math.Cos(lat*math.Pi/180))) * math.Sin(angle)
		coords = append(coords, []float64{lng + dLng, lat + dLat})
	}

	return map[string]interface{}{
		"type": "Feature",
		"geometry": map[string]interface{}{
			"type":        "Polygon",
			"coordinates": []interface{}{coords},
		},
		"properties": map[string]interface{}{},
	}
}