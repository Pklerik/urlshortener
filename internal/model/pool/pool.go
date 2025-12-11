// Package pool provide pool schema.
package pool

// Type provide interface for resettable types.
type Type interface {
	Reset()
}

// Pool generic struct for all Resettable types.
type Pool[T Type] struct {
	items []T
}

// New provide new Pool with size.
func New[T Type](size int) *Pool[T] {
	return &Pool[T]{items: make([]T, size)}
}

// Get take item from Pool.items []T.
func (p *Pool[T]) Get() *T {
	if len(p.items) == 0 {
		return nil
	}

	item := p.items[len(p.items)-1]
	p.items = p.items[:len(p.items)-1]

	return &item
}

// Put store item in Pool.items []T.
func (p *Pool[T]) Put(item T) {
	p.items = append(p.items, item)
}
