package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/alex/geopulse/server/internal/quota"
	"github.com/alex/geopulse/server/internal/spatial"
)

// ExportRequest describes an analysis to export.
type ExportRequest struct {
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
	Mode     string  `json:"mode"`
	Minutes  int     `json:"minutes"`
	Format   string  `json:"format"` // "geojson" | "json" | "csv"
}

// ExportHandler renders an analysis result as a downloadable artifact
// (GeoJSON FeatureCollection, full JSON, or CSV of band rows). Exporting does
// NOT consume a spatial run — the client already paid for the analysis.
type ExportHandler struct {
	quota quota.Tracker
}

// NewExportHandler builds the handler.
func NewExportHandler(quotaTracker quota.Tracker) *ExportHandler {
	return &ExportHandler{quota: quotaTracker}
}

func (h *ExportHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// Accept both POST (JSON body) and GET (query params) so the client can
	// trigger a browser download via window.open() without a CORS preflight.
	var req ExportRequest
	switch r.Method {
	case http.MethodPost:
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}
	case http.MethodGet:
		q := r.URL.Query()
		req.Lat = parseFloatParam(q, "lat")
		req.Lng = parseFloatParam(q, "lng")
		req.Mode = q.Get("mode")
		req.Minutes = parseIntParam(q, "minutes", 15)
		req.Format = q.Get("format")
	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if req.Lat < -90 || req.Lat > 90 || req.Lng < -180 || req.Lng > 180 {
		http.Error(w, `{"error":"coordinates out of range"}`, http.StatusBadRequest)
		return
	}
	if !spatial.ValidModes[req.Mode] {
		req.Mode = "walk"
	}

	result := spatial.RunAnalysis(req.Lat, req.Lng, req.Mode, req.Minutes, nil)

	format := req.Format
	switch format {
	case "geojson":
		w.Header().Set("Content-Type", "application/geo+json")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="geopulse-isochrone-%s.geojson"`, time.Now().Format("20060102-150405")))
		writeGeoJSON(w, result)
	case "csv":
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="geopulse-isochrone-%s.csv"`, time.Now().Format("20060102-150405")))
		writeCSV(w, result)
	default:
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="geopulse-analysis-%s.json"`, time.Now().Format("20060102-150405")))
		writeJSON(w, http.StatusOK, result)
	}
}

// writeGeoJSON emits the isochrone bands as a GeoJSON FeatureCollection.
func writeGeoJSON(w http.ResponseWriter, result *spatial.AnalysisResult) {
	features := make([]map[string]interface{}, 0, len(result.Bands))
	for _, b := range result.Bands {
		var f map[string]interface{}
		_ = json.Unmarshal(b.GeoJSON, &f)
		props := map[string]interface{}{
			"minutes":     b.Minutes,
			"area_km2":    b.Area,
			"fillColor":   b.FillColor,
			"strokeColor": b.StrokeColor,
			"fillOpacity": b.FillOpacity,
			"mode":        result.Mode,
		}
		if m, ok := f["properties"].(map[string]interface{}); ok {
			for k, v := range m {
				props[k] = v
			}
		}
		f["properties"] = props
		features = append(features, f)
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"type":     "FeatureCollection",
		"features": features,
		"properties": map[string]interface{}{
			"center":      []float64{result.Lng, result.Lat},
			"mode":        result.Mode,
			"minutes":     result.Minutes,
			"total_area":  result.TotalArea,
			"poi_count":   result.PoiCount,
			"score":       result.Score,
			"generated":   time.Now().UTC().Format(time.RFC3339),
		},
	})
}

// writeCSV emits one row per isochrone band.
func writeCSV(w http.ResponseWriter, result *spatial.AnalysisResult) {
	w.Write([]byte("minutes,area_km2,fillColor,strokeColor,fillOpacity\n"))
	for _, b := range result.Bands {
		w.Write([]byte(fmt.Sprintf("%d,%.2f,%s,%s,%.2f\n",
			b.Minutes, b.Area, b.FillColor, b.StrokeColor, b.FillOpacity)))
	}
	w.Write([]byte(fmt.Sprintf("total,,,,%.2f\n", result.TotalArea)))
	w.Write([]byte(fmt.Sprintf("poi_count,,,,%d\n", result.PoiCount)))
	w.Write([]byte(fmt.Sprintf("score,,,,%d\n", result.Score)))
}

func parseFloatParam(q url.Values, key string) float64 {
	v := q.Get(key)
	if v == "" {
		return 0
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0
	}
	return f
}

func parseIntParam(q url.Values, key string, def int) int {
	v := q.Get(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return def
	}
	return n
}

var _ = json.Marshal