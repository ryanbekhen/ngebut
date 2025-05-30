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

// TestDummyResponseWriter tests the dummyResponseWriter implementation
func TestDummyResponseWriter(t *testing.T) {
	// Create a new dummyResponseWriter
	d := &dummyResponseWriter{
		header: make(http.Header),
	}

	// Test Header method
	header := d.Header()
	assert.NotNil(t, header, "dummyResponseWriter.Header() returned nil")

	// Test setting a header value
	header.Set("Content-Type", "application/json")
	assert.Equal(t, "application/json", header.Get("Content-Type"), "Header value not set correctly")

	// Test Write method
	n, err := d.Write([]byte("test"))
	assert.NoError(t, err, "dummyResponseWriter.Write() returned error")
	assert.Equal(t, 4, n, "dummyResponseWriter.Write() returned incorrect byte count")

	// Test WriteHeader method (should be a no-op)
	d.WriteHeader(StatusOK)

	// Test Flush method (should be a no-op)
	d.Flush()
}

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
