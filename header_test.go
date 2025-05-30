package ngebut

import (
	"bytes"
	"strings"
	"testing"
)

// TestHeaderAdd tests the Add method of Header
func TestHeaderAdd(t *testing.T) {
	h := make(Header)

	// Test adding a single value
	h.Add("Content-Type", "application/json")
	if got := h.Get("Content-Type"); got != "application/json" {
		t.Errorf("h.Get(\"Content-Type\") = %q, want %q", got, "application/json")
	}

	// Test adding multiple values for the same key
	h.Add("Accept", "text/html")
	h.Add("Accept", "application/json")
	values := h.Values("Accept")
	if len(values) != 2 {
		t.Errorf("len(h.Values(\"Accept\")) = %d, want 2", len(values))
	}
	if values[0] != "text/html" || values[1] != "application/json" {
		t.Errorf("h.Values(\"Accept\") = %v, want [\"text/html\", \"application/json\"]", values)
	}

	// Test case insensitivity
	h.Add("content-type", "text/html")
	values = h.Values("Content-Type")
	if len(values) != 2 {
		t.Errorf("len(h.Values(\"Content-Type\")) = %d, want 2", len(values))
	}
	if values[0] != "application/json" || values[1] != "text/html" {
		t.Errorf("h.Values(\"Content-Type\") = %v, want [\"application/json\", \"text/html\"]", values)
	}
}

// TestHeaderSet tests the Set method of Header
func TestHeaderSet(t *testing.T) {
	h := make(Header)

	// Test setting a value
	h.Set("Content-Type", "application/json")
	if got := h.Get("Content-Type"); got != "application/json" {
		t.Errorf("h.Get(\"Content-Type\") = %q, want %q", got, "application/json")
	}

	// Test overwriting a value
	h.Set("Content-Type", "text/html")
	if got := h.Get("Content-Type"); got != "text/html" {
		t.Errorf("h.Get(\"Content-Type\") = %q, want %q", got, "text/html")
	}

	// Test case insensitivity
	h.Set("content-type", "application/xml")
	if got := h.Get("Content-Type"); got != "application/xml" {
		t.Errorf("h.Get(\"Content-Type\") = %q, want %q", got, "application/xml")
	}
}

// TestHeaderGet tests the Get method of Header
func TestHeaderGet(t *testing.T) {
	h := make(Header)

	// Test getting a non-existent key
	if got := h.Get("Content-Type"); got != "" {
		t.Errorf("h.Get(\"Content-Type\") = %q, want \"\"", got)
	}

	// Test getting an existing key
	h.Set("Content-Type", "application/json")
	if got := h.Get("Content-Type"); got != "application/json" {
		t.Errorf("h.Get(\"Content-Type\") = %q, want %q", got, "application/json")
	}

	// Test case insensitivity
	if got := h.Get("content-type"); got != "application/json" {
		t.Errorf("h.Get(\"content-type\") = %q, want %q", got, "application/json")
	}
}

// TestHeaderValues tests the Values method of Header
func TestHeaderValues(t *testing.T) {
	h := make(Header)

	// Test getting values for a non-existent key
	values := h.Values("Accept")
	if len(values) != 0 {
		t.Errorf("len(h.Values(\"Accept\")) = %d, want 0", len(values))
	}

	// Test getting values for a key with a single value
	h.Set("Content-Type", "application/json")
	values = h.Values("Content-Type")
	if len(values) != 1 || values[0] != "application/json" {
		t.Errorf("h.Values(\"Content-Type\") = %v, want [\"application/json\"]", values)
	}

	// Test getting values for a key with multiple values
	h.Add("Accept", "text/html")
	h.Add("Accept", "application/json")
	values = h.Values("Accept")
	if len(values) != 2 || values[0] != "text/html" || values[1] != "application/json" {
		t.Errorf("h.Values(\"Accept\") = %v, want [\"text/html\", \"application/json\"]", values)
	}

	// Test case insensitivity
	values = h.Values("accept")
	if len(values) != 2 || values[0] != "text/html" || values[1] != "application/json" {
		t.Errorf("h.Values(\"accept\") = %v, want [\"text/html\", \"application/json\"]", values)
	}
}

// TestHeaderDel tests the Del method of Header
func TestHeaderDel(t *testing.T) {
	h := make(Header)
	h.Set("Content-Type", "application/json")
	h.Add("Accept", "text/html")
	h.Add("Accept", "application/json")

	// Test deleting a key
	h.Del("Content-Type")
	if got := h.Get("Content-Type"); got != "" {
		t.Errorf("h.Get(\"Content-Type\") = %q, want \"\"", got)
	}

	// Test case insensitivity
	h.Del("accept")
	if got := h.Get("Accept"); got != "" {
		t.Errorf("h.Get(\"Accept\") = %q, want \"\"", got)
	}

	// Test deleting a non-existent key (should not panic)
	h.Del("X-Custom-Header")
}

// TestHeaderClone tests the Clone method of Header
func TestHeaderClone(t *testing.T) {
	// Test cloning nil header
	var h Header
	clone := h.Clone()
	if clone != nil {
		t.Errorf("nil.Clone() = %v, want nil", clone)
	}

	// Test cloning empty header
	h = make(Header)
	clone = h.Clone()
	if len(clone) != 0 {
		t.Errorf("len(empty.Clone()) = %d, want 0", len(clone))
	}

	// Test cloning header with values
	h.Set("Content-Type", "application/json")
	h.Add("Accept", "text/html")
	h.Add("Accept", "application/json")

	clone = h.Clone()

	// Check if clone has the same values
	if got := clone.Get("Content-Type"); got != "application/json" {
		t.Errorf("clone.Get(\"Content-Type\") = %q, want %q", got, "application/json")
	}

	values := clone.Values("Accept")
	if len(values) != 2 || values[0] != "text/html" || values[1] != "application/json" {
		t.Errorf("clone.Values(\"Accept\") = %v, want [\"text/html\", \"application/json\"]", values)
	}

	// Modify original and check if clone is affected
	h.Set("Content-Type", "text/html")
	if got := clone.Get("Content-Type"); got != "application/json" {
		t.Errorf("clone.Get(\"Content-Type\") = %q, want %q", got, "application/json")
	}

	// Modify clone and check if original is affected
	clone.Set("Content-Type", "application/xml")
	if got := h.Get("Content-Type"); got != "text/html" {
		t.Errorf("h.Get(\"Content-Type\") = %q, want %q", got, "text/html")
	}
}

// TestHeaderWrite tests the Write method of Header
func TestHeaderWrite(t *testing.T) {
	h := make(Header)
	h.Set("Content-Type", "application/json")
	h.Add("Accept", "text/html")
	h.Add("Accept", "application/json")

	var buf bytes.Buffer
	err := h.Write(&buf)
	if err != nil {
		t.Fatalf("h.Write() returned error: %v", err)
	}

	expected := "Content-Type: application/json\r\nAccept: text/html\r\nAccept: application/json\r\n"
	// Since map iteration order is not guaranteed, we need to check for both possible orders
	alt1 := "Accept: text/html\r\nAccept: application/json\r\nContent-Type: application/json\r\n"
	alt2 := "Accept: application/json\r\nAccept: text/html\r\nContent-Type: application/json\r\n"

	got := buf.String()
	if got != expected && got != alt1 && got != alt2 {
		t.Errorf("h.Write() produced %q, want one of:\n%q\n%q\n%q", got, expected, alt1, alt2)
	}
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
	if err != nil {
		t.Fatalf("h.WriteSubset() returned error: %v", err)
	}

	expected := "Accept: text/html\r\nAccept: application/json\r\n"
	alt := "Accept: application/json\r\nAccept: text/html\r\n"

	got := buf.String()
	if got != expected && got != alt {
		t.Errorf("h.WriteSubset() produced %q, want either %q or %q", got, expected, alt)
	}

	// Test with nil exclude
	buf.Reset()
	err = h.WriteSubset(&buf, nil)
	if err != nil {
		t.Fatalf("h.WriteSubset(nil) returned error: %v", err)
	}

	// Check that the output contains all headers
	got = buf.String()
	if !strings.Contains(got, "Content-Type: application/json\r\n") {
		t.Errorf("h.WriteSubset(nil) output doesn't contain Content-Type header")
	}
	if !strings.Contains(got, "X-Custom-Header: custom value\r\n") {
		t.Errorf("h.WriteSubset(nil) output doesn't contain X-Custom-Header header")
	}
	if !strings.Contains(got, "Accept: text/html\r\n") {
		t.Errorf("h.WriteSubset(nil) output doesn't contain Accept: text/html header")
	}
	if !strings.Contains(got, "Accept: application/json\r\n") {
		t.Errorf("h.WriteSubset(nil) output doesn't contain Accept: application/json header")
	}
}

// TestHeaderSanitization tests that headers are properly sanitized
func TestHeaderSanitization(t *testing.T) {
	h := make(Header)
	h.Set("Content-Type", "application/json\r\nX-Injected: value")
	h.Set("X-Header", "value with \n newline and \r carriage return")

	var buf bytes.Buffer
	err := h.Write(&buf)
	if err != nil {
		t.Fatalf("h.Write() returned error: %v", err)
	}

	// Check that newlines and carriage returns in values are replaced with spaces
	got := buf.String()

	// The header value should have spaces instead of newlines and carriage returns
	if strings.Contains(got, "Content-Type: application/json  X-Injected: value") {
		// This is expected behavior - the \r\n in the middle of the value is replaced with spaces
		// but it doesn't prevent the value from being written
	} else {
		t.Errorf("h.Write() didn't properly replace newlines with spaces in Content-Type header: %q", got)
	}

	if strings.Contains(got, "X-Header: value with   newline and   carriage return") {
		// This is expected behavior - \n and \r are replaced with spaces
	} else {
		t.Errorf("h.Write() didn't properly replace newlines with spaces in X-Header header: %q", got)
	}

	// The CRLF at the end of each header line should still be there
	if !strings.Contains(got, "\r\n") {
		t.Errorf("h.Write() didn't include CRLF line endings: %q", got)
	}
}
