package logger

import (
	"context"
	"io"
	"os"
	"runtime"

	"github.com/rs/zerolog"
)

var funcNameTag string = "func_name_tag"
var defaultFuncNameDepth = 2

func newBaseLogger() (*zerolog.Logger, func() error, error) {

	f, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)

	if err != nil {
		return nil, nil, err
	}

	logger := zerolog.
		New(io.MultiWriter(os.Stderr, f)).
		With().
		Timestamp().
		Logger()

	return &logger, f.Close, nil
}

func NewRequestLogger(ctx context.Context, spanID int64) (*zerolog.Logger, func() error, error) {

	log, close, err := newBaseLogger()

	if err != nil {
		return nil, nil, err
	}

	wrappedLog := log.With().
		Ctx(ctx).
		Int64("spanID", spanID).
		Stack().
		Logger()

	return &wrappedLog, close, nil
}

func FromContext(ctx context.Context) *zerolog.Logger {
	funcName := callerFuncName(defaultFuncNameDepth)

	logWithFnName := zerolog.Ctx(ctx).With().Str(funcNameTag, funcName).Logger()
	return &logWithFnName
}

func Middleware(ctx context.Context, name string) (*zerolog.Logger, func() error, error) {

	log, close, err := newBaseLogger()

	if err != nil {
		return nil, nil, err
	}

	wrappedLog := log.With().Str("middleware", name).Logger()

	return &wrappedLog, close, nil
}

func callerFuncName(callLevel int) string {
	pc, _, _, ok := runtime.Caller(callLevel)
	if !ok {
		return "?"
	}
	return runtime.FuncForPC(pc).Name()
}
