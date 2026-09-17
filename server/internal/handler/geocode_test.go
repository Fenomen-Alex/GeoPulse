package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alex/geopulse/server/internal/spatial"
)

type stubGeocoder struct {
	results []spatial.Place
	err     error
}

func (s stubGeocoder) Search(query string, limit int) ([]spatial.Place, error) {
	return s.results, s.err
}

func newTestGeocodeHandler(g Geocoder) *GeocodeHandler {
	return &GeocodeHandler{geocoder: g}
}

func TestGeocodeHandler_Success(t *testing.T) {
	h := newTestGeocodeHandler(stubGeocoder{
		results: []spatial.Place{
			{Lat: 52.52, Lng: 13.405, Label: "Berlin, Germany"},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/geocode?q=berlin&limit=5", nil)
	w := httptest.NewRecorder()

	h.Handle(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp GeocodeResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Results))
	}
	if resp.Results[0].Label != "Berlin, Germany" {
		t.Fatalf("unexpected label: %s", resp.Results[0].Label)
	}
}

func TestGeocodeHandler_QueryTooShort(t *testing.T) {
	h := newTestGeocodeHandler(stubGeocoder{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/geocode?q=be", nil)
	w := httptest.NewRecorder()

	h.Handle(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGeocodeHandler_UpstreamError(t *testing.T) {
	h := newTestGeocodeHandler(stubGeocoder{
		err: &upstreamError{},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/geocode?q=berlin", nil)
	w := httptest.NewRecorder()

	h.Handle(w, req)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGeocodeHandler_OfflineCityIndex(t *testing.T) {
	// Real constructor: local city index first, remote fallback second.
	// A city-level query must resolve fully offline without upstream calls.
	h := NewGeocodeHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/geocode?q=paris&limit=5", nil)
	w := httptest.NewRecorder()

	h.Handle(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp GeocodeResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp.Results) == 0 {
		t.Fatal("expected offline city matches for paris")
	}
	if !strings.Contains(resp.Results[0].Label, "Paris") {
		t.Errorf("expected Paris first, got %q", resp.Results[0].Label)
	}
}

type upstreamError struct{}

func (e *upstreamError) Error() string { return "upstream unavailable" }
