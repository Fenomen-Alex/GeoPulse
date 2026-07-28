package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// ContactRouter creates a new router for contact-related endpoints
func ContactRouter() http.Handler {
	r := chi.NewRouter()

	// Public endpoint for contact form
	r.Post("/", ContactHandler)

	return r
}