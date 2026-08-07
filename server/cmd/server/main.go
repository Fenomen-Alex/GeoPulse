package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"path"

	"github.com/go-chi/chi/v5"

	"github.com/alex/geopulse/server/internal/config"
	"github.com/alex/geopulse/server/internal/handler"
	"github.com/alex/geopulse/server/internal/middleware"
	"github.com/alex/geopulse/server/internal/quota"
)

//go:embed all:public
var publicFS embed.FS

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	quotaTracker := quota.New()

	analysisHandler := handler.NewAnalysisHandler(cfg, quotaTracker)
	routeHandler := handler.NewRouteHandler(cfg, quotaTracker)

	r := chi.NewRouter()

	// Apply middleware
	r.Use(middleware.CORS(cfg.AllowedOrigin))
	r.Use(middleware.RateLimitMiddleware)

	// Auth routes (public)
	r.Mount("/api/v1/auth", handler.AuthHandler(cfg.TestMode))

	// API Sub-router with auth middleware + test mode bypass
	r.Route("/api/v1", func(api chi.Router) {
		api.Use(middleware.RequireAuth(cfg.TestMode))
		api.Get("/health", handler.HealthCheck)
		api.Post("/analysis", analysisHandler.HandleAnalysis)
		api.Post("/routes", routeHandler.Handle)
	})

	// Contact router (public)
	r.Mount("/api/v1/contact", handler.ContactRouter())

	// Embedded Static Frontend Handler with SPA fallback
	contentFS, _ := fs.Sub(publicFS, "public")
	fileServer := http.FileServer(http.FS(contentFS))
	indexBytes, _ := fs.ReadFile(contentFS, "index.html")

	r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cleanPath := path.Clean(r.URL.Path)
		if cleanPath[0] == '/' {
			cleanPath = cleanPath[1:]
		}

		// 1. If the file exists in publicFS, serve it directly
		if cleanPath != "" {
			f, err := contentFS.Open(cleanPath)
			if err == nil {
				f.Close()
				fileServer.ServeHTTP(w, r)
				return
			}
		}

		// 2. SPA navigation paths — serve index.html for client-side routing
		//    All asset/file requests (with extensions) and tile/data paths
		//    that don't match embedded files get a 404 to prevent MIME mismatches.
		if path.Ext(cleanPath) != "" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(indexBytes)
	}))

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("GeoPulse server starting on %s", addr)
	log.Printf("Allowed origin: %s", cfg.AllowedOrigin)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
