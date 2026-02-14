package logger_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/jorgefuertes/martian-stack/pkg/helper"
	"github.com/jorgefuertes/martian-stack/pkg/service/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoggerService(t *testing.T) {
	const (
		testMsg = "test line"
	)
	fakeError := errors.New("fake error")

	wr := helper.NewWriter()

	t.Run("JSON Logger", func(t *testing.T) {
		l := logger.New(wr, logger.JsonFormat, logger.LevelDebug)

		l.From("test", "testing").Debug(testMsg)
		checkLine(
			t,
			wr,
			logLine{Level: slog.LevelDebug.String(), Component: "test", Action: "testing", Msg: testMsg},
		)

		l.From("test", "testing").Error(fakeError.Error())
		checkLine(
			t,
			wr,
			logLine{Level: slog.LevelError.String(), Component: "test", Action: "testing", Msg: fakeError.Error()},
		)
	})

	t.Run("Text Logger", func(t *testing.T) {
		l := logger.New(wr, logger.TextFormat, logger.LevelDebug)
		l.From("test", "testing").Debug(testMsg)
		textLine := string(getFirstLine(t, wr))
		require.NotEmpty(t, textLine)
		assert.Contains(t, textLine, testMsg)
		l.With("test", true).Info("Text Logger")
		textLine = string(getFirstLine(t, wr))
		assert.Contains(t, textLine, "test=true")
		// pairs
		l.From("test", "testing").Debug("test", "one", "two")
		textLine = string(getFirstLine(t, wr))
		assert.Contains(t, textLine, "msg=test")
		assert.Contains(t, textLine, "one=two")
		// sublogger reuse
		sub := l.From("test", "testing")
		for i := 0; i < 100; i++ {
			sub.Debug("reuse", "count", i)
			textLine = string(getFirstLine(t, wr))
			assert.Contains(t, textLine, "msg=reuse")
			assert.Contains(t, textLine, fmt.Sprintf("count=%d", i))
		}
	})

	t.Run("JSON Logger", func(t *testing.T) {
		l := logger.New(wr, logger.JsonFormat, logger.LevelDebug)
		for i := 0; i < 1000; i++ {
			l.From("test", "testing").Debug(testMsg)
			checkLine(
				t,
				wr,
				logLine{Level: slog.LevelDebug.String(), Component: "test", Action: "testing", Msg: testMsg},
			)
		}
	})
}

func getFirstLine(t *testing.T, w *helper.Writer) []byte {
	b, err := w.Read()
	require.NoError(t, err)
	assert.NotEmpty(t, b)
	return b
}

type logLine struct {
	Level     string    `json:"level"`
	Time      time.Time `json:"time"`
	Component string    `json:"component"`
	Action    string    `json:"action"`
	Msg       string    `json:"msg"`
}

func unmarshalLine(t *testing.T, data []byte) logLine {
	require.NotEmpty(t, data)
	var line logLine
	require.NoError(t, json.Unmarshal(data, &line))
	return line
}

func checkLine(t *testing.T, w *helper.Writer, expectedLine logLine) {
	b := getFirstLine(t, w)
	line := unmarshalLine(t, b)
	assert.NotZero(t, line.Time)
	expectedLine.Time = line.Time
	assert.EqualValues(t, expectedLine, line)
}
