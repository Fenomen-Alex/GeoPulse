package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alex/geopulse/server/internal/quota"
)

func TestExportHandler_JSON(t *testing.T) {
	h := NewExportHandler(quota.New())

	body := `{"lat":40.7128,"lng":-74.0060,"mode":"walk","minutes":15,"format":"json"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/export", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Handle(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	if !strings.Contains(w.Body.String(), "totalArea") {
		t.Errorf("expected GeoJSON-like response with totalArea")
	}
}

func TestExportHandler_GeoJSON(t *testing.T) {
	h := NewExportHandler(quota.New())

	body := `{"lat":40.7128,"lng":-74.0060,"mode":"walk","minutes":15,"format":"geojson"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/export", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Handle(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/geo+json" {
		t.Errorf("expected content-type application/geo+json, got %s", ct)
	}
	if !strings.Contains(w.Body.String(), "FeatureCollection") {
		t.Errorf("expected FeatureCollection")
	}
}

func TestExportHandler_CSV(t *testing.T) {
	h := NewExportHandler(quota.New())

	body := `{"lat":40.7128,"lng":-74.0060,"mode":"walk","minutes":15,"format":"csv"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/export", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Handle(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "text/csv" {
		t.Errorf("expected content-type text/csv, got %s", ct)
	}
	if !strings.Contains(w.Body.String(), "minutes,area_km2") {
		t.Errorf("expected CSV header row")
	}
}

func TestExportHandler_MethodNotAllowed(t *testing.T) {
	h := NewExportHandler(quota.New())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/export", nil)
	w := httptest.NewRecorder()

	h.Handle(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}