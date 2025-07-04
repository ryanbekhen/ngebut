package ngebut

import (
	"net/http/httptest"
	"testing"
)

func TestCompileMiddleware(t *testing.T) {
	// Create a test handler
	handler := func(c *Ctx) {
		c.String("Hello, World!")
	}

	// Create test middleware
	middleware1 := func(c *Ctx) {
		c.Set("X-Middleware-1", "true")
		// In compiled middleware, we don't call c.Next()
	}

	middleware2 := func(c *Ctx) {
		c.Set("X-Middleware-2", "true")
		// In compiled middleware, we don't call c.Next()
	}

	// Compile the middleware and handler
	compiledHandler := CompileMiddleware(handler, middleware1, middleware2)

	// Create a test request
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// Create a context
	ctx := GetContext(w, req)
	defer ReleaseContext(ctx)

	// Execute the compiled handler
	compiledHandler(ctx)

	// Check the response
	if w.Body.String() != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got '%s'", w.Body.String())
	}

	// Check that middleware was executed
	if ctx.Get("X-Middleware-1") != "true" {
		t.Errorf("Middleware 1 was not executed")
	}

	if ctx.Get("X-Middleware-2") != "true" {
		t.Errorf("Middleware 2 was not executed")
	}
}

func BenchmarkCompileMiddleware(b *testing.B) {
	// Create a test handler
	handler := func(c *Ctx) {
		c.String("Hello, World!")
	}

	// Create test middleware
	middleware1 := func(c *Ctx) {
		c.Set("X-Middleware-1", "true")
	}

	middleware2 := func(c *Ctx) {
		c.Set("X-Middleware-2", "true")
	}

	// Compile the middleware and handler
	compiledHandler := CompileMiddleware(handler, middleware1, middleware2)

	// Create a test request
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// Reset the timer
	b.ResetTimer()

	// Run the benchmark
	for i := 0; i < b.N; i++ {
		// Create a context
		ctx := GetContext(w, req)

		// Execute the compiled handler
		compiledHandler(ctx)

		// Release the context
		ReleaseContext(ctx)
	}
}

func BenchmarkDynamicMiddleware(b *testing.B) {
	// Create a test handler
	handler := func(c *Ctx) {
		c.String("Hello, World!")
	}

	// Create test middleware
	middleware1 := func(c *Ctx) {
		c.Set("X-Middleware-1", "true")
		c.Next()
	}

	middleware2 := func(c *Ctx) {
		c.Set("X-Middleware-2", "true")
		c.Next()
	}

	// Create a router
	router := NewRouter()

	// Add middleware
	router.Use(middleware1)
	router.Use(middleware2)

	// Reset the timer
	b.ResetTimer()

	// Run the benchmark
	for i := 0; i < b.N; i++ {
		// Create a test request
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		// Create a context
		ctx := GetContext(w, req)

		// Set up middleware
		ctx.middlewareStack = ctx.middlewareStack[:0]
		ctx.middlewareStack = append(ctx.middlewareStack, middleware1, middleware2)
		ctx.middlewareIndex = -1
		ctx.handler = handler

		// Execute middleware
		ctx.Next()

		// Release the context
		ReleaseContext(ctx)
	}
}
