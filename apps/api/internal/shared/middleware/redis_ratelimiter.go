package middleware

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisRateLimiter implements a distributed token-bucket-style rate limiter
// backed by Redis. It uses a Lua script so that INCR and PEXPIRE are atomic.
// If Redis is unavailable, it transparently falls back to the provided local
// rate limiter so that the API continues to be protected.
type RedisRateLimiter struct {
	client *redis.Client
	local  RateLimiter
}

var rateLimitScript = `-- 1: INCR counter for this key
local current = redis.call('INCR', KEYS[1])
if current == 1 then
    -- First request in the window: set the TTL.
    redis.call('PEXPIRE', KEYS[1], tonumber(ARGV[1]))
end
return current
`

// NewRedisRateLimiter creates a rate limiter backed by Redis at addr. If addr
// is empty, or if Redis becomes unavailable at runtime, the limiter will use
// the local fallback.
func NewRedisRateLimiter(addr string, local RateLimiter) *RedisRateLimiter {
	if addr == "" {
		return &RedisRateLimiter{local: local}
	}
	return &RedisRateLimiter{
		client: redis.NewClient(&redis.Options{
			Addr: addr,
			// Conservative timeouts: rate limiting should not block requests.
			ReadTimeout:  500 * time.Millisecond,
			WriteTimeout: 500 * time.Millisecond,
		}),
		local: local,
	}
}

// Allow increments the per-key counter in Redis and checks it against limit.
// On any Redis error the call transparently falls back to the local limiter.
func (r *RedisRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	if r.client == nil {
		return r.local.Allow(ctx, key, limit, window)
	}

	ms := window.Milliseconds()
	count, err := r.client.Eval(ctx, rateLimitScript, []string{key}, ms).Int64()
	if err != nil {
		return r.local.Allow(ctx, key, limit, window)
	}
	return count <= int64(limit), nil
}

func (r *RedisRateLimiter) Close() error {
	if r.client == nil {
		return nil
	}
	return r.client.Close()
}
