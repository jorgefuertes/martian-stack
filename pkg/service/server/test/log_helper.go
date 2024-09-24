package server_test

import (
	"testing"
	"time"

	"git.martianoids.com/martianoids/martian-stack/pkg/helper"
	"git.martianoids.com/martianoids/martian-stack/pkg/service/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type logLine struct {
	Time      time.Time `json:"time"`
	Level     string    `json:"level"`
	Msg       string    `json:"msg"`
	Component string    `json:"component"`
	Action    string    `json:"action"`
	IP        string    `json:"ip"`
	Path      string    `json:"path"`
	Code      int       `json:"code"`
	Status    string    `json:"status"`
	ErrMsg    string    `json:"error"`
}

func checkLogHas(t *testing.T, wr *helper.Writer, level logger.Level, code int, errMsg string) {
	t.Helper()

	line := new(logLine)
	err := wr.ReadJSON(line)
	require.NoError(t, err)

	assert.NotZero(t, line.Time, "log line time is zero")
	assert.Equal(t, line.Level, level.String(), "log line level is %s not %s", line.Level, level.String())
	assert.Equal(t, line.Component, "server", "log line component is not server")
	assert.NotEmpty(t, line.Action, "log line action is empty")
	assert.Equal(t, code, line.Code, "log line code is %d not %d", line.Code, code)
	assert.Contains(t, []string{"OK", "FAILED"}, line.Msg, "log line msg is not OK nor FAILED: %s", line.Msg)
	assert.NotEmpty(t, line.IP, "log line IP is empty")
	assert.NotEmpty(t, line.Path, "log line Path is empty")
	if errMsg != "" {
		assert.Equal(t, errMsg, line.ErrMsg, "log line error is %s not %s", line.ErrMsg, errMsg)
	}

	wr.Reset()
}
