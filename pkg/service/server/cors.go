package server

import (
	"net/http"
	"strings"
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
		HeaderContentType,
		HeaderAccept,
		HeaderAcceptLanguage,
		HeaderAcceptEncoding,
	}
}

func newCorsHandler(options CorsOptions) Handler {
	return func(c Ctx) error {
		if c.Method() == http.MethodOptions {
			c.WithHeader(HeaderAccessControlAllowOrigin, options.Origin).
				WithHeader(HeaderAccessControlAllowMethods, strings.Join(options.AllowedMethods, ",")).
				WithHeader(HeaderAccessControlAllowHeaders, strings.Join(options.AllowedMethods, ", ")).
				WithStatus(http.StatusNoContent).Next()
		}

		return nil
	}
}
