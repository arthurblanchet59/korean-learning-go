package api

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// rateLimiter is a small in-memory sliding-window limiter keyed by client IP,
// used to slow down brute-force attempts on the auth endpoints. State lives on
// the instance (one per router), so it is naturally reset between tests.
//
// Note: behind a proxy (e.g. Azure App Service) the key comes from
// gin's ClientIP(), which trusts X-Forwarded-For — configure trusted proxies
// before relying on this for anything stronger than basic throttling.
type rateLimiter struct {
	mu     sync.Mutex
	hits   map[string][]time.Time
	limit  int
	window time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{hits: make(map[string][]time.Time), limit: limit, window: window}
}

func (rl *rateLimiter) allow(key string, now time.Time) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := now.Add(-rl.window)
	recent := make([]time.Time, 0, len(rl.hits[key]))
	for _, t := range rl.hits[key] {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}

	if len(recent) >= rl.limit {
		rl.hits[key] = recent
		return false
	}

	rl.hits[key] = append(recent, now)
	return true
}

func (rl *rateLimiter) middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !rl.allow(ctx.ClientIP(), time.Now()) {
			writeError(ctx, http.StatusTooManyRequests, errors.New("too many requests, please try again later"))
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
