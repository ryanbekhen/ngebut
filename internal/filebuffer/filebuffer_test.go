package filebuffer

import (
	"bytes"
	"testing"

	"github.com/valyala/bytebufferpool"
)

func TestBufferPool(t *testing.T) {
	// Get a buffer from the pool
	var buf *bytebufferpool.ByteBuffer = GetBuffer()
	if buf == nil {
		t.Fatal("GetBuffer() returned nil")
	}

	// Check that the buffer is empty
	if buf.Len() != 0 {
		t.Errorf("Expected empty buffer, got length %d", buf.Len())
	}

	// Write some data to the buffer
	testData := []byte("test data")
	buf.Write(testData)

	// Check that the buffer contains the expected data
	if !bytes.Equal(buf.B, testData) {
		t.Errorf("Buffer contains unexpected data")
	}

	// Release the buffer back to the pool
	ReleaseBuffer(buf)

	// Get another buffer from the pool
	buf2 := GetBuffer()
	if buf2 == nil {
		t.Fatal("GetBuffer() returned nil on second call")
	}

	// Check that the buffer is empty (should have been reset)
	if buf2.Len() != 0 {
		t.Errorf("Expected empty buffer after release, got length %d", buf2.Len())
	}

	// Release the second buffer
	ReleaseBuffer(buf2)
}

func TestReadBufferPool(t *testing.T) {
	// Get a read buffer from the pool
	buf := GetReadBuffer()
	if buf == nil {
		t.Fatal("GetReadBuffer() returned nil")
	}

	// Check that the buffer has the expected length
	if len(buf) != 64*1024 {
		t.Errorf("Expected read buffer length 64KB, got %d bytes", len(buf))
	}

	// Write some data to the buffer
	testData := []byte("test data")
	copy(buf, testData)

	// Check that the buffer contains the expected data
	if !bytes.Equal(buf[:len(testData)], testData) {
		t.Errorf("Buffer contains unexpected data")
	}

	// Release the buffer back to the pool
	ReleaseReadBuffer(buf)

	// Get another buffer from the pool
	buf2 := GetReadBuffer()
	if buf2 == nil {
		t.Fatal("GetReadBuffer() returned nil on second call")
	}

	// Check that the buffer has the expected length
	if len(buf2) != 64*1024 {
		t.Errorf("Expected read buffer length 64KB, got %d bytes", len(buf2))
	}

	// Release the second buffer
	ReleaseReadBuffer(buf2)
}

func TestBufferPoolConcurrency(t *testing.T) {
	// Test that the buffer pool works correctly with concurrent access
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				// Get a buffer from the pool
				buf := GetBuffer()
				if buf == nil {
					t.Error("GetBuffer() returned nil in goroutine")
					done <- true
					return
				}

				// Write some data to the buffer
				buf.WriteString("test data")

				// Release the buffer back to the pool
				ReleaseBuffer(buf)
			}
			done <- true
		}()
	}

	// Wait for all goroutines to finish
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

func TestReadBufferPoolConcurrency(t *testing.T) {
	// Test that the read buffer pool works correctly with concurrent access
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				// Get a read buffer from the pool
				buf := GetReadBuffer()
				if buf == nil {
					t.Error("GetReadBuffer() returned nil in goroutine")
					done <- true
					return
				}

				// Write some data to the buffer
				copy(buf, []byte("test data"))

				// Release the buffer back to the pool
				ReleaseReadBuffer(buf)
			}
			done <- true
		}()
	}

	// Wait for all goroutines to finish
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}
