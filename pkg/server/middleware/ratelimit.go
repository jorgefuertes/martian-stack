package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/jorgefuertes/martian-stack/pkg/server/ctx"
)

// RateLimitConfig configures the rate limiter.
type RateLimitConfig struct {
	// Maximum number of requests allowed within the window
	Max int

	// Time window for counting requests
	Window time.Duration

	// CleanupInterval controls how often expired entries are removed.
	// Defaults to 2x Window if zero.
	CleanupInterval time.Duration
}

// DefaultRateLimitConfig returns a config allowing 60 requests per minute.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Max:    60,
		Window: 1 * time.Minute,
	}
}

type visitor struct {
	count    int
	windowAt time.Time
}

// NewRateLimit returns a middleware that limits requests per IP address
// using a fixed-window counter algorithm.
func NewRateLimit(cfg RateLimitConfig) ctx.Handler {
	if cfg.CleanupInterval == 0 {
		cfg.CleanupInterval = 2 * cfg.Window
	}

	var mu sync.Mutex
	visitors := make(map[string]*visitor)

	// background cleanup of expired entries
	go func() {
		ticker := time.NewTicker(cfg.CleanupInterval)
		defer ticker.Stop()

		for range ticker.C {
			mu.Lock()
			now := time.Now()
			for ip, v := range visitors {
				if now.Sub(v.windowAt) > cfg.Window {
					delete(visitors, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c ctx.Ctx) error {
		ip := c.UserIP()

		mu.Lock()
		v, exists := visitors[ip]
		now := time.Now()

		if !exists || now.Sub(v.windowAt) > cfg.Window {
			visitors[ip] = &visitor{count: 1, windowAt: now}
			mu.Unlock()
			return c.Next()
		}

		v.count++
		if v.count > cfg.Max {
			mu.Unlock()
			return c.Error(http.StatusTooManyRequests, "Rate limit exceeded")
		}

		mu.Unlock()
		return c.Next()
	}
}
