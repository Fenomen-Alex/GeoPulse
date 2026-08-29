package handler

import (
	"encoding/json"
	"net/http"

	"github.com/alex/geopulse/server/internal/auth"
	"github.com/alex/geopulse/server/internal/quota"
)

// NewQuotaStatusHandler reports the shared analysis and routing allowance
// without consuming a spatial run.
func NewQuotaStatusHandler(tracker *quota.Quota) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := auth.GetUserIDFromContext(r.Context())
		if !ok {
			userID = "test-user"
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{
			"daily_quota":     quota.DefaultDailyQuota,
			"remaining_quota": tracker.Remaining(userID),
		})
	}
}
