package ngebut

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestNewRouter tests the NewRouter function
func TestNewRouter(t *testing.T) {
	router := NewRouter()
	if router == nil {
		t.Fatal("NewRouter() returned nil")
	}

	if router.Routes == nil {
		t.Error("router.Routes is nil")
	}
	if len(router.Routes) != 0 {
		t.Errorf("len(router.Routes) = %d, want 0", len(router.Routes))
	}

	if router.middlewareFuncs == nil {
		t.Error("router.middlewareFuncs is nil")
	}
	if len(router.middlewareFuncs) != 0 {
		t.Errorf("len(router.middlewareFuncs) = %d, want 0", len(router.middlewareFuncs))
	}

	if router.NotFound == nil {
		t.Error("router.NotFound is nil")
	}
}

// TestRouterUse tests the Use method of Router
func TestRouterUse(t *testing.T) {
	router := NewRouter()

	// Test with middleware function
	middleware1 := func(c *Ctx) {
		c.Next()
	}

	router.Use(middleware1)
	if len(router.middlewareFuncs) != 1 {
		t.Errorf("len(router.middlewareFuncs) = %d, want 1", len(router.middlewareFuncs))
	}

	// Test with multiple middleware functions
	middleware2 := func(c *Ctx) {
		c.Next()
	}

	router.Use(middleware2)
	if len(router.middlewareFuncs) != 2 {
		t.Errorf("len(router.middlewareFuncs) = %d, want 2", len(router.middlewareFuncs))
	}

	// Test with invalid middleware (should panic)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Router.Use() with invalid middleware did not panic")
		}
	}()
	router.Use("not a middleware")
}

// TestRouterHandle tests the Handle method of Router
func TestRouterHandle(t *testing.T) {
	router := NewRouter()

	// Test with simple pattern
	handler := func(c *Ctx) {
		// Handler function
	}

	result := router.Handle("/users", "GET", handler)
	if result != router {
		t.Error("Router.Handle() did not return the router")
	}

	if len(router.Routes) != 1 {
		t.Errorf("len(router.Routes) = %d, want 1", len(router.Routes))
	}

	route := router.Routes[0]
	if route.Pattern != "/users" {
		t.Errorf("route.Pattern = %q, want %q", route.Pattern, "/users")
	}
	if route.Method != "GET" {
		t.Errorf("route.Method = %q, want %q", route.Method, "GET")
	}
	if len(route.Handlers) != 1 {
		t.Errorf("len(route.Handlers) = %d, want 1", len(route.Handlers))
	}
	if route.Regex == nil {
		t.Error("route.Regex is nil")
	}

	// Test with pattern containing parameters
	router.Handle("/users/:id", "POST", handler)
	if len(router.Routes) != 2 {
		t.Errorf("len(router.Routes) = %d, want 2", len(router.Routes))
	}

	route = router.Routes[1]
	if route.Pattern != "/users/:id" {
		t.Errorf("route.Pattern = %q, want %q", route.Pattern, "/users/:id")
	}
	if route.Method != "POST" {
		t.Errorf("route.Method = %q, want %q", route.Method, "POST")
	}
	if route.Regex == nil {
		t.Error("route.Regex is nil")
	}

	// Test with multiple handlers
	handler2 := func(c *Ctx) {
		// Another handler
	}
	router.Handle("/multi", "DELETE", handler, handler2)
	if len(router.Routes) != 3 {
		t.Errorf("len(router.Routes) = %d, want 3", len(router.Routes))
	}

	route = router.Routes[2]
	if route.Pattern != "/multi" {
		t.Errorf("route.Pattern = %q, want %q", route.Pattern, "/multi")
	}
	if route.Method != "DELETE" {
		t.Errorf("route.Method = %q, want %q", route.Method, "DELETE")
	}
	if len(route.Handlers) != 2 {
		t.Errorf("len(route.Handlers) = %d, want 2", len(route.Handlers))
	}
}

// TestRouterHTTPMethods tests the HTTP method registration methods of Router
func TestRouterHTTPMethods(t *testing.T) {
	router := NewRouter()
	handler := func(c *Ctx) {}

	// Test GET
	result := router.GET("/users", handler)
	if result != router {
		t.Error("Router.GET() did not return the router")
	}
	if len(router.Routes) != 1 {
		t.Errorf("len(router.Routes) = %d, want 1", len(router.Routes))
	}
	if router.Routes[0].Method != "GET" {
		t.Errorf("router.Routes[0].Method = %q, want %q", router.Routes[0].Method, "GET")
	}

	// Test HEAD
	router.HEAD("/users", handler)
	if router.Routes[1].Method != "HEAD" {
		t.Errorf("router.Routes[1].Method = %q, want %q", router.Routes[1].Method, "HEAD")
	}

	// Test POST
	router.POST("/users", handler)
	if router.Routes[2].Method != "POST" {
		t.Errorf("router.Routes[2].Method = %q, want %q", router.Routes[2].Method, "POST")
	}

	// Test PUT
	router.PUT("/users", handler)
	if router.Routes[3].Method != "PUT" {
		t.Errorf("router.Routes[3].Method = %q, want %q", router.Routes[3].Method, "PUT")
	}

	// Test DELETE
	router.DELETE("/users", handler)
	if router.Routes[4].Method != "DELETE" {
		t.Errorf("router.Routes[4].Method = %q, want %q", router.Routes[4].Method, "DELETE")
	}

	// Test CONNECT
	router.CONNECT("/users", handler)
	if router.Routes[5].Method != "CONNECT" {
		t.Errorf("router.Routes[5].Method = %q, want %q", router.Routes[5].Method, "CONNECT")
	}

	// Test OPTIONS
	router.OPTIONS("/users", handler)
	if router.Routes[6].Method != "OPTIONS" {
		t.Errorf("router.Routes[6].Method = %q, want %q", router.Routes[6].Method, "OPTIONS")
	}

	// Test TRACE
	router.TRACE("/users", handler)
	if router.Routes[7].Method != "TRACE" {
		t.Errorf("router.Routes[7].Method = %q, want %q", router.Routes[7].Method, "TRACE")
	}

	// Test PATCH
	router.PATCH("/users", handler)
	if router.Routes[8].Method != "PATCH" {
		t.Errorf("router.Routes[8].Method = %q, want %q", router.Routes[8].Method, "PATCH")
	}
}

// TestRouterServeHTTP tests the ServeHTTP method of Router
func TestRouterServeHTTP(t *testing.T) {
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
	if !handlerCalled {
		t.Error("Handler was not called")
	}

	// Check the response
	if w.Code != StatusOK {
		t.Errorf("w.Code = %d, want %d", w.Code, StatusOK)
	}
	if w.Body.String() != "OK" {
		t.Errorf("w.Body = %q, want %q", w.Body.String(), "OK")
	}
}

// TestRouterServeHTTPWithParams tests the ServeHTTP method of Router with URL parameters
func TestRouterServeHTTPWithParams(t *testing.T) {
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
	if paramValue != "123" {
		t.Errorf("paramValue = %q, want %q", paramValue, "123")
	}

	// Check the response
	if w.Code != StatusOK {
		t.Errorf("w.Code = %d, want %d", w.Code, StatusOK)
	}
	if w.Body.String() != "User ID: 123" {
		t.Errorf("w.Body = %q, want %q", w.Body.String(), "User ID: 123")
	}
}

// TestRouterServeHTTPNotFound tests the ServeHTTP method of Router with a non-existent route
func TestRouterServeHTTPNotFound(t *testing.T) {
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
	if w.Code != StatusNotFound {
		t.Errorf("w.Code = %d, want %d", w.Code, StatusNotFound)
	}
	if w.Body.String() != "404 page not found" {
		t.Errorf("w.Body = %q, want %q", w.Body.String(), "404 page not found")
	}
}

// TestRouterServeHTTPMethodNotAllowed tests the ServeHTTP method of Router with a method not allowed
func TestRouterServeHTTPMethodNotAllowed(t *testing.T) {
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
	if w.Code != StatusMethodNotAllowed {
		t.Errorf("w.Code = %d, want %d", w.Code, StatusMethodNotAllowed)
	}
	if w.Body.String() != "Method Not Allowed" {
		t.Errorf("w.Body = %q, want %q", w.Body.String(), "Method Not Allowed")
	}
	if w.Header().Get("Allow") != "GET" {
		t.Errorf("w.Header().Get(\"Allow\") = %q, want %q", w.Header().Get("Allow"), "GET")
	}
}

// TestRouterServeHTTPWithMiddleware tests the ServeHTTP method of Router with middleware
func TestRouterServeHTTPWithMiddleware(t *testing.T) {
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
	if !middlewareCalled {
		t.Error("Middleware was not called")
	}
	if !handlerCalled {
		t.Error("Handler was not called")
	}

	// Check the response
	if w.Code != StatusOK {
		t.Errorf("w.Code = %d, want %d", w.Code, StatusOK)
	}
	if w.Body.String() != "OK" {
		t.Errorf("w.Body = %q, want %q", w.Body.String(), "OK")
	}
}

// TestMiddlewareStackPool tests the middlewareStackPool
func TestMiddlewareStackPool(t *testing.T) {
	// Get a stack from the pool
	stack := middlewareStackPool.Get().([]MiddlewareFunc)
	if stack == nil {
		t.Fatal("middlewareStackPool.Get() returned nil")
	}

	// Reset the stack to ensure it's empty
	stack = stack[:0]

	// Check that the stack is empty
	if len(stack) != 0 {
		t.Errorf("len(stack) = %d, want 0", len(stack))
	}

	// Add a middleware function to the stack
	middleware := func(c *Ctx) {
		c.Next()
	}
	stack = append(stack, middleware)

	// Check that the middleware was added
	if len(stack) != 1 {
		t.Errorf("len(stack) = %d, want 1", len(stack))
	}

	// Put the stack back in the pool
	middlewareStackPool.Put(stack)

	// Get another stack from the pool (might be the same one)
	stack2 := middlewareStackPool.Get().([]MiddlewareFunc)
	if stack2 == nil {
		t.Fatal("middlewareStackPool.Get() returned nil on second call")
	}

	// Reset the stack to ensure it's empty
	stack2 = stack2[:0]

	// Check that the stack is empty after resetting
	if len(stack2) != 0 {
		t.Errorf("len(stack2) = %d, want 0", len(stack2))
	}

	// Put the stack back in the pool
	middlewareStackPool.Put(stack2)
}
