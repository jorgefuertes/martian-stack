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
	SessionID string    `json:"session_id"`
	Path      string    `json:"path"`
	Code      int       `json:"code"`
	Status    string    `json:"status"`
}

func checkLogHas(t *testing.T, wr *helper.Writer, level logger.Level, code int, msg string) {
	t.Helper()

	line := new(logLine)
	err := wr.ReadJSON(line)
	require.NoError(t, err)

	assert.NotZero(t, line.Time, "log line time is zero")
	assert.Equal(t, line.Level, level.String(), "log line level is %s not %s", line.Level, level.String())
	assert.Equal(t, line.Component, "server", "log line component is not server %+v", line)
	assert.NotEmpty(t, line.Action, "log line action is empty %+v", line)
	assert.Equal(t, code, line.Code, "log line code is %d not %d: %+v", line.Code, code, line)
	assert.NotEmpty(t, line.IP, "log line IP is empty %+v", line)
	assert.NotEmpty(t, line.Path, "log line Path is empty %+v", line)
	if msg != "" {
		assert.Equal(t, msg, line.Msg, "log line msg is '%s', not '%s': %+v", line.Msg, msg, line)
	}

	wr.Reset()
}
