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
