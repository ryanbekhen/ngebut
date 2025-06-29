package ngebut

import (
	"context"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
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
	// Convert URL parameters like :id and wildcards * to regex patterns
	var regexPattern string

	if strings.Contains(pattern, ":") || strings.Contains(pattern, "*") {
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
				// Parameter segment like :id
				sb.WriteString("([^/]+)")
			} else if segment == "*" {
				// Wildcard segment - matches everything including slashes
				sb.WriteString("(.*)")
			} else {
				// Regular segment - escape special regex characters
				escaped := regexp.QuoteMeta(segment)
				sb.WriteString(escaped)
			}
		}
		sb.WriteString("$")
		regexPattern = sb.String()
	} else {
		// Simple case, just add ^ and $ and escape special regex characters
		regexPattern = "^" + regexp.QuoteMeta(pattern) + "$"
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

// HandleStatic registers a new route for serving static files.
func (r *Router) HandleStatic(prefix, root string, config ...Static) *Router {
	// Use default config if none provided
	cfg := DefaultStaticConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	// Clean up the prefix to ensure it ends with /*
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	pattern := prefix + "*"

	// Create the static file handler
	handler := createStaticHandler(prefix, root, cfg)

	// Register the route
	return r.Handle(pattern, MethodGet, handler)
}

// createStaticHandler creates a handler function for serving static files
func createStaticHandler(prefix, root string, config Static) Handler {
	// Ensure root path is absolute and clean
	absRoot, err := filepath.Abs(root)
	if err != nil {
		absRoot = root
	}

	return func(c *Ctx) {
		// Skip if Next function returns true
		if config.Next != nil && config.Next(c) {
			c.Next()
			return
		}

		// Get the file path from the URL
		filePath := strings.TrimPrefix(c.Path(), strings.TrimSuffix(prefix, "/"))

		// Remove leading slash if present
		filePath = strings.TrimPrefix(filePath, "/")

		if filePath == "" {
			filePath = config.Index
		}

		// Clean the file path and join with root
		filePath = filepath.Clean(filePath)
		fullPath := filepath.Join(absRoot, filePath)

		// Security check: ensure the file path is within the root directory
		if !strings.HasPrefix(fullPath, absRoot) {
			c.Status(StatusForbidden)
			c.String("Forbidden")
			return
		}

		// Get file info
		fileInfo, err := os.Stat(fullPath)
		if err != nil {
			if os.IsNotExist(err) {
				c.Status(StatusNotFound)
				c.String("File not found")
				return
			}
			c.Status(StatusInternalServerError)
			c.String("Internal Server Error")
			return
		}

		// Handle directory requests
		if fileInfo.IsDir() {
			// Try to serve index file only if Index is specified
			if config.Index != "" {
				indexPath := filepath.Join(fullPath, config.Index)
				if indexInfo, err := os.Stat(indexPath); err == nil && !indexInfo.IsDir() {
					fullPath = indexPath
					fileInfo = indexInfo
				} else if config.Browse {
					// Serve directory listing
					serveDirectoryListing(c, fullPath, filePath, config)
					return
				} else {
					c.Status(StatusForbidden)
					c.String("Directory listing is disabled")
					return
				}
			} else if config.Browse {
				// No index file specified, serve directory listing
				serveDirectoryListing(c, fullPath, filePath, config)
				return
			} else {
				c.Status(StatusForbidden)
				c.String("Directory listing is disabled")
				return
			}
		}

		// Handle byte range requests
		if config.ByteRange && c.Get("Range") != "" {
			serveFileWithRange(c, fullPath, fileInfo, config)
			return
		}

		// Serve the file
		serveFile(c, fullPath, fileInfo, config)
	}
}

// serveFile serves a single file
func serveFile(c *Ctx, filePath string, fileInfo os.FileInfo, config Static) {
	// Open and read the file
	file, err := os.Open(filePath)
	if err != nil {
		c.Status(StatusInternalServerError)
		c.String("Error opening file")
		return
	}
	defer file.Close()

	// Set headers
	setFileHeaders(c, filePath, fileInfo, config)

	// Call ModifyResponse if provided
	if config.ModifyResponse != nil {
		config.ModifyResponse(c)
	}

	// Determine content type
	contentType := mime.TypeByExtension(filepath.Ext(filePath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Set content type header
	c.Set("Content-Type", contentType)

	// Stream the file directly to reduce memory usage and GC pressure
	_, err = io.Copy(c.Writer, file)
	if err != nil {
		logger.Error().Err(err).Msg("Error serving file")
	}
}

// serveFileWithRange serves a file with HTTP range support
func serveFileWithRange(c *Ctx, filePath string, fileInfo os.FileInfo, config Static) {
	rangeHeader := c.Get("Range")
	if !strings.HasPrefix(rangeHeader, "bytes=") {
		// Invalid range header, serve the whole file
		serveFile(c, filePath, fileInfo, config)
		return
	}

	fileSize := fileInfo.Size()
	ranges := parseRangeHeader(rangeHeader[6:], fileSize) // Remove "bytes=" prefix

	if len(ranges) == 0 {
		// Invalid range, return 416 Range Not Satisfiable
		c.Status(StatusRequestedRangeNotSatisfiable)
		c.Set("Content-Range", fmt.Sprintf("bytes */%d", fileSize))
		return
	}

	// For simplicity, only handle single range requests
	if len(ranges) > 1 {
		serveFile(c, filePath, fileInfo, config)
		return
	}

	r := ranges[0]

	// Open and seek to the range start
	file, err := os.Open(filePath)
	if err != nil {
		c.Status(StatusInternalServerError)
		c.String("Error opening file")
		return
	}
	defer file.Close()

	_, err = file.Seek(r.start, 0)
	if err != nil {
		c.Status(StatusInternalServerError)
		c.String("Error seeking file")
		return
	}

	// Read the range content
	rangeLength := r.end - r.start + 1
	content := make([]byte, rangeLength)
	_, err = io.ReadFull(file, content)
	if err != nil {
		c.Status(StatusInternalServerError)
		c.String("Error reading file range")
		return
	}

	// Set range headers
	c.Status(StatusPartialContent)
	c.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", r.start, r.end, fileSize))
	c.Set("Accept-Ranges", "bytes")
	c.Set("Content-Length", strconv.FormatInt(rangeLength, 10))

	// Set other headers
	setFileHeaders(c, filePath, fileInfo, config)

	// Call ModifyResponse if provided
	if config.ModifyResponse != nil {
		config.ModifyResponse(c)
	}

	// Determine content type
	contentType := mime.TypeByExtension(filepath.Ext(filePath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Send the range content
	c.Data(contentType, content)
}

// serveDirectoryListing serves a directory listing
func serveDirectoryListing(c *Ctx, dirPath, urlPath string, config Static) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		c.Status(StatusInternalServerError)
		c.String("Error reading directory")
		return
	}

	// Build HTML directory listing
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
	<title>Directory listing for %s</title>
	<style>
		body { font-family: Arial, sans-serif; margin: 20px; }
		table { border-collapse: collapse; width: 100%%; }
		th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
		th { background-color: #f2f2f2; }
		a { text-decoration: none; color: #0066cc; }
		a:hover { text-decoration: underline; }
	</style>
</head>
<body>
	<h1>Directory listing for %s</h1>
	<table>
		<tr><th>Name</th><th>Size</th><th>Modified</th></tr>`, urlPath, urlPath)

	// Add parent directory link if not at root
	if urlPath != "/" {
		html += `<tr><td><a href="../">../</a></td><td>-</td><td>-</td></tr>`
	}

	// Add entries
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		name := entry.Name()
		if entry.IsDir() {
			name += "/"
		}

		size := "-"
		if !entry.IsDir() {
			size = formatFileSize(info.Size())
		}

		modTime := info.ModTime().Format("2006-01-02 15:04:05")
		html += fmt.Sprintf(`<tr><td><a href="%s">%s</a></td><td>%s</td><td>%s</td></tr>`,
			name, name, size, modTime)
	}

	html += `</table></body></html>`

	c.HTML(html)
}

// setFileHeaders sets common headers for file responses
func setFileHeaders(c *Ctx, filePath string, fileInfo os.FileInfo, config Static) {
	// Set Last-Modified header
	c.Set("Last-Modified", fileInfo.ModTime().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT"))

	// Set Cache-Control header
	if config.MaxAge > 0 {
		c.Set("Cache-Control", fmt.Sprintf("public, max-age=%d", config.MaxAge))
	}

	// Set Content-Length header
	c.Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))

	// Set Content-Disposition for downloads
	if config.Download {
		filename := filepath.Base(filePath)
		c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	}

	// Set Accept-Ranges header if byte range is supported
	if config.ByteRange {
		c.Set("Accept-Ranges", "bytes")
	}
}

// httpRange represents a byte range request
type httpRange struct {
	start, end int64
}

// parseRangeHeader parses the Range header value
func parseRangeHeader(rangeSpec string, fileSize int64) []httpRange {
	var ranges []httpRange

	// Split multiple ranges by comma
	parts := strings.Split(rangeSpec, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)

		if strings.Contains(part, "-") {
			rangeParts := strings.SplitN(part, "-", 2)

			var start, end int64
			var err error

			if rangeParts[0] == "" {
				// Suffix-byte-range-spec (e.g., "-500")
				if rangeParts[1] == "" {
					continue // Invalid range
				}
				suffixLength, err := strconv.ParseInt(rangeParts[1], 10, 64)
				if err != nil || suffixLength >= fileSize {
					continue
				}
				start = fileSize - suffixLength
				end = fileSize - 1
			} else if rangeParts[1] == "" {
				// Range from start to end (e.g., "500-")
				start, err = strconv.ParseInt(rangeParts[0], 10, 64)
				if err != nil || start >= fileSize {
					continue
				}
				end = fileSize - 1
			} else {
				// Full range (e.g., "0-499")
				start, err = strconv.ParseInt(rangeParts[0], 10, 64)
				if err != nil {
					continue
				}
				end, err = strconv.ParseInt(rangeParts[1], 10, 64)
				if err != nil || start > end || start >= fileSize {
					continue
				}
				if end >= fileSize {
					end = fileSize - 1
				}
			}

			ranges = append(ranges, httpRange{start: start, end: end})
		}
	}

	return ranges
}

// formatFileSize formats file size in human-readable format
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
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

// STATIC registers a new route with the GET method.
func (r *Router) STATIC(prefix, root string, config ...Static) *Router {
	return r.HandleStatic(prefix, root, config...)
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
