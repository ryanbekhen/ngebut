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

// Error variables for HTTP parsing
var (
	// ErrIncompleteBody is returned when the request body is incomplete
	ErrIncompleteBody = errors.New("incomplete body")
	// ErrInvalidChunk is returned when a chunk in a chunked request is invalid
	ErrInvalidChunk = errors.New("invalid chunk")
)

// Parse parses HTTP request data.
func (hc *Codec) Parse(data []byte) (int, []byte, error) {
	bodyOffset, err := hc.Parser.Parse(data)
	if err != nil {
		return 0, nil, err
	}

	// Fast path: Check for requests without a body first (GET, HEAD, etc.)
	// This is the most common case for HTTP requests
	if bodyOffset < len(data) && data[bodyOffset] == '\r' &&
		bodyOffset+3 < len(data) && data[bodyOffset+1] == '\n' &&
		data[bodyOffset+2] == '\r' && data[bodyOffset+3] == '\n' {
		return bodyOffset + 4, nil, nil
	}

	// Check for Content-Length header (common case for requests with bodies)
	contentLength := hc.GetContentLength()
	if contentLength > -1 {
		bodyEnd := bodyOffset + contentLength
		if len(data) >= bodyEnd {
			// Zero-copy slice of the body
			return bodyEnd, data[bodyOffset:bodyEnd], nil
		}
		return 0, nil, ErrIncompleteBody
	}

	// Transfer-Encoding: chunked (less common case)
	if idx := bytes.Index(data[bodyOffset:], lastChunk); idx != -1 {
		bodyEnd := bodyOffset + idx + 5
		if len(data) < bodyEnd {
			return 0, nil, ErrIncompleteBody
		}

		// Try the optimized chunked body parser first
		chunkedBody, err := parseChunkedBody(data[bodyOffset : bodyEnd-5])
		if err == nil {
			return bodyEnd, chunkedBody, nil
		}

		// Fallback to standard library for complex cases
		body, err := parseChunkedBodyFallback(data[:bodyEnd])
		return bodyEnd, body, err
	}

	// Fallback check for requests without a body using bytes.Index
	// This is slower but more thorough
	if idx := bytes.Index(data, crlf); idx != -1 {
		return idx + 4, nil, nil
	}

	return 0, nil, errors.New("invalid http request")
}

// parseChunkedBody parses a chunked HTTP body more efficiently than the standard library
func parseChunkedBody(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, nil
	}

	// Ultra-fast path for the most common case: a single small chunk
	// This avoids all allocations for simple cases
	if len(data) < 32 {
		// For very small data, it's likely just a single small chunk
		// Check for single-digit chunk size (0-9) followed by \r\n
		if len(data) >= 3 && data[0] >= '0' && data[0] <= '9' &&
			data[1] == '\r' && data[2] == '\n' {
			size := int(data[0] - '0')
			if 3+size+2 <= len(data) && // 3 for chunk header, size for data, 2 for trailing CRLF
				data[3+size] == '\r' && data[3+size+1] == '\n' {
				// For single chunks, we can return a slice of the original data
				// This avoids allocation completely
				return data[3 : 3+size], nil
			}
		}

		// Check for two-digit chunk size (10-99) followed by \r\n
		if len(data) >= 4 && data[0] >= '1' && data[0] <= '9' &&
			data[1] >= '0' && data[1] <= '9' &&
			data[2] == '\r' && data[3] == '\n' {
			size := (int(data[0]-'0') * 10) + int(data[1]-'0')
			if 4+size+2 <= len(data) && // 4 for chunk header, size for data, 2 for trailing CRLF
				data[4+size] == '\r' && data[4+size+1] == '\n' {
				return data[4 : 4+size], nil
			}
		}
	}

	// Estimate the result size to avoid reallocations
	// Most chunked bodies are smaller than the original data
	resultCap := len(data) / 2
	if resultCap < 1024 {
		resultCap = 1024 // Minimum capacity of 1KB
	}
	result := make([]byte, 0, resultCap)

	var i int
	for i < len(data) {
		// Fast path for common chunk sizes (0-9) with direct character checks
		// This avoids the expensive bytes.IndexByte call for simple cases
		if i+2 < len(data) && data[i] >= '0' && data[i] <= '9' &&
			data[i+1] == '\r' && data[i+2] == '\n' {
			size := int(data[i] - '0')
			i += 3 // Move past the chunk size and CRLF

			// If chunk size is 0, we're done
			if size == 0 {
				break
			}

			// Make sure we have enough data
			if i+size+2 > len(data) {
				return nil, ErrInvalidChunk
			}

			// Append the chunk data to the result
			result = append(result, data[i:i+size]...)

			// Move past the chunk data and the trailing CRLF
			i += size
			if data[i] == '\r' && data[i+1] == '\n' {
				i += 2
				continue
			} else {
				return nil, ErrInvalidChunk
			}
		}

		// Fast path for common chunk sizes (10-99) with direct character checks
		if i+3 < len(data) && data[i] >= '1' && data[i] <= '9' &&
			data[i+1] >= '0' && data[i+1] <= '9' &&
			data[i+2] == '\r' && data[i+3] == '\n' {
			size := (int(data[i]-'0') * 10) + int(data[i+1]-'0')
			i += 4 // Move past the chunk size and CRLF

			// Make sure we have enough data
			if i+size+2 > len(data) {
				return nil, ErrInvalidChunk
			}

			// Append the chunk data to the result
			result = append(result, data[i:i+size]...)

			// Move past the chunk data and the trailing CRLF
			i += size
			if data[i] == '\r' && data[i+1] == '\n' {
				i += 2
				continue
			} else {
				return nil, ErrInvalidChunk
			}
		}

		// Standard path for other chunk sizes
		// Find the end of the chunk size line
		lineEnd := bytes.IndexByte(data[i:], '\n')
		if lineEnd == -1 {
			return nil, ErrInvalidChunk
		}
		lineEnd += i // Adjust to absolute position

		// Parse the chunk size (in hex)
		line := data[i:lineEnd]
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}

		// Find the end of the chunk size (before any chunk extension)
		sizeEnd := bytes.IndexByte(line, ';')
		if sizeEnd == -1 {
			sizeEnd = len(line)
		}

		// Parse the chunk size
		size, err := strconv.ParseInt(unsafeByteToString(line[:sizeEnd]), 16, 32)
		if err != nil || size < 0 {
			return nil, ErrInvalidChunk
		}

		// Move past the chunk size line
		i = lineEnd + 1

		// If chunk size is 0, we're done
		if size == 0 {
			break
		}

		// Make sure we have enough data
		if i+int(size) > len(data) {
			return nil, ErrInvalidChunk
		}

		// Append the chunk data to the result
		// Pre-grow the result slice if needed to avoid multiple small allocations
		if cap(result)-len(result) < int(size) {
			newResult := make([]byte, len(result), len(result)+int(size)+1024)
			copy(newResult, result)
			result = newResult
		}
		result = append(result, data[i:i+int(size)]...)

		// Move past the chunk data and the trailing CRLF
		i += int(size)
		if i+2 <= len(data) && data[i] == '\r' && data[i+1] == '\n' {
			i += 2
		} else {
			return nil, ErrInvalidChunk
		}
	}

	return result, nil
}

// Helper function to parse chunked body using standard library as a fallback
func parseChunkedBodyFallback(data []byte) ([]byte, error) {
	// Get a reader from the pool
	reader := GetReader()
	defer ReleaseReader(reader)

	// Get a bytes reader from the pool
	bytesReader := bytesReaderPool.Get().(*bytes.Reader)
	bytesReader.Reset(data)
	defer bytesReaderPool.Put(bytesReader)

	// Reset the reader with the bytes reader
	reader.Reset(bytesReader)

	req, err := http.ReadRequest(reader)
	if err != nil {
		return nil, err
	}
	defer func() {
		if req.Body != nil {
			req.Body.Close()
		}
	}()

	// Read the body without allocations if possible
	if req.ContentLength > 0 {
		body := make([]byte, req.ContentLength)
		_, _ = req.Body.Read(body)
		return body, nil
	} else if req.Body != nil {
		// For chunked encoding, we still need to read the body
		return io.ReadAll(req.Body)
	}

	return nil, nil
}

// GetContentLength gets the content length from the HTTP headers.
func (hc *Codec) GetContentLength() int {
	// Fast path: return cached value if available
	if hc.ContentLength != -1 {
		return hc.ContentLength
	}

	// Use pre-allocated byte slice for header name to avoid allocations
	val := hc.Parser.FindHeader(contentLengthBytes)
	if val == nil {
		hc.ContentLength = -1 // Cache the result
		return -1             // No Content-Length header
	}

	// Fast path for common content lengths (0-9)
	if len(val) == 1 && val[0] >= '0' && val[0] <= '9' {
		hc.ContentLength = int(val[0] - '0')
		return hc.ContentLength
	}

	// Fast path for common content lengths (10-99)
	if len(val) == 2 && val[0] >= '1' && val[0] <= '9' && val[1] >= '0' && val[1] <= '9' {
		hc.ContentLength = (int(val[0]-'0') * 10) + int(val[1]-'0')
		return hc.ContentLength
	}

	// Fast path for common content lengths (100-999)
	if len(val) == 3 && val[0] >= '1' && val[0] <= '9' &&
		val[1] >= '0' && val[1] <= '9' && val[2] >= '0' && val[2] <= '9' {
		hc.ContentLength = (int(val[0]-'0') * 100) + (int(val[1]-'0') * 10) + int(val[2]-'0')
		return hc.ContentLength
	}

	// Use unsafeByteToString to avoid allocation for larger numbers
	i, err := strconv.ParseInt(unsafeByteToString(val), 10, 31)
	if err == nil {
		hc.ContentLength = int(i)
		return hc.ContentLength
	}

	// If parsing failed, set to -1 to indicate no valid Content-Length
	hc.ContentLength = -1
	return -1
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

// Common header constants to avoid allocations
var (
	// contentLengthPrefix is the prefix for the Content-Length header
	contentLengthPrefix = []byte("Content-Length: ")

	// contentLengthBytes is the byte representation of "Content-Length" header name
	contentLengthBytes = []byte("Content-Length")

	// httpVersion is the HTTP version string
	httpVersion = []byte("HTTP/1.1 ")

	// dateHeaderPrefix is the prefix for the Date header
	dateHeaderPrefix = []byte("Date: ")

	// crlfBytes is the CRLF byte sequence
	crlfBytes = []byte("\r\n")

	// doubleCrlfBytes is the double CRLF byte sequence that ends headers
	doubleCrlfBytes = []byte("\r\n\r\n")

	// colonSpace is the colon+space sequence used in headers
	colonSpace = []byte(": ")
)

// Date header caching to avoid expensive time formatting on every response
var (
	// cachedDateHeader stores the pre-formatted Date header
	cachedDateHeader []byte

	// lastDateUpdate tracks when the cached date was last updated
	lastDateUpdate int64

	// dateMutex protects access to the cached date header
	dateMutex sync.RWMutex
)

// getDateHeader returns a formatted date header, using a cached version if possible
// RFC7232 recommends date precision to the second, so we update at most once per second
func getDateHeader() []byte {
	// Fast path: use cached date if it's recent enough (less than 1 second old)
	dateMutex.RLock()
	now := time.Now().Unix()
	if now == lastDateUpdate && cachedDateHeader != nil {
		header := cachedDateHeader
		dateMutex.RUnlock()
		return header
	}
	dateMutex.RUnlock()

	// Slow path: update the cached date
	dateMutex.Lock()
	defer dateMutex.Unlock()

	// Check again in case another goroutine updated while we were waiting
	if now == lastDateUpdate && cachedDateHeader != nil {
		return cachedDateHeader
	}

	// Format the current time according to HTTP spec
	t := time.Now()
	if cachedDateHeader == nil {
		cachedDateHeader = make([]byte, 0, len(dateHeaderPrefix)+30+len(crlfBytes))
	} else {
		cachedDateHeader = cachedDateHeader[:0]
	}

	cachedDateHeader = append(cachedDateHeader, dateHeaderPrefix...)
	cachedDateHeader = t.AppendFormat(cachedDateHeader, "Mon, 02 Jan 2006 15:04:05 GMT")
	cachedDateHeader = append(cachedDateHeader, crlfBytes...)

	lastDateUpdate = now
	return cachedDateHeader
}

// WriteResponse writes an HTTP response to the codec's buffer.
func (hc *Codec) WriteResponse(statusCode int, header Header, body []byte) {
	// Get a more accurate estimate of the response size
	estimatedSize := EstimateResponseSize(statusCode, header, body)

	// Get buffer from pool if needed - optimized buffer management
	var buf []byte
	if cap(hc.Buf) < estimatedSize {
		// Return current buffer to pool if it exists and is worth pooling
		if cap(hc.Buf) > 0 && cap(hc.Buf) <= 32*1024 { // Don't pool buffers larger than 32KB
			ResponseBufferPool.Put(hc.Buf[:0])
		}

		// Get a buffer from the pool
		poolBuf := ResponseBufferPool.Get().([]byte)
		if cap(poolBuf) < estimatedSize {
			// If the pool buffer is too small, create a new one with exact capacity
			ResponseBufferPool.Put(poolBuf[:0])
			buf = make([]byte, 0, estimatedSize)
		} else {
			buf = poolBuf[:0]
		}
	} else {
		// Reuse existing buffer
		buf = hc.Buf[:0]
	}

	// Write HTTP response - use pre-computed byte slices for common parts
	buf = append(buf, httpVersion...)

	// Use pre-computed status code bytes if available
	if codeBytes, ok := statusCodeBytes[statusCode]; ok {
		buf = append(buf, codeBytes...)
	} else {
		buf = strconv.AppendInt(buf, int64(statusCode), 10)
	}

	buf = append(buf, ' ')
	buf = append(buf, StatusText(statusCode)...)
	buf = append(buf, crlfBytes...)

	// Add Date header - use cached version to avoid expensive time formatting
	buf = append(buf, getDateHeader()...)

	// Fast path for empty headers (common case)
	if len(header) == 0 {
		// Add Content-Length header
		buf = append(buf, contentLengthPrefix...)
		buf = strconv.AppendInt(buf, int64(len(body)), 10)
		buf = append(buf, doubleCrlfBytes...)

		// Add body
		if len(body) > 0 {
			buf = append(buf, body...)
		}

		// Store the buffer for writing
		hc.Buf = buf
		return
	}

	// Add custom headers - optimize for common cases
	if len(header) <= 4 { // Small number of headers is common
		// Direct iteration for small maps is faster than range
		for k, values := range header {
			if len(values) == 1 { // Most common case: single value per header
				buf = append(buf, k...)
				buf = append(buf, colonSpace...)
				buf = append(buf, values[0]...)
				buf = append(buf, crlfBytes...)
			} else {
				for _, v := range values {
					buf = append(buf, k...)
					buf = append(buf, colonSpace...)
					buf = append(buf, v...)
					buf = append(buf, crlfBytes...)
				}
			}
		}
	} else {
		// Standard path for many headers
		for k, values := range header {
			for _, v := range values {
				buf = append(buf, k...)
				buf = append(buf, colonSpace...)
				buf = append(buf, v...)
				buf = append(buf, crlfBytes...)
			}
		}
	}

	// Add Content-Length header
	buf = append(buf, contentLengthPrefix...)
	buf = strconv.AppendInt(buf, int64(len(body)), 10)
	buf = append(buf, doubleCrlfBytes...)

	// Add body
	if len(body) > 0 {
		buf = append(buf, body...)
	}

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
