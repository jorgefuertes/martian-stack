package server_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/jorgefuertes/martian-stack/pkg/helper"
	"github.com/jorgefuertes/martian-stack/pkg/server"
	"github.com/jorgefuertes/martian-stack/pkg/server/middleware"
	"github.com/jorgefuertes/martian-stack/pkg/server/servererror"
	"github.com/jorgefuertes/martian-stack/pkg/server/web"
	"github.com/jorgefuertes/martian-stack/pkg/service/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	logWriter := helper.NewWriter()
	l := logger.New(logWriter, logger.JsonFormat, logger.LevelInfo)
	srv := server.New(host, port, timeoutSeconds)
	logMw := middleware.NewLog(l)
	srv.Use(middleware.NewCors(middleware.NewCorsOptions()), logMw)
	srv.ErrorHandler(testErrorHandlerfunc)

	// test routes
	registerRoutes(srv)

	// background start
	go func() {
		t.Log("starting server")
		err := srv.Start()
		if err != nil {
			t.Logf("Server: %s", err.Error())
		}
	}()
	srv.WaitUntilReady()

	t.Cleanup(func() {
		// stop gracefully
		err := srv.Stop()
		require.NoError(t, err, "stopping server")
	})

	t.Run("request", func(t *testing.T) {
		t.Run("Home Page", func(t *testing.T) {
			res, err := call(web.MethodGet, "", nil, "/", nil)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, res.StatusCode)
			body := bodyAsString(t, res)
			assert.Equal(t, "Welcome to the Home Page", body)
			checkLogHas(t, logWriter, logger.LevelInfo, http.StatusOK, "")
		})

		t.Run("hello world", func(t *testing.T) {
			res, err := call(web.MethodGet, "", nil, "/hello", nil)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, res.StatusCode)
			body := bodyAsString(t, res)
			assert.Equal(t, "Hello, World!", body)
			checkLogHas(t, logWriter, logger.LevelInfo, http.StatusOK, "")
		})
	})

	t.Run("not found", func(t *testing.T) {
		res, err := call(http.MethodGet, "", nil, "/not-found", nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, res.StatusCode)
		body := bodyAsString(t, res)
		assert.Equal(t, "TestErrorHandler: 404 Resource not found", body)
		checkLogHas(t, logWriter, logger.LevelError, http.StatusNotFound, "Resource not found")
	})

	t.Run("error 500", func(t *testing.T) {
		res, err := call(http.MethodGet, "", nil, "/error/500", nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
		body := bodyAsString(t, res)
		assert.Equal(t, "TestErrorHandler: 500 Internal Server Error", body)
		checkLogHas(t, logWriter, logger.LevelError, http.StatusInternalServerError, "Internal Server Error")

		t.Run("json error", func(t *testing.T) {
			res, err := call(http.MethodGet, web.MIMEApplicationJSON, nil, "/error/500", nil)
			require.NoError(t, err)
			assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

			e := servererror.New()
			err = json.NewDecoder(res.Body).Decode(&e)
			require.NoError(t, err)
			assert.Equal(t, http.StatusInternalServerError, e.Code)
			assert.Equal(t, "TestErrorHandler: 500 Internal Server Error", e.Msg)
			checkLogHas(t, logWriter, logger.LevelError, http.StatusInternalServerError, "Internal Server Error")
		})
	})

	t.Run("path params", func(t *testing.T) {
		res, err := call(http.MethodGet, "", nil, "/param-test/John/30", nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		body := bodyAsString(t, res)
		assert.Equal(t, "Hello, John! You are 30 years old.", body)
		checkLogHas(t, logWriter, logger.LevelInfo, http.StatusOK, "")
	})

	t.Run("path params with url encoded", func(t *testing.T) {
		res, err := call(http.MethodGet, "", nil, "/param-test/John%20Smith/30", nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		body := bodyAsString(t, res)
		assert.Equal(t, "Hello, John Smith! You are 30 years old.", body)
		checkLogHas(t, logWriter, logger.LevelInfo, http.StatusOK, "")
	})

	t.Run("query params", func(t *testing.T) {
		res, err := call(http.MethodGet, "", nil, "/param-query-test?name=John&age=30", nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		body := bodyAsString(t, res)
		assert.Equal(t, "Hello, John! You are 30 years old.", body)
		checkLogHas(t, logWriter, logger.LevelInfo, http.StatusOK, "")
	})

	t.Run("json post", func(t *testing.T) {
		u := user{Name: "John", Age: 30}

		res, err := call(http.MethodPost, "", nil, "/post-json-test", u)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		body := bodyAsString(t, res)
		assert.Equal(t, "Hello, John! You are 30 years old.", body)
		checkLogHas(t, logWriter, logger.LevelInfo, http.StatusOK, "")
	})

	t.Run("json reply", func(t *testing.T) {
		res, err := call(http.MethodGet, "", nil, "/json-reply-test", nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, web.MIMEApplicationJSON, res.Header.Get(web.HeaderContentType))
		body := bodyAsString(t, res)
		assert.Equal(t, `{"name":"John","age":30}`, body)
		checkLogHas(t, logWriter, logger.LevelInfo, http.StatusOK, "")
	})
}
