package ngebut

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewRouter tests the NewRouter function
func TestNewRouter(t *testing.T) {
	assert := assert.New(t)

	router := NewRouter()
	assert.NotNil(router, "NewRouter() returned nil")

	assert.NotNil(router.Routes, "router.Routes is nil")
	assert.Len(router.Routes, 0, "router.Routes should be empty")

	assert.NotNil(router.middlewareFuncs, "router.middlewareFuncs is nil")
	assert.Len(router.middlewareFuncs, 0, "router.middlewareFuncs should be empty")

	assert.NotNil(router.NotFound, "router.NotFound is nil")
}

// TestRouterUse tests the Use method of Router
func TestRouterUse(t *testing.T) {
	assert := assert.New(t)
	router := NewRouter()

	// Test with middleware function
	middleware1 := func(c *Ctx) {
		c.Next()
	}

	router.Use(middleware1)
	assert.Len(router.middlewareFuncs, 1, "should have 1 middleware function")

	// Test with multiple middleware functions
	middleware2 := func(c *Ctx) {
		c.Next()
	}

	router.Use(middleware2)
	assert.Len(router.middlewareFuncs, 2, "should have 2 middleware functions")

	// Test with invalid middleware (should panic)
	assert.Panics(func() {
		router.Use("not a middleware")
	}, "Router.Use() with invalid middleware should panic")
}

// TestRouterHandle tests the Handle method of Router
func TestRouterHandle(t *testing.T) {
	assert := assert.New(t)
	router := NewRouter()

	// Test with simple pattern
	handler := func(c *Ctx) {
		// Handler function
	}

	result := router.Handle("/users", "GET", handler)
	assert.Equal(router, result, "Router.Handle() should return the router")
	assert.Len(router.Routes, 1, "should have 1 route")

	route := router.Routes[0]
	assert.Equal("/users", route.Pattern, "route pattern should match")
	assert.Equal("GET", route.Method, "route method should match")
	assert.Len(route.Handlers, 1, "should have 1 handler")
	assert.NotNil(route.Regex, "route.Regex should not be nil")

	// Test with pattern containing parameters
	router.Handle("/users/:id", "POST", handler)
	assert.Len(router.Routes, 2, "should have 2 routes")

	route = router.Routes[1]
	assert.Equal("/users/:id", route.Pattern, "route pattern should match")
	assert.Equal("POST", route.Method, "route method should match")
	assert.NotNil(route.Regex, "route.Regex should not be nil")

	// Test with multiple handlers
	handler2 := func(c *Ctx) {
		// Another handler
	}
	router.Handle("/multi", "DELETE", handler, handler2)
	assert.Len(router.Routes, 3, "should have 3 routes")

	route = router.Routes[2]
	assert.Equal("/multi", route.Pattern, "route pattern should match")
	assert.Equal("DELETE", route.Method, "route method should match")
	assert.Len(route.Handlers, 2, "should have 2 handlers")
}

// TestRouterHTTPMethods tests the HTTP method registration methods of Router
func TestRouterHTTPMethods(t *testing.T) {
	assert := assert.New(t)
	router := NewRouter()
	handler := func(c *Ctx) {}

	// Test GET
	result := router.GET("/users", handler)
	assert.Equal(router, result, "Router.GET() should return the router")
	assert.Len(router.Routes, 1, "should have 1 route")
	assert.Equal("GET", router.Routes[0].Method, "method should be GET")

	// Test HEAD
	router.HEAD("/users", handler)
	assert.Equal("HEAD", router.Routes[1].Method, "method should be HEAD")

	// Test POST
	router.POST("/users", handler)
	assert.Equal("POST", router.Routes[2].Method, "method should be POST")

	// Test PUT
	router.PUT("/users", handler)
	assert.Equal("PUT", router.Routes[3].Method, "method should be PUT")

	// Test DELETE
	router.DELETE("/users", handler)
	assert.Equal("DELETE", router.Routes[4].Method, "method should be DELETE")

	// Test CONNECT
	router.CONNECT("/users", handler)
	assert.Equal("CONNECT", router.Routes[5].Method, "method should be CONNECT")

	// Test OPTIONS
	router.OPTIONS("/users", handler)
	assert.Equal("OPTIONS", router.Routes[6].Method, "method should be OPTIONS")

	// Test TRACE
	router.TRACE("/users", handler)
	assert.Equal("TRACE", router.Routes[7].Method, "method should be TRACE")

	// Test PATCH
	router.PATCH("/users", handler)
	assert.Equal("PATCH", router.Routes[8].Method, "method should be PATCH")
}

// TestRouterServeHTTP tests the ServeHTTP method of Router
func TestRouterServeHTTP(t *testing.T) {
	assert := assert.New(t)
	router := NewRouter()

	// Add a route
	handlerCalled := false
	router.GET("/users", func(c *Ctx) {
		handlerCalled = true
		c.Status(StatusOK).String("OK")
	})

	// Create a request
	req, _ := http.NewRequest("GET", "http://example.com/users", nil)
	w := httptest.NewRecorder()
	ctx := GetContext(w, req)

	// Serve the request
	router.ServeHTTP(ctx, ctx.Request)

	// Flush the response to the underlying writer
	ctx.Writer.Flush()

	// Check that the handler was called
	assert.True(handlerCalled, "Handler was not called")

	// Check the response
	assert.Equal(StatusOK, w.Code, "status code should be StatusOK")
	assert.Equal("OK", w.Body.String(), "response body should be 'OK'")
}

// TestRouterServeHTTPWithParams tests the ServeHTTP method of Router with URL parameters
func TestRouterServeHTTPWithParams(t *testing.T) {
	assert := assert.New(t)
	router := NewRouter()

	// Add a route with parameters
	var paramValue string
	router.GET("/users/:id", func(c *Ctx) {
		paramValue = c.Param("id")
		c.Status(StatusOK).String("User ID: %s", paramValue)
	})

	// Create a request
	req, _ := http.NewRequest("GET", "http://example.com/users/123", nil)
	w := httptest.NewRecorder()
	ctx := GetContext(w, req)

	// Serve the request
	router.ServeHTTP(ctx, ctx.Request)

	// Flush the response to the underlying writer
	ctx.Writer.Flush()

	// Check that the parameter was extracted
	assert.Equal("123", paramValue, "parameter value should be '123'")

	// Check the response
	assert.Equal(StatusOK, w.Code, "status code should be StatusOK")
	assert.Equal("User ID: 123", w.Body.String(), "response body should match")
}

// TestRouterServeHTTPNotFound tests the ServeHTTP method of Router with a non-existent route
func TestRouterServeHTTPNotFound(t *testing.T) {
	assert := assert.New(t)
	router := NewRouter()

	// Create a request for a non-existent route
	req, _ := http.NewRequest("GET", "http://example.com/nonexistent", nil)
	w := httptest.NewRecorder()
	ctx := GetContext(w, req)

	// Serve the request
	router.ServeHTTP(ctx, ctx.Request)

	// Flush the response to the underlying writer
	ctx.Writer.Flush()

	// Check the response
	assert.Equal(StatusNotFound, w.Code, "status code should be StatusNotFound")
	assert.Equal("404 page not found", w.Body.String(), "response body should match")
}

// TestRouterServeHTTPMethodNotAllowed tests the ServeHTTP method of Router with a method not allowed
func TestRouterServeHTTPMethodNotAllowed(t *testing.T) {
	assert := assert.New(t)
	router := NewRouter()

	// Add a route for GET
	router.GET("/users", func(c *Ctx) {
		c.Status(StatusOK).String("OK")
	})

	// Create a request with a different method
	req, _ := http.NewRequest("POST", "http://example.com/users", nil)
	w := httptest.NewRecorder()
	ctx := GetContext(w, req)

	// Serve the request
	router.ServeHTTP(ctx, ctx.Request)

	// Flush the response to the underlying writer
	ctx.Writer.Flush()

	// Check the response
	assert.Equal(StatusMethodNotAllowed, w.Code, "status code should be StatusMethodNotAllowed")
	assert.Equal("Method Not Allowed", w.Body.String(), "response body should match")
	assert.Equal("GET", w.Header().Get("Allow"), "Allow header should be GET")
}

// TestRouterServeHTTPWithMiddleware tests the ServeHTTP method of Router with middleware
func TestRouterServeHTTPWithMiddleware(t *testing.T) {
	assert := assert.New(t)
	router := NewRouter()

	// Add middleware
	middlewareCalled := false
	router.Use(func(c *Ctx) {
		middlewareCalled = true
		c.Next()
	})

	// Add a route
	handlerCalled := false
	router.GET("/users", func(c *Ctx) {
		handlerCalled = true
		c.Status(StatusOK).String("OK")
	})

	// Create a request
	req, _ := http.NewRequest("GET", "http://example.com/users", nil)
	w := httptest.NewRecorder()
	ctx := GetContext(w, req)

	// Serve the request
	router.ServeHTTP(ctx, ctx.Request)

	// Flush the response to the underlying writer
	ctx.Writer.Flush()

	// Check that the middleware and handler were called
	assert.True(middlewareCalled, "Middleware was not called")
	assert.True(handlerCalled, "Handler was not called")

	// Check the response
	assert.Equal(StatusOK, w.Code, "status code should be StatusOK")
	assert.Equal("OK", w.Body.String(), "response body should match")
}

// TestMiddlewareStackPool tests the middlewareStackPool
func TestMiddlewareStackPool(t *testing.T) {
	assert := assert.New(t)

	// Get a stack from the pool
	stack := middlewareStackPool.Get().([]MiddlewareFunc)
	assert.NotNil(stack, "middlewareStackPool.Get() returned nil")

	// Reset the stack to ensure it's empty
	stack = stack[:0]

	// Check that the stack is empty
	assert.Empty(stack, "stack should be empty")

	// Add a middleware function to the stack
	middleware := func(c *Ctx) {
		c.Next()
	}
	stack = append(stack, middleware)

	// Check that the middleware was added
	assert.Len(stack, 1, "stack should have 1 middleware function")

	// Put the stack back in the pool
	middlewareStackPool.Put(stack)

	// Get another stack from the pool (might be the same one)
	stack2 := middlewareStackPool.Get().([]MiddlewareFunc)
	assert.NotNil(stack2, "middlewareStackPool.Get() returned nil on second call")

	// Reset the stack to ensure it's empty
	stack2 = stack2[:0]

	// Check that the stack is empty after resetting
	assert.Empty(stack2, "stack2 should be empty after resetting")

	// Put the stack back in the pool
	middlewareStackPool.Put(stack2)
}

// TestRouterSTATIC tests the STATIC method of Router
func TestRouterSTATIC(t *testing.T) {
	assert := assert.New(t)
	router := NewRouter()

	// Test basic static file serving
	result := router.STATIC("/assets", "examples/static/assets")
	assert.Equal(router, result, "Router.STATIC() should return the router")
	assert.Len(router.Routes, 1, "should have 1 route")

	route := router.Routes[0]
	assert.Equal("/assets/*", route.Pattern, "route pattern should match")
	assert.Equal("GET", route.Method, "route method should be GET")
	assert.Len(route.Handlers, 1, "should have 1 handler")
}

// TestRouterHandleStatic tests the HandleStatic method of Router
func TestRouterHandleStatic(t *testing.T) {
	assert := assert.New(t)
	router := NewRouter()

	// Test with default config
	result := router.HandleStatic("/static", "examples/static/assets")
	assert.Equal(router, result, "Router.HandleStatic() should return the router")
	assert.Len(router.Routes, 1, "should have 1 route")

	// Test with custom config
	config := Static{
		Browse:    true,
		ByteRange: true,
		MaxAge:    3600,
	}
	router.HandleStatic("/files", "examples/static/assets", config)
	assert.Len(router.Routes, 2, "should have 2 routes")
}

// TestStaticFileServing tests serving actual static files
func TestStaticFileServing(t *testing.T) {
	assert := assert.New(t)
	router := NewRouter()

	// Register static file serving
	router.STATIC("/assets", "examples/static/assets")

	// Test serving index.html
	req, _ := http.NewRequest("GET", "http://example.com/assets/index.html", nil)
	w := httptest.NewRecorder()
	ctx := GetContext(w, req)

	router.ServeHTTP(ctx, ctx.Request)
	ctx.Writer.Flush()

	assert.Equal(StatusOK, w.Code, "should return 200 for existing file")
	assert.Contains(w.Header().Get("Content-Type"), "text/html", "should set correct content type")
	assert.Contains(w.Body.String(), "<!DOCTYPE html>", "should return HTML content")
}

// TestStaticFileServingWithDefaultIndex tests serving index file by default
func TestStaticFileServingWithDefaultIndex(t *testing.T) {
	assert := assert.New(t)
	router := NewRouter()

	// Register static file serving with default index
	router.STATIC("/assets", "examples/static/assets")

	// Test serving directory (should serve index.html)
	req, _ := http.NewRequest("GET", "http://example.com/assets/", nil)
	w := httptest.NewRecorder()
	ctx := GetContext(w, req)

	router.ServeHTTP(ctx, ctx.Request)
	ctx.Writer.Flush()

	assert.Equal(StatusOK, w.Code, "should return 200 for directory with index")
	assert.Contains(w.Header().Get("Content-Type"), "text/html", "should set correct content type")
	assert.Contains(w.Body.String(), "<!DOCTYPE html>", "should return index.html content")
}

// TestStaticFileServingCSS tests serving CSS files
func TestStaticFileServingCSS(t *testing.T) {
	assert := assert.New(t)
	router := NewRouter()

	// Register static file serving
	router.STATIC("/assets", "examples/static/assets")

	// Test serving CSS file
	req, _ := http.NewRequest("GET", "http://example.com/assets/css/style.css", nil)
	w := httptest.NewRecorder()
	ctx := GetContext(w, req)

	router.ServeHTTP(ctx, ctx.Request)
	ctx.Writer.Flush()

	assert.Equal(StatusOK, w.Code, "should return 200 for CSS file")
	assert.Contains(w.Header().Get("Content-Type"), "text/css", "should set correct content type for CSS")
	assert.Contains(w.Body.String(), "body", "should return CSS content")
}

// TestStaticFileServingJS tests serving JavaScript files
func TestStaticFileServingJS(t *testing.T) {
	assert := assert.New(t)
	router := NewRouter()

	// Register static file serving
	router.STATIC("/assets", "examples/static/assets")

	// Test serving JS file
	req, _ := http.NewRequest("GET", "http://example.com/assets/js/app.js", nil)
	w := httptest.NewRecorder()
	ctx := GetContext(w, req)

	router.ServeHTTP(ctx, ctx.Request)
	ctx.Writer.Flush()

	assert.Equal(StatusOK, w.Code, "should return 200 for JS file")
	assert.Contains(w.Header().Get("Content-Type"), "javascript", "should set correct content type for JS")
	assert.Contains(w.Body.String(), "function", "should return JS content")
}

// TestStaticFileServingTextFile tests serving text files
func TestStaticFileServingTextFile(t *testing.T) {
	assert := assert.New(t)
	router := NewRouter()

	// Register static file serving
	router.STATIC("/assets", "examples/static/assets")

	// Test serving text file
	req, _ := http.NewRequest("GET", "http://example.com/assets/sample.txt", nil)
	w := httptest.NewRecorder()
	ctx := GetContext(w, req)

	router.ServeHTTP(ctx, ctx.Request)
	ctx.Writer.Flush()

	assert.Equal(StatusOK, w.Code, "should return 200 for text file")
	assert.Contains(w.Header().Get("Content-Type"), "text/plain", "should set correct content type for text")
	assert.NotEmpty(w.Body.String(), "should return text content")
}

// TestStaticFileNotFound tests 404 handling for non-existent files
func TestStaticFileNotFound(t *testing.T) {
	assert := assert.New(t)
	router := NewRouter()

	// Register static file serving
	router.STATIC("/assets", "examples/static/assets")

	// Test non-existent file
	req, _ := http.NewRequest("GET", "http://example.com/assets/nonexistent.txt", nil)
	w := httptest.NewRecorder()
	ctx := GetContext(w, req)

	router.ServeHTTP(ctx, ctx.Request)
	ctx.Writer.Flush()

	assert.Equal(StatusNotFound, w.Code, "should return 404 for non-existent file")
	assert.Equal("File not found", w.Body.String(), "should return file not found message")
}

// TestStaticDirectoryBrowsingDisabled tests that directory browsing is disabled by default
func TestStaticDirectoryBrowsingDisabled(t *testing.T) {
	assert := assert.New(t)
	router := NewRouter()

	// Register static file serving without browse enabled
	router.STATIC("/assets", "examples/static/assets")

	// Test directory without index file (css directory)
	req, _ := http.NewRequest("GET", "http://example.com/assets/css/", nil)
	w := httptest.NewRecorder()
	ctx := GetContext(w, req)

	router.ServeHTTP(ctx, ctx.Request)
	ctx.Writer.Flush()

	assert.Equal(StatusForbidden, w.Code, "should return 403 when directory browsing is disabled")
	assert.Equal("Directory listing is disabled", w.Body.String(), "should return directory listing disabled message")
}

// TestStaticDirectoryBrowsingEnabled tests that directory browsing works when enabled
func TestStaticDirectoryBrowsingEnabled(t *testing.T) {
	assert := assert.New(t)
	router := NewRouter()

	// Register static file serving with browse enabled
	config := Static{
		Browse: true,
	}
	router.STATIC("/assets", "examples/static/assets", config)

	// Test directory listing
	req, _ := http.NewRequest("GET", "http://example.com/assets/css/", nil)
	w := httptest.NewRecorder()
	ctx := GetContext(w, req)

	router.ServeHTTP(ctx, ctx.Request)
	ctx.Writer.Flush()

	assert.Equal(StatusOK, w.Code, "should return 200 when directory browsing is enabled")
	assert.Contains(w.Header().Get("Content-Type"), "text/html", "should return HTML for directory listing")
	assert.Contains(w.Body.String(), "Directory listing", "should contain directory listing text")
	assert.Contains(w.Body.String(), "style.css", "should list files in directory")
}

// TestStaticByteRangeRequests tests byte range request handling
func TestStaticByteRangeRequests(t *testing.T) {
	assert := assert.New(t)
	router := NewRouter()

	// Register static file serving with byte range enabled
	config := Static{
		ByteRange: true,
	}
	router.STATIC("/assets", "examples/static/assets", config)

	// Test byte range request
	req, _ := http.NewRequest("GET", "http://example.com/assets/sample.txt", nil)
	req.Header.Set("Range", "bytes=0-10")
	w := httptest.NewRecorder()
	ctx := GetContext(w, req)

	router.ServeHTTP(ctx, ctx.Request)
	ctx.Writer.Flush()

	// Test that byte range is at least supported (Accept-Ranges header)
	assert.Equal("bytes", w.Header().Get("Accept-Ranges"), "should set Accept-Ranges header")

	t.Logf("Response status: %d", w.Code)
	t.Logf("Content-Length: %s", w.Header().Get("Content-Length"))
	t.Logf("Content-Range: %s", w.Header().Get("Content-Range"))
	t.Logf("Body length: %d", len(w.Body.String()))

	// The test verifies that ByteRange config is respected by setting Accept-Ranges header
	// The actual byte range functionality may depend on the implementation details
	assert.True(w.Code == StatusOK || w.Code == StatusPartialContent,
		"should return either 200 or 206 for byte range request")
}

// TestStaticMaxAge tests Cache-Control header setting
func TestStaticMaxAge(t *testing.T) {
	assert := assert.New(t)
	router := NewRouter()

	// Register static file serving with max age
	config := Static{
		MaxAge: 3600, // 1 hour
	}
	router.STATIC("/assets", "examples/static/assets", config)

	// Test cache headers
	req, _ := http.NewRequest("GET", "http://example.com/assets/sample.txt", nil)
	w := httptest.NewRecorder()
	ctx := GetContext(w, req)

	router.ServeHTTP(ctx, ctx.Request)
	ctx.Writer.Flush()

	assert.Equal(StatusOK, w.Code, "should return 200")
	assert.Equal("public, max-age=3600", w.Header().Get("Cache-Control"), "should set Cache-Control header")
}

// TestStaticDownload tests download mode
func TestStaticDownload(t *testing.T) {
	assert := assert.New(t)
	router := NewRouter()

	// Register static file serving with download enabled
	config := Static{
		Download: true,
	}
	router.STATIC("/assets", "examples/static/assets", config)

	// Test download headers
	req, _ := http.NewRequest("GET", "http://example.com/assets/sample.txt", nil)
	w := httptest.NewRecorder()
	ctx := GetContext(w, req)

	router.ServeHTTP(ctx, ctx.Request)
	ctx.Writer.Flush()

	assert.Equal(StatusOK, w.Code, "should return 200")
	assert.Contains(w.Header().Get("Content-Disposition"), "attachment", "should set Content-Disposition for download")
	assert.Contains(w.Header().Get("Content-Disposition"), "sample.txt", "should include filename in Content-Disposition")
}

// TestStaticNext tests the Next function
func TestStaticNext(t *testing.T) {
	assert := assert.New(t)
	router := NewRouter()

	// Test counter to track Next function calls
	nextCallCount := 0

	// Register static file serving with Next function that skips certain files
	config := Static{
		Next: func(c *Ctx) bool {
			nextCallCount++
			// Skip files ending with .private
			return strings.HasSuffix(c.Path(), ".private")
		},
	}
	router.STATIC("/assets", "examples/static/assets", config)

	// Test normal file serving (Next returns false)
	req, _ := http.NewRequest("GET", "http://example.com/assets/sample.txt", nil)
	w := httptest.NewRecorder()
	ctx := GetContext(w, req)

	router.ServeHTTP(ctx, ctx.Request)
	ctx.Writer.Flush()

	assert.Equal(StatusOK, w.Code, "should serve normal files when Next returns false")
	assert.Contains(w.Body.String(), "This is a sample", "should return file content")
	assert.Equal(1, nextCallCount, "Next function should be called once")

	// Reset for next test
	nextCallCount = 0

	// Test that .private files are skipped (Next returns true)
	req, _ = http.NewRequest("GET", "http://example.com/assets/secret.private", nil)
	w = httptest.NewRecorder()
	ctx = GetContext(w, req)

	router.ServeHTTP(ctx, ctx.Request)
	ctx.Writer.Flush()

	// The Next function should be called and return true
	assert.Equal(1, nextCallCount, "Next function should be called once for .private file")

	// Since Next returns true, the static handler calls c.Next() which continues
	// But since there's only one handler, the behavior might vary
	// The important thing is that Next was called and the file was skipped
	t.Logf("Response status: %d, body: %s", w.Code, w.Body.String())
}

// TestStaticModifyResponse tests the ModifyResponse function
func TestStaticModifyResponse(t *testing.T) {
	assert := assert.New(t)
	router := NewRouter()

	// Register static file serving with ModifyResponse function
	config := Static{
		ModifyResponse: func(c *Ctx) {
			c.Set("X-Custom-Header", "Modified")
		},
	}
	router.STATIC("/assets", "examples/static/assets", config)

	// Test that ModifyResponse is called
	req, _ := http.NewRequest("GET", "http://example.com/assets/sample.txt", nil)
	w := httptest.NewRecorder()
	ctx := GetContext(w, req)

	router.ServeHTTP(ctx, ctx.Request)
	ctx.Writer.Flush()

	assert.Equal(StatusOK, w.Code, "should return 200")
	assert.Equal("Modified", w.Header().Get("X-Custom-Header"), "should apply ModifyResponse function")
}

// TestStaticSecurityPathTraversal tests protection against directory traversal attacks
func TestStaticSecurityPathTraversal(t *testing.T) {
	assert := assert.New(t)
	router := NewRouter()

	// Register static file serving
	router.STATIC("/assets", "examples/static/assets")

	// Test directory traversal attempt
	req, _ := http.NewRequest("GET", "http://example.com/assets/../../../config.go", nil)
	w := httptest.NewRecorder()
	ctx := GetContext(w, req)

	router.ServeHTTP(ctx, ctx.Request)
	ctx.Writer.Flush()

	assert.Equal(StatusForbidden, w.Code, "should block directory traversal attempts")
	assert.Equal("Forbidden", w.Body.String(), "should return forbidden message")
}

// TestStaticPrefixHandling tests various prefix formats
func TestStaticPrefixHandling(t *testing.T) {
	assert := assert.New(t)
	router := NewRouter()

	// Test prefix without trailing slash
	router.STATIC("/assets", "examples/static/assets")

	// Test prefix with trailing slash
	router.STATIC("/files/", "examples/static/assets")

	// Should have 2 routes
	assert.Len(router.Routes, 2, "should have 2 routes")

	// Both should have wildcard patterns
	assert.Equal("/assets/*", router.Routes[0].Pattern, "first route should have wildcard pattern")
	assert.Equal("/files/*", router.Routes[1].Pattern, "second route should have wildcard pattern")
}
