package handler

import (
	"encoding/json"
	"net/http"

	"github.com/alex/geopulse/server/internal/auth"
	"github.com/alex/geopulse/server/internal/config"
	"github.com/alex/geopulse/server/internal/quota"
	"github.com/alex/geopulse/server/internal/spatial"
)

// POICategories is the set of POI categories the client can filter by. The
// server distributes a simulated POI count across these categories so the
// client can render a breakdown chart.
var POICategories = []string{
	"amenity", "food", "transit", "shopping", "education", "health",
}

// AnalysisRequest is the client payload for a reachability analysis.
type AnalysisRequest struct {
	Lat          float64  `json:"lat"`
	Lng          float64  `json:"lng"`
	Mode         string   `json:"mode"`
	Minutes      int      `json:"minutes"`
	POICategories []string `json:"poi_categories,omitempty"`
}

// AnalysisResponse is the server's reply to an analysis request.
type AnalysisResponse struct {
	Lat            float64             `json:"lat"`
	Lng            float64             `json:"lng"`
	Mode           string              `json:"mode"`
	Minutes        int                 `json:"minutes"`
	TotalArea      float64             `json:"totalArea"`
	PoiCount       int                 `json:"poiCount"`
	Score          int                 `json:"score"`
	PoiBreakdown   map[string]int      `json:"poiBreakdown,omitempty"`
	Bands          []spatial.IsochroneBand `json:"bands"`
	RemainingQuota int                 `json:"remaining_quota"`
}

// AnalysisHandler answers reachability queries. It is quota-aware: each call
// consumes one of the user's 15 daily spatial runs.
type AnalysisHandler struct {
	cfg   *config.Config
	quota quota.Tracker
}

func NewAnalysisHandler(cfg *config.Config, quotaTracker quota.Tracker) *AnalysisHandler {
	return &AnalysisHandler{cfg: cfg, quota: quotaTracker}
}

func (h *AnalysisHandler) HandleAnalysis(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req AnalysisRequest
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Lat < -90 || req.Lat > 90 || req.Lng < -180 || req.Lng > 180 {
		http.Error(w, `{"error":"coordinates out of range"}`, http.StatusBadRequest)
		return
	}

	// Filter and cap POI categories so a malicious payload can't blow up the
	// breakdown map.
	poiCats := make([]string, 0, len(req.POICategories))
	seen := map[string]bool{}
	for _, c := range req.POICategories {
		for _, valid := range POICategories {
			if c == valid && !seen[c] {
				seen[c] = true
				poiCats = append(poiCats, c)
				break
			}
		}
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

	result := spatial.RunAnalysis(req.Lat, req.Lng, req.Mode, req.Minutes, poiCats)

	resp := AnalysisResponse{
		Lat:            result.Lat,
		Lng:            result.Lng,
		Mode:           result.Mode,
		Minutes:        result.Minutes,
		TotalArea:      result.TotalArea,
		PoiCount:       result.PoiCount,
		Score:          result.Score,
		PoiBreakdown:   result.PoiBreakdown,
		Bands:          result.Bands,
		RemainingQuota: remaining,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}