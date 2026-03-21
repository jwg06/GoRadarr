package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(60, 3) // 1 req/sec, burst 3

	// First 3 requests should be allowed (burst)
	for i := 0; i < 3; i++ {
		if !rl.Allow("10.0.0.1") {
			t.Fatalf("request %d should be allowed (burst), was denied", i+1)
		}
	}

	// 4th request should be denied (bucket empty)
	if rl.Allow("10.0.0.1") {
		t.Fatal("4th request should be rate-limited but was allowed")
	}

	// Different IP should be independent
	if !rl.Allow("10.0.0.2") {
		t.Fatal("different IP should be allowed independently")
	}
}

func TestRateLimiter_Middleware_TooManyRequests(t *testing.T) {
	rl := NewRateLimiter(60, 1) // burst=1
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	makeReq := func() int {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		return rr.Code
	}

	if code := makeReq(); code != http.StatusOK {
		t.Fatalf("first request: want 200, got %d", code)
	}
	if code := makeReq(); code != http.StatusTooManyRequests {
		t.Fatalf("second request: want 429, got %d", code)
	}
}

func TestCSRFGuard_GET_passes(t *testing.T) {
	h := CSRFGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/movie", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET without X-Api-Key should pass CSRF guard, got %d", rr.Code)
	}
}

func TestCSRFGuard_POST_without_header_rejected(t *testing.T) {
	h := CSRFGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/movie", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("POST without X-Api-Key should be rejected by CSRF guard, got %d", rr.Code)
	}
}

func TestCSRFGuard_POST_with_api_key_allowed(t *testing.T) {
	h := CSRFGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/movie", nil)
	req.Header.Set("X-Api-Key", "test-key")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("POST with X-Api-Key should pass CSRF guard, got %d", rr.Code)
	}
}

func TestCSRFGuard_POST_with_authorization_allowed(t *testing.T) {
	h := CSRFGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	req.Header.Set("Authorization", "Bearer sometoken")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("POST with Authorization should pass CSRF guard, got %d", rr.Code)
	}
}

func TestExtractIP(t *testing.T) {
	cases := []struct{ in, want string }{
		{"192.168.1.1:5432", "192.168.1.1"},
		{"10.0.0.1:80", "10.0.0.1"},
		{"[::1]:8080", "[::1]"},
		{"nodport", "nodport"},
	}
	for _, c := range cases {
		if got := extractIP(c.in); got != c.want {
			t.Errorf("extractIP(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
