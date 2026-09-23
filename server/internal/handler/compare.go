package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/alex/geopulse/server/internal/auth"
	"github.com/alex/geopulse/server/internal/quota"
	"github.com/alex/geopulse/server/internal/spatial"
)

// CompareRequest is a pair of reachability analyses to compare side-by-side.
type CompareRequest struct {
	Left  AnalysisRequest `json:"left"`
	Right AnalysisRequest `json:"right"`
}

// CompareResponse holds both analyses plus a computed delta summary.
type CompareResponse struct {
	Left  *spatial.AnalysisResult `json:"left"`
	Right *spatial.AnalysisResult `json:"right"`
	Delta CompareDelta           `json:"delta"`
}

// CompareDelta quantifies the difference between two analyses.
type CompareDelta struct {
	AreaDiff     float64 `json:"areaDiff"`     // (right - left) in km²
	AreaDiffPct  float64 `json:"areaDiffPct"`  // percent change
	PoiDiff      int     `json:"poiDiff"`
	ScoreDiff    int     `json:"scoreDiff"`
	Winner       string  `json:"winner"`       // "left" | "right" | "tie"
}

// CompareHandler computes two reachability analyses and returns them together
// so the client can render a side-by-side or overlay comparison. Both runs
// share a single quota charge.
type CompareHandler struct {
	quota quota.Tracker
}

// NewCompareHandler builds a handler. The handler is quota-aware: a single
// comparison consumes one spatial run.
func NewCompareHandler(quotaTracker quota.Tracker) *CompareHandler {
	return &CompareHandler{quota: quotaTracker}
}

func (h *CompareHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req CompareRequest
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Validate both halves.
	if req.Left.Lat < -90 || req.Left.Lat > 90 || req.Left.Lng < -180 || req.Left.Lng > 180 ||
		req.Right.Lat < -90 || req.Right.Lat > 90 || req.Right.Lng < -180 || req.Right.Lng > 180 {
		http.Error(w, `{"error":"coordinates out of range"}`, http.StatusBadRequest)
		return
	}

	// Normalize modes.
	if !spatial.ValidModes[req.Left.Mode] {
		req.Left.Mode = "walk"
	}
	if !spatial.ValidModes[req.Right.Mode] {
		req.Right.Mode = "walk"
	}

	// Quota: one charge for the whole comparison.
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		userID = "test-user"
	}
	_, allowed := h.quota.Consume(userID)
	if !allowed {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":"Daily quota exceeded","remaining_quota":0,"message":"You have reached your limit of 15 queries today. Request an extension to get additional runs."}`))
		return
	}

	left := spatial.RunAnalysis(req.Left.Lat, req.Left.Lng, req.Left.Mode, req.Left.Minutes, nil)
	right := spatial.RunAnalysis(req.Right.Lat, req.Right.Lng, req.Right.Mode, req.Right.Minutes, nil)

	delta := CompareDelta{
		AreaDiff:    mathRound2(right.TotalArea - left.TotalArea),
		AreaDiffPct: mathRound2(pctChange(left.TotalArea, right.TotalArea)),
		PoiDiff:     right.PoiCount - left.PoiCount,
		ScoreDiff:   right.Score - left.Score,
	}
	switch {
	case delta.AreaDiff > 0.1:
		delta.Winner = "right"
	case delta.AreaDiff < -0.1:
		delta.Winner = "left"
	default:
		delta.Winner = "tie"
	}

	resp := CompareResponse{Left: left, Right: right, Delta: delta}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func pctChange(a, b float64) float64 {
	if a == 0 {
		return 0
	}
	return (b - a) / a * 100
}

func mathRound2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

var _ = fmt.Sprintf