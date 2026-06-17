package pool

import (
	"sync"
)

type Resetable interface {
	Reset()
}

type Pool[T Resetable] struct {
	pool []T
	newT func() T

	mux sync.Mutex
}

func New[T Resetable](newT func() T) *Pool[T] {
	return &Pool[T]{
		pool: make([]T, 0),
		newT: newT,
	}
}

func (p *Pool[T]) Put(v T) {

	p.mux.Lock()
	defer p.mux.Unlock()

	v.Reset()
	p.pool = append(p.pool, v)
}

func (p *Pool[T]) Get() T {
	p.mux.Lock()
	defer p.mux.Unlock()

	if len(p.pool) == 0 {
		t := p.newT()
		return t
	}

	lastId := len(p.pool) - 1
	v := p.pool[lastId]

	p.pool = p.pool[:lastId]

	return v
}
