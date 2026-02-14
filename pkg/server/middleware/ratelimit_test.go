package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jorgefuertes/martian-stack/pkg/server/ctx"
	"github.com/jorgefuertes/martian-stack/pkg/server/middleware"
	"github.com/jorgefuertes/martian-stack/pkg/server/servererror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimit_AllowsUnderLimit(t *testing.T) {
	cfg := middleware.RateLimitConfig{
		Max:    3,
		Window: 1 * time.Second,
	}

	handler := func(c ctx.Ctx) error {
		return c.SendString("ok")
	}

	rl := middleware.NewRateLimit(cfg)

	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		c := ctx.New(rec, req, rl, handler)
		err := c.Next()
		require.NoError(t, err, "request %d should succeed", i+1)
	}
}

func TestRateLimit_BlocksOverLimit(t *testing.T) {
	cfg := middleware.RateLimitConfig{
		Max:    2,
		Window: 1 * time.Second,
	}

	handler := func(c ctx.Ctx) error {
		return c.SendString("ok")
	}

	rl := middleware.NewRateLimit(cfg)

	// First 2 succeed
	for i := 0; i < 2; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.1:12345"
		c := ctx.New(rec, req, rl, handler)
		err := c.Next()
		require.NoError(t, err)
	}

	// 3rd should be rate limited
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	c := ctx.New(rec, req, rl, handler)
	err := c.Next()
	require.Error(t, err)

	sErr, ok := err.(servererror.Error)
	require.True(t, ok)
	assert.Equal(t, http.StatusTooManyRequests, sErr.Code)
}

func TestRateLimit_DifferentIPsIndependent(t *testing.T) {
	cfg := middleware.RateLimitConfig{
		Max:    1,
		Window: 1 * time.Second,
	}

	handler := func(c ctx.Ctx) error {
		return c.SendString("ok")
	}

	rl := middleware.NewRateLimit(cfg)

	// IP 1: 1 request should succeed
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	c := ctx.New(rec, req, rl, handler)
	err := c.Next()
	require.NoError(t, err)

	// IP 2: 1 request should also succeed (different IP)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.2:12345"
	c = ctx.New(rec, req, rl, handler)
	err = c.Next()
	require.NoError(t, err)
}

func TestRateLimit_WindowResets(t *testing.T) {
	cfg := middleware.RateLimitConfig{
		Max:    1,
		Window: 50 * time.Millisecond,
	}

	handler := func(c ctx.Ctx) error {
		return c.SendString("ok")
	}

	rl := middleware.NewRateLimit(cfg)

	// 1st request succeeds
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	c := ctx.New(rec, req, rl, handler)
	err := c.Next()
	require.NoError(t, err)

	// 2nd request fails (over limit)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	c = ctx.New(rec, req, rl, handler)
	err = c.Next()
	require.Error(t, err)

	// Wait for window to expire
	time.Sleep(60 * time.Millisecond)

	// 3rd request succeeds (new window)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	c = ctx.New(rec, req, rl, handler)
	err = c.Next()
	require.NoError(t, err)
}
