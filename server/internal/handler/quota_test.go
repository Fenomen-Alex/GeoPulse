package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alex/geopulse/server/internal/quota"
)

func TestQuotaStatusHandler(t *testing.T) {
	tracker := quota.New()
	tracker.Consume("test-user")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/quota", nil)
	w := httptest.NewRecorder()
	NewQuotaStatusHandler(tracker).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var response struct {
		DailyQuota     int `json:"daily_quota"`
		RemainingQuota int `json:"remaining_quota"`
	}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.DailyQuota != quota.DefaultDailyQuota {
		t.Errorf("expected daily quota %d, got %d", quota.DefaultDailyQuota, response.DailyQuota)
	}
	if response.RemainingQuota != quota.DefaultDailyQuota-1 {
		t.Errorf("expected remaining quota %d, got %d", quota.DefaultDailyQuota-1, response.RemainingQuota)
	}
}
