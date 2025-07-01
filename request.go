package ngebut

import (
	"bytes"
	"context"
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
		return bytes.NewBuffer(make([]byte, 0, 8192)) // 8KB initial capacity to reduce reallocations
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
	Header *Header

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
// It uses the requestPool to reuse Request objects when possible.
func NewRequest(r *http.Request) *Request {
	// Use getRequest to get a Request from the pool and initialize it
	return getRequest(r)
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
	return r.Header.Get(HeaderUserAgent)
}

func (r *Request) SetContext(ctx context.Context) {
	if ctx == nil {
		panic("nil context")
	}
	r.ctx = ctx
}
