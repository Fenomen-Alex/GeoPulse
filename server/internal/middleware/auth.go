package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// AuthContextKey is the key used to store the user ID in the context
type AuthContextKey string

const (
	// UserIDKey is the key used to store the user ID in the context
	UserIDKey AuthContextKey = "user_id"
)

// RequireAuth middleware rejects unauthenticated requests to API endpoints.
// When testMode is true, requests with ?test=true query parameter bypass auth.
func RequireAuth(testMode bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if testMode && r.URL.Query().Get("test") == "true" {
				ctx := context.WithValue(r.Context(), UserIDKey, "test-user")
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			if !strings.HasPrefix(r.URL.Path, "/api/v1/") {
				next.ServeHTTP(w, r)
				return
			}

			if r.URL.Path == "/api/v1/health" || r.URL.Path == "/api/v1/auth/login" || r.URL.Path == "/api/v1/auth/callback" || strings.HasPrefix(r.URL.Path, "/api/v1/auth/") {
				next.ServeHTTP(w, r)
				return
			}

			cookie, err := r.Cookie("geopulse_session")
			if err != nil {
				http.Error(w, `{"error":"Authentication required. Please log in."}`, http.StatusUnauthorized)
				return
			}

			token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(os.Getenv("JWT_SECRET")), nil
			})

			if err != nil || !token.Valid {
				http.Error(w, `{"error":"Invalid or expired session token."}`, http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, `{"error":"Invalid token claims"}`, http.StatusUnauthorized)
				return
			}

			userID, ok := claims["user_id"].(string)
			if !ok {
				http.Error(w, `{"error":"Invalid token claims"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}