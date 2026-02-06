package logger

import (
	"context"
	"os"

	"github.com/rs/zerolog"
)

func NewLogger() zerolog.Logger {
	return zerolog.
		New(os.Stderr).
		With().
		Timestamp().
		Logger()
}

func NewRequestLogger(ctx context.Context, spanID int64) zerolog.Logger {
	return zerolog.
		New(os.Stderr).
		With().
		Timestamp().
		Ctx(ctx).
		Int64("spanID", spanID).
		Stack().
		Logger()
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
