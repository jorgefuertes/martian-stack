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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	host           = "localhost"
	port           = "8080"
	timeoutSeconds = 10
)

func composeURL(path string) string {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return fmt.Sprintf("http://%s:%s%s", host, port, path)
}

func registerRoutes(srv *server.Server) {
	srv.Route("GET", "/hello", func(c server.Ctx) error {
		return c.SendString("Hello, World!")
	})
}

func call(method, path string, obj any) (*http.Response, error) {
	var req *http.Request
	var err error
	if obj != nil {
		b, _ := json.Marshal(obj)
		reqBodyReader := bytes.NewReader(b)
		req, err = http.NewRequest(method, composeURL(path), reqBodyReader)
	} else {
		req, err = http.NewRequest(method, composeURL(path), nil)
	}
	if err != nil {
		return nil, err
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

func TestServer(t *testing.T) {
	l := logger.New(os.Stdout, logger.TextFormat, logger.LevelInfo)
	srv := server.New(host, port, timeoutSeconds)
	logMw := middleware.NewLogMiddleware(l)
	srv.Use(middleware.NewCorsHandler(middleware.NewCorsOptions()))
	srv.Use(logMw)

	registerRoutes(srv)

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
		t.Run("hello world", func(t *testing.T) {
			res, err := call(http.MethodGet, "/hello", nil)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, res.StatusCode)
			body := bodyAsString(t, res)
			assert.Equal(t, "Hello, World!", body)
		})
	})
}
