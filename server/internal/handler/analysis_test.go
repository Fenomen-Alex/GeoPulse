package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alex/geopulse/server/internal/config"
)

func TestHandleAnalysis_Success(t *testing.T) {
	cfg := &config.Config{}
	h := NewAnalysisHandler(cfg)

	body := `{"lat":40.7128,"lng":-74.0060,"mode":"walk","minutes":15}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/analysis", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleAnalysis(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp AnalysisResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Lat != 40.7128 {
		t.Errorf("expected lat 40.7128, got %f", resp.Lat)
	}
	if resp.Mode != "walk" {
		t.Errorf("expected mode walk, got %s", resp.Mode)
	}
	if resp.Minutes != 15 {
		t.Errorf("expected minutes 15, got %d", resp.Minutes)
	}
	if resp.TotalArea <= 0 {
		t.Errorf("expected positive totalArea, got %f", resp.TotalArea)
	}
	if len(resp.Bands) == 0 {
		t.Error("expected at least one isochrone band")
	}
	if resp.Score <= 0 || resp.Score > 100 {
		t.Errorf("expected score between 1-100, got %d", resp.Score)
	}
}

func TestHandleAnalysis_InvalidMethod(t *testing.T) {
	cfg := &config.Config{}
	h := NewAnalysisHandler(cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analysis", nil)
	w := httptest.NewRecorder()

	h.HandleAnalysis(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestHandleAnalysis_InvalidBody(t *testing.T) {
	cfg := &config.Config{}
	h := NewAnalysisHandler(cfg)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/analysis", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleAnalysis(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleAnalysis_OutOfRangeCoords(t *testing.T) {
	cfg := &config.Config{}
	h := NewAnalysisHandler(cfg)

	body := `{"lat":100,"lng":200,"mode":"walk","minutes":15}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/analysis", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleAnalysis(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleAnalysis_DefaultsInvalidMode(t *testing.T) {
	cfg := &config.Config{}
	h := NewAnalysisHandler(cfg)

	body := `{"lat":40.7128,"lng":-74.0060,"mode":"helicopter","minutes":15}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/analysis", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleAnalysis(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp AnalysisResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Mode != "walk" {
		t.Errorf("expected mode to default to walk, got %s", resp.Mode)
	}
}

func TestGetBandMinutes(t *testing.T) {
	tests := []struct {
		input    int
		expected []int
	}{
		{15, []int{5, 15}},
		{30, []int{5, 15, 30}},
		{60, []int{5, 15, 30, 45, 60}},
		{3, []int{3}},
	}
	for _, tt := range tests {
		result := getBandMinutes(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("getBandMinutes(%d): expected %v, got %v", tt.input, tt.expected, result)
		}
	}
}
