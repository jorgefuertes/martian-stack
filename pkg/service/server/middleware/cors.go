package middleware

import (
	"net/http"
	"strings"

	"git.martianoids.com/martianoids/martian-stack/pkg/service/server"
	"git.martianoids.com/martianoids/martian-stack/pkg/service/server/web"
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

func NewCorsHandler(options CorsOptions) server.Handler {
	return func(c server.Ctx) error {
		if c.Method() == http.MethodOptions {
			c.WithHeader(web.HeaderAccessControlAllowOrigin, options.Origin).
				WithHeader(web.HeaderAccessControlAllowMethods, strings.Join(options.AllowedMethods, ",")).
				WithHeader(web.HeaderAccessControlAllowHeaders, strings.Join(options.AllowedMethods, ", ")).
				WithStatus(http.StatusNoContent)

			return nil
		}

		return c.Next()
	}
}
