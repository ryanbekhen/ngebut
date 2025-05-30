package ngebut

import (
	"net/http"
	"net/http/httptest"
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
