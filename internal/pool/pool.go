package pool

import (
	"sync"
)

// Pool is a generic sync.Pool wrapper that provides type safety and convenience methods.
// It's designed to be used with any type, reducing boilerplate code and making pool usage safer.
type Pool[T any] struct {
	pool sync.Pool
}

// New creates a new Pool with the given factory function.
// The factory function is called when the pool needs to create a new item.
func New[T any](factory func() T) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() interface{} {
				return factory()
			},
		},
	}
}

// Get retrieves an item from the pool, or creates a new one if the pool is empty.
func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

// Put returns an item to the pool.
func (p *Pool[T]) Put(x T) {
	p.pool.Put(x)
}

// BufferPool is a specialized pool for byte buffers.
// It provides additional methods for working with buffers.
type BufferPool[T ~[]byte] struct {
	Pool[T]
	size int
}

// Get retrieves a buffer from the pool.
// The buffer's length is reset to 0 but its capacity is preserved.
func (p *BufferPool[T]) Get() T {
	buf := p.Pool.Get()
	return buf[:0] // Reset the buffer length to 0 but keep the capacity
}

// NewBuffer creates a new BufferPool with the given size.
// The size is used as the initial capacity for new buffers.
func NewBuffer[T ~[]byte](size int, factory func(size int) T) *BufferPool[T] {
	return &BufferPool[T]{
		Pool: Pool[T]{
			pool: sync.Pool{
				New: func() interface{} {
					return factory(size)
				},
			},
		},
		size: size,
	}
}

// GetWithSize retrieves a buffer from the pool with at least the specified size.
// If the buffer from the pool is too small, a new one is created.
func (p *BufferPool[T]) GetWithSize(size int) T {
	buf := p.Get()
	if cap(buf) < size {
		// Return the too-small buffer to the pool
		p.Put(buf)
		// Create a new buffer with the requested size
		return make(T, 0, size)
	}
	return buf[:0] // Reset the buffer length to 0 but keep the capacity
}
