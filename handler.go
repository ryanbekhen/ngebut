package ngebut

// Handler is a function that handles an HTTP request with a Ctx.
// This is the same signature as middleware functions, making them interchangeable.
type Handler func(c *Ctx)

// Middleware is a function that can be used as middleware.
// It has the same signature as Handler, making them interchangeable.
// The function should call c.Next() to continue to the next middleware or handler.
type Middleware func(c *Ctx)

// MiddlewareFunc is an alias for Middleware for backward compatibility.
// It's similar to the middleware pattern used in gofiber.
type MiddlewareFunc = Middleware

// CompileMiddleware precomposes multiple middleware functions into a single handler function.
// This eliminates the need for runtime middleware chaining, reducing allocations and function call overhead.
// The middleware functions are executed in the order they are provided.
//
// Unlike the dynamic middleware approach that uses c.Next(), this function creates a static chain
// where each middleware is directly called in sequence, eliminating the overhead of dynamic dispatch.
//
// Note: This approach is most effective for middleware that doesn't rely on c.Next() for complex
// behavior like executing code after the next middleware completes.
func CompileMiddleware(handler Handler, middleware ...Middleware) Handler {
	// If there's no middleware, just return the handler
	if len(middleware) == 0 {
		return handler
	}

	// Create a new handler that executes all middleware in sequence
	return func(c *Ctx) {
		// Create a chain of function calls without using c.Next()
		// This eliminates the overhead of dynamic dispatch

		// Execute all middleware in sequence
		for _, m := range middleware {
			// Call each middleware directly
			// If the middleware sets an error or writes a response, we should stop
			m(c)

			// Check if an error was set or if a response was written
			if c.GetError() != nil || (c.Writer != nil && c.statusCode >= 400) {
				return
			}
		}

		// Finally, call the handler
		handler(c)
	}
}
