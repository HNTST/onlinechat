package core_logger

import (
	"context"
	"log/slog"
	"os"
)

type Logger struct {
	*slog.Logger
}

type contextKey struct{}

var loggerKey = contextKey{}

func NewLogger() *Logger {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: false,
	})

	slogLogger := slog.New(handler)

	return &Logger{
		Logger: slogLogger,
	}
}

func FromContext(ctx context.Context) *Logger {
	log, ok := ctx.Value(loggerKey).(*Logger)
	if !ok {
		return NewLogger()
	}
	return log
}

func ToContext(ctx context.Context, l *Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

func (l *Logger) With(fields ...slog.Attr) *Logger {
	args := make([]any, len(fields))
	for i, v := range fields {
		args[i] = v
	}
	return &Logger{
		Logger: l.Logger.With(args...),
	}
}
