package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	googleidtoken "google.golang.org/api/idtoken"
)

var (
	googleOAuthConfig *oauth2.Config
	jwtSecret         []byte
)

func init() {
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))
	googleOAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("VITE_GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Endpoint:     google.Endpoint,
		RedirectURL:  fmt.Sprintf("http://localhost:%s/api/v1/auth/callback", os.Getenv("PORT")),
		Scopes: []string{
			"openid",
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	state := fmt.Sprintf("state-%d", time.Now().Unix())
	url := googleOAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "select_account"))
	http.Redirect(w, r, url, http.StatusFound)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
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

	tokenString, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claimsJWT).SignedString(jwtSecret)

	http.SetCookie(w, &http.Cookie{
		Name:     "geopulse_session",
		Value:    tokenString,
		HttpOnly: true,
		Secure:   false,
		MaxAge:   3600 * 24,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/", http.StatusFound)
}

func AuthHandler() http.Handler {
	r := chi.NewRouter()

	r.Get("/login", loginHandler)
	r.Get("/callback", callbackHandler)

	r.Get("/status", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("geopulse_session")
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
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
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"authenticated":false,"user":null}`))
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"authenticated":false,"user":null}`))
			return
		}

		userID, _ := claims["user_id"].(string)
		email, _ := claims["email"].(string)
		name, _ := claims["name"].(string)

		w.Header().Set("Content-Type", "application/json")
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
			Secure:   false,
			Path:     "/",
			SameSite: http.SameSiteLaxMode,
		})

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"success":true}`))
	})

	return r
}