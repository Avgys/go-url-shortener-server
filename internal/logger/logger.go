package logger

import (
	"context"
	"os"

	"github.com/rs/zerolog"
)

var DefaulLogger = zerolog.New(os.Stderr).With().Timestamp().Logger()

func NewLogger(ctx context.Context, spanID int64) zerolog.Logger {
	return zerolog.
		New(os.Stderr).
		With().
		Timestamp().
		Ctx(ctx).
		Int64("spanID", spanID).
		Logger()
}

func FromContext(ctx context.Context) *zerolog.Logger {
	return zerolog.Ctx(ctx)
}
