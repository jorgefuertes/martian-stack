package middleware

import (
	"net/http"
	"strings"

	"github.com/jorgefuertes/martian-stack/pkg/server/ctx"
	"github.com/jorgefuertes/martian-stack/pkg/server/web"
)

type CorsOptions struct {
	Origin         string
	AllowedMethods []string
	AllowedHeaders []string
}

func NewCorsOptions() CorsOptions {
	return CorsOptions{
		Origin:         "same-origin",
		AllowedMethods: DefaultAllowedMethods(),
		AllowedHeaders: DefaultAllowedHeaders(),
	}
}

func DefaultAllowedMethods() []string {
	return []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodOptions,
	}
}

func DefaultAllowedHeaders() []string {
	return []string{
		web.HeaderContentType,
		web.HeaderAccept,
		web.HeaderAcceptLanguage,
		web.HeaderAcceptEncoding,
	}
}

func NewCors(options CorsOptions) ctx.Handler {
	return func(c ctx.Ctx) error {
		c.SetHeader(web.HeaderAccessControlAllowOrigin, options.Origin)

		if c.Method() == http.MethodOptions {
			c.WithHeader(web.HeaderAccessControlAllowMethods, strings.Join(options.AllowedMethods, ", ")).
				WithHeader(web.HeaderAccessControlAllowHeaders, strings.Join(options.AllowedHeaders, ", ")).
				WithStatus(http.StatusNoContent)

			return nil
		}

		return c.Next()
	}
}
