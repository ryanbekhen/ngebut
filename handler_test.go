package ngebut

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestHandlerType tests that a function with the correct signature can be assigned to a Handler variable
func TestHandlerType(t *testing.T) {
	// Define a function with the Handler signature
	handlerFunc := func(c *Ctx) {
		c.Status(StatusOK).String("Handler called")
	}

	// Assign the function to a Handler variable
	var handler Handler = handlerFunc

	// Verify the handler can be called
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	res := httptest.NewRecorder()
	ctx := GetContext(res, req)

	// Call the handler
	handler(ctx)

	// Flush the response to the underlying writer
	ctx.Writer.Flush()

	// Check the response
	if res.Code != StatusOK {
		t.Errorf("Expected status code %d, got %d", StatusOK, res.Code)
	}
	if res.Body.String() != "Handler called" {
		t.Errorf("Expected body 'Handler called', got '%s'", res.Body.String())
	}
}

// TestMiddlewareType tests that a function with the Middleware signature can be assigned to a Middleware variable
func TestMiddlewareType(t *testing.T) {
	// Define a function with the Middleware signature
	middlewareFunc := func(c *Ctx) {
		c.Set("middleware", "called")
		// In a real application, c.Next() would be called here
	}

	// Assign the function to a Middleware variable
	var middleware Middleware = middlewareFunc

	// Verify the middleware can be called
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	res := httptest.NewRecorder()
	ctx := GetContext(res, req)

	// Call the middleware
	middleware(ctx)

	// Check the middleware set the expected value
	if ctx.Get("middleware") != "called" {
		t.Errorf("Expected middleware to set 'middleware' to 'called', got '%s'", ctx.Get("middleware"))
	}
}

// TestMiddlewareFuncAlias tests that MiddlewareFunc is an alias for Middleware
func TestMiddlewareFuncAlias(t *testing.T) {
	// Define a function with the Middleware signature
	middlewareFunc := func(c *Ctx) {
		c.Set("middleware", "called")
		// In a real application, c.Next() would be called here
	}

	// Assign the function to both Middleware and MiddlewareFunc variables
	var middleware Middleware = middlewareFunc
	var middlewareFunc2 MiddlewareFunc = middlewareFunc

	// Verify both can be assigned the same function
	if &middleware == nil || &middlewareFunc2 == nil {
		t.Error("Failed to assign function to Middleware and MiddlewareFunc variables")
	}
}

// TestHandlerAndMiddlewareSignature tests that Handler and Middleware have the same signature
func TestHandlerAndMiddlewareSignature(t *testing.T) {
	// Define a function that can be used as both Handler and Middleware
	fn := func(c *Ctx) {
		c.Status(StatusOK).String("Function called")
	}

	// Assign the function to both Handler and Middleware variables
	var handler Handler = fn
	var middleware Middleware = fn

	// Verify both can be called with the same context
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Test handler
	res1 := httptest.NewRecorder()
	ctx1 := GetContext(res1, req)
	handler(ctx1)
	ctx1.Writer.Flush()

	// Test middleware
	res2 := httptest.NewRecorder()
	ctx2 := GetContext(res2, req)
	middleware(ctx2)
	ctx2.Writer.Flush()

	// Both should set the same status and body
	if res1.Code != StatusOK || res2.Code != StatusOK {
		t.Errorf("Expected both to set status %d, got %d and %d", StatusOK, res1.Code, res2.Code)
	}

	if res1.Body.String() != "Function called" || res2.Body.String() != "Function called" {
		t.Errorf("Expected both to set body 'Function called', got '%s' and '%s'",
			res1.Body.String(), res2.Body.String())
	}
}
