package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"git.martianoids.com/martianoids/martian-stack/pkg/service/logger"
	"git.martianoids.com/martianoids/martian-stack/pkg/service/server"
	"git.martianoids.com/martianoids/martian-stack/pkg/service/server/middleware"
	"git.martianoids.com/martianoids/martian-stack/pkg/service/server/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	host           = "localhost"
	port           = "8080"
	timeoutSeconds = 300 // high timeout to allow debugging
)

func composeURL(path string) string {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return fmt.Sprintf("http://%s:%s%s", host, port, path)
}

func call(method web.Method, acceptJSON bool, path string, obj any) (*http.Response, error) {
	var req *http.Request
	var err error
	if obj != nil {
		b, _ := json.Marshal(obj)
		reqBodyReader := bytes.NewReader(b)
		req, err = http.NewRequest(method.String(), composeURL(path), reqBodyReader)
		req.Header.Set(web.HeaderContentType, "application/json")
	} else {
		req, err = http.NewRequest(method.String(), composeURL(path), nil)
	}
	if err != nil {
		return nil, err
	}

	if acceptJSON {
		req.Header.Set(web.HeaderAccept, "application/json")
	}

	client := &http.Client{Timeout: timeoutSeconds * time.Second}
	return client.Do(req)
}

func bodyAsString(t *testing.T, res *http.Response) string {
	t.Helper()

	b, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	return string(b)
}

func testErrorHandlerfunc(c server.Ctx, err error) {
	var e server.HttpError
	e, ok := err.(server.HttpError)
	if !ok {
		e = server.HttpError{Code: http.StatusInternalServerError, Msg: err.Error()}
	}
	e.Msg = fmt.Sprintf("TestErrorHandler: %d %s", e.Code, e.Msg)

	if c.AcceptsJSON() {
		_ = c.WithStatus(e.Code).SendJSON(e)
	} else {
		_ = c.WithStatus(e.Code).SendString(e.Error())
	}
}

func TestServer(t *testing.T) {
	l := logger.New(os.Stdout, logger.JsonFormat, logger.LevelInfo)
	srv := server.New(host, port, timeoutSeconds)
	logMw := middleware.NewLogMiddleware(l)
	srv.Use(middleware.NewCorsHandler(middleware.NewCorsOptions()), logMw)
	srv.ErrorHandler(testErrorHandlerfunc)

	// define homepage route
	srv.Route(web.MethodGet, "/", func(c server.Ctx) error {
		return c.SendString("Welcome to the Home Page")
	})

	srv.Route(web.MethodGet, "/hello", func(c server.Ctx) error {
		return c.SendString("Hello, World!")
	})

	srv.Route(web.MethodGet, "/error/500", func(c server.Ctx) error {
		return c.Error(http.StatusInternalServerError, nil)
	})

	// background start
	go func() {
		t.Log("starting server")
		err := srv.Start()
		if err != nil {
			t.Logf("Server: %s", err.Error())
		}
	}()
	time.Sleep(1 * time.Second)

	t.Cleanup(func() {
		// stop gracefully
		err := srv.Stop()
		require.NoError(t, err, "stopping server")
	})

	t.Run("request", func(t *testing.T) {
		t.Run("Home Page", func(t *testing.T) {
			res, err := call(web.MethodGet, false, "/", nil)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, res.StatusCode)
			body := bodyAsString(t, res)
			assert.Equal(t, "Welcome to the Home Page", body)
		})

		t.Run("hello world", func(t *testing.T) {
			res, err := call(web.MethodGet, false, "/hello", nil)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, res.StatusCode)
			body := bodyAsString(t, res)
			assert.Equal(t, "Hello, World!", body)
		})
	})

	t.Run("not found", func(t *testing.T) {
		res, err := call(http.MethodGet, false, "/not-found", nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, res.StatusCode)
		body := bodyAsString(t, res)
		assert.Equal(t, "TestErrorHandler: 404 Resource not found", body)
	})

	t.Run("error 500", func(t *testing.T) {
		res, err := call(http.MethodGet, false, "/error/500", nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
		body := bodyAsString(t, res)
		assert.Equal(t, "TestErrorHandler: 500 Internal Server Error", body)

		t.Run("json error", func(t *testing.T) {
			res, err := call(http.MethodGet, true, "/error/500", nil)
			require.NoError(t, err)
			assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

			e := server.HttpError{}
			err = json.NewDecoder(res.Body).Decode(&e)
			require.NoError(t, err)
			assert.Equal(t, http.StatusInternalServerError, e.Code)
			assert.Equal(t, "TestErrorHandler: 500 Internal Server Error", e.Msg)
		})
	})
}
