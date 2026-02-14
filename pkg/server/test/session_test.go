package server_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/jorgefuertes/martian-stack/pkg/helper"
	"github.com/jorgefuertes/martian-stack/pkg/server"
	"github.com/jorgefuertes/martian-stack/pkg/server/ctx"
	"github.com/jorgefuertes/martian-stack/pkg/server/middleware"
	"github.com/jorgefuertes/martian-stack/pkg/server/web"
	"github.com/jorgefuertes/martian-stack/pkg/service/cache/memory"
	"github.com/jorgefuertes/martian-stack/pkg/service/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerSession(t *testing.T) {
	cacheSvc := memory.New()
	logWriter := helper.NewWriter()
	l := logger.New(logWriter, logger.JsonFormat, logger.LevelInfo)
	srv := server.New(host, port, timeoutSeconds)
	sessMw := middleware.NewSession(cacheSvc, middleware.SessionAutoStart)
	logMw := middleware.NewLog(l)
	srv.Use(logMw, middleware.NewCors(middleware.NewCorsOptions()), sessMw)
	srv.ErrorHandler(testErrorHandlerfunc)

	type sessionResponse struct {
		SessionID string `json:"session_id"`
		Counter   int    `json:"counter"`
		Name      string `json:"name"`
	}

	// routes
	srv.Route(web.MethodGet, "/session/:name", func(c ctx.Ctx) error {
		name := c.Param("name")
		sess := c.Session()

		const counterKey = "counter"
		counter := sess.Data().GetInt(counterKey)

		sess.Data().Set(counterKey, counter+1)

		return c.SendJSON(sessionResponse{SessionID: sess.ID, Counter: counter, Name: name})
	})

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

	var sessionID string
	var cookies []*http.Cookie

	for i := range 10 {
		t.Run(fmt.Sprintf("session %d", i), func(t *testing.T) {
			name := fmt.Sprintf("test_%d", i)
			res, err := call(web.MethodGet, web.MIMEApplicationJSON, cookies, "/session/"+name, nil)
			require.NoError(t, err)
			if res.StatusCode != http.StatusOK {
				t.Logf("RESPONSE: %s", bodyAsString(t, res))
				require.Equal(t, http.StatusOK, res.StatusCode)
			}

			checkLogHas(t, logWriter, logger.LevelInfo, http.StatusOK, "")

			if len(cookies) == 0 {
				assert.NotEmpty(t, res.Cookies())
			}

			sr := new(sessionResponse)
			bodyAsJSON(t, res, sr)

			assert.NotEmpty(t, sr.SessionID)
			if sessionID == "" {
				sessionID = sr.SessionID
			} else {
				assert.Equal(t, sessionID, sr.SessionID)
			}

			if len(res.Cookies()) > 0 {
				cookies = res.Cookies()
			}

			assert.Equal(t, i, sr.Counter)
			assert.Equal(t, name, sr.Name)
		})
	}
}
