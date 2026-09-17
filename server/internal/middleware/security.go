package middleware

import "net/http"

// SecurityHeaders middleware sets sensible browser security headers on every
// response. The CSP below reflects the (current) client-side third-party
// surface: bundled scripts/fonts are served same-origin, and the map still
// fetches its basemap from CARTO's CDN. Once basemaps move to self-hosted
// PMTiles (roadmap), the basemaps.cartocdn.com entry drops out entirely.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self'; "+
				"style-src 'self' 'unsafe-inline'; "+
				"font-src 'self' data:; "+
				"img-src 'self' data: blob:; "+
				"connect-src 'self' https://basemaps.cartocdn.com https://*.basemaps.cartocdn.com; "+
				"worker-src 'self' blob:; "+
				"base-uri 'self'; form-action 'self'; frame-ancestors 'none'")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Permissions-Policy", "geolocation=(), camera=(), microphone=()")
		next.ServeHTTP(w, r)
	})
}