package ngebut

import (
	"net/http"
	"sync"
)

// ResponseWriter is the interface used by an HTTP handler to construct an HTTP response.
type ResponseWriter interface {
	// Header returns the header map that will be sent by WriteHeader.
	// The Header map also is the mechanism with which
	// Handlers can set HTTP trailers.
	Header() Header

	// Write writes the data to the connection as part of an HTTP reply.
	Write([]byte) (int, error)

	// WriteHeader sends an HTTP response header with the provided
	// status code.
	WriteHeader(statusCode int)

	// Flush writes the buffered response to the underlying writer.
	Flush()
}

// headerAdapter adapts http.Header to our Header type
// This is a zero-allocation wrapper around http.Header
type headerAdapter map[string][]string

// httpResponseWriterAdapter adapts http.ResponseWriter to our ResponseWriter interface
type httpResponseWriterAdapter struct {
	writer     http.ResponseWriter
	header     headerAdapter
	statusCode int
	written    bool
	body       []byte
}

// responseWriterPool is a pool of httpResponseWriterAdapter objects for reuse
var responseWriterPool = sync.Pool{
	New: func() interface{} {
		return &httpResponseWriterAdapter{
			statusCode: http.StatusOK,
			written:    false,
			body:       make([]byte, 0, 512), // Pre-allocate with capacity
		}
	},
}

// NewResponseWriter creates a new ResponseWriter from an http.ResponseWriter
func NewResponseWriter(w http.ResponseWriter) ResponseWriter {
	// Get an adapter from the pool
	adapter := responseWriterPool.Get().(*httpResponseWriterAdapter)

	// Initialize with the current writer and header
	adapter.writer = w
	adapter.header = headerAdapter(w.Header())
	adapter.statusCode = http.StatusOK
	adapter.written = false

	// Reuse the existing body buffer if possible
	if cap(adapter.body) < 512 {
		adapter.body = make([]byte, 0, 512)
	} else {
		adapter.body = adapter.body[:0]
	}

	return adapter
}

// ReleaseResponseWriter returns a ResponseWriter to the pool
func ReleaseResponseWriter(w ResponseWriter) {
	if adapter, ok := w.(*httpResponseWriterAdapter); ok {
		adapter.writer = nil
		adapter.header = nil
		responseWriterPool.Put(adapter)
	}
}

// Header returns the header map that will be sent by WriteHeader
func (a *httpResponseWriterAdapter) Header() Header {
	// Cast our headerAdapter to Header type without allocation
	return Header(a.header)
}

// Write writes the data to the connection as part of an HTTP reply
func (a *httpResponseWriterAdapter) Write(b []byte) (int, error) {
	// Buffer the response body
	a.body = append(a.body, b...)
	return len(b), nil
}

// WriteHeader sends an HTTP response header with the provided status code
func (a *httpResponseWriterAdapter) WriteHeader(statusCode int) {
	// Store the status code for later
	a.statusCode = statusCode
}

// Flush writes the buffered response to the underlying writer
func (a *httpResponseWriterAdapter) Flush() {
	if !a.written {
		// Write the status code
		a.writer.WriteHeader(a.statusCode)

		// Write the body
		if len(a.body) > 0 {
			a.writer.Write(a.body)
		}

		a.written = true
	}
}
