package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"git.martianoids.com/martianoids/martian-stack/pkg/server/ctx"
	"git.martianoids.com/martianoids/martian-stack/pkg/server/middleware"
	"git.martianoids.com/martianoids/martian-stack/pkg/server/servererror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecovery_NoPanic(t *testing.T) {
	handler := func(c ctx.Ctx) error {
		return c.SendString("ok")
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c := ctx.New(rec, req, middleware.NewRecovery(), handler)

	err := c.Next()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRecovery_WithPanic(t *testing.T) {
	handler := func(c ctx.Ctx) error {
		panic("something went wrong")
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c := ctx.New(rec, req, middleware.NewRecovery(), handler)

	err := c.Next()
	require.Error(t, err)

	sErr, ok := err.(servererror.Error)
	require.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, sErr.Code)
	assert.Contains(t, sErr.Msg, "something went wrong")
}

func TestRecovery_WithPanicError(t *testing.T) {
	handler := func(c ctx.Ctx) error {
		panic(42)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c := ctx.New(rec, req, middleware.NewRecovery(), handler)

	err := c.Next()
	require.Error(t, err)

	sErr, ok := err.(servererror.Error)
	require.True(t, ok)
	assert.Contains(t, sErr.Msg, "42")
}

func TestRecovery_HandlerError(t *testing.T) {
	handler := func(c ctx.Ctx) error {
		return c.Error(http.StatusBadRequest, "bad request")
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c := ctx.New(rec, req, middleware.NewRecovery(), handler)

	err := c.Next()
	require.Error(t, err)

	sErr, ok := err.(servererror.Error)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, sErr.Code)
}
