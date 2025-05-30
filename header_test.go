package ngebut

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestHeaderAdd tests the Add method of Header
func TestHeaderAdd(t *testing.T) {
	h := make(Header)

	// Test adding a single value
	h.Add("Content-Type", "application/json")
	require.Equal(t, "application/json", h.Get("Content-Type"))

	// Test adding multiple values for the same key
	h.Add("Accept", "text/html")
	h.Add("Accept", "application/json")
	values := h.Values("Accept")
	require.Len(t, values, 2)
	require.Equal(t, "text/html", values[0])
	require.Equal(t, "application/json", values[1])

	// Test case insensitivity
	h.Add("content-type", "text/html")
	values = h.Values("Content-Type")
	require.Len(t, values, 2)
	require.Equal(t, "application/json", values[0])
	require.Equal(t, "text/html", values[1])
}

// TestHeaderSet tests the Set method of Header
func TestHeaderSet(t *testing.T) {
	h := make(Header)

	// Test setting a value
	h.Set("Content-Type", "application/json")
	require.Equal(t, "application/json", h.Get("Content-Type"))

	// Test overwriting a value
	h.Set("Content-Type", "text/html")
	require.Equal(t, "text/html", h.Get("Content-Type"))

	// Test case insensitivity
	h.Set("content-type", "application/xml")
	require.Equal(t, "application/xml", h.Get("Content-Type"))
}

// TestHeaderGet tests the Get method of Header
func TestHeaderGet(t *testing.T) {
	h := make(Header)

	// Test getting a non-existent key
	require.Empty(t, h.Get("Content-Type"))

	// Test getting an existing key
	h.Set("Content-Type", "application/json")
	require.Equal(t, "application/json", h.Get("Content-Type"))

	// Test case insensitivity
	require.Equal(t, "application/json", h.Get("content-type"))
}

// TestHeaderValues tests the Values method of Header
func TestHeaderValues(t *testing.T) {
	h := make(Header)

	// Test getting values for a non-existent key
	values := h.Values("Accept")
	require.Empty(t, values)

	// Test getting values for a key with a single value
	h.Set("Content-Type", "application/json")
	values = h.Values("Content-Type")
	require.Len(t, values, 1)
	require.Equal(t, "application/json", values[0])

	// Test getting values for a key with multiple values
	h.Add("Accept", "text/html")
	h.Add("Accept", "application/json")
	values = h.Values("Accept")
	require.Len(t, values, 2)
	require.Equal(t, "text/html", values[0])
	require.Equal(t, "application/json", values[1])

	// Test case insensitivity
	values = h.Values("accept")
	require.Len(t, values, 2)
	require.Equal(t, "text/html", values[0])
	require.Equal(t, "application/json", values[1])
}

// TestHeaderDel tests the Del method of Header
func TestHeaderDel(t *testing.T) {
	h := make(Header)
	h.Set("Content-Type", "application/json")
	h.Add("Accept", "text/html")
	h.Add("Accept", "application/json")

	// Test deleting a key
	h.Del("Content-Type")
	require.Empty(t, h.Get("Content-Type"))

	// Test case insensitivity
	h.Del("accept")
	require.Empty(t, h.Get("Accept"))

	// Test deleting a non-existent key (should not panic)
	require.NotPanics(t, func() {
		h.Del("X-Custom-Header")
	})
}

// TestHeaderClone tests the Clone method of Header
func TestHeaderClone(t *testing.T) {
	// Test cloning nil header
	var h Header
	clone := h.Clone()
	require.Nil(t, clone)

	// Test cloning empty header
	h = make(Header)
	clone = h.Clone()
	require.Empty(t, clone)

	// Test cloning header with values
	h.Set("Content-Type", "application/json")
	h.Add("Accept", "text/html")
	h.Add("Accept", "application/json")

	clone = h.Clone()

	// Check if clone has the same values
	require.Equal(t, "application/json", clone.Get("Content-Type"))

	values := clone.Values("Accept")
	require.Len(t, values, 2)
	require.Equal(t, "text/html", values[0])
	require.Equal(t, "application/json", values[1])

	// Modify original and check if clone is affected
	h.Set("Content-Type", "text/html")
	require.Equal(t, "application/json", clone.Get("Content-Type"))

	// Modify clone and check if original is affected
	clone.Set("Content-Type", "application/xml")
	require.Equal(t, "text/html", h.Get("Content-Type"))
}

// TestHeaderWrite tests the Write method of Header
func TestHeaderWrite(t *testing.T) {
	h := make(Header)
	h.Set("Content-Type", "application/json")
	h.Add("Accept", "text/html")
	h.Add("Accept", "application/json")

	var buf bytes.Buffer
	err := h.Write(&buf)
	require.NoError(t, err)

	got := buf.String()

	// Since map iteration order is not guaranteed, we check for individual headers
	require.Contains(t, got, "Content-Type: application/json\r\n")
	require.Contains(t, got, "Accept: text/html\r\n")
	require.Contains(t, got, "Accept: application/json\r\n")
}

// TestHeaderWriteSubset tests the WriteSubset method of Header
func TestHeaderWriteSubset(t *testing.T) {
	h := make(Header)
	h.Set("Content-Type", "application/json")
	h.Add("Accept", "text/html")
	h.Add("Accept", "application/json")
	h.Set("X-Custom-Header", "custom value")

	// Exclude Content-Type and X-Custom-Header
	exclude := map[string]bool{
		"Content-Type":    true,
		"X-Custom-Header": true,
	}

	var buf bytes.Buffer
	err := h.WriteSubset(&buf, exclude)
	require.NoError(t, err)

	got := buf.String()

	// Check that excluded headers are not present
	require.NotContains(t, got, "Content-Type")
	require.NotContains(t, got, "X-Custom-Header")

	// Check that non-excluded headers are present
	require.Contains(t, got, "Accept: text/html\r\n")
	require.Contains(t, got, "Accept: application/json\r\n")

	// Test with nil exclude
	buf.Reset()
	err = h.WriteSubset(&buf, nil)
	require.NoError(t, err)

	// Check that the output contains all headers
	got = buf.String()
	require.Contains(t, got, "Content-Type: application/json\r\n")
	require.Contains(t, got, "X-Custom-Header: custom value\r\n")
	require.Contains(t, got, "Accept: text/html\r\n")
	require.Contains(t, got, "Accept: application/json\r\n")
}

// TestHeaderSanitization tests that headers are properly sanitized
func TestHeaderSanitization(t *testing.T) {
	h := make(Header)
	h.Set("Content-Type", "application/json\r\nX-Injected: value")
	h.Set("X-Header", "value with \n newline and \r carriage return")

	var buf bytes.Buffer
	err := h.Write(&buf)
	require.NoError(t, err)

	// Check that newlines and carriage returns in values are replaced with spaces
	got := buf.String()

	// The header value should have spaces instead of newlines and carriage returns
	if strings.Contains(got, "application/json") {
		// Check that CR/LF are replaced with spaces
		require.Contains(t, got, "application/json  X-Injected: value")
	}

	if strings.Contains(got, "X-Header") {
		// Check that \n and \r are replaced with spaces
		require.Contains(t, got, "value with   newline and   carriage return")
	}

	// The CRLF at the end of each header line should still be there
	require.Contains(t, got, "\r\n")
}
