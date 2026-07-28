package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
)

// AuthHandler handles Google OIDC authentication
func AuthHandler() http.Handler {
	r := chi.NewRouter()

	// Google OIDC login handler
	r.Get("/login", func(w http.ResponseWriter, r *http.Request) {
		testMode := r.URL.Query().Get("test") == "true"
		userID := "demo-user-id"
		userName := "Demo User"
		if testMode {
			userID = "test-user"
			userName = "Test User"
		}

		// In a real implementation, this would initiate the Google OIDC flow
		// For now, we'll simulate a successful login
		claims := jwt.MapClaims{
			"user_id": userID,
			"exp":     time.Now().Add(24 * time.Hour).Unix(),
		}

		// Create a new token with the claims
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		// Sign and get the complete encoded token as a string
		tokenString, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

		// Set the token as a cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "geopulse_session",
			Value:    tokenString,
			HttpOnly: true,
			Secure:   false,
			MaxAge:   3600 * 24,
			Path:     "/",
			SameSite: http.SameSiteLaxMode,
		})

		// Redirect to the app after login
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"authenticated": true,
			"redirect":      "/app",
			"user": map[string]interface{}{
				"id":          userID,
				"name":        userName,
				"email":       "demo@example.com",
				"daily_quota": 15,
			},
		})
	})

	// Google OIDC callback handler
	r.Get("/callback", func(w http.ResponseWriter, r *http.Request) {
		// In a real implementation, this would handle the Google OIDC callback
		// For now, we'll simulate a successful login
		claims := jwt.MapClaims{
			"user_id": "demo-user-id",
			"exp": time.Now().Add(24 * time.Hour).Unix(),
		}

		// Create a new token with the claims
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		// Sign and get the complete encoded token as a string
		tokenString, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

		// Set the token as a cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "geopulse_session",
			Value:    tokenString,
			HttpOnly: true,
			Secure:   true,
			MaxAge:   3600 * 24,
			Path:     "/",
			SameSite: http.SameSiteLaxMode,
		})

		// Redirect to the app after login
		http.Redirect(w, r, "/app", http.StatusFound)
	})

	// Auth status handler
	r.Get("/status", func(w http.ResponseWriter, r *http.Request) {
		// Get the session cookie
		cookie, err := r.Cookie("geopulse_session")
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{
			"authenticated": false,
			"user": null
		}`, http.StatusOK)
		return
	}

		// Validate the JWT token
		token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
			// Check signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

			// Return the secret key
			return []byte(os.Getenv("JWT_SECRET")), nil
	})

		if err != nil || !token.Valid {
			w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{
			"authenticated": false,
			"user": null
		}`, http.StatusOK)
		return
	}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, `{
			"authenticated": false,
			"user": null
		}`, http.StatusOK)
		return
	}

		userID, ok := claims["user_id"].(string)
		if !ok {
			http.Error(w, `{
			"authenticated": false,
			"user": null
		}`, http.StatusOK)
		return
	}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"authenticated": true,
			"user": map[string]interface{}{
				"id":        userID,
				"name":      "Demo User",
				"email":     "demo@example.com",
				"daily_quota": 15,
			},
		})
	})

	// Auth logout handler
	r.Post("/logout", func(w http.ResponseWriter, r *http.Request) {
		// Clear the session cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "geopulse_session",
			Value:    "",
			Expires:  time.Unix(0, 0),
			HttpOnly: true,
			Secure:   true,
			Path:     "/",
			SameSite: http.SameSiteLaxMode,
		})

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"success":true}`))
	})

	return r
}