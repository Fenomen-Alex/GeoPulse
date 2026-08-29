package handler

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	googleidtoken "google.golang.org/api/idtoken"

	"github.com/alex/geopulse/server/internal/config"
)

var (
	googleOAuthConfig *oauth2.Config
	jwtSecret         []byte
)

const oauthStateCookie = "geopulse_oauth_state"

func loginHandler(w http.ResponseWriter, r *http.Request, secureCookies bool) {
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		http.Error(w, `{"error":"authentication unavailable"}`, http.StatusInternalServerError)
		return
	}
	state := base64.RawURLEncoding.EncodeToString(stateBytes)
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookie,
		Value:    state,
		Path:     "/api/v1/auth/callback",
		MaxAge:   600,
		HttpOnly: true,
		Secure:   secureCookies,
		SameSite: http.SameSiteLaxMode,
	})
	url := googleOAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "select_account"))
	http.Redirect(w, r, url, http.StatusFound)
}

func callbackHandler(w http.ResponseWriter, r *http.Request, secureCookies bool) {
	stateCookie, err := r.Cookie(oauthStateCookie)
	requestState := r.URL.Query().Get("state")
	if err != nil || requestState == "" || subtle.ConstantTimeCompare([]byte(stateCookie.Value), []byte(requestState)) != 1 {
		http.Error(w, `{"error":"invalid authentication state"}`, http.StatusUnauthorized)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookie,
		Value:    "",
		Path:     "/api/v1/auth/callback",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secureCookies,
		SameSite: http.SameSiteLaxMode,
	})

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, `{"error":"no authorization code"}`, http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	oauthToken, err := googleOAuthConfig.Exchange(ctx, code)
	if err != nil {
		http.Error(w, `{"error":"token exchange failed"}`, http.StatusInternalServerError)
		return
	}

	rawIDToken, ok := oauthToken.Extra("id_token").(string)
	if !ok {
		http.Error(w, `{"error":"no id_token in response"}`, http.StatusInternalServerError)
		return
	}

	payload, err := googleidtoken.Validate(ctx, rawIDToken, googleOAuthConfig.ClientID)
	if err != nil {
		http.Error(w, `{"error":"id_token validation failed"}`, http.StatusUnauthorized)
		return
	}

	userID, _ := payload.Claims["sub"].(string)
	email, _ := payload.Claims["email"].(string)
	name, _ := payload.Claims["name"].(string)

	nameStr := "User"
	if name != "" {
		nameStr = name
	}

	claimsJWT := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"name":    nameStr,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claimsJWT).SignedString(jwtSecret)
	if err != nil {
		http.Error(w, `{"error":"session creation failed"}`, http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "geopulse_session",
		Value:    tokenString,
		HttpOnly: true,
		Secure:   secureCookies,
		MaxAge:   3600 * 24,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/", http.StatusFound)
}

func AuthHandler(cfg *config.Config) http.Handler {
	jwtSecret = []byte(cfg.JWTSecret)
	googleOAuthConfig = &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  fmt.Sprintf("%s/api/v1/auth/callback", strings.TrimSuffix(cfg.AllowedOrigin, "/")),
		Scopes: []string{
			"openid",
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
	}

	r := chi.NewRouter()
	secureCookies := strings.HasPrefix(strings.ToLower(cfg.AllowedOrigin), "https://")

	r.Get("/login", func(w http.ResponseWriter, r *http.Request) {
		if cfg.TestMode {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		loginHandler(w, r, secureCookies)
	})
	r.Get("/callback", func(w http.ResponseWriter, r *http.Request) {
		callbackHandler(w, r, secureCookies)
	})

	r.Get("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if cfg.TestMode {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"authenticated": true,
				"test_mode":     true,
				"user": map[string]interface{}{
					"id":          "test-user",
					"name":        "Test User",
					"email":       "test@geopulse.local",
					"daily_quota": 15,
				},
			})
			return
		}

		cookie, err := r.Cookie("geopulse_session")
		if err != nil {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"authenticated":false,"user":null}`))
			return
		}

		token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"authenticated":false,"user":null}`))
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"authenticated":false,"user":null}`))
			return
		}

		userID, _ := claims["user_id"].(string)
		email, _ := claims["email"].(string)
		name, _ := claims["name"].(string)

		json.NewEncoder(w).Encode(map[string]interface{}{
			"authenticated": true,
			"user": map[string]interface{}{
				"id":          userID,
				"name":        name,
				"email":       email,
				"daily_quota": 15,
			},
		})
	})

	r.Post("/logout", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:     "geopulse_session",
			Value:    "",
			Expires:  time.Unix(0, 0),
			HttpOnly: true,
			Secure:   secureCookies,
			Path:     "/",
			SameSite: http.SameSiteLaxMode,
		})

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"success":true}`))
	})

	return r
}
