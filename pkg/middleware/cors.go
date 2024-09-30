package middleware

import (
	"net/http"
	"strings"

	"git.martianoids.com/martianoids/martian-stack/pkg/httpconst"
	"git.martianoids.com/martianoids/martian-stack/pkg/server"
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
		httpconst.HeaderContentType,
		httpconst.HeaderAccept,
		httpconst.HeaderAcceptLanguage,
		httpconst.HeaderAcceptEncoding,
	}
}

func NewCorsHandler(options CorsOptions) server.Handler {
	return func(c server.Ctx) error {
		if c.Method() == http.MethodOptions {
			c.WithHeader(httpconst.HeaderAccessControlAllowOrigin, options.Origin).
				WithHeader(httpconst.HeaderAccessControlAllowMethods, strings.Join(options.AllowedMethods, ",")).
				WithHeader(httpconst.HeaderAccessControlAllowHeaders, strings.Join(options.AllowedMethods, ", ")).
				WithStatus(http.StatusNoContent)

			return nil
		}

		return c.Next()
	}
}
