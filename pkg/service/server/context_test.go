package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCtxMiddleware(t *testing.T) {
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
	for i := range 3 {
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

func TestResponseWriting(t *testing.T) {
	getNewCtx := func() (*httptest.ResponseRecorder, *http.Request, Ctx) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		c := newCtx(w, req)

		return w, req, c
	}

	t.Run("HTML with status 400", func(t *testing.T) {
		w, _, c := getNewCtx()
		html := "<p>Hello world</p>"
		err := c.WithStatus(http.StatusBadRequest).SendHtml(html)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
		assert.Equal(t, html, w.Body.String())
	})

	t.Run("TEXT with status OK", func(t *testing.T) {
		w, _, c := getNewCtx()
		s := "Hello world!"
		err := c.SendString(s)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, s, w.Body.String())
	})

	t.Run("JSON with status OK", func(t *testing.T) {
		w, _, c := getNewCtx()
		type testStruct struct {
			Num int    `json:"num"`
			Str string `json:"str"`
		}
		obj := &testStruct{Num: 1, Str: "one"}
		err := c.SendJSON(obj)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Result().StatusCode)

		obj2 := new(testStruct)
		err = json.Unmarshal(w.Body.Bytes(), obj2)
		require.NoError(t, err)
		assert.EqualValues(t, obj, obj2)
	})
}
