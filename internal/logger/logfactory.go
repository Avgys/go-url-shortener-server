package logger

import (
	"context"
	"io"
	"os"
	"runtime"
	"sync/atomic"

	"github.com/rs/zerolog"
)

var funcNameTag string = "func_name_tag"

var discardOutput atomic.Bool

// SetDiscardOutput disables file/stderr logging; NewBaseLogger returns a nop logger.
func SetDiscardOutput(discard bool) {
	discardOutput.Store(discard)
}

func NewBaseLogger(funcName string) (*zerolog.Logger, func() error, error) {
	if discardOutput.Load() {
		log := zerolog.Nop()
		return &log, func() error { return nil }, nil
	}

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

	log, close, err := NewBaseLogger(GetFuncName())

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

func FromContext(ctx context.Context, funcName string) *zerolog.Logger {

	logWithFnName := zerolog.Ctx(ctx).With().Str(funcNameTag, funcName).Logger()
	return &logWithFnName
}

func Middleware(ctx context.Context, name string) (*zerolog.Logger, func() error, error) {

	log, close, err := NewBaseLogger(name)

	if err != nil {
		return nil, nil, err
	}

	wrappedLog := log.With().Str("middleware", name).Logger()

	return &wrappedLog, close, nil
}

func GetFuncName() string {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return "?"
	}
	return runtime.FuncForPC(pc).Name()
}
