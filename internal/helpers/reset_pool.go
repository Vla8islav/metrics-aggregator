package helpers

type ResetCapable interface {
	Reset()
}
type Pool[T ResetCapable] struct {
	items []T
}

func New[T ResetCapable]() *Pool[T] {
	return &Pool[T]{}
}

func (p *Pool[T]) Get() T {
	if len(p.items) == 0 {
		var zero T
		return zero
	}

	last := len(p.items) - 1
	item := p.items[last]
	// cutting the pool
	p.items = p.items[:last]

	return item
}

func (p *Pool[T]) Put(item T) {
	item.Reset()
	p.items = append(p.items, item)
}
