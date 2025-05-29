// Package httpparser provides HTTP parsing functionality for the ngebut framework.
package httpparser

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
	"unsafe"

	"github.com/evanphx/wildcat"
)

// Constants for HTTP parsing
var (
	crlf      = []byte("\r\n\r\n")
	lastChunk = []byte("0\r\n\r\n")
)

// Object pools for reusing frequently created objects
var (
	// parserPool reuses HTTP parsers
	parserPool = sync.Pool{
		New: func() interface{} {
			return wildcat.NewHTTPParser()
		},
	}

	// readerPool reuses bufio.Reader objects
	readerPool = sync.Pool{
		New: func() interface{} {
			return bufio.NewReaderSize(nil, 4096)
		},
	}

	// bytesReaderPool reuses bytes.Reader objects
	bytesReaderPool = sync.Pool{
		New: func() interface{} {
			return bytes.NewReader(nil)
		},
	}

	// bodyReaderPool reuses io.ReadCloser objects for request bodies
	bodyReaderPool = sync.Pool{
		New: func() interface{} {
			return &bodyReader{
				reader: bytes.NewReader(nil),
			}
		},
	}

	// Common status codes as byte slices to avoid conversions
	statusCodeBytes = make(map[int][]byte)
)

// Initialize common status codes as byte slices
func init() {
	// Pre-compute common status codes as byte slices
	for i := 100; i < 600; i++ {
		statusCodeBytes[i] = []byte(strconv.Itoa(i))
	}
}

// bodyReader is a reusable io.ReadCloser for request bodies
type bodyReader struct {
	reader *bytes.Reader
	data   []byte
}

// Read implements io.Reader
func (b *bodyReader) Read(p []byte) (n int, err error) {
	return b.reader.Read(p)
}

// Close implements io.Closer
func (b *bodyReader) Close() error {
	// Reset the reader position to the beginning
	b.reader.Reset(b.data)
	return nil
}

// Reset resets the bodyReader with new data
func (b *bodyReader) Reset(data []byte) {
	b.data = data
	b.reader.Reset(data)
}

// GetBodyReader returns a reusable io.ReadCloser for request bodies
func GetBodyReader(data []byte) io.ReadCloser {
	br := bodyReaderPool.Get().(*bodyReader)
	br.Reset(data)
	return br
}

// ReleaseBodyReader returns a bodyReader to the pool
func ReleaseBodyReader(rc io.ReadCloser) {
	if br, ok := rc.(*bodyReader); ok {
		bodyReaderPool.Put(br)
	}
}

// GetReader returns a reusable bufio.Reader
func GetReader() *bufio.Reader {
	return readerPool.Get().(*bufio.Reader)
}

// ReleaseReader returns a bufio.Reader to the pool
func ReleaseReader(r *bufio.Reader) {
	r.Reset(nil)
	readerPool.Put(r)
}

// GetBytesReader returns a reusable bytes.Reader
func GetBytesReader() *bytes.Reader {
	return bytesReaderPool.Get().(*bytes.Reader)
}

// ReleaseBytesReader returns a bytes.Reader to the pool
func ReleaseBytesReader(r *bytes.Reader) {
	r.Reset(nil)
	bytesReaderPool.Put(r)
}

// unsafeByteToString converts a byte slice to a string without allocation
// This is safe only if the byte slice is not modified after the conversion
func unsafeByteToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// Header represents HTTP headers.
type Header map[string][]string

// Codec represents an HTTP codec for parsing requests and writing responses.
type Codec struct {
	Parser        *wildcat.HTTPParser
	ContentLength int
	Buf           []byte
	Router        interface{} // Using interface{} to avoid cyclic imports
}

// StatusText returns a text for the HTTP status code.
func StatusText(code int) string {
	return http.StatusText(code)
}

// Common constants for response size estimation
const (
	// Size of fixed parts of the response
	statusLineBaseSize    = 9  // "HTTP/1.1 " + " " + "\r\n"
	dateHeaderBaseSize    = 8  // "Date: " + "\r\n"
	serverHeaderSize      = 16 // "Server: ngebut\r\n"
	contentLengthBaseSize = 18 // "Content-Length: " + "\r\n\r\n"
	crlfSize              = 2  // "\r\n"

	// Standard date format length
	dateFormatSize = 29 // "Mon, 02 Jan 2006 15:04:05 GMT"

	// Average size of a status code (3 digits)
	statusCodeAvgSize = 3

	// Average size of a content length (assuming most responses < 10KB)
	contentLengthAvgSize = 4
)

// EstimateResponseSize calculates the estimated size of an HTTP response.
func EstimateResponseSize(statusCode int, header Header, body []byte) int {
	// Calculate size for status line, standard headers, and body
	size := statusLineBaseSize + statusCodeAvgSize + len(StatusText(statusCode)) // Status line
	size += dateHeaderBaseSize + dateFormatSize                                  // Date header
	size += serverHeaderSize                                                     // Server header

	// Content-Length header - use a more efficient approach
	contentLengthSize := contentLengthBaseSize
	if len(body) > 0 {
		// Estimate the size of the content length value
		if len(body) < 10 {
			contentLengthSize += 1 // 1 digit
		} else if len(body) < 100 {
			contentLengthSize += 2 // 2 digits
		} else if len(body) < 1000 {
			contentLengthSize += 3 // 3 digits
		} else if len(body) < 10000 {
			contentLengthSize += 4 // 4 digits
		} else {
			contentLengthSize += 6 // Assume 6 digits for larger bodies
		}
	} else {
		contentLengthSize += 1 // Just "0"
	}
	size += contentLengthSize

	// Add size for custom headers
	for k, values := range header {
		for _, v := range values {
			size += len(k) + 2 + len(v) + crlfSize // "key: value\r\n"
		}
	}

	size += len(body) // Body

	return size
}

// Parse parses HTTP request data.
func (hc *Codec) Parse(data []byte) (int, []byte, error) {
	bodyOffset, err := hc.Parser.Parse(data)
	if err != nil {
		return 0, nil, err
	}

	contentLength := hc.GetContentLength()
	if contentLength > -1 {
		bodyEnd := bodyOffset + contentLength
		var body []byte
		if len(data) >= bodyEnd {
			body = data[bodyOffset:bodyEnd]
		}
		return bodyEnd, body, nil
	}

	// Transfer-Encoding: chunked
	if idx := bytes.Index(data[bodyOffset:], lastChunk); idx != -1 {
		bodyEnd := bodyOffset + idx + 5
		var body []byte
		if len(data) >= bodyEnd {
			// Get a reader from the pool
			reader := GetReader()
			defer ReleaseReader(reader)

			// Get a bytes reader from the pool
			bytesReader := bytesReaderPool.Get().(*bytes.Reader)
			bytesReader.Reset(data[:bodyEnd])
			defer bytesReaderPool.Put(bytesReader)

			// Reset the reader with the bytes reader
			reader.Reset(bytesReader)

			req, err := http.ReadRequest(reader)
			if err != nil {
				return bodyEnd, nil, err
			}

			// Read the body without allocations if possible
			if req.ContentLength > 0 {
				body = make([]byte, req.ContentLength)
				_, _ = req.Body.Read(body)
			} else if req.Body != nil {
				// For chunked encoding, we still need to read the body
				// This is a rare case, so the allocation is acceptable
				body, _ = io.ReadAll(req.Body)
			}

			if req.Body != nil {
				req.Body.Close()
			}
		}
		return bodyEnd, body, nil
	}

	// Requests without a body.
	if idx := bytes.Index(data, crlf); idx != -1 {
		return idx + 4, nil, nil
	}

	return 0, nil, errors.New("invalid http request")
}

// GetContentLength gets the content length from the HTTP headers.
func (hc *Codec) GetContentLength() int {
	if hc.ContentLength != -1 {
		return hc.ContentLength
	}

	val := hc.Parser.FindHeader([]byte("Content-Length"))
	if val != nil {
		// Use unsafeByteToString to avoid allocation
		i, err := strconv.ParseInt(unsafeByteToString(val), 10, 0)
		if err == nil {
			hc.ContentLength = int(i)
		}
	}

	return hc.ContentLength
}

// ResetParser resets the HTTP parser.
func (hc *Codec) ResetParser() {
	// Reset content length
	hc.ContentLength = -1
}

// Reset resets the HTTP codec.
func (hc *Codec) Reset() {
	// Reset the parser
	hc.ResetParser()

	// Clear the buffer
	hc.Buf = hc.Buf[:0]

	// Ensure we have a valid parser
	if hc.Parser == nil {
		hc.Parser = parserPool.Get().(*wildcat.HTTPParser)
	}
}

// ResponseBufferPool is a pool of byte slices for response buffers.
var ResponseBufferPool = sync.Pool{
	New: func() interface{} {
		// Start with a reasonable size buffer
		return make([]byte, 0, 4096)
	},
}

// contentLengthPrefix is the prefix for the Content-Length header
var contentLengthPrefix = []byte("Content-Length: ")

// WriteResponse writes an HTTP response to the codec's buffer.
func (hc *Codec) WriteResponse(statusCode int, header Header, body []byte) {
	estimatedSize := EstimateResponseSize(statusCode, header, body)

	// Get buffer from pool if needed
	var buf []byte
	if cap(hc.Buf) < estimatedSize {
		// Return current buffer to pool if it exists
		if cap(hc.Buf) > 0 {
			tmp := hc.Buf[:0]
			ResponseBufferPool.Put(tmp)
		}

		// Get a buffer from the pool
		poolBuf := ResponseBufferPool.Get().([]byte)
		if cap(poolBuf) < estimatedSize {
			// If the pool buffer is too small, create a new one and discard the pool buffer
			ResponseBufferPool.Put(poolBuf[:0])
			buf = make([]byte, 0, estimatedSize)
		} else {
			buf = poolBuf[:0]
		}
	} else {
		// Reuse existing buffer
		buf = hc.Buf[:0]
	}

	// Write HTTP response - use direct string constants where possible
	buf = append(buf, "HTTP/1.1 "...)

	// Use pre-computed status code bytes if available
	if codeBytes, ok := statusCodeBytes[statusCode]; ok {
		buf = append(buf, codeBytes...)
	} else {
		buf = strconv.AppendInt(buf, int64(statusCode), 10)
	}

	buf = append(buf, ' ')
	buf = append(buf, StatusText(statusCode)...)
	buf = append(buf, "\r\n"...)

	// Add Date header
	buf = append(buf, "Date: "...)
	buf = time.Now().AppendFormat(buf, "Mon, 02 Jan 2006 15:04:05 GMT")
	buf = append(buf, "\r\n"...)

	// Add custom headers
	for k, values := range header {
		for _, v := range values {
			buf = append(buf, k...)
			buf = append(buf, ": "...)
			buf = append(buf, v...)
			buf = append(buf, "\r\n"...)
		}
	}

	// Add Content-Length header
	buf = append(buf, contentLengthPrefix...)
	buf = strconv.AppendInt(buf, int64(len(body)), 10)
	buf = append(buf, "\r\n\r\n"...)

	// Add body
	buf = append(buf, body...)

	// Store the buffer for writing
	hc.Buf = buf
}

// codecPool is a pool of Codec objects for reuse
var codecPool = sync.Pool{
	New: func() interface{} {
		return &Codec{
			Parser:        parserPool.Get().(*wildcat.HTTPParser),
			ContentLength: -1,
		}
	},
}

// NewCodec creates a new HTTP codec.
func NewCodec(router interface{}) *Codec {
	// Get a codec from the pool
	codec := codecPool.Get().(*Codec)

	// Reset content length
	codec.ContentLength = -1

	// Ensure we have a valid parser
	if codec.Parser == nil {
		codec.Parser = parserPool.Get().(*wildcat.HTTPParser)
	}

	// Set the router
	codec.Router = router

	// Clear the buffer
	codec.Buf = codec.Buf[:0]

	return codec
}

// ReleaseCodec returns a codec to the pool.
func ReleaseCodec(codec *Codec) {
	// Reset the codec to ensure it's in a clean state
	// This resets the content length and clears the buffer
	codec.Reset()

	// Clear the router reference to avoid memory leaks
	codec.Router = nil

	// Return to the pool
	codecPool.Put(codec)
}
