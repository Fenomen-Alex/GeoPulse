package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alex/geopulse/server/internal/config"
)

func TestAuthLoginSetsStateCookie(t *testing.T) {
	router := AuthHandler(&config.Config{
		AllowedOrigin:      "http://localhost:8080",
		GoogleClientID:     "client-id",
		GoogleClientSecret: "client-secret",
		JWTSecret:          "test-secret",
	})
	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Fatalf("expected redirect, got %d", w.Code)
	}
	var found bool
	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == oauthStateCookie && cookie.Value != "" && cookie.HttpOnly {
			found = true
		}
	}
	if !found {
		t.Fatal("expected an HttpOnly OAuth state cookie")
	}
}

func TestAuthCallbackRejectsInvalidState(t *testing.T) {
	router := AuthHandler(&config.Config{
		AllowedOrigin:      "http://localhost:8080",
		GoogleClientID:     "client-id",
		GoogleClientSecret: "client-secret",
		JWTSecret:          "test-secret",
	})
	req := httptest.NewRequest(http.MethodGet, "/callback?code=code&state=wrong", nil)
	req.AddCookie(&http.Cookie{Name: oauthStateCookie, Value: "expected"})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
