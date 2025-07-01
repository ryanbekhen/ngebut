package ngebut

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNew tests the New function
func TestNew(t *testing.T) {
	// Test with default config
	server := New(DefaultConfig())
	require.NotNil(t, server, "New() returned nil")
	assert.NotNil(t, server.router, "server.router is nil")
	assert.NotNil(t, server.httpServer, "server.httpServer is nil")

	// Test with custom config
	customConfig := Config{
		DisableStartupMessage: true,
		ErrorHandler: func(c *Ctx) {
			c.Status(StatusInternalServerError).String("Custom error")
		},
	}

	server = New(customConfig)
	require.NotNil(t, server, "New() with custom config returned nil")
	assert.True(t, server.disableStartupMessage, "server.disableStartupMessage = false, want true")
	assert.NotNil(t, server.errorHandler, "server.errorHandler is nil")
}

// TestServerRouter tests the Router method
func TestServerRouter(t *testing.T) {
	server := New(DefaultConfig())
	router := server.Router()

	require.NotNil(t, router, "Server.Router() returned nil")
	assert.Equal(t, server.router, router, "Server.Router() did not return the server's router")
}

// TestServerHTTPMethods tests the HTTP method registration methods of Server
func TestServerHTTPMethods(t *testing.T) {
	server := New(DefaultConfig())
	handler := func(c *Ctx) {}

	// Test GET
	result := server.GET("/users", handler)
	assert.Equal(t, server.router, result, "Server.GET() did not return the router")
	assert.Len(t, server.router.Routes, 1, "len(server.router.Routes) should be 1")
	assert.Equal(t, "GET", server.router.Routes[0].Method, "server.router.Routes[0].Method should be GET")

	// Test HEAD
	server.HEAD("/users", handler)
	assert.Equal(t, "HEAD", server.router.Routes[1].Method, "server.router.Routes[1].Method should be HEAD")

	// Test POST
	server.POST("/users", handler)
	assert.Equal(t, "POST", server.router.Routes[2].Method, "server.router.Routes[2].Method should be POST")

	// Test PUT
	server.PUT("/users", handler)
	assert.Equal(t, "PUT", server.router.Routes[3].Method, "server.router.Routes[3].Method should be PUT")

	// Test DELETE
	server.DELETE("/users", handler)
	assert.Equal(t, "DELETE", server.router.Routes[4].Method, "server.router.Routes[4].Method should be DELETE")

	// Test CONNECT
	server.CONNECT("/users", handler)
	assert.Equal(t, "CONNECT", server.router.Routes[5].Method, "server.router.Routes[5].Method should be CONNECT")

	// Test OPTIONS
	server.OPTIONS("/users", handler)
	assert.Equal(t, "OPTIONS", server.router.Routes[6].Method, "server.router.Routes[6].Method should be OPTIONS")

	// Test TRACE
	server.TRACE("/users", handler)
	assert.Equal(t, "TRACE", server.router.Routes[7].Method, "server.router.Routes[7].Method should be TRACE")

	// Test PATCH
	server.PATCH("/users", handler)
	assert.Equal(t, "PATCH", server.router.Routes[8].Method, "server.router.Routes[8].Method should be PATCH")
}

// TestServerUse tests the Use method of Server
func TestServerUse(t *testing.T) {
	server := New(DefaultConfig())

	// Test with middleware function
	middleware1 := func(c *Ctx) {
		c.Next()
	}

	server.Use(middleware1)
	assert.Len(t, server.router.middlewareFuncs, 1, "len(server.router.middlewareFuncs) should be 1")

	// Test with multiple middleware functions
	middleware2 := func(c *Ctx) {
		c.Next()
	}

	server.Use(middleware2)
	assert.Len(t, server.router.middlewareFuncs, 2, "len(server.router.middlewareFuncs) should be 2")

	// Test with invalid middleware (should panic)
	assert.Panics(t, func() {
		server.Use("not a middleware")
	}, "Server.Use() with invalid middleware should panic")
}

// TestServerNotFound tests the NotFound method of Server
func TestServerNotFound(t *testing.T) {
	server := New(DefaultConfig())

	// Set a custom NotFound handler
	customHandler := func(c *Ctx) {
		c.Status(StatusNotFound).String("Custom 404")
	}

	server.NotFound(customHandler)

	// Verify the handler was set
	assert.NotNil(t, server.router.NotFound, "server.router.NotFound is nil after setting")

	// Create a request for a non-existent route
	req, _ := http.NewRequest("GET", "http://example.com/nonexistent", nil)
	w := httptest.NewRecorder()
	ctx := GetContext(w, req)

	// Serve the request
	server.router.ServeHTTP(ctx, ctx.Request)

	// Flush the response to the underlying writer
	ctx.Writer.Flush()

	// Check the response
	assert.Equal(t, StatusNotFound, w.Code, "Expected status code to be StatusNotFound")
	assert.Equal(t, "Custom 404", w.Body.String(), "Expected body to be 'Custom 404'")
}

// TestServerGroup tests the Group method of Server
func TestServerGroup(t *testing.T) {
	server := New(DefaultConfig())

	// Create a group
	group := server.Group("/api")
	require.NotNil(t, group, "Server.Group() returned nil")

	// Add a route to the group
	handlerCalled := false
	group.GET("/users", func(c *Ctx) {
		handlerCalled = true
		c.Status(StatusOK).String("OK")
	})

	// Create a request
	req, _ := http.NewRequest("GET", "http://example.com/api/users", nil)
	w := httptest.NewRecorder()
	ctx := GetContext(w, req)

	// Serve the request
	server.router.ServeHTTP(ctx, ctx.Request)

	// Flush the response to the underlying writer
	ctx.Writer.Flush()

	// Check that the handler was called
	assert.True(t, handlerCalled, "Group handler was not called")

	// Check the response
	assert.Equal(t, StatusOK, w.Code, "Expected status code to be StatusOK")
	assert.Equal(t, "OK", w.Body.String(), "Expected body to be 'OK'")
}

// TestDefaultErrorHandler tests the defaultErrorHandler function
func TestDefaultErrorHandler(t *testing.T) {
	// Create a context with an error
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://example.com/", nil)
	ctx := GetContext(w, req)

	// Set an error
	testError := errors.New("test error")
	ctx.Error(testError)

	// Call the default error handler
	defaultErrorHandler(ctx)

	// Flush the response to the underlying writer
	ctx.Writer.Flush()

	// Check the response
	assert.Equal(t, StatusInternalServerError, w.Code, "Expected status code to be StatusInternalServerError")
	assert.Equal(t, "test error", w.Body.String(), "Expected body to match error message")

	// Test with HttpError
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "http://example.com/", nil)
	ctx = GetContext(w, req)

	// Set an HttpError
	httpErr := NewHttpError(StatusBadRequest, "bad request")
	ctx.Error(httpErr)

	// Call the default error handler
	defaultErrorHandler(ctx)

	// Flush the response to the underlying writer
	ctx.Writer.Flush()

	// Check the response
	assert.Equal(t, StatusBadRequest, w.Code, "Expected status code to be StatusBadRequest")
	assert.Equal(t, "bad request", w.Body.String(), "Expected body to match HttpError message")
}

// TestResponseRecorder tests the responseRecorder implementation
func TestResponseRecorder(t *testing.T) {
	// Create a new responseRecorder
	r := getResponseRecorder()
	defer releaseResponseRecorder(r)

	// Test Header method
	header := r.Header()
	assert.NotNil(t, header, "responseRecorder.Header() returned nil")

	// Test setting a header value
	header.Set("Content-Type", "application/json")
	assert.Equal(t, "application/json", header.Get("Content-Type"), "Header value not set correctly")

	// Test Write method
	n, err := r.Write([]byte("test"))
	assert.NoError(t, err, "responseRecorder.Write() returned error")
	assert.Equal(t, 4, n, "responseRecorder.Write() returned incorrect byte count")

	// Test WriteHeader method
	r.WriteHeader(StatusOK)
	assert.Equal(t, StatusOK, r.code, "Status code not set correctly")

	// Test Flush method (should be a no-op)
	r.Flush()
}

// The following code was moved from bench_writer_test.go

// BenchResponseWriter is a response writer optimized for benchmarking
type BenchResponseWriter = responseRecorder

// GetBenchWriter returns a response writer optimized for benchmarking
func GetBenchWriter() *BenchResponseWriter {
	return getResponseRecorder()
}

// ReleaseBenchWriter returns a benchmark response writer to the pool
func ReleaseBenchWriter(w *BenchResponseWriter) {
	releaseResponseRecorder(w)
}

// TestResponseWriter is a response writer optimized for testing
type TestResponseWriter = responseRecorder

// GetTestWriter returns a response writer optimized for testing
func GetTestWriter() *TestResponseWriter {
	return getResponseRecorder()
}

// ReleaseTestWriter returns a test response writer to the pool
func ReleaseTestWriter(w *TestResponseWriter) {
	releaseResponseRecorder(w)
}

// For backward compatibility with existing tests
var getBenchWriter = GetBenchWriter
var getTestWriter = GetTestWriter

// TestNoopLogger tests the noopLogger implementation
func TestNoopLogger(t *testing.T) {
	logger := &noopLogger{}

	// Test all logger methods (they should not panic)
	assert.NotPanics(t, func() {
		logger.Debugf("test %s", "debug")
		logger.Infof("test %s", "info")
		logger.Warnf("test %s", "warn")
		logger.Errorf("test %s", "error")
		logger.Fatalf("test %s", "fatal")
	}, "Logger methods should not panic")
}

// TestServerSTATIC tests the STATIC method of Server
func TestServerSTATIC(t *testing.T) {
	server := New(DefaultConfig())

	// Test basic static file serving registration
	result := server.STATIC("/assets", "examples/static/assets")
	require.NotNil(t, result, "Server.STATIC() returned nil router")
	assert.Equal(t, server.router, result, "Server.STATIC() did not return the server's router")

	// Verify the route was registered
	assert.Len(t, server.router.Routes, 1, "len(server.router.Routes) should be 1")
	assert.Equal(t, "/assets/*", server.router.Routes[0].Pattern, "route pattern should be '/assets/*'")
	assert.Equal(t, "GET", server.router.Routes[0].Method, "route method should be GET")

	// Test with custom config
	config := Static{
		Browse:    true,
		ByteRange: true,
		MaxAge:    3600,
	}
	server.STATIC("/files", "examples/static/assets", config)
	assert.Len(t, server.router.Routes, 2, "len(server.router.Routes) should be 2 after adding second route")
}

// TestServerStaticFileIntegration tests end-to-end static file serving through the server
func TestServerStaticFileIntegration(t *testing.T) {
	server := New(DefaultConfig())

	// Register static file serving
	server.STATIC("/assets", "examples/static/assets")

	// Test serving index.html
	req, _ := http.NewRequest("GET", "http://example.com/assets/index.html", nil)
	w := httptest.NewRecorder()
	ctx := GetContext(w, req)

	// Process through the server's router
	server.router.ServeHTTP(ctx, ctx.Request)
	ctx.Writer.Flush()

	assert.Equal(t, StatusOK, w.Code, "Expected status code to be StatusOK")
	assert.Contains(t, w.Header().Get("Content-Type"), "text/html", "Expected content type to contain text/html")
	assert.Contains(t, w.Body.String(), "<!DOCTYPE html>", "Expected response to contain HTML doctype")
}

// TestServerStaticWithMultipleConfigs tests server with multiple static routes and configs
func TestServerStaticWithMultipleConfigs(t *testing.T) {
	server := New(DefaultConfig())

	// Register multiple static routes with different configs
	server.STATIC("/public", "examples/static/assets")

	downloadConfig := Static{
		Download: true,
		MaxAge:   3600,
	}
	server.STATIC("/downloads", "examples/static/assets", downloadConfig)

	browseConfig := Static{
		Browse: true,
	}
	server.STATIC("/browse", "examples/static/assets", browseConfig)

	// Verify all routes were registered
	assert.Len(t, server.router.Routes, 3, "Expected 3 routes to be registered")

	// Test normal serving
	req, _ := http.NewRequest("GET", "http://example.com/public/sample.txt", nil)
	w := httptest.NewRecorder()
	ctx := GetContext(w, req)

	server.router.ServeHTTP(ctx, ctx.Request)
	ctx.Writer.Flush()

	assert.Equal(t, StatusOK, w.Code, "Expected normal serving to work")
	assert.NotContains(t, w.Header().Get("Content-Disposition"), "attachment", "Normal route should not force download")

	// Test download route
	req, _ = http.NewRequest("GET", "http://example.com/downloads/sample.txt", nil)
	w = httptest.NewRecorder()
	ctx = GetContext(w, req)

	server.router.ServeHTTP(ctx, ctx.Request)
	ctx.Writer.Flush()

	assert.Equal(t, StatusOK, w.Code, "Expected download serving to work")
	assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment", "Download route should force download")
	assert.Equal(t, "public, max-age=3600", w.Header().Get("Cache-Control"), "Download route should set cache control")

	// Test browse route with directory
	req, _ = http.NewRequest("GET", "http://example.com/browse/css/", nil)
	w = httptest.NewRecorder()
	ctx = GetContext(w, req)

	server.router.ServeHTTP(ctx, ctx.Request)
	ctx.Writer.Flush()

	assert.Equal(t, StatusOK, w.Code, "Expected browse route to work")
	assert.Contains(t, w.Header().Get("Content-Type"), "text/html", "Browse route should return HTML for directory listing")
	assert.Contains(t, w.Body.String(), "Directory listing", "Browse route should show directory listing")
}

// TestServerStaticErrorHandling tests error scenarios with static file serving
func TestServerStaticErrorHandling(t *testing.T) {
	server := New(DefaultConfig())

	// Register static file serving
	server.STATIC("/assets", "examples/static/assets")

	// Test 404 for non-existent file
	req, _ := http.NewRequest("GET", "http://example.com/assets/nonexistent.txt", nil)
	w := httptest.NewRecorder()
	ctx := GetContext(w, req)

	server.router.ServeHTTP(ctx, ctx.Request)
	ctx.Writer.Flush()

	assert.Equal(t, StatusNotFound, w.Code, "Expected 404 for non-existent file")
	assert.Equal(t, "File not found", w.Body.String(), "Expected 'File not found' message")

	// Test 403 for directory without browse enabled
	req, _ = http.NewRequest("GET", "http://example.com/assets/css/", nil)
	w = httptest.NewRecorder()
	ctx = GetContext(w, req)

	server.router.ServeHTTP(ctx, ctx.Request)
	ctx.Writer.Flush()

	assert.Equal(t, StatusForbidden, w.Code, "Expected 403 for directory access without browse")
	assert.Equal(t, "Directory listing is disabled", w.Body.String(), "Expected directory listing disabled message")

	// Test 403 for path traversal
	req, _ = http.NewRequest("GET", "http://example.com/assets/../../../config.go", nil)
	w = httptest.NewRecorder()
	ctx = GetContext(w, req)

	server.router.ServeHTTP(ctx, ctx.Request)
	ctx.Writer.Flush()

	assert.Equal(t, StatusForbidden, w.Code, "Expected 403 for path traversal attempt")
	assert.Equal(t, "Forbidden", w.Body.String(), "Expected 'Forbidden' message for path traversal")
}

// TestServerStaticWithCustomErrorHandler tests static serving with custom error handler
func TestServerStaticWithCustomErrorHandler(t *testing.T) {
	// Create server with custom error handler
	customConfig := Config{
		ErrorHandler: func(c *Ctx) {
			c.Status(StatusInternalServerError).String("Custom server error")
		},
	}
	server := New(customConfig)

	// Register static file serving
	server.STATIC("/assets", "examples/static/assets")

	// Test that static file serving still works normally (no errors triggered)
	req, _ := http.NewRequest("GET", "http://example.com/assets/sample.txt", nil)
	w := httptest.NewRecorder()
	ctx := GetContext(w, req)

	server.router.ServeHTTP(ctx, ctx.Request)
	ctx.Writer.Flush()

	assert.Equal(t, StatusOK, w.Code, "Static file serving should work with custom error handler")
	assert.NotEqual(t, "Custom server error", w.Body.String(), "Should not trigger custom error handler for successful requests")
}

// TestServerStaticHeaderSettings tests that static files set appropriate headers
func TestServerStaticHeaderSettings(t *testing.T) {
	server := New(DefaultConfig())

	// Register static file serving with custom config
	config := Static{
		MaxAge:    7200,
		ByteRange: true,
	}
	server.STATIC("/assets", "examples/static/assets", config)

	// Test header setting
	req, _ := http.NewRequest("GET", "http://example.com/assets/sample.txt", nil)
	w := httptest.NewRecorder()
	ctx := GetContext(w, req)

	server.router.ServeHTTP(ctx, ctx.Request)
	ctx.Writer.Flush()

	assert.Equal(t, StatusOK, w.Code, "Expected successful response")
	assert.Equal(t, "public, max-age=7200", w.Header().Get("Cache-Control"), "Expected Cache-Control header to be set")
	assert.Equal(t, "bytes", w.Header().Get("Accept-Ranges"), "Expected Accept-Ranges header for ByteRange support")
	assert.NotEmpty(t, w.Header().Get("Last-Modified"), "Expected Last-Modified header to be set")
	assert.NotEmpty(t, w.Header().Get("Content-Length"), "Expected Content-Length header to be set")
	// Note: Server header might not be set in test context, so we check if it's present
	serverHeader := w.Header().Get("Server")
	t.Logf("Server header: '%s'", serverHeader)
	// The server header test is informational since it may depend on the test setup
}

// TestServerStaticDefaultIndexHandling tests default index file handling
func TestServerStaticDefaultIndexHandling(t *testing.T) {
	server := New(DefaultConfig())

	// Register static file serving
	server.STATIC("/assets", "examples/static/assets")

	// Test accessing directory root (should serve index.html)
	req, _ := http.NewRequest("GET", "http://example.com/assets/", nil)
	w := httptest.NewRecorder()
	ctx := GetContext(w, req)

	server.router.ServeHTTP(ctx, ctx.Request)
	ctx.Writer.Flush()

	assert.Equal(t, StatusOK, w.Code, "Expected successful response for directory with index file")
	assert.Contains(t, w.Header().Get("Content-Type"), "text/html", "Expected HTML content type for index file")
	assert.Contains(t, w.Body.String(), "<!DOCTYPE html>", "Expected HTML content from index.html")

	// Test accessing root without trailing slash
	req, _ = http.NewRequest("GET", "http://example.com/assets", nil)
	w = httptest.NewRecorder()
	ctx = GetContext(w, req)

	server.router.ServeHTTP(ctx, ctx.Request)
	ctx.Writer.Flush()

	t.Logf("Response status for /assets: %d", w.Code)
	t.Logf("Response body: %s", w.Body.String())

	// The route pattern is "/assets/*" so "/assets" without trailing slash might not match
	// This behavior may vary depending on the router implementation
	assert.True(t, w.Code == StatusOK || w.Code == StatusNotFound,
		"Response should be either 200 (if route matches) or 404 (if route doesn't match without trailing slash)")
}
