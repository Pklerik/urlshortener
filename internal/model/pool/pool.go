// Package pool provide pool schema.
package pool

import "sync"

// Type provide interface for resettable types.
type Type interface {
	Reset()
}

// Pool generic struct for all Resettable types.
type Pool[T Type] struct {
	pool sync.Pool
}

// New provide new Pool with size.
func New[T Type]() *Pool[T] {
	return &Pool[T]{pool: sync.Pool{New: func() any {
		return new(T)
	}}}
}

// Get take item from Pool.pool []T.
func (p *Pool[T]) Get() *T {
	item := p.pool.Get()
	if item != nil {
		return item.(*T)
	}

	return new(T)
}

// Put store item in Pool.pool []T.
func (p *Pool[T]) Put(item T) {
	p.pool.Put(item)
}
