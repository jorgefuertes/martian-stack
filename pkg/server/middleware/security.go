package middleware

import (
	"github.com/jorgefuertes/martian-stack/pkg/server/ctx"
	"github.com/jorgefuertes/martian-stack/pkg/server/web"
)

// NewSecurityHeaders adds common security headers to all responses.
func NewSecurityHeaders() ctx.Handler {
	return func(c ctx.Ctx) error {
		c.SetHeader(web.HeaderXContentTypeOptions, "nosniff")
		c.SetHeader(web.HeaderXFrameOptions, "DENY")
		c.SetHeader(web.HeaderReferrerPolicy, "strict-origin-when-cross-origin")
		c.SetHeader(web.HeaderPermissionsPolicy, "geolocation=(), camera=(), microphone=()")
		c.SetHeader(web.HeaderContentSecurityPolicy, "default-src 'self'")
		c.SetHeader(web.HeaderCrossOriginOpenerPolicy, "same-origin")

		return c.Next()
	}
}
