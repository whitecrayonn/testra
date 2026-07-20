package middleware

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	apihttp "github.com/testra/testra/apps/api/internal/shared/http"
)

type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}

type RateLimitConfig struct {
	Limiter    RateLimiter
	FailClosed bool
}

type RateLimitRule struct {
	Limit  int
	Window time.Duration
}

func RateLimit(cfg RateLimitConfig, keyFn func(*http.Request) string, rule RateLimitRule) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFn(r)

			limiter := cfg.Limiter
			if limiter == nil {
				if cfg.FailClosed {
					apihttp.ErrorJSON(w, http.StatusServiceUnavailable, "RATE_LIMIT_UNAVAILABLE", "rate limiter unavailable")
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			allowed, err := limiter.Allow(r.Context(), key, rule.Limit, rule.Window)
			if err != nil {
				if cfg.FailClosed {
					apihttp.ErrorJSON(w, http.StatusServiceUnavailable, "RATE_LIMIT_UNAVAILABLE", "rate limiter unavailable")
					return
				}
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
		return "rl:ip:" + realClientIP(r)
	}
}

func RateLimitByEmail(field string) func(*http.Request) string {
	return func(r *http.Request) string {
		email := r.URL.Query().Get(field)
		if email == "" {
			return "rl:ip:" + realClientIP(r)
		}
		return "rl:email:" + email
	}
}

func RateLimitByAPIKey() func(*http.Request) string {
	return func(r *http.Request) string {
		if key := extractAPIKey(r); key != "" {
			return "rl:apikey:" + hashAPIKey(key)
		}
		return "rl:ip:" + realClientIP(r)
	}
}

// realClientIP returns the best-effort client IP for rate limiting. It trusts
// X-Forwarded-For and X-Real-Ip when present, and strips the source port from
// RemoteAddr so that ephemeral ports do not fragment per-client buckets.
func realClientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		if i := strings.Index(forwarded, ","); i >= 0 {
			forwarded = strings.TrimSpace(forwarded[:i])
		}
		if host := stripPort(forwarded); host != "" {
			return host
		}
	}
	if realIP := r.Header.Get("X-Real-Ip"); realIP != "" {
		if host := stripPort(realIP); host != "" {
			return host
		}
	}
	return stripPort(r.RemoteAddr)
}

func stripPort(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return host
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
