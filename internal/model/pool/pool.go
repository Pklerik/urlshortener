package pool

type PoolType interface {
	Reset()
}

type Pool[T PoolType] struct {
	items []T
}

func New[T PoolType](size int) *Pool[T] {
	return &Pool[T]{items: make([]T, size)}
}

func (p *Pool[T]) Get() *T {
	if len(p.items) == 0 {
		return nil
	}
	item := p.items[len(p.items)-1]
	p.items = p.items[:len(p.items)-1]
	return &item
}

func (p *Pool[T]) Put(item T) {
	p.items = append(p.items, item)
}
