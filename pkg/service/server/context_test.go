package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCtx(t *testing.T) {
	makeMw := func(t *testing.T, order int) Handler {
		return func(c Ctx) error {
			msg := fmt.Sprintf("mw %d in\n", order)
			require.NoError(t, c.SendString(msg))
			err := c.Next()
			msg = fmt.Sprintf("mw %d out\n", order)
			require.NoError(t, c.SendString(msg))
			return err
		}
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c := newCtx(w, req, makeMw(t, 0), makeMw(t, 1), makeMw(t, 2))
	// execute the whole chain
	require.NoError(t, c.Next())
	for i := 0; i < 3; i++ {
		line, err := w.Body.ReadString('\n')
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("mw %d in\n", i), line)
	}
	for i := 2; i >= 0; i-- {
		line, err := w.Body.ReadString('\n')
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("mw %d out\n", i), line)
	}
}
