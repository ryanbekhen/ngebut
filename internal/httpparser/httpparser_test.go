package httpparser

import (
	"bytes"
	"testing"

	"github.com/evanphx/wildcat"
)

// TestEstimateResponseSize tests the EstimateResponseSize function
func TestEstimateResponseSize(t *testing.T) {
	// Test with empty body and headers
	size := EstimateResponseSize(200, Header{}, []byte{})
	if size <= 0 {
		t.Errorf("EstimateResponseSize() = %d, want > 0", size)
	}

	// Test with body
	body := []byte("Hello, World!")
	size = EstimateResponseSize(200, Header{}, body)
	if size <= len(body) {
		t.Errorf("EstimateResponseSize() = %d, want > %d", size, len(body))
	}

	// Test with headers
	header := Header{
		"Content-Type": []string{"application/json"},
		"X-Custom":     []string{"value1", "value2"},
	}
	size = EstimateResponseSize(200, header, body)
	if size <= 0 {
		t.Errorf("EstimateResponseSize() = %d, want > 0", size)
	}
}

// TestCodec tests the Codec implementation
func TestCodec(t *testing.T) {
	// Create a new Codec
	hc := NewCodec(nil)

	// Test ResetParser method
	hc.ContentLength = 100
	hc.ResetParser()
	if hc.ContentLength != -1 {
		t.Errorf("hc.ContentLength = %d, want %d", hc.ContentLength, -1)
	}

	// Test Reset method
	hc.Buf = append(hc.Buf, []byte("test")...)
	hc.ContentLength = 100
	hc.Reset()
	if hc.ContentLength != -1 {
		t.Errorf("hc.ContentLength = %d, want %d", hc.ContentLength, -1)
	}
	if len(hc.Buf) != 0 {
		t.Errorf("len(hc.Buf) = %d, want %d", len(hc.Buf), 0)
	}

	// Test WriteResponse method
	statusCode := 200
	header := Header{
		"Content-Type": []string{"text/plain"},
	}
	body := []byte("Hello, World!")

	hc.WriteResponse(statusCode, header, body)

	// Check that the buffer contains the expected response
	if len(hc.Buf) == 0 {
		t.Error("WriteResponse() did not write anything to the buffer")
	}

	// Check for status line
	statusLine := "HTTP/1.1 200 OK\r\n"
	if !bytes.Contains(hc.Buf, []byte(statusLine)) {
		t.Errorf("Response does not contain status line: %q", statusLine)
	}

	// Check for header
	headerLine := "Content-Type: text/plain\r\n"
	if !bytes.Contains(hc.Buf, []byte(headerLine)) {
		t.Errorf("Response does not contain header: %q", headerLine)
	}

	// Check for body
	if !bytes.Contains(hc.Buf, body) {
		t.Errorf("Response does not contain body: %q", body)
	}
}

// TestCodecGetContentLength tests the GetContentLength method of Codec
func TestCodecGetContentLength(t *testing.T) {
	// Create a new Codec
	hc := NewCodec(nil)

	// Test when ContentLength is already set
	hc.ContentLength = 100
	length := hc.GetContentLength()
	if length != 100 {
		t.Errorf("GetContentLength() = %d, want %d", length, 100)
	}

	// Test when ContentLength is not set and Content-Length header is present
	hc.ContentLength = -1
	hc.Parser.Parse([]byte("GET / HTTP/1.1\r\nContent-Length: 42\r\n\r\n"))

	length = hc.GetContentLength()
	if length != 42 {
		t.Errorf("GetContentLength() = %d, want %d", length, 42)
	}

	// Test when ContentLength is not set and Content-Length header is not present
	hc.ContentLength = -1
	hc.Parser = parserPool.Get().(*wildcat.HTTPParser)

	// Simulate a request without Content-Length header
	hc.Parser.Parse([]byte("GET / HTTP/1.1\r\n\r\n"))

	length = hc.GetContentLength()
	if length != -1 {
		t.Errorf("GetContentLength() = %d, want %d", length, -1)
	}
}

// TestCodecParse tests the Parse method of Codec
func TestCodecParse(t *testing.T) {
	// Create a new Codec
	hc := NewCodec(nil)

	// Test parsing a simple GET request without body
	simpleReq := "GET / HTTP/1.1\r\nHost: example.com\r\n\r\n"
	n, body, err := hc.Parse([]byte(simpleReq))

	if err != nil {
		t.Errorf("Parse() returned error: %v", err)
	}
	if n != len(simpleReq) {
		t.Errorf("Parse() = %d, want %d", n, len(simpleReq))
	}
	// For a GET request, body should be nil or empty
	if body != nil && len(body) > 0 {
		t.Errorf("Parse() returned non-empty body for GET request: %v", body)
	}

	// Test parsing a POST request with Content-Length
	postReq := "POST / HTTP/1.1\r\nHost: example.com\r\nContent-Length: 11\r\n\r\nHello World"
	hc.ResetParser()
	n, body, err = hc.Parse([]byte(postReq))

	if err != nil {
		t.Errorf("Parse() returned error: %v", err)
	}
	if n != len(postReq) {
		t.Errorf("Parse() = %d, want %d", n, len(postReq))
	}
	// For a POST request with body, body should not be nil
	if body == nil {
		t.Error("Parse() returned nil body for POST request with body")
	} else if string(body) != "Hello World" {
		t.Errorf("Parse() returned body = %q, want %q", string(body), "Hello World")
	}

	// Test parsing an incomplete request
	incompleteReq := "GET / HTTP/1.1\r\nHost: example.com\r\n"
	hc.ResetParser()
	_, _, err = hc.Parse([]byte(incompleteReq))

	if err == nil {
		t.Error("Parse() did not return error for incomplete request")
	}

	// Test parsing a chunked request
	chunkedReq := "POST / HTTP/1.1\r\nHost: example.com\r\nTransfer-Encoding: chunked\r\n\r\n5\r\nHello\r\n0\r\n\r\n"
	hc.ResetParser()
	n, body, err = hc.Parse([]byte(chunkedReq))

	if err != nil {
		t.Errorf("Parse() returned error: %v", err)
	}
	// For a chunked request, body should not be nil
	if body == nil {
		t.Error("Parse() returned nil body for chunked request")
	} else if string(body) != "Hello" {
		t.Errorf("Parse() returned body = %q, want %q", string(body), "Hello")
	}
}

// TestParserReset tests that the parser can be reset
func TestParserReset(t *testing.T) {
	// Create a new Codec
	hc := NewCodec(nil)

	// First, parse an incomplete request which should fail
	incompleteReq := "GET / HTTP/1.1\r\nHost: example.com\r\n"
	_, _, err := hc.Parse([]byte(incompleteReq))
	if err == nil {
		t.Error("Parse() did not return error for incomplete request")
	}

	// Reset the parser
	hc.ResetParser()

	// Now try to parse a complete request
	completeReq := "GET / HTTP/1.1\r\nHost: example.com\r\n\r\n"
	n, body, err := hc.Parse([]byte(completeReq))

	// This should succeed
	if err != nil {
		t.Errorf("Parse() returned error after reset: %v", err)
	}
	if n != len(completeReq) {
		t.Errorf("Parse() = %d, want %d", n, len(completeReq))
	}
	// For a GET request, body should be nil or empty
	if body != nil && len(body) > 0 {
		t.Errorf("Parse() returned non-empty body for GET request: %v", body)
	}
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
	if err != nil {
		t.Errorf("bodyReader.Read() returned error: %v", err)
	}
	if n != len(testData) {
		t.Errorf("bodyReader.Read() = %d, want %d", n, len(testData))
	}
	if !bytes.Equal(buf, testData) {
		t.Errorf("bodyReader.Read() = %q, want %q", buf, testData)
	}

	// Close the reader
	err = br.Close()
	if err != nil {
		t.Errorf("bodyReader.Close() returned error: %v", err)
	}

	// Read again after close (should work because Close resets position)
	n, err = br.Read(buf)
	if err != nil {
		t.Errorf("bodyReader.Read() after Close returned error: %v", err)
	}
	if n != len(testData) {
		t.Errorf("bodyReader.Read() after Close = %d, want %d", n, len(testData))
	}

	// Release the reader back to the pool
	ReleaseBodyReader(br)
}

// TestPooledReaders tests the reader pool functions
func TestPooledReaders(t *testing.T) {
	// Test GetReader and ReleaseReader
	reader := GetReader()
	if reader == nil {
		t.Error("GetReader() returned nil")
	}
	ReleaseReader(reader)

	// Test GetBytesReader and ReleaseBytesReader
	bytesReader := GetBytesReader()
	if bytesReader == nil {
		t.Error("GetBytesReader() returned nil")
	}
	ReleaseBytesReader(bytesReader)

	// Test GetBodyReader and ReleaseBodyReader
	bodyReader := GetBodyReader([]byte("test"))
	if bodyReader == nil {
		t.Error("GetBodyReader() returned nil")
	}
	ReleaseBodyReader(bodyReader)
}
