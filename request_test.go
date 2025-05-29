package ngebut

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// TestNewRequest tests the NewRequest function
func TestNewRequest(t *testing.T) {
	// Test with nil http.Request
	req := NewRequest(nil)
	if req == nil {
		t.Fatal("NewRequest(nil) returned nil, expected a new Request")
	}
	if req.Header == nil {
		t.Error("NewRequest(nil) returned a Request with nil Header")
	}
	if req.ctx == nil {
		t.Error("NewRequest(nil) returned a Request with nil ctx")
	}

	// Test with a valid http.Request
	httpReq, err := http.NewRequest("GET", "http://example.com/path?query=value", nil)
	if err != nil {
		t.Fatalf("Failed to create http.Request: %v", err)
	}
	httpReq.Header.Set("User-Agent", "test-agent")
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Host = "example.com"
	httpReq.RemoteAddr = "192.168.1.1:1234"

	req = NewRequest(httpReq)
	if req == nil {
		t.Fatal("NewRequest(httpReq) returned nil")
	}

	// Check that fields were copied correctly
	if req.Method != "GET" {
		t.Errorf("req.Method = %q, want %q", req.Method, "GET")
	}
	if req.URL.String() != "http://example.com/path?query=value" {
		t.Errorf("req.URL = %q, want %q", req.URL.String(), "http://example.com/path?query=value")
	}
	if req.Header.Get("User-Agent") != "test-agent" {
		t.Errorf("req.Header.Get(\"User-Agent\") = %q, want %q", req.Header.Get("User-Agent"), "test-agent")
	}
	if req.Header.Get("Content-Type") != "application/json" {
		t.Errorf("req.Header.Get(\"Content-Type\") = %q, want %q", req.Header.Get("Content-Type"), "application/json")
	}
	if req.Host != "example.com" {
		t.Errorf("req.Host = %q, want %q", req.Host, "example.com")
	}
	if req.RemoteAddr != "192.168.1.1:1234" {
		t.Errorf("req.RemoteAddr = %q, want %q", req.RemoteAddr, "192.168.1.1:1234")
	}

	// Test with a request that has a body
	body := strings.NewReader(`{"key":"value"}`)
	httpReq, err = http.NewRequest("POST", "http://example.com", body)
	if err != nil {
		t.Fatalf("Failed to create http.Request with body: %v", err)
	}

	req = NewRequest(httpReq)
	if req == nil {
		t.Fatal("NewRequest(httpReq) with body returned nil")
	}

	// Check that the body was read correctly
	if string(req.Body) != `{"key":"value"}` {
		t.Errorf("req.Body = %q, want %q", string(req.Body), `{"key":"value"}`)
	}

	// Check that the original body can still be read
	bodyBytes, err := io.ReadAll(httpReq.Body)
	if err != nil {
		t.Fatalf("Failed to read httpReq.Body: %v", err)
	}
	if string(bodyBytes) != `{"key":"value"}` {
		t.Errorf("httpReq.Body = %q, want %q", string(bodyBytes), `{"key":"value"}`)
	}
}

// TestRequestContext tests the Context method
func TestRequestContext(t *testing.T) {
	// Test with nil context
	req := &Request{
		ctx: nil,
	}
	ctx := req.Context()
	if ctx == nil {
		t.Fatal("req.Context() returned nil, expected background context")
	}
	if ctx != context.Background() {
		t.Error("req.Context() didn't return background context when ctx is nil")
	}

	// Test with non-nil context
	customCtx := context.WithValue(context.Background(), "key", "value")
	req.ctx = customCtx
	ctx = req.Context()
	if ctx != customCtx {
		t.Error("req.Context() didn't return the expected context")
	}
	if val, ok := ctx.Value("key").(string); !ok || val != "value" {
		t.Errorf("ctx.Value(\"key\") = %v, want %q", ctx.Value("key"), "value")
	}
}

// TestRequestWithContext tests the WithContext method
func TestRequestWithContext(t *testing.T) {
	// Create a request
	req := &Request{
		Method: "GET",
		URL: &url.URL{
			Scheme: "http",
			Host:   "example.com",
			Path:   "/path",
		},
		Header: make(Header),
		Body:   []byte("test body"),
		ctx:    context.Background(),
	}
	req.Header.Set("User-Agent", "test-agent")

	// Test with a new context
	customCtx := context.WithValue(context.Background(), "key", "value")
	newReq := req.WithContext(customCtx)

	// Check that the new request has the new context
	if newReq.ctx != customCtx {
		t.Error("newReq.ctx is not the custom context")
	}

	// Check that other fields were copied correctly
	if newReq.Method != req.Method {
		t.Errorf("newReq.Method = %q, want %q", newReq.Method, req.Method)
	}
	if newReq.URL.String() != req.URL.String() {
		t.Errorf("newReq.URL = %q, want %q", newReq.URL.String(), req.URL.String())
	}
	if newReq.Header.Get("User-Agent") != req.Header.Get("User-Agent") {
		t.Errorf("newReq.Header.Get(\"User-Agent\") = %q, want %q", newReq.Header.Get("User-Agent"), req.Header.Get("User-Agent"))
	}
	if string(newReq.Body) != string(req.Body) {
		t.Errorf("newReq.Body = %q, want %q", string(newReq.Body), string(req.Body))
	}

	// Test with nil context (should panic)
	defer func() {
		if r := recover(); r == nil {
			t.Error("WithContext(nil) did not panic")
		}
	}()
	req.WithContext(nil)
}

// TestRequestUserAgent tests the UserAgent method
func TestRequestUserAgent(t *testing.T) {
	// Test with no User-Agent header
	req := &Request{
		Header: make(Header),
	}
	if ua := req.UserAgent(); ua != "" {
		t.Errorf("req.UserAgent() = %q, want \"\"", ua)
	}

	// Test with User-Agent header
	req.Header.Set("User-Agent", "test-agent")
	if ua := req.UserAgent(); ua != "test-agent" {
		t.Errorf("req.UserAgent() = %q, want %q", ua, "test-agent")
	}
}

// TestRequestBodyBufferPool tests the requestBodyBufferPool functionality
func TestRequestBodyBufferPool(t *testing.T) {
	// Get a buffer from the pool
	buf := requestBodyBufferPool.Get().(*bytes.Buffer)
	if buf == nil {
		t.Fatal("requestBodyBufferPool.Get() returned nil")
	}

	// Reset the buffer to ensure it's empty
	buf.Reset()

	// Check initial capacity
	if buf.Cap() < 4096 {
		t.Errorf("buf.Cap() = %d, want at least 4096", buf.Cap())
	}

	// Write to the buffer
	testData := "test data"
	buf.WriteString(testData)
	if buf.String() != testData {
		t.Errorf("buf.String() = %q, want %q", buf.String(), testData)
	}

	// Reset and return to the pool
	buf.Reset()
	requestBodyBufferPool.Put(buf)

	// Get another buffer from the pool (might be the same one)
	buf2 := requestBodyBufferPool.Get().(*bytes.Buffer)
	if buf2 == nil {
		t.Fatal("requestBodyBufferPool.Get() returned nil on second call")
	}

	// Check that it's empty
	if buf2.Len() != 0 {
		t.Errorf("buf2.Len() = %d, want 0", buf2.Len())
	}

	// Return to the pool
	requestBodyBufferPool.Put(buf2)
}
