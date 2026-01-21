package models

import "sync"

type Resettable interface {
	Reset()
}

type Pool[T Resettable] struct {
	pool *sync.Pool
}

func New[T Resettable]() *Pool[T] {
	return &Pool[T]{
		pool: &sync.Pool{New: func() any {
			var zero T
			return zero
		}},
	}
}

func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

func (p *Pool[T]) Put(v T) {
	v.Reset()
	p.pool.Put(v)
}
