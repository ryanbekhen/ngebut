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
	Header() *Header

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

// syncingHeader is a Header implementation that syncs changes to the underlying writer
type syncingHeader struct {
	header headerAdapter
	writer http.ResponseWriter
}

// Add adds a value to the header and syncs it to the underlying writer
func (h *syncingHeader) Add(key, value string) {
	// Add to our header
	(*Header)(&h.header).Add(key, value)

	// Sync to the underlying writer
	if h.writer != nil {
		h.writer.Header().Add(key, value)
	}
}

// Set sets a value in the header and syncs it to the underlying writer
func (h *syncingHeader) Set(key, value string) {
	// Set in our header
	(*Header)(&h.header).Set(key, value)

	// Sync to the underlying writer
	if h.writer != nil {
		h.writer.Header().Set(key, value)
	}
}

// Get gets a value from the header
func (h *syncingHeader) Get(key string) string {
	return (*Header)(&h.header).Get(key)
}

// Values gets all values for a key from the header
func (h *syncingHeader) Values(key string) []string {
	return (*Header)(&h.header).Values(key)
}

// Del deletes a key from the header and syncs it to the underlying writer
func (h *syncingHeader) Del(key string) {
	// Delete from our header
	(*Header)(&h.header).Del(key)

	// Sync to the underlying writer
	if h.writer != nil {
		h.writer.Header().Del(key)
	}
}

// Clone clones the header
func (h *syncingHeader) Clone() Header {
	return (*Header)(&h.header).Clone()
}

// Write writes the header to a writer
func (h *syncingHeader) Write(w stringWriter) error {
	return (*Header)(&h.header).Write(w)
}

// WriteSubset writes a subset of the header to a writer
func (h *syncingHeader) WriteSubset(w stringWriter, exclude map[string]bool) error {
	return (*Header)(&h.header).WriteSubset(w, exclude)
}

// copyHeadersToWriter copies all headers from the headerAdapter to the underlying writer's header
func (a *httpResponseWriterAdapter) copyHeadersToWriter() {
	// No need to copy headers since we're using the writer's header directly
	// This method is kept for backward compatibility
	return
}

// httpResponseWriterAdapter adapts http.ResponseWriter to our ResponseWriter interface
type httpResponseWriterAdapter struct {
	writer     http.ResponseWriter
	header     headerAdapter
	statusCode int
	written    bool
	// Cache for the header to avoid creating a new one on each Header() call
	headerCache *Header
}

// responseWriterPool is a pool of httpResponseWriterAdapter objects for reuse
var responseWriterPool = sync.Pool{
	New: func() interface{} {
		return &httpResponseWriterAdapter{
			statusCode: StatusOK,
			written:    false,
		}
	},
}

// headerAdapterPool is a pool of headerAdapter objects for reuse
var headerAdapterPool = sync.Pool{
	New: func() interface{} {
		return make(headerAdapter)
	},
}

// NewResponseWriter creates a new ResponseWriter from an http.ResponseWriter
func NewResponseWriter(w http.ResponseWriter) ResponseWriter {
	// Get an adapter from the pool
	adapter := responseWriterPool.Get().(*httpResponseWriterAdapter)

	// Initialize with the current writer
	adapter.writer = w

	// We don't need to use our own header map anymore
	// Just set it to nil to indicate we're using the writer's header directly
	adapter.header = nil

	adapter.statusCode = StatusOK
	adapter.written = false
	adapter.headerCache = nil

	return adapter
}

// ReleaseResponseWriter returns a ResponseWriter to the pool
func ReleaseResponseWriter(w ResponseWriter) {
	if adapter, ok := w.(*httpResponseWriterAdapter); ok {
		// We don't need to do anything with the header since we're using the writer's header directly

		adapter.writer = nil
		adapter.header = nil // Ensure header is nil
		adapter.statusCode = StatusOK
		adapter.written = false
		adapter.headerCache = nil // Reset the header cache

		// Return the adapter to the pool
		responseWriterPool.Put(adapter)
	}
}

// Header returns the header map that will be sent by WriteHeader
func (a *httpResponseWriterAdapter) Header() *Header {
	// Use cached header if available
	if a.headerCache != nil {
		return a.headerCache
	}

	// Use the underlying writer's header directly
	if a.writer != nil {
		// Create a headerAdapter that wraps the writer's header
		// Store it directly in a.header to avoid allocation
		a.header = headerAdapter(a.writer.Header())
		a.headerCache = (*Header)(&a.header)
		return a.headerCache
	}

	// Create an empty header if no writer is available
	h := NewHeader()
	a.headerCache = h
	return h
}

// Write writes the data to the connection as part of an HTTP reply
func (a *httpResponseWriterAdapter) Write(b []byte) (int, error) {
	// Fast path: if there's no data to write, return immediately
	if len(b) == 0 {
		// Still need to write the header if not written yet
		if !a.written {
			// Write the header directly
			a.writer.WriteHeader(a.statusCode)
			a.written = true
		}
		return 0, nil
	}

	// Write directly to the underlying writer if we've already written the header
	if a.written {
		return a.writer.Write(b)
	}

	// Otherwise, write the header first, then write the data
	a.writer.WriteHeader(a.statusCode)
	a.written = true

	// Write the data directly to avoid extra allocations
	return a.writer.Write(b)
}

// WriteHeader sends an HTTP response header with the provided status code
func (a *httpResponseWriterAdapter) WriteHeader(statusCode int) {
	// Store the status code for later
	a.statusCode = statusCode
}

// Flush writes the buffered response to the underlying writer
func (a *httpResponseWriterAdapter) Flush() {
	// If we haven't written anything yet, write the status code
	if !a.written {
		a.writer.WriteHeader(a.statusCode)
		a.written = true
	}

	// If the underlying writer implements http.Flusher, call Flush()
	if flusher, ok := a.writer.(http.Flusher); ok {
		flusher.Flush()
	}
}
