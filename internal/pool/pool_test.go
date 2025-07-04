package pool

import (
	"testing"
)

func TestPool(t *testing.T) {
	// Test creating a new pool with a factory function
	factory := func() int { return 42 }
	p := New(factory)

	// Test Get returns the expected value
	val := p.Get()
	if val != 42 {
		t.Errorf("Expected 42, got %d", val)
	}

	// Test Put and Get work together
	p.Put(100)
	val = p.Get()
	if val != 100 {
		t.Errorf("Expected 100, got %d", val)
	}

	// Test that after exhausting the pool, the factory is called again
	val = p.Get()
	if val != 42 {
		t.Errorf("Expected factory value 42, got %d", val)
	}
}

func TestPoolWithStruct(t *testing.T) {
	type testStruct struct {
		Value string
		Count int
	}

	// Test with a struct type
	factory := func() *testStruct { return &testStruct{Value: "initial", Count: 0} }
	p := New(factory)

	// Get an item and modify it
	item := p.Get()
	item.Value = "modified"
	item.Count = 5

	// Put it back in the pool
	p.Put(item)

	// Get it again and verify it's the same modified item
	item2 := p.Get()
	if item2.Value != "modified" || item2.Count != 5 {
		t.Errorf("Expected modified values, got Value=%s, Count=%d", item2.Value, item2.Count)
	}
}

func TestBufferPool(t *testing.T) {
	// Test creating a new buffer pool
	factory := func(size int) []byte { return make([]byte, 0, size) }
	bp := NewBuffer(10, factory)

	// Test Get returns a buffer with the expected capacity
	buf := bp.Get()
	if cap(buf) != 10 {
		t.Errorf("Expected capacity 10, got %d", cap(buf))
	}
	if len(buf) != 0 {
		t.Errorf("Expected length 0, got %d", len(buf))
	}

	// Test Put and Get work together
	buf = append(buf, []byte("test data")...)
	bp.Put(buf)

	buf2 := bp.Get()
	// Length should be reset to 0
	if len(buf2) != 0 {
		t.Errorf("Expected length 0, got %d", len(buf2))
	}
	// Capacity should still be at least 10
	if cap(buf2) < 10 {
		t.Errorf("Expected capacity at least 10, got %d", cap(buf2))
	}
}

func TestBufferPoolGetWithSize(t *testing.T) {
	// Test creating a new buffer pool
	factory := func(size int) []byte { return make([]byte, 0, size) }
	bp := NewBuffer(10, factory)

	// Test GetWithSize returns a buffer with at least the requested capacity
	buf := bp.GetWithSize(5)
	if cap(buf) < 5 {
		t.Errorf("Expected capacity at least 5, got %d", cap(buf))
	}

	// Test GetWithSize with a size larger than the default
	buf = bp.GetWithSize(20)
	if cap(buf) < 20 {
		t.Errorf("Expected capacity at least 20, got %d", cap(buf))
	}

	// Put a buffer back and get one with a smaller size
	bp.Put(buf)
	buf = bp.GetWithSize(15)
	// Should reuse the buffer with capacity 20
	if cap(buf) < 15 {
		t.Errorf("Expected capacity at least 15, got %d", cap(buf))
	}
}

func BenchmarkPool(b *testing.B) {
	factory := func() int { return 42 }
	p := New(factory)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		val := p.Get()
		p.Put(val)
	}
}

func BenchmarkBufferPool(b *testing.B) {
	factory := func(size int) []byte { return make([]byte, 0, size) }
	bp := NewBuffer(1024, factory)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := bp.Get()
		buf = append(buf, []byte("test data")...)
		bp.Put(buf)
	}
}

func BenchmarkBufferPoolGetWithSize(b *testing.B) {
	factory := func(size int) []byte { return make([]byte, 0, size) }
	bp := NewBuffer(1024, factory)

	sizes := []int{512, 1024, 2048, 4096}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		size := sizes[i%len(sizes)]
		buf := bp.GetWithSize(size)
		buf = append(buf, []byte("test data")...)
		bp.Put(buf)
	}
}
