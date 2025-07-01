package filebuffer

import (
	"bytes"
	"sync"
)

// BufferPool is a pool of bytes.Buffer objects for reuse when reading files
// This pool helps reduce memory allocations by reusing buffers instead of creating
// new ones for each file read. The buffers are used to read file content and then
// returned to the pool for future use.
var BufferPool = sync.Pool{
	New: func() interface{} {
		// Pre-allocate with a larger capacity for better performance with larger files
		// 64KB reduces reallocations for common web assets
		return bytes.NewBuffer(make([]byte, 0, 64*1024))
	},
}

// GetBuffer gets a buffer from the pool
func GetBuffer() *bytes.Buffer {
	return BufferPool.Get().(*bytes.Buffer)
}

// ReleaseBuffer returns a buffer to the pool
func ReleaseBuffer(buf *bytes.Buffer) {
	// Reset the buffer to clear its contents
	buf.Reset()
	BufferPool.Put(buf)
}

// ReadBufferPool is a pool of byte slices for reuse when reading files
// This pool helps reduce memory allocations by reusing byte slices instead of creating
// new ones for each file read. The byte slices are used as read buffers and then
// returned to the pool for future use.
var ReadBufferPool = sync.Pool{
	New: func() interface{} {
		// Pre-allocate with a larger capacity for better performance with larger files
		// 64KB reduces reallocations for common web assets
		return make([]byte, 64*1024)
	},
}

// GetReadBuffer gets a read buffer from the pool
func GetReadBuffer() []byte {
	return ReadBufferPool.Get().([]byte)
}

// ReleaseReadBuffer returns a read buffer to the pool
func ReleaseReadBuffer(buf []byte) {
	ReadBufferPool.Put(buf)
}
