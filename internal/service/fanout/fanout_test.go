package fanout_test

import (
	"context"
	"testing"

	"github.com/rs/zerolog"

	"go-url-shortener/internal/service/fanout"
)

func testLogger() *zerolog.Logger {
	log := zerolog.Nop()
	return &log
}

func TestFanout_deliversToSubscribers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	f := fanout.New[int](ctx, 2, testLogger())
	sub := f.Take("test", 1)

	f.In <- 42

	if got := <-sub; got != 42 {
		t.Fatalf("got %d, want 42", got)
	}
}

func TestFanout_Take_returnsDistinctChannels(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	f := fanout.New[int](ctx, 0, testLogger())

	a := f.Take("a", 1)
	b := f.Take("b", 1)

	if a == b {
		t.Fatal("expected distinct subscriber channels")
	}
}

func TestFanout_Close(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	f := fanout.New[int](ctx, 1, testLogger())
	sub := f.Take("test", 1)

	if err := f.Close(sub); err != nil {
		t.Fatal(err)
	}

	f.In <- 1

	select {
	case _, ok := <-sub:
		if ok {
			t.Fatal("closed subscriber channel should not deliver values")
		}
	default:
	}
}
