package ngebut

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestNew tests the New function
func TestNew(t *testing.T) {
	// Test with default config
	server := New(DefaultConfig())
	if server == nil {
		t.Fatal("New() returned nil")
	}

	if server.router == nil {
		t.Error("server.router is nil")
	}

	if server.httpServer == nil {
		t.Error("server.httpServer is nil")
	}

	// Test with custom config
	customConfig := Config{
		DisableStartupMessage: true,
		ErrorHandler: func(c *Ctx) {
			c.Status(StatusInternalServerError).String("Custom error")
		},
	}

	server = New(customConfig)
	if server == nil {
		t.Fatal("New() with custom config returned nil")
	}

	if !server.disableStartupMessage {
		t.Error("server.disableStartupMessage = false, want true")
	}

	if server.errorHandler == nil {
		t.Error("server.errorHandler is nil")
	}
}

// TestServerRouter tests the Router method
func TestServerRouter(t *testing.T) {
	server := New(DefaultConfig())
	router := server.Router()

	if router == nil {
		t.Fatal("Server.Router() returned nil")
	}

	if router != server.router {
		t.Error("Server.Router() did not return the server's router")
	}
}

// TestServerHTTPMethods tests the HTTP method registration methods of Server
func TestServerHTTPMethods(t *testing.T) {
	server := New(DefaultConfig())
	handler := func(c *Ctx) {}

	// Test GET
	result := server.GET("/users", handler)
	if result != server.router {
		t.Error("Server.GET() did not return the router")
	}
	if len(server.router.Routes) != 1 {
		t.Errorf("len(server.router.Routes) = %d, want 1", len(server.router.Routes))
	}
	if server.router.Routes[0].Method != "GET" {
		t.Errorf("server.router.Routes[0].Method = %q, want %q", server.router.Routes[0].Method, "GET")
	}

	// Test HEAD
	server.HEAD("/users", handler)
	if server.router.Routes[1].Method != "HEAD" {
		t.Errorf("server.router.Routes[1].Method = %q, want %q", server.router.Routes[1].Method, "HEAD")
	}

	// Test POST
	server.POST("/users", handler)
	if server.router.Routes[2].Method != "POST" {
		t.Errorf("server.router.Routes[2].Method = %q, want %q", server.router.Routes[2].Method, "POST")
	}

	// Test PUT
	server.PUT("/users", handler)
	if server.router.Routes[3].Method != "PUT" {
		t.Errorf("server.router.Routes[3].Method = %q, want %q", server.router.Routes[3].Method, "PUT")
	}

	// Test DELETE
	server.DELETE("/users", handler)
	if server.router.Routes[4].Method != "DELETE" {
		t.Errorf("server.router.Routes[4].Method = %q, want %q", server.router.Routes[4].Method, "DELETE")
	}

	// Test CONNECT
	server.CONNECT("/users", handler)
	if server.router.Routes[5].Method != "CONNECT" {
		t.Errorf("server.router.Routes[5].Method = %q, want %q", server.router.Routes[5].Method, "CONNECT")
	}

	// Test OPTIONS
	server.OPTIONS("/users", handler)
	if server.router.Routes[6].Method != "OPTIONS" {
		t.Errorf("server.router.Routes[6].Method = %q, want %q", server.router.Routes[6].Method, "OPTIONS")
	}

	// Test TRACE
	server.TRACE("/users", handler)
	if server.router.Routes[7].Method != "TRACE" {
		t.Errorf("server.router.Routes[7].Method = %q, want %q", server.router.Routes[7].Method, "TRACE")
	}

	// Test PATCH
	server.PATCH("/users", handler)
	if server.router.Routes[8].Method != "PATCH" {
		t.Errorf("server.router.Routes[8].Method = %q, want %q", server.router.Routes[8].Method, "PATCH")
	}
}

// TestServerUse tests the Use method of Server
func TestServerUse(t *testing.T) {
	server := New(DefaultConfig())

	// Test with middleware function
	middleware1 := func(c *Ctx) {
		c.Next()
	}

	server.Use(middleware1)
	if len(server.router.middlewareFuncs) != 1 {
		t.Errorf("len(server.router.middlewareFuncs) = %d, want 1", len(server.router.middlewareFuncs))
	}

	// Test with multiple middleware functions
	middleware2 := func(c *Ctx) {
		c.Next()
	}

	server.Use(middleware2)
	if len(server.router.middlewareFuncs) != 2 {
		t.Errorf("len(server.router.middlewareFuncs) = %d, want 2", len(server.router.middlewareFuncs))
	}

	// Test with invalid middleware (should panic)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Server.Use() with invalid middleware did not panic")
		}
	}()
	server.Use("not a middleware")
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
	if server.router.NotFound == nil {
		t.Error("server.router.NotFound is nil after setting")
	}

	// Create a request for a non-existent route
	req, _ := http.NewRequest("GET", "http://example.com/nonexistent", nil)
	w := httptest.NewRecorder()
	ctx := GetContext(w, req)

	// Serve the request
	server.router.ServeHTTP(ctx, ctx.Request)

	// Flush the response to the underlying writer
	ctx.Writer.Flush()

	// Check the response
	if w.Code != StatusNotFound {
		t.Errorf("w.Code = %d, want %d", w.Code, StatusNotFound)
	}
	if w.Body.String() != "Custom 404" {
		t.Errorf("w.Body = %q, want %q", w.Body.String(), "Custom 404")
	}
}

// TestServerGroup tests the Group method of Server
func TestServerGroup(t *testing.T) {
	server := New(DefaultConfig())

	// Create a group
	group := server.Group("/api")

	if group == nil {
		t.Fatal("Server.Group() returned nil")
	}

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
	if !handlerCalled {
		t.Error("Group handler was not called")
	}

	// Check the response
	if w.Code != StatusOK {
		t.Errorf("w.Code = %d, want %d", w.Code, StatusOK)
	}
	if w.Body.String() != "OK" {
		t.Errorf("w.Body = %q, want %q", w.Body.String(), "OK")
	}
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
	if w.Code != StatusInternalServerError {
		t.Errorf("w.Code = %d, want %d", w.Code, StatusInternalServerError)
	}
	if w.Body.String() != "test error" {
		t.Errorf("w.Body = %q, want %q", w.Body.String(), "test error")
	}

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
	if w.Code != StatusBadRequest {
		t.Errorf("w.Code = %d, want %d", w.Code, StatusBadRequest)
	}
	if w.Body.String() != "bad request" {
		t.Errorf("w.Body = %q, want %q", w.Body.String(), "bad request")
	}
}

// TestEstimateResponseSize has been moved to internal/httpparser/httpparser_test.go

// TestDummyResponseWriter tests the dummyResponseWriter implementation
func TestDummyResponseWriter(t *testing.T) {
	// Create a new dummyResponseWriter
	d := &dummyResponseWriter{
		header: make(http.Header),
	}

	// Test Header method
	header := d.Header()
	if header == nil {
		t.Error("dummyResponseWriter.Header() returned nil")
	}

	// Test setting a header value
	header.Set("Content-Type", "application/json")
	if header.Get("Content-Type") != "application/json" {
		t.Errorf("header.Get(\"Content-Type\") = %q, want %q", header.Get("Content-Type"), "application/json")
	}

	// Test Write method
	n, err := d.Write([]byte("test"))
	if err != nil {
		t.Errorf("dummyResponseWriter.Write() returned error: %v", err)
	}
	if n != 4 {
		t.Errorf("dummyResponseWriter.Write() = %d, want %d", n, 4)
	}

	// Test WriteHeader method (should be a no-op)
	d.WriteHeader(StatusOK)

	// Test Flush method (should be a no-op)
	d.Flush()
}

// TestHttpCodec has been moved to internal/httpparser/httpparser_test.go

// TestHttpCodecGetContentLength has been moved to internal/httpparser/httpparser_test.go

// TestHttpCodecParse has been moved to internal/httpparser/httpparser_test.go

// TestNoopLogger tests the noopLogger implementation
func TestNoopLogger(t *testing.T) {
	logger := &noopLogger{}

	// Test all logger methods (they should not panic)
	logger.Debugf("test %s", "debug")
	logger.Infof("test %s", "info")
	logger.Warnf("test %s", "warn")
	logger.Errorf("test %s", "error")

	// Test Fatalf separately with defer/recover to catch potential panics
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("logger.Fatalf() panicked: %v", r)
		}
	}()
	logger.Fatalf("test %s", "fatal")
}

// TestParserResetAfterError has been moved to internal/httpparser/httpparser_test.go
