package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/alex/geopulse/server/internal/config"
	"github.com/alex/geopulse/server/internal/middleware"
)

// ContactRouter creates a new router for contact-related endpoints
func ContactRouter(cfg *config.Config) http.Handler {
	r := chi.NewRouter()

	// Per-IP burst limiter independent of authentication: a real person files
	// one request; bots/spammers get throttled fast and cannot bypass via a
	// session cookie.
	r.Use(middleware.PerIPRateLimitMiddleware(time.Minute, 5))

	// Public endpoint for contact form
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		handleContact(w, r, cfg.TestMode)
	})

	return r
}