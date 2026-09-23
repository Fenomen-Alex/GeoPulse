package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/alex/geopulse/server/internal/config"
	"github.com/alex/geopulse/server/internal/db"
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

	// Managed edge persistence (embedded SQLite, single-binary friendly).
	// In test mode a missing/unusable store degrades gracefully to the
	// in-memory tracker so dev/demo flows never hard-crash on DB issues.
	var store *db.DB
	store, err = db.InitDB(cfg.DBPath)
	if err == nil {
		if err := db.Migrate(context.Background(), store); err != nil {
			log.Fatalf("Failed to apply migrations: %v", err)
		}
		log.Printf("Database ready at %s", cfg.DBPath)
	} else {
		if !cfg.TestMode {
			log.Fatalf("Failed to initialize database: %v", err)
		}
		log.Printf("Database unavailable, falling back to in-memory stores: %v", err)
	}

	// Persist quota counters in the DB when available; otherwise keep the
	// in-memory tracker so behavior is identical when no store is configured.
	var quotaTracker quota.Tracker = quota.New()
	if store != nil {
		quotaTracker = quota.NewPersistent(store)
	}

	analysisHandler := handler.NewAnalysisHandler(cfg, quotaTracker)
	routeHandler := handler.NewRouteHandler(cfg, quotaTracker)
	geocodeHandler := handler.NewGeocodeHandler()
	workspaceHandler := handler.NewWorkspaceHandler(store)
	compareHandler := handler.NewCompareHandler(quotaTracker)
	exportHandler := handler.NewExportHandler(quotaTracker)

	r := chi.NewRouter()

	// Apply middleware
	r.Use(middleware.CORS(cfg.AllowedOrigin))
	r.Use(middleware.SecurityHeaders)
	r.Use(middleware.RateLimitMiddleware)

	// Auth routes (public)
	r.Mount("/api/v1/auth", handler.AuthHandler(cfg))

	// API Sub-router with auth middleware + test mode bypass
	r.Route("/api/v1", func(api chi.Router) {
		api.Use(middleware.RequireAuth(cfg.TestMode))
		api.Get("/health", handler.HealthCheck)
		api.Post("/analysis", analysisHandler.HandleAnalysis)
		api.Post("/routes", routeHandler.Handle)
		api.With(middleware.UserRateLimitMiddleware(500*time.Millisecond, 40)).Get("/geocode", geocodeHandler.Handle)
		api.Get("/quota", handler.NewQuotaStatusHandler(quotaTracker))
		api.Get("/workspaces", workspaceHandler.List)
		api.Post("/workspaces", workspaceHandler.Save)
		// Reachability comparison: two analyses, one quota charge.
		api.Post("/compare", compareHandler.Handle)
		// Export: render an analysis as GeoJSON/JSON/CSV (no quota charge).
		api.Post("/export", exportHandler.Handle)
	})

	// Contact router (public)
	r.Mount("/api/v1/contact", handler.ContactRouter(cfg))

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
				// Vite emits content-hashed filenames under /assets/, so those are
				// immutable: cache them for a year and instruct clients/proxies
				// not to revalidate. HTML and everything else stays no-store so a
				// deploy never serves a stale index referencing old hashes.
				if strings.HasPrefix("/"+cleanPath, "/assets/") {
					w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
				} else {
					w.Header().Set("Cache-Control", "no-store")
				}
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

	srv := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 16, // 64 KiB – large enough, prevents header bloat
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
