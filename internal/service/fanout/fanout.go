package fanout

import (
	"context"
	"fmt"
	"sync"

	"github.com/rs/zerolog"
	"github.com/samber/lo"
)

const defaultBuffer = 10

// Fanout fans out items received on In to subscriber channels created by [Fanout.Take].
type Fanout[T any] struct {
	subs []chan T
	In   chan T

	logger *zerolog.Logger
	mu     sync.RWMutex
}

// New creates a Fanout with a buffered input channel and starts delivery until ctx is done.
// buffer <= 0 uses [defaultBuffer].
func New[T any](ctx context.Context, buffer int, logger *zerolog.Logger) *Fanout[T] {
	if buffer <= 0 {
		buffer = defaultBuffer
	}

	innerLog := logger.With().Str("service", "Fanout").Logger()

	f := &Fanout[T]{
		subs:   make([]chan T, 0),
		In:     make(chan T, buffer),
		logger: &innerLog,
	}

	go f.run(ctx)

	return f
}

// Take creates a buffered subscriber channel and registers it for delivery.
// subBuffer <= 0 uses [defaultBuffer].
func (f *Fanout[T]) Take(subBuffer int) chan T {
	if subBuffer <= 0 {
		subBuffer = defaultBuffer
	}

	ch := make(chan T, subBuffer)

	f.mu.Lock()
	f.subs = append(f.subs, ch)
	f.mu.Unlock()

	return ch
}

// Close removes a subscriber channel created by [Fanout.Take].
func (f *Fanout[T]) Close(ch chan T) error {
	if ch == nil {
		return fmt.Errorf("subscriber channel is nil")
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	next := lo.Filter(f.subs, func(sub chan T, _ int) bool {
		return sub != ch
	})

	if len(next) == len(f.subs) {
		return fmt.Errorf("subscriber not found")
	}

	f.subs = next
	close(ch)

	return nil
}

func (f *Fanout[T]) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case item := <-f.In:
			f.deliver(item)
		}
	}
}

func (f *Fanout[T]) deliver(item T) {
	f.mu.RLock()
	subs := append([]chan T(nil), f.subs...)
	f.mu.RUnlock()

	for _, sub := range subs {
		select {
		case sub <- item:
		default:
		}
	}
}
