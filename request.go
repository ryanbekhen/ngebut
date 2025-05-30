package ngebut

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"sync"
)

// requestBodyBufferPool is a pool of bytes.Buffer objects for reuse when reading request bodies
// This pool helps reduce memory allocations by reusing buffers instead of creating
// new ones for each request. The buffers are used to read the request body and then
// returned to the pool for future use.
var requestBodyBufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 4096)) // 4KB initial capacity
	},
}

// Request represents an HTTP request.
type Request struct {
	// Method specifies the HTTP method (GET, POST, PUT, etc.).
	Method string

	// URL specifies the URL being requested.
	URL *url.URL

	// Proto is the protocol version.
	Proto string

	// Header contains the request header fields.
	Header Header

	// Body is the request's body.
	Body []byte

	// ContentLength records the length of the associated content.
	ContentLength int64

	// Host specifies the host on which the URL is sought.
	Host string

	// RemoteAddr allows HTTP servers and other software to record
	// the network address that sent the request.
	RemoteAddr string

	// RequestURI is the unmodified request-target as sent by the client
	// to a server.
	RequestURI string

	// ctx is the request's context.
	ctx context.Context
}

// NewRequest creates a new Request from an http.Request.
func NewRequest(r *http.Request) *Request {
	if r == nil {
		return &Request{
			Header: make(Header),
			ctx:    context.Background(),
		}
	}

	// Read the request body if it exists
	var body []byte
	if r.Body != nil {
		// Get a buffer from the pool
		buf := requestBodyBufferPool.Get().(*bytes.Buffer)
		buf.Reset()

		// Read the body into the buffer
		_, err := io.Copy(buf, r.Body)
		if err == nil {
			// Close the original body
			_ = r.Body.Close()

			// Get the bytes from the buffer
			body = buf.Bytes()

			// Create a new ReadCloser so the body can be read again if needed
			// Use the same buffer to avoid allocation
			r.Body = io.NopCloser(bytes.NewReader(body))
		}

		// Return the buffer to the pool
		requestBodyBufferPool.Put(buf)
	}

	// Convert http.Header to our Header type without allocation
	// by casting the map directly
	return &Request{
		Method:        r.Method,
		URL:           r.URL,
		Proto:         r.Proto,
		Header:        Header(r.Header),
		Body:          body,
		ContentLength: r.ContentLength,
		Host:          r.Host,
		RemoteAddr:    r.RemoteAddr,
		RequestURI:    r.RequestURI,
		ctx:           r.Context(),
	}
}

// Context returns the request's context.
func (r *Request) Context() context.Context {
	if r.ctx == nil {
		return context.Background()
	}
	return r.ctx
}

// WithContext returns a shallow copy of r with its context changed to ctx.
func (r *Request) WithContext(ctx context.Context) *Request {
	if ctx == nil {
		panic("nil context")
	}
	r2 := new(Request)
	*r2 = *r
	r2.ctx = ctx
	return r2
}

// UserAgent returns the client's User-Agent header.
func (r *Request) UserAgent() string {
	return r.Header.Get("User-Agent")
}
