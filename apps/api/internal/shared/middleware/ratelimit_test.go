package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
)

func TestLocalRateLimiterAllowsWithinLimit(t *testing.T) {
	limiter := NewLocalRateLimiter()
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		allowed, err := limiter.Allow(ctx, "key", 5, time.Minute)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !allowed {
			t.Fatalf("expected request %d to be allowed", i+1)
		}
	}
	allowed, err := limiter.Allow(ctx, "key", 5, time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Fatal("expected request over limit to be denied")
	}
}

func TestRedisRateLimiterAllowsWithinLimit(t *testing.T) {
	s := miniredis.RunT(t)
	defer s.Close()

	limiter := NewRedisRateLimiter(s.Addr(), NewLocalRateLimiter())
	defer limiter.Close()

	ctx := context.Background()
	for i := 0; i < 5; i++ {
		allowed, err := limiter.Allow(ctx, "key", 5, time.Minute)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !allowed {
			t.Fatalf("expected request %d to be allowed", i+1)
		}
	}
	allowed, err := limiter.Allow(ctx, "key", 5, time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Fatal("expected request over limit to be denied")
	}
}

func TestRateLimitFailClosed(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	cfg := RateLimitConfig{Limiter: nil, FailClosed: true}
	wrapped := RateLimit(cfg, RateLimitByIP(), RateLimitRule{Limit: 5, Window: time.Minute})(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 for fail-closed, got %d", rr.Code)
	}
}

func TestRedisRateLimiterFallsBackToLocalWhenRedisUnavailable(t *testing.T) {
	// 0.0.0.0:1 is a guaranteed unreachable address for the test environment.
	limiter := NewRedisRateLimiter("0.0.0.0:1", NewLocalRateLimiter())
	defer limiter.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	allowed, err := limiter.Allow(ctx, "key", 1, time.Minute)
	if err != nil {
		t.Fatalf("expected fallback to local limiter, got error: %v", err)
	}
	if !allowed {
		t.Fatal("expected first request to be allowed by fallback limiter")
	}
}
