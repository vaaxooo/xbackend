package log

import (
	"context"
	"log/slog"
	"os"
)

type Logger interface {
	Debug(ctx context.Context, msg string, args ...any)
	Info(ctx context.Context, msg string, args ...any)
	Warn(ctx context.Context, msg string, args ...any)
	Error(ctx context.Context, msg string, err error, args ...any)
}

type slogLogger struct {
	log *slog.Logger
}

func New(env string) Logger {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	if env == "dev" {
		opts.Level = slog.LevelDebug
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return &slogLogger{
		log: slog.New(handler),
	}
}

func (l *slogLogger) Debug(ctx context.Context, msg string, args ...any) {
	l.log.DebugContext(ctx, msg, args...)
}

func (l *slogLogger) Info(ctx context.Context, msg string, args ...any) {
	l.log.InfoContext(ctx, msg, args...)
}

func (l *slogLogger) Warn(ctx context.Context, msg string, args ...any) {
	l.log.WarnContext(ctx, msg, args...)
}

func (l *slogLogger) Error(ctx context.Context, msg string, err error, args ...any) {
	args = append(args, slog.Any("error", err))
	l.log.ErrorContext(ctx, msg, args...)
}
