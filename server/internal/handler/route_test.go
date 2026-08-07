package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alex/geopulse/server/internal/config"
	"github.com/alex/geopulse/server/internal/quota"
	"github.com/alex/geopulse/server/internal/spatial"
)

type stubRouteService struct{}

func (s stubRouteService) Directions(start, end spatial.Point, mode string) (*spatial.Route, error) {
	return &spatial.Route{
		DistanceM:   2500,
		DurationS:   420,
		Coordinates: [][]float64{{start.Lng, start.Lat}, {end.Lng, end.Lat}},
		Steps: []spatial.Step{
			{
				Instruction: "Head north on Main Street",
				Name:        "Main Street",
				Distance:    2500,
				Duration:    420,
				Type:        11,
				Location:    []float64{start.Lng, start.Lat},
			},
		},
	}, nil
}

func newTestRouteHandler() *RouteHandler {
	return &RouteHandler{
		cfg:    &config.Config{},
		routes: stubRouteService{},
		quota:  quota.New(),
	}
}

func TestRouteHandler_Success(t *testing.T) {
	h := newTestRouteHandler()

	body := `{"start":{"lat":40.7128,"lng":-74.0060},"end":{"lat":40.7580,"lng":-73.9855},"mode":"drive"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/routes", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Handle(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp RouteResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.DistanceM != 2500 {
		t.Errorf("expected distance 2500, got %f", resp.DistanceM)
	}
	if resp.DurationS != 420 {
		t.Errorf("expected duration 420, got %f", resp.DurationS)
	}
	if resp.Mode != "drive" {
		t.Errorf("expected mode drive, got %s", resp.Mode)
	}
	if len(resp.Steps) == 0 {
		t.Error("expected at least one navigation step")
	}
	if resp.GeoJSON == nil {
		t.Error("expected geojson route geometry")
	}
	if resp.RemainingQuota != quota.DefaultDailyQuota-1 {
		t.Errorf("expected remaining quota %d, got %d", quota.DefaultDailyQuota-1, resp.RemainingQuota)
	}
}

func TestRouteHandler_MethodNotAllowed(t *testing.T) {
	h := newTestRouteHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/routes", nil)
	w := httptest.NewRecorder()

	h.Handle(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestRouteHandler_InvalidBody(t *testing.T) {
	h := newTestRouteHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/routes", bytes.NewBufferString("not json"))
	w := httptest.NewRecorder()

	h.Handle(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestRouteHandler_SamePoint(t *testing.T) {
	h := newTestRouteHandler()

	body := `{"start":{"lat":40.7128,"lng":-74.0060},"end":{"lat":40.7128,"lng":-74.0060},"mode":"walk"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/routes", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Handle(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for identical points, got %d", w.Code)
	}
}

func TestRouteHandler_OutOfRangeCoords(t *testing.T) {
	h := newTestRouteHandler()

	body := `{"start":{"lat":100,"lng":200},"end":{"lat":40.7580,"lng":-73.9855},"mode":"walk"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/routes", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Handle(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestRouteHandler_QuotaExceeded(t *testing.T) {
	h := newTestRouteHandler()

	body := `{"start":{"lat":40.7128,"lng":-74.0060},"end":{"lat":40.7580,"lng":-73.9855},"mode":"walk"}`

	for i := 0; i < quota.DefaultDailyQuota; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/routes", bytes.NewBufferString(body))
		w := httptest.NewRecorder()
		h.Handle(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("run %d: expected 200, got %d", i+1, w.Code)
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/routes", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	h.Handle(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after quota exhausted, got %d", w.Code)
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("Daily quota exceeded")) {
		t.Errorf("expected quota error payload, got %s", w.Body.String())
	}
}
