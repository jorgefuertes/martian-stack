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

func TestSecurityHeaders(t *testing.T) {
	handler := func(c ctx.Ctx) error {
		return c.SendString("ok")
	}

	mw := middleware.NewSecurityHeaders()

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err)

	c := ctx.New(w, req, mw, handler)
	err = c.Next()
	assert.NoError(t, err)

	headers := w.Result().Header
	assert.Equal(t, "nosniff", headers.Get(web.HeaderXContentTypeOptions))
	assert.Equal(t, "DENY", headers.Get(web.HeaderXFrameOptions))
	assert.Equal(t, "strict-origin-when-cross-origin", headers.Get(web.HeaderReferrerPolicy))
	assert.Contains(t, headers.Get(web.HeaderPermissionsPolicy), "geolocation=()")
	assert.Equal(t, "default-src 'self'", headers.Get(web.HeaderContentSecurityPolicy))
	assert.Equal(t, "same-origin", headers.Get(web.HeaderCrossOriginOpenerPolicy))
}
