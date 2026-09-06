package logger

import (
	"io"
	"log/slog"
	"os"
)

type Config struct {
	Output io.Writer
	JSON   bool
	Level  slog.Level
}

func New(cfg Config) *slog.Logger {
	if cfg.Output == nil {
		cfg.Output = os.Stdout
	}
	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: cfg.Level}
	if cfg.JSON {
		handler = slog.NewJSONHandler(cfg.Output, opts)
	} else {
		handler = slog.NewTextHandler(cfg.Output, opts)
	}
	return slog.New(handler)
}
func With(args ...any) *slog.Logger { return defaultLogger.With(args...) }

var defaultLogger = slog.New(
	slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}),
)

func Info(msg string, args ...any) {
	defaultLogger.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	defaultLogger.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
}

func Debug(msg string, args ...any) {
	defaultLogger.Debug(msg, args...)
}
