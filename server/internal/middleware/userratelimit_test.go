package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestUserRateLimitMiddleware_AllowsBurst(t *testing.T) {
	handler := UserRateLimitMiddleware(time.Hour, 3)(okHandler())

	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/geocode", nil)
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("burst request %d: expected 200, got %d", i+1, rec.Code)
		}
	}
}

func TestUserRateLimitMiddleware_BlocksAfterBurst(t *testing.T) {
	handler := UserRateLimitMiddleware(time.Hour, 3)(okHandler())

	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/geocode", nil)
		handler.ServeHTTP(rec, req)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/geocode", nil)
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after burst, got %d", rec.Code)
	}
}

func TestUserRateLimitMiddleware_KeysPerUser(t *testing.T) {
	handler := UserRateLimitMiddleware(time.Hour, 1)(okHandler())

	asUser := func(id string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/geocode", nil)
		req = req.WithContext(context.WithValue(req.Context(), UserIDKey, id))
		handler.ServeHTTP(rec, req)
		return rec
	}

	if rec := asUser("alice"); rec.Code != http.StatusOK {
		t.Fatalf("alice request 1: expected 200, got %d", rec.Code)
	}
	if rec := asUser("alice"); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("alice request 2 should be throttled, got %d", rec.Code)
	}
	if rec := asUser("bob"); rec.Code != http.StatusOK {
		t.Fatalf("bob should have an independent bucket, got %d", rec.Code)
	}
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}