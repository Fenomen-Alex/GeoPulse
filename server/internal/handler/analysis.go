package handler

import (
	"encoding/json"
	"math"
	"net/http"

	"github.com/alex/geopulse/server/internal/auth"
	"github.com/alex/geopulse/server/internal/config"
	"github.com/alex/geopulse/server/internal/quota"
)

type AnalysisRequest struct {
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	Mode    string  `json:"mode"`
	Minutes int     `json:"minutes"`
}

type IsochroneBand struct {
	Minutes     int         `json:"minutes"`
	Area        float64     `json:"area"`
	FillColor   string      `json:"fillColor"`
	StrokeColor string      `json:"strokeColor"`
	FillOpacity float64     `json:"fillOpacity"`
	GeoJSON     interface{} `json:"geojson"`
}

type AnalysisResponse struct {
	Lat            float64         `json:"lat"`
	Lng            float64         `json:"lng"`
	Mode           string          `json:"mode"`
	Minutes        int             `json:"minutes"`
	TotalArea      float64         `json:"totalArea"`
	PoiCount       int             `json:"poiCount"`
	Score          int             `json:"score"`
	Bands          []IsochroneBand `json:"bands"`
	RemainingQuota int             `json:"remaining_quota"`
}

type AnalysisHandler struct {
	cfg   *config.Config
	quota *quota.Quota
}

func NewAnalysisHandler(cfg *config.Config, quotaTracker *quota.Quota) *AnalysisHandler {
	return &AnalysisHandler{cfg: cfg, quota: quotaTracker}
}

func (h *AnalysisHandler) HandleAnalysis(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req AnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Lat < -90 || req.Lat > 90 || req.Lng < -180 || req.Lng > 180 {
		http.Error(w, `{"error":"coordinates out of range"}`, http.StatusBadRequest)
		return
	}

	if req.Minutes <= 0 || req.Minutes > 60 {
		req.Minutes = 15
	}

	validModes := map[string]bool{"walk": true, "bike": true, "drive": true}
	if !validModes[req.Mode] {
		req.Mode = "walk"
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		userID = "test-user"
	}

	remaining, allowed := h.quota.Consume(userID)
	if !allowed {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":"Daily quota exceeded","remaining_quota":0,"message":"You have reached your limit of 15 queries today. Request an extension to get additional runs."}`))
		return
	}

	// Speed in km/h by travel mode
	speeds := map[string]float64{
		"walk":  5.0,
		"bike":  15.0,
		"drive": 40.0,
	}
	speed := speeds[req.Mode]

	// Generate isochrone bands
	bandMinutes := getBandMinutes(req.Minutes)
	bands := make([]IsochroneBand, 0, len(bandMinutes))

	colors := []struct {
		fill        string
		stroke      string
		fillOpacity float64
	}{
		{"#10b981", "#34d399", 0.40},
		{"#f59e0b", "#fbbf24", 0.35},
		{"#ef4444", "#f87171", 0.25},
		{"#8b5cf6", "#a78bfa", 0.20},
		{"#3b82f6", "#60a5fa", 0.15},
	}

	var totalArea float64
	for i, mins := range bandMinutes {
		radiusKm := speed * (float64(mins) / 60.0)
		area := math.Pi * radiusKm * radiusKm
		totalArea = area // outermost band area is total

		colorIdx := i
		if colorIdx >= len(colors) {
			colorIdx = len(colors) - 1
		}

		polygon := generateCircleGeoJSON(req.Lat, req.Lng, radiusKm, 64)

		bands = append(bands, IsochroneBand{
			Minutes:     mins,
			Area:        math.Round(area*100) / 100,
			FillColor:   colors[colorIdx].fill,
			StrokeColor: colors[colorIdx].stroke,
			FillOpacity: colors[colorIdx].fillOpacity,
			GeoJSON:     polygon,
		})
	}

	// Simulated POI count and score
	poiCount := int(totalArea * 12)
	score := min(100, int(totalArea*8)+10)

	resp := AnalysisResponse{
		Lat:            req.Lat,
		Lng:            req.Lng,
		Mode:           req.Mode,
		Minutes:        req.Minutes,
		TotalArea:      math.Round(totalArea*10) / 10,
		PoiCount:       poiCount,
		Score:          score,
		Bands:          bands,
		RemainingQuota: remaining,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func getBandMinutes(maxMinutes int) []int {
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

func generateCircleGeoJSON(lat, lng, radiusKm float64, numPoints int) map[string]interface{} {
	coords := make([][]float64, 0, numPoints+1)
	for i := 0; i <= numPoints; i++ {
		angle := 2 * math.Pi * float64(i) / float64(numPoints)
		// Approximate offset in degrees
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
