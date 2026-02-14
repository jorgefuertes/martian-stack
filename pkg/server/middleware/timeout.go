package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/jorgefuertes/martian-stack/pkg/server/ctx"
)

// NewTimeout returns a middleware that cancels the request context after the
// given duration. If the handler exceeds the deadline, a 503 Service Unavailable
// error is returned.
func NewTimeout(d time.Duration) ctx.Handler {
	return func(c ctx.Ctx) error {
		reqCtx, cancel := context.WithTimeout(c.Context(), d)
		defer cancel()

		c = c.WithContext(reqCtx)

		done := make(chan error, 1)
		go func() {
			done <- c.Next()
		}()

		select {
		case err := <-done:
			return err
		case <-reqCtx.Done():
			return c.Error(http.StatusServiceUnavailable, "Request timed out")
		}
	}
}
