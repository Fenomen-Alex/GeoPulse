package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alex/geopulse/server/internal/quota"
	"github.com/alex/geopulse/server/internal/spatial"
)

func TestCompareHandler_Success(t *testing.T) {
	h := NewCompareHandler(quota.New())

	body := `{"left":{"lat":40.7128,"lng":-74.0060,"mode":"walk","minutes":15},"right":{"lat":40.75,"lng":-73.98,"mode":"bike","minutes":30}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/compare", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Handle(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp CompareResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Left.Mode != "walk" || resp.Right.Mode != "bike" {
		t.Errorf("expected modes walk/bike, got %s/%s", resp.Left.Mode, resp.Right.Mode)
	}

	if resp.Delta.Winner == "" {
		t.Error("expected delta winner to be set")
	}
}

func TestCompareHandler_InvalidBody(t *testing.T) {
	h := NewCompareHandler(quota.New())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/compare", bytes.NewBufferString("not json"))
	w := httptest.NewRecorder()

	h.Handle(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCompareHandler_InvalidCoords(t *testing.T) {
	h := NewCompareHandler(quota.New())

	body := `{"left":{"lat":100,"lng":-74.0060},"right":{"lat":40.75,"lng":-73.98}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/compare", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Handle(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCompareHandler_QuotaExceeded(t *testing.T) {
	h := NewCompareHandler(quota.New())

	body := `{"left":{"lat":40.7128,"lng":-74.0060,"mode":"walk","minutes":15},"right":{"lat":40.75,"lng":-73.98,"mode":"bike","minutes":30}}`

	for i := 0; i < quota.DefaultDailyQuota; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/compare", bytes.NewBufferString(body))
		w := httptest.NewRecorder()
		h.Handle(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("run %d: expected 200, got %d", i+1, w.Code)
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/compare", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	h.Handle(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after quota exhausted, got %d", w.Code)
	}
}

func TestSpatialAnalysisStandardOutput(t *testing.T) {
	result := spatial.RunAnalysis(40.7128, -74.0060, "walk", 30, nil)

	if result.Lat != 40.7128 || result.Lng != -74.0060 {
		t.Errorf("coordinates not preserved: got (%f, %f)", result.Lat, result.Lng)
	}
	if result.Mode != "walk" {
		t.Errorf("expected mode walk, got %s", result.Mode)
	}
	if result.Minutes != 30 {
		t.Errorf("expected 30 minutes, got %d", result.Minutes)
	}
	if len(result.Bands) != 3 {
		t.Errorf("expected 3 bands for 30 min, got %d", len(result.Bands))
	}
	if result.Score <= 0 || result.Score > 100 {
		t.Errorf("score out of range: %d", result.Score)
	}
}