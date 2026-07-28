package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/alex/geopulse/server/internal/config"
	"github.com/alex/geopulse/server/internal/handler"
	"github.com/alex/geopulse/server/internal/middleware"
)

//go:embed all:public
var publicFS embed.FS

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	analysisHandler := handler.NewAnalysisHandler(cfg)

	r := chi.NewRouter()

	// Apply middleware
	r.Use(middleware.CORS(cfg.AllowedOrigin))

	// API Sub-router
	r.Route("/api/v1", func(api chi.Router) {
		api.Get("/health", handler.HealthCheck)
		api.Post("/analysis", analysisHandler.HandleAnalysis)
	})

	// Embedded Static Frontend Handler
	contentFS, _ := fs.Sub(publicFS, "public")
	fileServer := http.FileServer(http.FS(contentFS))
	r.Handle("/*", fileServer)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("GeoPulse server starting on %s", addr)
	log.Printf("Allowed origin: %s", cfg.AllowedOrigin)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
