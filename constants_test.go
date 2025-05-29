package ngebut

import (
	"bytes"
	"testing"
)

// TestCRLF tests that the crlf constant has the correct value
func TestCRLF(t *testing.T) {
	expected := []byte{0x0d, 0x0a, 0x0d, 0x0a} // "\r\n\r\n"

	if !bytes.Equal(crlf, expected) {
		t.Errorf("crlf = %v, want %v", crlf, expected)
	}

	// Also check that it's equivalent to the string representation
	stringRepresentation := []byte("\r\n\r\n")
	if !bytes.Equal(crlf, stringRepresentation) {
		t.Errorf("crlf = %v, want %v", crlf, stringRepresentation)
	}
}

// TestLastChunk tests that the lastChunk constant has the correct value
func TestLastChunk(t *testing.T) {
	expected := []byte{0x30, 0x0d, 0x0a, 0x0d, 0x0a} // "0\r\n\r\n"

	if !bytes.Equal(lastChunk, expected) {
		t.Errorf("lastChunk = %v, want %v", lastChunk, expected)
	}

	// Also check that it's equivalent to the string representation
	stringRepresentation := []byte("0\r\n\r\n")
	if !bytes.Equal(lastChunk, stringRepresentation) {
		t.Errorf("lastChunk = %v, want %v", lastChunk, stringRepresentation)
	}
}

// TestConstantsUsage tests how these constants might be used in practice
func TestConstantsUsage(t *testing.T) {
	// Test that crlf can be used to detect the end of HTTP headers
	headers := []byte("Content-Type: application/json\r\nContent-Length: 123\r\n\r\n")
	if !bytes.HasSuffix(headers, crlf) {
		t.Errorf("crlf doesn't match the end of HTTP headers")
	}

	// Test that lastChunk can be used to detect the end of a chunked HTTP response
	chunkedResponse := []byte("10\r\n0123456789abcdef\r\n0\r\n\r\n")
	if !bytes.HasSuffix(chunkedResponse, lastChunk) {
		t.Errorf("lastChunk doesn't match the end of a chunked HTTP response")
	}
}
