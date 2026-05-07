package closer

import (
	"context"
	"errors"
	"io"
	"sync"
)

type closer func(ctx context.Context) error

// Closer is a helper to close multiple closers.
type Closer struct {
	closers  []io.Closer
	mu       sync.Mutex
	isDone   chan struct{}
	isClosed bool
	once     sync.Once
}

// NewCloser creates a new Closer.
func NewCloser() *Closer {
	return &Closer{
		closers:  make([]io.Closer, 0),
		isDone:   make(chan struct{}),
		isClosed: false,
	}
}

// Add adds a closer to the Closer.
func (c *Closer) Add(cl io.Closer) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.isClosed {
		return
	}
	c.closers = append(c.closers, cl)
}

// Done returns a channel that's closed when Close is called.
func (c *Closer) Done() chan struct{} {
	return c.isDone
}

// Close closes all closers.
func (c *Closer) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isClosed {
		return nil
	}
	c.isClosed = true

	defer func() {
		c.once.Do(func() {
			close(c.isDone)
		})
	}()

	var resultErr []error
	for i := len(c.closers) - 1; i >= 0; i-- {
		if err := c.closers[i].Close(); err != nil {
			resultErr = append(resultErr, err)
		}
	}

	return errors.Join(resultErr...)
}
