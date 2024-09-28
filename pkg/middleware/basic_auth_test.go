package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"git.martianoids.com/martianoids/martian-stack/pkg/middleware"
	"git.martianoids.com/martianoids/martian-stack/pkg/service/server"
	"git.martianoids.com/martianoids/martian-stack/pkg/service/server/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicAuthMw(t *testing.T) {
	mw := middleware.NewBasicAuthMw("user", "pass")

	t.Run("no auth", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/", nil)
		require.NoError(t, err)
		c := server.NewCtx(w, req)

		err = mw(c)
		assert.Error(t, err)
		httpErr := &server.HttpError{}
		assert.ErrorAs(t, err, httpErr)
		assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
		assert.Equal(t, http.StatusText(http.StatusUnauthorized), httpErr.Msg)
		assert.Equal(t, w.Result().Header.Get(web.HeaderWWWAuthenticate), "Basic realm=\"Restricted\"")
	})

	t.Run("invalid auth", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/", nil)
		require.NoError(t, err)
		req.SetBasicAuth("user", "pass2")
		c := server.NewCtx(w, req)

		err = mw(c)
		assert.Error(t, err)
		httpErr := &server.HttpError{}
		assert.ErrorAs(t, err, httpErr)
		assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
		assert.Equal(t, http.StatusText(http.StatusUnauthorized), httpErr.Msg)
		assert.Equal(t, w.Result().Header.Get(web.HeaderWWWAuthenticate), "Basic realm=\"Restricted\"")
	})

	t.Run("valid auth", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/", nil)
		require.NoError(t, err)
		req.SetBasicAuth("user", "pass")
		c := server.NewCtx(w, req)

		err = mw(c)
		assert.NoError(t, err)
		assert.Equal(t, w.Result().Header.Get(web.HeaderWWWAuthenticate), "")
	})

	t.Run("bad header", func(t *testing.T) {
		w := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodGet, "/", nil)
		require.NoError(t, err)

		testCases := []struct {
			name   string
			header string
		}{
			{"invalid header", "NonValidBasic header"},
			{"invalid encodign", "Basic EIULKJkjj39143"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				req.Header.Set(web.HeaderAuthorization, tc.header)

				c := server.NewCtx(w, req)
				err = mw(c)
				require.Error(t, err)
				httpErr := &server.HttpError{}
				assert.ErrorAs(t, err, httpErr)
				assert.Equal(t, http.StatusBadRequest, httpErr.Code)
				assert.Equal(t, http.StatusText(http.StatusBadRequest), httpErr.Msg)
			})
		}
	})
}
