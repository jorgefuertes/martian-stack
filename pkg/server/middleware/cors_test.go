package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jorgefuertes/martian-stack/pkg/server/ctx"
	"github.com/jorgefuertes/martian-stack/pkg/server/middleware"
	"github.com/jorgefuertes/martian-stack/pkg/server/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCors_Preflight(t *testing.T) {
	opts := middleware.NewCorsOptions()
	mw := middleware.NewCors(opts)

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodOptions, "/api/test", nil)
	require.NoError(t, err)

	c := ctx.New(w, req, mw)
	err = mw(c)
	assert.NoError(t, err)

	headers := w.Result().Header
	assert.Equal(t, "same-origin", headers.Get(web.HeaderAccessControlAllowOrigin))
	assert.Contains(t, headers.Get(web.HeaderAccessControlAllowMethods), http.MethodGet)
	assert.Contains(t, headers.Get(web.HeaderAccessControlAllowHeaders), web.HeaderContentType)
}

func TestCors_ActualRequest(t *testing.T) {
	opts := middleware.NewCorsOptions()
	opts.Origin = "https://example.com"

	handler := func(c ctx.Ctx) error {
		return c.SendString("ok")
	}

	mw := middleware.NewCors(opts)

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/api/test", nil)
	require.NoError(t, err)

	c := ctx.New(w, req, mw, handler)
	err = c.Next()
	assert.NoError(t, err)

	headers := w.Result().Header
	// CORS origin header should be set on actual requests too
	assert.Equal(t, "https://example.com", headers.Get(web.HeaderAccessControlAllowOrigin))
}

func TestCors_CustomOptions(t *testing.T) {
	opts := middleware.CorsOptions{
		Origin:         "https://app.example.com",
		AllowedMethods: []string{http.MethodGet, http.MethodPost},
		AllowedHeaders: []string{"Authorization", "Content-Type", "X-Custom-Header"},
	}
	mw := middleware.NewCors(opts)

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodOptions, "/api/test", nil)
	require.NoError(t, err)

	c := ctx.New(w, req, mw)
	err = mw(c)
	assert.NoError(t, err)

	headers := w.Result().Header
	assert.Equal(t, "https://app.example.com", headers.Get(web.HeaderAccessControlAllowOrigin))
	assert.Contains(t, headers.Get(web.HeaderAccessControlAllowMethods), http.MethodGet)
	assert.Contains(t, headers.Get(web.HeaderAccessControlAllowMethods), http.MethodPost)
	assert.Contains(t, headers.Get(web.HeaderAccessControlAllowHeaders), "Authorization")
	assert.Contains(t, headers.Get(web.HeaderAccessControlAllowHeaders), "X-Custom-Header")
	// Must NOT contain method names in the headers
	assert.NotContains(t, headers.Get(web.HeaderAccessControlAllowHeaders), http.MethodGet)
}
