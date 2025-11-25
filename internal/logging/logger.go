package logging

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"
)

// New creates a structured slog.Logger with sane defaults for cloud logging.
func New(env string) *slog.Logger {
	var handler slog.Handler
	level := slog.LevelInfo
	if strings.EqualFold(env, "development") {
		level = slog.LevelDebug
	}

	handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	})
	return slog.New(&callerHandler{Handler: handler})
}

type callerHandler struct {
	slog.Handler
}

func (h *callerHandler) Handle(ctx context.Context, record slog.Record) error {
	if record.PC != 0 {
		frame, _ := runtime.CallersFrames([]uintptr{record.PC}).Next()
		record = record.Clone()
		if frame.Function != "" {
			record.AddAttrs(slog.String("caller", frame.Function))
		}
		if frame.File != "" {
			record.AddAttrs(slog.String("file", fmt.Sprintf("%s:%d", frame.File, frame.Line)))
		}
	}
	return h.Handler.Handle(ctx, record)
}

type contextKey struct{}

// WithContext adds a logger to the context for downstream use.
func WithContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, logger)
}

// FromContext extracts a logger stored via WithContext, falling back to slog.Default().
func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(contextKey{}).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}
