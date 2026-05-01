package core_logger

import (
	"context"
	"log/slog"
	"os"
)

type Logger struct {
	*slog.Logger
}

func NewLogger() *Logger {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: false,
	})

	// Создаем slog.Logger и оборачиваем в наш Logger
	slogLogger := slog.New(handler)

	return &Logger{
		Logger: slogLogger,
	}
}
func FromContext(ctx context.Context) *Logger {
	log, ok := ctx.Value("log").(*Logger)
	if !ok {
		panic("no logger in context")
	}
	return log
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
