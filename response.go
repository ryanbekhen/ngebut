package ngebut

import (
	"bytes"
	"github.com/panjf2000/gnet/v2"
	"io"
	"net/http"
	"strconv"
)

type Response struct {
	Status     string // e.g. "200 OK"
	StatusCode int    // e.g. 200
	Proto      string // e.g. "HTTP/1.0"
	ProtoMajor int    // e.g. 1
	ProtoMinor int    // e.g. 0

	// Header maps header keys to values. If the response had multiple
	// headers with the same key, they may be concatenated, with comma
	// delimiters.  (RFC 7230, section 3.2.2 requires that multiple headers
	// be semantically equivalent to a comma-delimited sequence.) When
	// Header values are duplicated by other fields in this struct (e.g.,
	// ContentLength, TransferEncoding, Trailer), the field values are
	// authoritative.
	//
	// Keys in the map are canonicalized (see CanonicalHeaderKey).
	Header Header

	// Body represents the response body.
	//
	// The response body is streamed on demand as the Body field
	// is read. If the network connection fails or the server
	// terminates the response, Body.Read calls return an error.
	//
	// The http Client and Transport guarantee that Body is always
	// non-nil, even on responses without a body or responses with
	// a zero-length body. It is the caller's responsibility to
	// close Body. The default HTTP client's Transport may not
	// reuse HTTP/1.x "keep-alive" TCP connections if the Body is
	// not read to completion and closed.
	//
	// The Body is automatically dechunked if the server replied
	// with a "chunked" Transfer-Encoding.
	//
	// As of Go 1.12, the Body will also implement io.Writer
	// on a successful "101 Switching Protocols" response,
	// as used by WebSockets and HTTP/2's "h2c" mode.
	Body io.ReadCloser

	// ContentLength records the length of the associated content. The
	// value -1 indicates that the length is unknown. Unless Request.Method
	// is "HEAD", values >= 0 indicate that the given number of bytes may
	// be read from Body.
	ContentLength int64

	// Contains transfer encodings from outer-most to inner-most. Value is
	// nil, means that "identity" encoding is used.
	TransferEncoding []string

	// Close records whether the header directed that the connection be
	// closed after reading Body. The value is advice for clients: neither
	// ReadResponse nor Response.Write ever closes a connection.
	Close bool

	// Uncompressed reports whether the response was sent compressed but
	// was decompressed by the http package. When true, reading from
	// Body yields the uncompressed content instead of the compressed
	// content actually set from the server, ContentLength is set to -1,
	// and the "Content-Length" and "Content-Encoding" fields are deleted
	// from the responseHeader. To get the original response from
	// the server, set Transport.DisableCompression to true.
	Uncompressed bool

	// Trailer maps trailer keys to values in the same
	// format as Header.
	//
	// The Trailer initially contains only nil values, one for
	// each key specified in the server's "Trailer" header
	// value. Those values are not added to Header.
	//
	// Trailer must not be accessed concurrently with Read calls
	// on the Body.
	//
	// After Body.Read has returned io.EOF, Trailer will contain
	// any trailer values sent by the server.
	Trailer Header

	// Request is the request that was sent to obtain this Response.
	// Request's Body is nil (having already been consumed).
	// This is only populated for Client requests.
	Request *Request
}

type responseWriter struct {
	conn     gnet.Conn
	Response *Response
}

func (w *responseWriter) Header() Header {
	return w.Response.Header
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.Response.Header.Set("Content-Length", strconv.Itoa(len(b)))
	totalSize := len(w.Response.Proto) + 1 + len(strconv.Itoa(w.Response.StatusCode)) + 1 + len(w.Response.Status) + 2
	for key, values := range w.Response.Header {
		for _, value := range values {
			totalSize += len(key) + 2 + len(value) + 2
		}
	}
	totalSize += 2
	totalSize += len(b)

	response := make([]byte, 0, totalSize)

	response = append(response, w.Response.Proto...)
	response = append(response, ' ')
	response = append(response, strconv.Itoa(w.Response.StatusCode)...)
	response = append(response, ' ')
	response = append(response, w.Response.Status...)
	response = append(response, '\r', '\n')

	for key, values := range w.Response.Header {
		for _, value := range values {
			response = append(response, key...)
			response = append(response, ':', ' ')
			response = append(response, value...)
			response = append(response, '\r', '\n')
		}
	}

	response = append(response, '\r', '\n')

	response = append(response, b...)
	w.Response.Body = io.NopCloser(bytes.NewReader(b))
	return w.conn.Write(response)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.Response.StatusCode = statusCode
	w.Response.Status = http.StatusText(statusCode)
}
