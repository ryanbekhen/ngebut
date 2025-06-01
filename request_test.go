package ngebut

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewRequest tests the NewRequest function
func TestNewRequest(t *testing.T) {
	// Test with nil http.Request
	req := NewRequest(nil)
	require.NotNil(t, req, "NewRequest(nil) returned nil, expected a new Request")
	assert.NotNil(t, req.Header, "NewRequest(nil) returned a Request with nil Header")
	assert.NotNil(t, req.ctx, "NewRequest(nil) returned a Request with nil ctx")

	// Test with a valid http.Request
	httpReq, err := http.NewRequest("GET", "http://example.com/path?query=value", nil)
	require.NoError(t, err, "Failed to create http.Request")

	httpReq.Header.Set("User-Agent", "test-agent")
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Host = "example.com"
	httpReq.RemoteAddr = "192.168.1.1:1234"

	req = NewRequest(httpReq)
	require.NotNil(t, req, "NewRequest(httpReq) returned nil")

	// Check that fields were copied correctly
	assert.Equal(t, "GET", req.Method, "req.Method was not copied correctly")
	assert.Equal(t, "http://example.com/path?query=value", req.URL.String(), "req.URL was not copied correctly")
	assert.Equal(t, "test-agent", req.Header.Get("User-Agent"), "User-Agent header was not copied correctly")
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"), "Content-Type header was not copied correctly")
	assert.Equal(t, "example.com", req.Host, "req.Host was not copied correctly")
	assert.Equal(t, "192.168.1.1:1234", req.RemoteAddr, "req.RemoteAddr was not copied correctly")

	// Test with a request that has a body
	body := strings.NewReader(`{"key":"value"}`)
	httpReq, err = http.NewRequest("POST", "http://example.com", body)
	require.NoError(t, err, "Failed to create http.Request with body")

	req = NewRequest(httpReq)
	require.NotNil(t, req, "NewRequest(httpReq) with body returned nil")

	// Check that the body was read correctly
	assert.Equal(t, `{"key":"value"}`, string(req.Body), "req.Body was not copied correctly")

	// Check that the original body can still be read
	bodyBytes, err := io.ReadAll(httpReq.Body)
	require.NoError(t, err, "Failed to read httpReq.Body")
	assert.Equal(t, `{"key":"value"}`, string(bodyBytes), "httpReq.Body was not readable after NewRequest")
}

// TestRequestContext tests the Context method
func TestRequestContext(t *testing.T) {
	// Test with nil context
	req := &Request{
		ctx: nil,
	}
	ctx := req.Context()
	require.NotNil(t, ctx, "req.Context() returned nil, expected background context")
	assert.Equal(t, context.Background(), ctx, "req.Context() didn't return background context when ctx is nil")

	// Test with non-nil context
	customCtx := context.WithValue(context.Background(), "key", "value")
	req.ctx = customCtx
	ctx = req.Context()
	assert.Equal(t, customCtx, ctx, "req.Context() didn't return the expected context")
	val, ok := ctx.Value("key").(string)
	assert.True(t, ok, "ctx.Value(\"key\") is not a string")
	assert.Equal(t, "value", val, "ctx.Value(\"key\") doesn't have expected value")
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
		Header: NewHeader(),
		Body:   []byte("test body"),
		ctx:    context.Background(),
	}
	req.Header.Set("User-Agent", "test-agent")

	// Test with a new context
	customCtx := context.WithValue(context.Background(), "key", "value")
	newReq := req.WithContext(customCtx)

	// Check that the new request has the new context
	assert.Equal(t, customCtx, newReq.ctx, "newReq.ctx is not the custom context")

	// Check that other fields were copied correctly
	assert.Equal(t, req.Method, newReq.Method, "newReq.Method was not copied correctly")
	assert.Equal(t, req.URL.String(), newReq.URL.String(), "newReq.URL was not copied correctly")
	assert.Equal(t, req.Header.Get("User-Agent"), newReq.Header.Get("User-Agent"), "newReq.Header was not copied correctly")
	assert.Equal(t, string(req.Body), string(newReq.Body), "newReq.Body was not copied correctly")

	// Test with nil context (should panic)
	assert.Panics(t, func() {
		req.WithContext(nil)
	}, "WithContext(nil) did not panic")
}

// TestRequestUserAgent tests the UserAgent method
func TestRequestUserAgent(t *testing.T) {
	// Test with no User-Agent header
	req := &Request{
		Header: NewHeader(),
	}
	assert.Equal(t, "", req.UserAgent(), "req.UserAgent() should return empty string when no User-Agent header")

	// Test with User-Agent header
	req.Header.Set("User-Agent", "test-agent")
	assert.Equal(t, "test-agent", req.UserAgent(), "req.UserAgent() returned incorrect value")
}

// TestRequestSetContext tests the SetContext method
func TestRequestSetContext(t *testing.T) {
	// Create a request
	req := &Request{
		Method: "GET",
		URL: &url.URL{
			Scheme: "http",
			Host:   "example.com",
			Path:   "/path",
		},
		Header: NewHeader(),
		Body:   []byte("test body"),
		ctx:    context.Background(),
	}

	// Test with a new context
	customCtx := context.WithValue(context.Background(), "key", "value")
	req.SetContext(customCtx)

	// Check that the context was set correctly
	assert.Equal(t, customCtx, req.ctx, "req.ctx is not the custom context after SetContext")

	// Check that the context can be retrieved correctly
	ctx := req.Context()
	assert.Equal(t, customCtx, ctx, "req.Context() didn't return the expected context after SetContext")
	val, ok := ctx.Value("key").(string)
	assert.True(t, ok, "ctx.Value(\"key\") is not a string")
	assert.Equal(t, "value", val, "ctx.Value(\"key\") doesn't have expected value")

	// Test with nil context (should panic)
	assert.Panics(t, func() {
		req.SetContext(nil)
	}, "SetContext(nil) did not panic")
}

// TestRequestBodyBufferPool tests the requestBodyBufferPool functionality
func TestRequestBodyBufferPool(t *testing.T) {
	// Get a buffer from the pool
	buf := requestBodyBufferPool.Get().(*bytes.Buffer)
	require.NotNil(t, buf, "requestBodyBufferPool.Get() returned nil")

	// Reset the buffer to ensure it's empty
	buf.Reset()

	// Check initial capacity
	assert.GreaterOrEqual(t, buf.Cap(), 4096, "buf.Cap() should be at least 4096")

	// Write to the buffer
	testData := "test data"
	buf.WriteString(testData)
	assert.Equal(t, testData, buf.String(), "Buffer content doesn't match written data")

	// Reset and return to the pool
	buf.Reset()
	requestBodyBufferPool.Put(buf)

	// Get another buffer from the pool (might be the same one)
	buf2 := requestBodyBufferPool.Get().(*bytes.Buffer)
	require.NotNil(t, buf2, "requestBodyBufferPool.Get() returned nil on second call")

	// Check that it's empty
	assert.Equal(t, 0, buf2.Len(), "buf2.Len() should be 0")

	// Return to the pool
	requestBodyBufferPool.Put(buf2)
}
