package logger

import (
	"context"
	"io"
	"os"

	"github.com/rs/zerolog"
)

func NewLogger() (*zerolog.Logger, func()) {

	f, _ := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)

	logger := zerolog.
		New(io.MultiWriter(os.Stderr, f)).
		With().
		Timestamp().
		Logger()

	return &logger, func() { f.Close() }
}

func NewRequestLogger(ctx context.Context, spanID int64) (*zerolog.Logger, func()) {

	log, close := NewLogger()

	wrappedLog := log.With().
		Timestamp().
		Ctx(ctx).
		Int64("spanID", spanID).
		Stack().
		Logger()

	return &wrappedLog, close
}

func FromContext(ctx context.Context) *zerolog.Logger {
	return zerolog.Ctx(ctx)
}

func Middleware(ctx context.Context, name string) *zerolog.Logger {
	log := zerolog.Ctx(ctx).With().Str("middleware", name).Logger()
	return &log
}

func Endpoint(ctx context.Context, name string) *zerolog.Logger {
	log := zerolog.Ctx(ctx).With().Str("endpoint", name).Logger()
	return &log
}
