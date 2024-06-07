package logger

import (
	"io"
	"log/slog"
)

type Format uint8
type Level slog.Level

func (level Level) Level() slog.Level { return slog.Level(level) }

const (
	TextFormat Format = 0
	JsonFormat Format = 1
	LevelDebug Level  = Level(slog.LevelDebug)
	LevelInfo  Level  = Level(slog.LevelInfo)
	LevelWarn  Level  = Level(slog.LevelWarn)
	LevelError Level  = Level(slog.LevelError)
)

type LogKey string

const (
	Component LogKey = "component"
	Action    LogKey = "action"
)

func (l LogKey) String() string {
	return string(l)
}

type Service struct {
	handler slog.Handler
}

func New(wr io.Writer, format Format, level Level) *Service {
	return &Service{handler: newHandlerFor(wr, format, level)}
}

func newHandlerFor(wr io.Writer, format Format, level Level) slog.Handler {
	if format == JsonFormat {
		return slog.NewJSONHandler(wr, &slog.HandlerOptions{AddSource: false, Level: level})
	}

	return slog.NewTextHandler(wr, &slog.HandlerOptions{AddSource: false, Level: LevelDebug})
}

func (s *Service) logger() *slog.Logger {
	return slog.New(s.handler)
}

func (s *Service) From(component, action string) *slog.Logger {
	return s.logger().With(Component.String(), component, Action.String(), action)
}

func (s *Service) With(args ...any) *slog.Logger {
	return s.logger().With(args...)
}
