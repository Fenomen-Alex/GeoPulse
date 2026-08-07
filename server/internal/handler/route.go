package handler

import (
	"encoding/json"
	"net/http"

	"github.com/alex/geopulse/server/internal/auth"
	"github.com/alex/geopulse/server/internal/config"
	"github.com/alex/geopulse/server/internal/quota"
	"github.com/alex/geopulse/server/internal/spatial"
)

// RouteService is the street-navigation backend used by the route handler.
// It is an interface so tests can inject a stub and avoid network calls.
type RouteService interface {
	Directions(start, end spatial.Point, mode string) (*spatial.Route, error)
}

type RouteRequest struct {
	Start spatial.Point `json:"start"`
	End   spatial.Point `json:"end"`
	Mode  string        `json:"mode"`
}

type RouteResponse struct {
	Start          spatial.Point  `json:"start"`
	End            spatial.Point  `json:"end"`
	Mode           string         `json:"mode"`
	DistanceM      float64        `json:"distance_m"`
	DurationS      float64        `json:"duration_s"`
	GeoJSON        map[string]any `json:"geojson"`
	Steps          []spatial.Step `json:"steps"`
	RemainingQuota int            `json:"remaining_quota"`
}

type RouteHandler struct {
	cfg    *config.Config
	routes RouteService
	quota  *quota.Quota
}

// NewRouteHandler builds a handler backed by the real OpenRouteService client.
func NewRouteHandler(cfg *config.Config, quotaTracker *quota.Quota) *RouteHandler {
	return &RouteHandler{
		cfg:    cfg,
		routes: spatial.NewORSClient(cfg.ORSAPIKey),
		quota:  quotaTracker,
	}
}

func (h *RouteHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req RouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if !validPoint(req.Start) || !validPoint(req.End) {
		http.Error(w, `{"error":"coordinates out of range"}`, http.StatusBadRequest)
		return
	}

	if req.Start == req.End {
		http.Error(w, `{"error":"start and end must differ"}`, http.StatusBadRequest)
		return
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

	route, err := h.routes.Directions(req.Start, req.End, req.Mode)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Routing service unavailable. Please try again later.",
		})
		return
	}

	resp := RouteResponse{
		Start:          req.Start,
		End:            req.End,
		Mode:           req.Mode,
		DistanceM:      route.DistanceM,
		DurationS:      route.DurationS,
		GeoJSON:        buildRouteGeoJSON(route),
		Steps:          route.Steps,
		RemainingQuota: remaining,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func validPoint(p spatial.Point) bool {
	return p.Lat >= -90 && p.Lat <= 90 && p.Lng >= -180 && p.Lng <= 180
}

func buildRouteGeoJSON(route *spatial.Route) map[string]any {
	return map[string]any{
		"type": "Feature",
		"geometry": map[string]any{
			"type":        "LineString",
			"coordinates": route.Coordinates,
		},
		"properties": map[string]any{
			"distance": route.DistanceM,
			"duration": route.DurationS,
		},
	}
}
