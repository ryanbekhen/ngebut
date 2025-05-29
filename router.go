package ngebut

import (
	"context"
	"regexp"
	"strings"
	"sync"
)

// route represents a route with a pattern, method, and handlers.
type route struct {
	Pattern  string
	Method   string
	Handlers []Handler
	Regex    *regexp.Regexp
}

// middlewareStackPool is a pool of middleware stacks for reuse
// This pool helps reduce memory allocations by reusing middleware stacks
// instead of creating new ones for each request.
var middlewareStackPool = sync.Pool{
	New: func() interface{} {
		// Create a middleware stack with a reasonable initial capacity
		return make([]MiddlewareFunc, 0, 16)
	},
}

// routeSegmentPool is a pool for route segments to reduce allocations
var routeSegmentPool = sync.Pool{
	New: func() interface{} {
		return make([]string, 0, 8)
	},
}

// stringBuilderPool is a pool for string builders to reduce allocations
var stringBuilderPool = sync.Pool{
	New: func() interface{} {
		return new(strings.Builder)
	},
}

// allowedMethodsPool is a pool for allowed methods slices to reduce allocations
var allowedMethodsPool = sync.Pool{
	New: func() interface{} {
		return make([]string, 0, 8)
	},
}

// Router is an HTTP request router.
type Router struct {
	Routes          []route
	routesByMethod  map[string][]route // Routes indexed by method for faster lookup
	middlewareFuncs []MiddlewareFunc
	NotFound        Handler
}

// NewRouter creates a new Router.
func NewRouter() *Router {
	return &Router{
		Routes:          []route{},
		routesByMethod:  make(map[string][]route),
		middlewareFuncs: []MiddlewareFunc{},
		NotFound: func(c *Ctx) {
			c.Status(StatusNotFound)
			c.String("404 page not found")
		},
	}
}

// Use adds middleware to the router.
// It accepts middleware functions that take a context parameter.
func (r *Router) Use(middleware ...interface{}) {
	for _, m := range middleware {
		switch m := m.(type) {
		case Middleware:
			r.middlewareFuncs = append(r.middlewareFuncs, m)
		case func(*Ctx):
			r.middlewareFuncs = append(r.middlewareFuncs, m)
		default:
			panic("middleware must be a function that takes a *Ctx parameter")
		}
	}
}

// Handle registers a new route with the given pattern and method.
func (r *Router) Handle(pattern, method string, handlers ...Handler) *Router {
	// Convert URL parameters like :id to regex patterns
	var regexPattern string

	if strings.Contains(pattern, ":") {
		// Get a string builder from the pool
		sb := stringBuilderPool.Get().(*strings.Builder)
		sb.Reset()
		defer stringBuilderPool.Put(sb)

		// Get a segments slice from the pool
		segments := routeSegmentPool.Get().([]string)
		segments = segments[:0]
		defer routeSegmentPool.Put(segments)

		// Split the pattern into segments
		start := 0
		for i := 0; i < len(pattern); i++ {
			if pattern[i] == '/' {
				if i > start {
					segments = append(segments, pattern[start:i])
				}
				start = i + 1
			}
		}
		if start < len(pattern) {
			segments = append(segments, pattern[start:])
		}

		// Build the regex pattern
		sb.WriteString("^")
		// Add leading slash
		sb.WriteString("/")
		for i, segment := range segments {
			if i > 0 {
				sb.WriteString("/")
			}
			if len(segment) > 0 && segment[0] == ':' {
				sb.WriteString("([^/]+)")
			} else {
				sb.WriteString(segment)
			}
		}
		sb.WriteString("$")
		regexPattern = sb.String()
	} else {
		// Simple case, just add ^ and $
		regexPattern = "^" + pattern + "$"
	}

	regex := regexp.MustCompile(regexPattern)
	newRoute := route{
		Pattern:  pattern,
		Method:   method,
		Handlers: handlers,
		Regex:    regex,
	}

	// Add to the main routes slice
	r.Routes = append(r.Routes, newRoute)

	// Add to the method-specific routes map for faster lookup
	r.routesByMethod[method] = append(r.routesByMethod[method], newRoute)

	// For HEAD requests, we can also use GET handlers (HTTP spec)
	if method == MethodGet {
		r.routesByMethod[MethodHead] = append(r.routesByMethod[MethodHead], newRoute)
	}

	return r
}

// GET registers a new route with the GET method.
func (r *Router) GET(pattern string, handlers ...Handler) *Router {
	return r.Handle(pattern, MethodGet, handlers...)
}

// HEAD registers a new route with the HEAD method.
func (r *Router) HEAD(pattern string, handlers ...Handler) *Router {
	return r.Handle(pattern, MethodHead, handlers...)
}

// POST registers a new route with the POST method.
func (r *Router) POST(pattern string, handlers ...Handler) *Router {
	return r.Handle(pattern, MethodPost, handlers...)
}

// PUT registers a new route with the PUT method.
func (r *Router) PUT(pattern string, handlers ...Handler) *Router {
	return r.Handle(pattern, MethodPut, handlers...)
}

// DELETE registers a new route with the DELETE method.
func (r *Router) DELETE(pattern string, handlers ...Handler) *Router {
	return r.Handle(pattern, MethodDelete, handlers...)
}

// CONNECT registers a new route with the CONNECT method.
func (r *Router) CONNECT(pattern string, handlers ...Handler) *Router {
	return r.Handle(pattern, MethodConnect, handlers...)
}

// OPTIONS registers a new route with the OPTIONS method.
func (r *Router) OPTIONS(pattern string, handlers ...Handler) *Router {
	return r.Handle(pattern, MethodOptions, handlers...)
}

// TRACE registers a new route with the TRACE method.
func (r *Router) TRACE(pattern string, handlers ...Handler) *Router {
	return r.Handle(pattern, MethodTrace, handlers...)
}

// PATCH registers a new route with the PATCH method.
func (r *Router) PATCH(pattern string, handlers ...Handler) *Router {
	return r.Handle(pattern, MethodPatch, handlers...)
}

// paramContextPool is a pool for request contexts with parameters
var paramContextPool = sync.Pool{
	New: func() interface{} {
		return make(map[paramKey]string, 8)
	},
}

// getParamContext gets a parameter context map from the pool
func getParamContext() map[paramKey]string {
	return paramContextPool.Get().(map[paramKey]string)
}

// releaseParamContext returns a parameter context map to the pool
func releaseParamContext(m map[paramKey]string) {
	for k := range m {
		delete(m, k)
	}
	paramContextPool.Put(m)
}

// releaseParamContextKey is the key used to store the function that releases the parameter context
type releaseParamContextKey struct{}

// handleMatchedRoute handles a route that matched the path and method
func (r *Router) handleMatchedRoute(ctx *Ctx, req *Request, route route, matches []string, path string) {
	// Extract URL parameters
	if strings.Contains(route.Pattern, ":") {
		// Get a parameter context map from the pool
		paramCtx := getParamContext()

		// We need to release the parameter context map when the request is done
		// We'll do this by wrapping the request context in a context that will release the map on Done
		originalReqCtx := req.Context()

		// Extract parameters directly from regex matches
		// The regex pattern is created to capture each parameter in order
		paramIndex := 1 // Start from 1 to skip the full match

		// Fast path for common patterns with a small number of parameters
		// This avoids the character-by-character scanning for most routes
		colonCount := strings.Count(route.Pattern, ":")
		if colonCount <= 4 { // Arbitrary threshold, adjust based on your routes
			// Use a more efficient approach for routes with few parameters
			start := 0
			for i := 0; i < len(route.Pattern); i++ {
				if route.Pattern[i] == ':' {
					// Found a parameter
					start = i + 1 // Skip the colon

					// Find the end of the parameter (next slash or end of string)
					end := strings.IndexByte(route.Pattern[start:], '/')
					if end == -1 {
						// Parameter extends to the end of the pattern
						paramName := route.Pattern[start:]
						if paramIndex < len(matches) {
							paramCtx[paramKey(paramName)] = matches[paramIndex]
							paramIndex++
						}
					} else {
						// Parameter ends at a slash
						paramName := route.Pattern[start : start+end]
						if paramIndex < len(matches) {
							paramCtx[paramKey(paramName)] = matches[paramIndex]
							paramIndex++
						}
						// Skip to after the slash to avoid finding it in the next iteration
						i = start + end
					}
				}
			}
		} else {
			// Use the character-by-character approach for complex patterns
			inParam := false
			paramStart := 0

			for i := 0; i < len(route.Pattern); i++ {
				if route.Pattern[i] == '/' {
					if inParam {
						// End of parameter
						paramName := route.Pattern[paramStart+1 : i] // +1 to skip the ':'
						if paramIndex < len(matches) {
							paramCtx[paramKey(paramName)] = matches[paramIndex]
							paramIndex++
						}
						inParam = false
					}
				} else if route.Pattern[i] == ':' && (i == 0 || route.Pattern[i-1] == '/') {
					// Start of parameter
					inParam = true
					paramStart = i
				}
			}

			// Check for parameter at the end of the pattern
			if inParam && paramStart < len(route.Pattern) {
				paramName := route.Pattern[paramStart+1:] // +1 to skip the ':'
				if paramIndex < len(matches) {
					paramCtx[paramKey(paramName)] = matches[paramIndex]
				}
			}
		}

		// Store parameters in request context
		reqCtx := context.WithValue(originalReqCtx, paramContextKey{}, paramCtx)

		// Create a context that will release the parameter context map when done
		reqCtx = context.WithValue(reqCtx, releaseParamContextKey{}, func() {
			releaseParamContext(paramCtx)
		})

		req = req.WithContext(reqCtx)
	}

	// Update the request in the context
	ctx.Request = req

	// Set up the middleware stack with both global middleware and route handlers
	r.setupMiddleware(ctx, route.Handlers)
}

// setupMiddleware sets up the middleware stack for a request
func (r *Router) setupMiddleware(ctx *Ctx, handlers []Handler) {
	// Fast path: if we have no middleware and only one handler, call it directly
	if len(r.middlewareFuncs) == 0 && len(handlers) == 1 {
		handlers[0](ctx)
		return
	}

	// Calculate the total middleware size
	totalMiddleware := len(r.middlewareFuncs) + len(handlers) - 1
	if totalMiddleware <= 0 {
		// No middleware and no handlers, or just one handler
		if len(handlers) == 1 {
			handlers[0](ctx)
		}
		return
	}

	// Prepare the middleware stack
	// Try to reuse the existing slice first
	if cap(ctx.middlewareStack) >= totalMiddleware {
		// Reuse the existing slice
		ctx.middlewareStack = ctx.middlewareStack[:totalMiddleware]
	} else {
		// Get a middleware stack from the pool with pre-check for capacity
		stack := middlewareStackPool.Get().([]MiddlewareFunc)

		if cap(stack) >= totalMiddleware {
			// Reuse the stack with the right size
			ctx.middlewareStack = stack[:totalMiddleware]
		} else {
			// Return the too-small stack to the pool
			middlewareStackPool.Put(stack)
			// Create a new stack with sufficient capacity
			// Add some extra capacity to reduce future allocations
			ctx.middlewareStack = make([]MiddlewareFunc, totalMiddleware, totalMiddleware+8)
		}
	}

	// Copy the global middleware functions in one operation
	globalMiddlewareCount := len(r.middlewareFuncs)
	if globalMiddlewareCount > 0 {
		copy(ctx.middlewareStack[:globalMiddlewareCount], r.middlewareFuncs)
	}

	// Add all but the last handler as middleware
	// Use direct indexing for better performance
	handlerCount := len(handlers)
	if handlerCount > 1 {
		for i := 0; i < handlerCount-1; i++ {
			ctx.middlewareStack[globalMiddlewareCount+i] = MiddlewareFunc(handlers[i])
		}
	}

	// Set the last handler as the final handler
	ctx.middlewareIndex = -1
	ctx.handler = handlers[handlerCount-1]

	// Call the first middleware function
	ctx.Next()
}

// Pre-allocated handler for method not allowed responses
var methodNotAllowedHandler = func(c *Ctx) {
	c.Status(StatusMethodNotAllowed)
	// The Allow header will be set before this handler is called
	c.String("Method Not Allowed")
}

// ServeHTTP implements a modified http.Handler interface that accepts a Ctx.
func (r *Router) ServeHTTP(ctx *Ctx, req *Request) {
	path := req.URL.Path
	method := req.Method

	// Fast path: check if we have routes for this method
	methodRoutes, hasMethodRoutes := r.routesByMethod[method]
	if hasMethodRoutes {
		// Use a more efficient loop with index for better performance
		for i := 0; i < len(methodRoutes); i++ {
			route := &methodRoutes[i]
			matches := route.Regex.FindStringSubmatch(path)
			if len(matches) > 0 {
				// We found a match, handle it
				r.handleMatchedRoute(ctx, req, *route, matches, path)
				return
			}
		}
	}

	// If we didn't find a match, check for method not allowed
	// Get allowed methods from the pool
	allowedMethods := allowedMethodsPool.Get().([]string)
	allowedMethods = allowedMethods[:0]
	methodNotAllowed := false

	// Use a more efficient approach to find allowed methods
	// Pre-allocate a map to track methods we've already seen
	methodSeen := make(map[string]bool, 8)

	// Check all routes for path matches with different methods
	for i := 0; i < len(r.Routes); i++ {
		route := &r.Routes[i]
		// Skip routes we already checked
		if route.Method == method {
			continue
		}

		// Skip methods we've already seen
		if methodSeen[route.Method] {
			continue
		}

		matches := route.Regex.FindStringSubmatch(path)
		if len(matches) > 0 {
			// Path matches but method doesn't match
			methodNotAllowed = true
			methodSeen[route.Method] = true
			allowedMethods = append(allowedMethods, route.Method)
		}
	}

	// If we found a matching path but method was not allowed, return 405 Method Not Allowed
	if methodNotAllowed {
		// Set the Allow header
		// Use a string builder to avoid allocations when joining allowed methods
		var allowHeader string
		if len(allowedMethods) == 1 {
			// Fast path for single method
			allowHeader = allowedMethods[0]
		} else {
			// Use a string builder for multiple methods
			sb := stringBuilderPool.Get().(*strings.Builder)
			sb.Reset()

			for i, m := range allowedMethods {
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(m)
			}

			allowHeader = sb.String()
			stringBuilderPool.Put(sb)
		}

		ctx.Set("Allow", allowHeader)

		// Return allowed methods to the pool
		allowedMethodsPool.Put(allowedMethods)

		// Set up middleware and call the handler
		r.setupMiddleware(ctx, []Handler{methodNotAllowedHandler})
		return
	}

	// Return allowed methods to the pool if we didn't use them
	allowedMethodsPool.Put(allowedMethods)

	// No route matched, use the NotFound handler
	r.setupMiddleware(ctx, []Handler{r.NotFound})
}
