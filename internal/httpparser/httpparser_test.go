package httpparser

import (
	"testing"

	"github.com/evanphx/wildcat"
	"github.com/stretchr/testify/assert"
)

// TestEstimateResponseSize tests the EstimateResponseSize function
func TestEstimateResponseSize(t *testing.T) {
	// Test with empty body and headers
	size := EstimateResponseSize(200, Header{}, []byte{})
	assert.Greater(t, size, 0, "Size should be greater than 0 for empty body and headers")

	// Test with body
	body := []byte("Hello, World!")
	size = EstimateResponseSize(200, Header{}, body)
	assert.Greater(t, size, len(body), "Size should be greater than body length")

	// Test with headers
	header := Header{
		"Content-Type": []string{"application/json"},
		"X-Custom":     []string{"value1", "value2"},
	}
	size = EstimateResponseSize(200, header, body)
	assert.Greater(t, size, 0, "Size should be greater than 0 with headers")
}

// TestCodec tests the Codec implementation
func TestCodec(t *testing.T) {
	// Create a new Codec
	hc := NewCodec(nil)

	// Test ResetParser method
	hc.ContentLength = 100
	hc.ResetParser()
	assert.Equal(t, -1, hc.ContentLength, "ContentLength should be -1 after ResetParser")

	// Test Reset method
	hc.Buf = append(hc.Buf, []byte("test")...)
	hc.ContentLength = 100
	hc.Reset()
	assert.Equal(t, -1, hc.ContentLength, "ContentLength should be -1 after Reset")
	assert.Empty(t, hc.Buf, "Buffer should be empty after Reset")

	// Test WriteResponse method
	statusCode := 200
	header := Header{
		"Content-Type": []string{"text/plain"},
	}
	body := []byte("Hello, World!")

	hc.WriteResponse(statusCode, header, body)

	// Check that the buffer contains the expected response
	assert.NotEmpty(t, hc.Buf, "WriteResponse should write data to the buffer")

	// Check for status line
	statusLine := "HTTP/1.1 200 OK\r\n"
	assert.Contains(t, string(hc.Buf), statusLine, "Response should contain status line")

	// Check for header
	headerLine := "Content-Type: text/plain\r\n"
	assert.Contains(t, string(hc.Buf), headerLine, "Response should contain header")

	// Check for body
	assert.Contains(t, string(hc.Buf), string(body), "Response should contain body")
}

// TestCodecGetContentLength tests the GetContentLength method of Codec
func TestCodecGetContentLength(t *testing.T) {
	// Create a new Codec
	hc := NewCodec(nil)

	// Test when ContentLength is already set
	hc.ContentLength = 100
	length := hc.GetContentLength()
	assert.Equal(t, 100, length, "GetContentLength should return the set ContentLength")

	// Test when ContentLength is not set and Content-Length header is present
	hc.ContentLength = -1
	hc.Parser.Parse([]byte("GET / HTTP/1.1\r\nContent-Length: 42\r\n\r\n"))

	length = hc.GetContentLength()
	assert.Equal(t, 42, length, "GetContentLength should return header Content-Length value")

	// Test when ContentLength is not set and Content-Length header is not present
	hc.ContentLength = -1
	hc.Parser = parserPool.Get().(*wildcat.HTTPParser)

	// Simulate a request without Content-Length header
	hc.Parser.Parse([]byte("GET / HTTP/1.1\r\n\r\n"))

	length = hc.GetContentLength()
	assert.Equal(t, -1, length, "GetContentLength should return -1 when no Content-Length is set")
}

// TestCodecParse tests the Parse method of Codec
func TestCodecParse(t *testing.T) {
	// Create a new Codec
	hc := NewCodec(nil)

	// Test parsing a simple GET request without body
	simpleReq := "GET / HTTP/1.1\r\nHost: example.com\r\n\r\n"
	n, body, err := hc.Parse([]byte(simpleReq))

	assert.NoError(t, err, "Parse should not return error for valid request")
	assert.Equal(t, len(simpleReq), n, "Parse should return the correct number of bytes read")
	// For a GET request, body should be nil or empty
	assert.True(t, body == nil || len(body) == 0, "Body should be nil or empty for GET request")

	// Test parsing a POST request with Content-Length
	postReq := "POST / HTTP/1.1\r\nHost: example.com\r\nContent-Length: 11\r\n\r\nHello World"
	hc.ResetParser()
	n, body, err = hc.Parse([]byte(postReq))

	assert.NoError(t, err, "Parse should not return error for valid POST request")
	assert.Equal(t, len(postReq), n, "Parse should return the correct number of bytes read")
	assert.NotNil(t, body, "Body should not be nil for POST request with body")
	assert.Equal(t, "Hello World", string(body), "Body content should match")

	// Test parsing an incomplete request
	incompleteReq := "GET / HTTP/1.1\r\nHost: example.com\r\n"
	hc.ResetParser()
	_, _, err = hc.Parse([]byte(incompleteReq))

	assert.Error(t, err, "Parse should return error for incomplete request")

	// Test parsing a chunked request
	chunkedReq := "POST / HTTP/1.1\r\nHost: example.com\r\nTransfer-Encoding: chunked\r\n\r\n5\r\nHello\r\n0\r\n\r\n"
	hc.ResetParser()
	n, body, err = hc.Parse([]byte(chunkedReq))

	assert.NoError(t, err, "Parse should not return error for valid chunked request")
	assert.NotNil(t, body, "Body should not be nil for chunked request")
	assert.Equal(t, "Hello", string(body), "Body content should match")
}

// TestParserReset tests that the parser can be reset
func TestParserReset(t *testing.T) {
	// Create a new Codec
	hc := NewCodec(nil)

	// First, parse an incomplete request which should fail
	incompleteReq := "GET / HTTP/1.1\r\nHost: example.com\r\n"
	_, _, err := hc.Parse([]byte(incompleteReq))
	assert.Error(t, err, "Parse should return error for incomplete request")

	// Reset the parser
	hc.ResetParser()

	// Now try to parse a complete request
	completeReq := "GET / HTTP/1.1\r\nHost: example.com\r\n\r\n"
	n, body, err := hc.Parse([]byte(completeReq))

	// This should succeed
	assert.NoError(t, err, "Parse should not return error after reset")
	assert.Equal(t, len(completeReq), n, "Parse should return the correct number of bytes read")
	assert.True(t, body == nil || len(body) == 0, "Body should be nil or empty for GET request")
}

// TestBodyReader tests the bodyReader implementation
func TestBodyReader(t *testing.T) {
	// Create test data
	testData := []byte("Hello, World!")

	// Get a bodyReader from the pool
	br := GetBodyReader(testData)

	// Read the data
	buf := make([]byte, len(testData))
	n, err := br.Read(buf)

	// Check results
	assert.NoError(t, err, "bodyReader.Read should not return error")
	assert.Equal(t, len(testData), n, "bodyReader.Read should read the correct number of bytes")
	assert.Equal(t, testData, buf, "bodyReader.Read should return the correct data")

	// Close the reader
	err = br.Close()
	assert.NoError(t, err, "bodyReader.Close should not return error")

	// Read again after close (should work because Close resets position)
	n, err = br.Read(buf)
	assert.NoError(t, err, "bodyReader.Read after Close should not return error")
	assert.Equal(t, len(testData), n, "bodyReader.Read after Close should read the correct number of bytes")

	// Release the reader back to the pool
	ReleaseBodyReader(br)
}

// TestPooledReaders tests the reader pool functions
func TestPooledReaders(t *testing.T) {
	// Test GetReader and ReleaseReader
	reader := GetReader()
	assert.NotNil(t, reader, "GetReader should not return nil")
	ReleaseReader(reader)

	// Test GetBytesReader and ReleaseBytesReader
	bytesReader := GetBytesReader()
	assert.NotNil(t, bytesReader, "GetBytesReader should not return nil")
	ReleaseBytesReader(bytesReader)

	// Test GetBodyReader and ReleaseBodyReader
	bodyReader := GetBodyReader([]byte("test"))
	assert.NotNil(t, bodyReader, "GetBodyReader should not return nil")
	ReleaseBodyReader(bodyReader)
}
