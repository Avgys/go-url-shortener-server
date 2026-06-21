package pool_test

import (
	"sync"
	"testing"

	"go-url-shortener/internal/service/pool"
)

type counter struct {
	n int
}

func (c *counter) Reset() {
	c.n = 0
}

func newCounter() *counter {
	return &counter{}
}

func TestPool_Get_emptyAllocatesAndResets(t *testing.T) {
	t.Parallel()

	p := pool.New(newCounter)

	got := p.Get()
	if got == nil {
		t.Fatal("Get returned nil")
	}
	if got.n != 0 {
		t.Fatalf("got n=%d, want 0", got.n)
	}
}

func TestPool_Put_resetsBeforeStore(t *testing.T) {
	t.Parallel()

	p := pool.New(newCounter)

	c := &counter{n: 42}
	p.Put(c)

	got := p.Get()
	if got != c {
		t.Fatal("expected same instance from pool")
	}
	if got.n != 0 {
		t.Fatalf("got n=%d, want 0 after Put reset", got.n)
	}
}

func TestPool_PutGet_lifoOrder(t *testing.T) {
	t.Parallel()

	p := pool.New(newCounter)

	a := &counter{n: 1}
	b := &counter{n: 2}
	c := &counter{n: 3}

	p.Put(a)
	p.Put(b)
	p.Put(c)

	if got := p.Get(); got != c {
		t.Fatalf("first Get: got %p, want %p", got, c)
	}
	if got := p.Get(); got != b {
		t.Fatalf("second Get: got %p, want %p", got, b)
	}
	if got := p.Get(); got != a {
		t.Fatalf("third Get: got %p, want %p", got, a)
	}
}

func TestPool_Get_afterDrainAllocatesNew(t *testing.T) {
	t.Parallel()

	p := pool.New(newCounter)

	orig := &counter{n: 7}
	p.Put(orig)
	if got := p.Get(); got != orig {
		t.Fatalf("got %p, want %p", got, orig)
	}

	fresh := p.Get()
	if fresh == nil {
		t.Fatal("Get returned nil")
	}
	if fresh == orig {
		t.Fatal("expected new allocation after pool drained")
	}
	if fresh.n != 0 {
		t.Fatalf("got n=%d, want 0", fresh.n)
	}
}

func TestPool_concurrentPutGet(t *testing.T) {
	t.Parallel()

	p := pool.New(newCounter)

	const goroutines = 32
	const rounds = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for range goroutines {
		go func() {
			defer wg.Done()
			for range rounds {
				c := p.Get()
				c.n++
				p.Put(c)
			}
		}()
	}

	wg.Wait()
}
