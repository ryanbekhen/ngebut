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

	"github.com/evanphx/wildcat"
	"github.com/ryanbekhen/ngebut/internal/pool"
	internalunsafe "github.com/ryanbekhen/ngebut/internal/unsafe"
	"github.com/valyala/bytebufferpool"
)

// Constants for HTTP parsing
var (
	doubleCRLF = []byte("\r\n\r\n")
	lastChunk  = []byte("0\r\n\r\n")
)

// Object pools for reusing frequently created objects
var (
	// parserPool reuses HTTP parsers
	parserPool = pool.New(func() *wildcat.HTTPParser {
		return wildcat.NewHTTPParser()
	})

	// readerPool reuses bufio.Reader objects
	readerPool = pool.New(func() *bufio.Reader {
		return bufio.NewReaderSize(nil, 16384) // Increased to 16KB for better performance
	})

	// bytesReaderPool reuses bytes.Reader objects
	bytesReaderPool = pool.New(func() *bytes.Reader {
		return bytes.NewReader(nil)
	})

	// bodyReaderPool reuses io.ReadCloser objects for request bodies
	bodyReaderPool = pool.New(func() *bodyReader {
		return &bodyReader{
			reader: bytes.NewReader(nil),
		}
	})

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
	br := bodyReaderPool.Get()
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
	return readerPool.Get()
}

// ReleaseReader returns a bufio.Reader to the pool
func ReleaseReader(r *bufio.Reader) {
	r.Reset(nil)
	readerPool.Put(r)
}

// GetBytesReader returns a reusable bytes.Reader
func GetBytesReader() *bytes.Reader {
	return bytesReaderPool.Get()
}

// ReleaseBytesReader returns a bytes.Reader to the pool
func ReleaseBytesReader(r *bytes.Reader) {
	r.Reset(nil)
	bytesReaderPool.Put(r)
}

// unsafeByteToString converts a byte slice to a string without allocation
// This is safe only if the byte slice is not modified after the conversion
func unsafeByteToString(b []byte) string {
	return internalunsafe.B2S(b)
}

// Header represents HTTP headers.
type Header map[string][]string

// Codec represents an HTTP codec for parsing requests and writing responses.
type Codec struct {
	Parser        *wildcat.HTTPParser
	ContentLength int
	Buf           *bytebufferpool.ByteBuffer
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
	// Use a more efficient approach to find the last chunk marker
	dataLen := len(data)
	bodyData := data[bodyOffset:]
	bodyLen := len(bodyData)

	// Fast path for small bodies - direct search for "0\r\n\r\n"
	if bodyLen < 256 {
		for i := 0; i <= bodyLen-5; i++ {
			if bodyData[i] == '0' &&
				bodyData[i+1] == '\r' &&
				bodyData[i+2] == '\n' &&
				bodyData[i+3] == '\r' &&
				bodyData[i+4] == '\n' {
				bodyEnd := bodyOffset + i + 5
				// Try the optimized chunked body parser first
				chunkedBody, err := parseChunkedBody(data[bodyOffset : bodyEnd-5])
				if err == nil {
					return bodyEnd, chunkedBody, nil
				}

				// Fallback to standard library for complex cases
				body, err := parseChunkedBodyFallback(data[:bodyEnd])
				return bodyEnd, body, err
			}
		}
	} else {
		// For larger bodies, use bytes.Index which is optimized for larger searches
		if idx := bytes.Index(bodyData, lastChunk); idx != -1 {
			bodyEnd := bodyOffset + idx + 5
			if dataLen < bodyEnd {
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
	}

	// Fallback check for requests without a body
	// First try a direct search for double CRLF which is faster for small data
	if dataLen < 256 {
		for i := 0; i <= dataLen-4; i++ {
			if data[i] == '\r' && data[i+1] == '\n' &&
				data[i+2] == '\r' && data[i+3] == '\n' {
				return i + 4, nil, nil
			}
		}
	} else {
		// For larger data, use bytes.Index
		if idx := bytes.Index(data, doubleCRLF); idx != -1 {
			return idx + 4, nil, nil
		}
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

	// First pass: calculate total size of all chunks to avoid reallocations
	totalSize := 0
	var chunkSizes []int   // Store chunk sizes to avoid recalculating
	var chunkOffsets []int // Store chunk offsets to avoid recalculating

	// Pre-allocate for common case (usually less than 8 chunks)
	chunkSizes = make([]int, 0, 8)
	chunkOffsets = make([]int, 0, 8)

	var i int
	for i < len(data) {
		// Fast path for common chunk sizes (0-9) with direct character checks
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

			// Store chunk info
			chunkSizes = append(chunkSizes, size)
			chunkOffsets = append(chunkOffsets, i)
			totalSize += size

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

			// Store chunk info
			chunkSizes = append(chunkSizes, size)
			chunkOffsets = append(chunkOffsets, i)
			totalSize += size

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

		// Store chunk info
		chunkSizes = append(chunkSizes, int(size))
		chunkOffsets = append(chunkOffsets, i)
		totalSize += int(size)

		// Move past the chunk data and the trailing CRLF
		i += int(size)
		if i+2 <= len(data) && data[i] == '\r' && data[i+1] == '\n' {
			i += 2
		} else {
			return nil, ErrInvalidChunk
		}
	}

	// If we have only one chunk, return a slice of the original data to avoid allocation
	if len(chunkSizes) == 1 {
		offset := chunkOffsets[0]
		size := chunkSizes[0]
		return data[offset : offset+size], nil
	}

	// Allocate result buffer with exact size needed
	result := make([]byte, 0, totalSize)

	// Second pass: copy chunks to result buffer
	for i := 0; i < len(chunkSizes); i++ {
		offset := chunkOffsets[i]
		size := chunkSizes[i]
		result = append(result, data[offset:offset+size]...)
	}

	return result, nil
}

// Helper function to parse chunked body using standard library as a fallback
func parseChunkedBodyFallback(data []byte) ([]byte, error) {
	// Get a reader from the pool
	reader := GetReader()
	defer ReleaseReader(reader)

	// Get a bytes reader from the pool
	bytesReader := bytesReaderPool.Get()
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
	val := hc.Parser.FindHeader([]byte("Content-Length"))
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

	// Return the current parser to the pool and get a new one
	if hc.Parser != nil {
		parserPool.Put(hc.Parser)
	}
	hc.Parser = parserPool.Get()
}

// Reset resets the HTTP codec.
func (hc *Codec) Reset() {
	// Reset the parser
	hc.ResetParser()

	// Clear the buffer
	if hc.Buf != nil {
		hc.Buf.Reset()
		ResponseBufferPool.Put(hc.Buf)
		hc.Buf = nil
	}

	// Ensure we have a valid parser
	if hc.Parser == nil {
		hc.Parser = parserPool.Get()
	}
}

// ResponseBufferPool is a pool of byte buffers for response buffers.
// Using valyala's bytebufferpool for better performance
var ResponseBufferPool bytebufferpool.Pool

// Common header constants to avoid allocations
var (
	// contentLengthPrefix is the prefix for the Content-Length header
	contentLengthPrefix = []byte("Content-Length: ")

	// httpVersion is the HTTP version string
	httpVersion = []byte("HTTP/1.1 ")

	// dateHeaderPrefix is the prefix for the Date header
	dateHeaderPrefix = []byte("Date: ")

	// crlfBytes is the CRLF byte sequence
	crlfBytes = []byte("\r\n")

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
	// If we don't have a buffer or it's too small, get a new one
	if hc.Buf == nil {
		// Get a buffer from the pool
		hc.Buf = ResponseBufferPool.Get()
	} else {
		// Reset the existing buffer
		hc.Buf.Reset()
	}

	// ByteBuffer will automatically grow as needed

	// Write HTTP response - use pre-computed byte slices for common parts
	hc.Buf.Write(httpVersion)

	// Use pre-computed status code bytes if available
	if codeBytes, ok := statusCodeBytes[statusCode]; ok {
		hc.Buf.Write(codeBytes)
	} else {
		hc.Buf.B = strconv.AppendInt(hc.Buf.B, int64(statusCode), 10)
	}

	hc.Buf.WriteByte(' ')
	hc.Buf.WriteString(StatusText(statusCode))
	hc.Buf.Write(crlfBytes)

	// Add Date header - use cached version to avoid expensive time formatting
	hc.Buf.Write(getDateHeader())

	// Add custom headers
	for k, values := range header {
		if len(values) == 1 { // Most common case: single value per header
			hc.Buf.WriteString(k)
			hc.Buf.Write(colonSpace)
			hc.Buf.WriteString(values[0])
			hc.Buf.Write(crlfBytes)
		} else {
			for _, v := range values {
				hc.Buf.WriteString(k)
				hc.Buf.Write(colonSpace)
				hc.Buf.WriteString(v)
				hc.Buf.Write(crlfBytes)
			}
		}
	}

	// Add Content-Length header
	hc.Buf.Write(contentLengthPrefix)
	hc.Buf.B = strconv.AppendInt(hc.Buf.B, int64(len(body)), 10)
	hc.Buf.Write(crlfBytes)

	// Add an additional CRLF to separate headers from body
	hc.Buf.Write(crlfBytes)

	// Add body
	if len(body) > 0 {
		hc.Buf.Write(body)
	}
}

// codecPool is a pool of Codec objects for reuse
var codecPool = pool.New(func() *Codec {
	return &Codec{
		Parser:        parserPool.Get(),
		ContentLength: -1,
	}
})

// NewCodec creates a new HTTP codec.
func NewCodec(router interface{}) *Codec {
	// Get a codec from the pool
	codec := codecPool.Get()

	// Reset content length
	codec.ContentLength = -1

	// Ensure we have a valid parser
	if codec.Parser == nil {
		codec.Parser = parserPool.Get()
	}

	// Set the router
	codec.Router = router

	// Get a new buffer from the pool
	codec.Buf = ResponseBufferPool.Get()

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
