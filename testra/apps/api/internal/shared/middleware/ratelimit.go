package middleware

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	apihttp "github.com/testra/testra/apps/api/internal/shared/http"
)

type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}

type RateLimitConfig struct {
	Limiter RateLimiter
}

type RateLimitRule struct {
	Limit  int
	Window time.Duration
}

func RateLimit(cfg RateLimitConfig, keyFn func(*http.Request) string, rule RateLimitRule) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.Limiter == nil {
				next.ServeHTTP(w, r)
				return
			}

			key := keyFn(r)
			allowed, err := cfg.Limiter.Allow(r.Context(), key, rule.Limit, rule.Window)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			if !allowed {
				w.Header().Set("Retry-After", strconv.Itoa(int(rule.Window.Seconds())))
				apihttp.ErrorJSON(w, http.StatusTooManyRequests, "TOO_MANY_REQUESTS", "rate limit exceeded")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RateLimitByIP() func(*http.Request) string {
	return func(r *http.Request) string {
		return "rl:ip:" + r.RemoteAddr
	}
}

func RateLimitByEmail(field string) func(*http.Request) string {
	return func(r *http.Request) string {
		email := r.URL.Query().Get(field)
		if email == "" {
			return "rl:ip:" + r.RemoteAddr
		}
		return "rl:email:" + email
	}
}

func RateLimitByAPIKey() func(*http.Request) string {
	return func(r *http.Request) string {
		if key := extractAPIKey(r); key != "" {
			return "rl:apikey:" + hashAPIKey(key)
		}
		return "rl:ip:" + r.RemoteAddr
	}
}

type LocalRateLimiter struct {
	mu    sync.Mutex
	store map[string]*rateBucket
}

type rateBucket struct {
	count   int
	expires time.Time
}

func NewLocalRateLimiter() *LocalRateLimiter {
	return &LocalRateLimiter{store: make(map[string]*rateBucket)}
}

func (l *LocalRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	b, ok := l.store[key]
	if !ok || now.After(b.expires) {
		l.store[key] = &rateBucket{count: 1, expires: now.Add(window)}
		return true, nil
	}

	if b.count >= limit {
		return false, nil
	}

	b.count++
	return true, nil
}
