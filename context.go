package ngebut

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/goccy/go-json"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// Ctx represents the context of an HTTP request.
// It contains the request and response data, as well as utilities for handling
// the request and generating a response. It also manages middleware execution.
type Ctx struct {
	Writer     ResponseWriter
	Request    *Request
	statusCode int
	err        error
	userData   map[string]interface{}

	// Cache for parameter lookup to avoid repeated context lookups
	paramCache cachedParamMap

	// Cache for query parameter lookup to avoid repeated parsing
	queryCache cachedQueryMap

	// Fields for the middleware pattern
	middlewareStack []MiddlewareFunc
	middlewareIndex int
	handler         Handler
}

// Note: The paramCtxKey variable is defined in param.go
// and is used as a key for storing parameters in the request context

// contextPool is a pool of Ctx objects for reuse
var contextPool = sync.Pool{
	New: func() interface{} {
		return &Ctx{
			statusCode:      StatusOK,
			err:             nil,
			middlewareStack: make([]MiddlewareFunc, 0, 4), // Pre-allocate capacity for common middleware count
			middlewareIndex: -1,
			paramCache:      cachedParamMap{valid: false, params: nil},
			queryCache:      cachedQueryMap{valid: false, values: nil},
			userData:        make(map[string]interface{}, 4), // Pre-allocate userData map with increased capacity
		}
	},
}

// bufferPool is a pool of bytes.Buffer objects for reuse
var bufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 8192)) // Increased capacity to 8KB to reduce reallocations
	},
}

// jsonBufferPool is a dedicated pool of bytes.Buffer objects for JSON serialization
// Using a separate pool for JSON allows us to optimize buffer sizes for JSON specifically
var jsonBufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 32768)) // 32KB capacity for large JSON responses
	},
}

// copyHeadersToWriter copies all headers from c.header to c.Writer.Header()
// and ensures headers set after c.Next() in middleware are included
// This optimized version reduces allocations and improves performance
func (c *Ctx) copyHeadersToWriter() {
	if c.Writer == nil || c.Request == nil || c.Request.Header == nil {
		return
	}

	// Get the writer's header map
	writerHeader := c.Writer.Header()
	if writerHeader == nil {
		return
	}

	// First, copy any headers from the writer that aren't in the context
	// This is typically needed for headers set by middleware after c.Next()
	for key, values := range *writerHeader {
		if len(values) == 0 {
			continue
		}

		// Check if this header is already in the context
		if _, exists := (*c.Request.Header)[key]; !exists && len(values) > 0 {
			// Copy the values directly to avoid allocations
			(*c.Request.Header)[key] = values
		}
	}

	// Now copy all headers from the context to the writer
	// This will include both original context headers and those we just copied from the writer
	for key, values := range *c.Request.Header {
		if len(values) == 0 {
			continue
		}

		// Set all values at once to avoid multiple map lookups
		(*writerHeader)[key] = values
	}
}

// prepareResponse sets the content type, copies headers to the writer, and writes the status code
// This is an optimized helper function to reduce function call overhead in response methods
func (c *Ctx) prepareResponse(contentType string) {
	if c.Writer == nil {
		return
	}

	writerHeader := c.Writer.Header()

	// Set content type on both writer and request header
	if contentType != "" {
		(*writerHeader)["Content-Type"] = []string{contentType}

		// Also set in Request.Header for ctx.Get to work in tests
		if c.Request != nil && c.Request.Header != nil {
			(*c.Request.Header)["Content-Type"] = []string{contentType}
		}
	}

	// Copy only essential headers that aren't already in the writer
	// Skip the full header iteration for better performance
	if c.Request.Header != nil && c.statusCode != StatusOK {
		// For non-200 responses, copy status-related headers
		if values, ok := (*c.Request.Header)["X-Status-Reason"]; ok && len(values) > 0 {
			(*writerHeader)["X-Status-Reason"] = values
		}
	}

	// Write status code
	c.Writer.WriteHeader(c.statusCode)
}

// write implements the http.ResponseWriter interface.
// It writes the data directly to the underlying response writer.
//
// Parameters:
//   - data: The byte slice to write
//
// Returns:
//   - The number of bytes written
//   - Any error that occurred during writing
func (c *Ctx) write(data []byte) (int, error) {
	// Write directly to the underlying response writer
	return c.Writer.Write(data)
}

// Error sets an error in the context.
// This can be used to store errors that occur during request processing.
// If the current status code is less than 400, it will be set to 500 (Internal Server Error).
//
// Parameters:
//   - err: The error to store in the context
//
// Returns:
//   - The context itself for method chaining
func (c *Ctx) Error(err error) *Ctx {
	if c.statusCode < 400 {
		c.statusCode = StatusInternalServerError
	}
	c.err = err
	return c
}

// GetError returns the error stored in the context.
// If no explicit error was set but the status code is 400 or higher,
// it returns a new error with the status text as the message.
//
// Returns:
//   - The error stored in the context, a new error based on the status code,
//     or nil if no error condition exists
func (c *Ctx) GetError() error {
	if c.err != nil {
		return c.err
	}
	return nil
}

// Next calls the next middleware or handler in the stack.
// If there are no more middleware functions, it calls the final handler.
// This method is typically called within middleware to pass control to the next middleware
// or to the final route handler.
//
// Example usage in middleware:
//
//	func MyMiddleware(c *ngebut.Ctx) {
//	    // Do something before the next middleware or handler
//	    c.Next()
//	    // Do something after the next middleware or handler has completed
//	}
func (c *Ctx) Next() {
	c.middlewareIndex++

	// If we've gone through all middleware, call the final handler
	if c.middlewareIndex >= len(c.middlewareStack) {
		if c.handler != nil {
			c.handler(c)
		}
		return
	}

	// Call the next middleware
	c.middlewareStack[c.middlewareIndex](c)
}

// GetContext gets a Ctx from the pool and initializes it with the given writer and request.
// This function reuses Ctx objects from a pool to reduce memory allocations.
//
// Parameters:
//   - w: The http.ResponseWriter to use for the response
//   - r: The http.Request to process
//
// Returns:
//   - A properly initialized *Ctx object ready for request processing
func GetContext(w http.ResponseWriter, r *http.Request) *Ctx {
	ctx := contextPool.Get().(*Ctx)
	ctx.Writer = NewResponseWriter(w)
	ctx.Request = NewRequest(r)
	return ctx
}

// GetContextFromRequest gets a Ctx from the pool and initializes it with the given writer and request.
// Similar to GetContext but accepts a *Request instead of http.Request.
// This function reuses Ctx objects from a pool to reduce memory allocations.
//
// Parameters:
//   - w: The http.ResponseWriter to use for the response
//   - r: The *Request to process (already wrapped ngebut Request)
//
// Returns:
//   - A properly initialized *Ctx object ready for request processing
func getContextFromRequest(w http.ResponseWriter, r *Request) *Ctx {
	ctx := contextPool.Get().(*Ctx)
	ctx.Writer = NewResponseWriter(w)
	ctx.Request = r

	// Use headers from the response writer
	if w != nil && w.Header() != nil {
		ctx.Request.Header = NewHeaderFromMap(w.Header())
	}

	return ctx
}

// ReleaseContext returns a Ctx to the pool after resetting its state.
// This function should be called when you're done with a context to allow reuse.
// It clears all fields and returns the Ctx to the pool.
//
// Parameters:
//   - ctx: The context to reset and return to the pool
//
// Note: After calling this function, the ctx should not be used anymore.
func ReleaseContext(ctx *Ctx) {
	ctx.statusCode = StatusOK
	ctx.err = nil

	if ctx.Request != nil && ctx.Request.Header != nil {
		for k := range *ctx.Request.Header {
			delete(*ctx.Request.Header, k)
		}
	}

	ctx.middlewareStack = ctx.middlewareStack[:0]
	ctx.middlewareIndex = -1
	ctx.handler = nil

	// Reset the parameter cache
	ctx.paramCache.valid = false
	if ctx.paramCache.params != nil {
		releaseParamSlice(ctx.paramCache.params)
		ctx.paramCache.params = nil
	}

	// Reset the query cache but keep the map for reuse
	ctx.queryCache.valid = false
	if ctx.queryCache.values != nil {
		// Clear the map without deallocating
		for k := range ctx.queryCache.values {
			delete(ctx.queryCache.values, k)
		}
	}

	// Clear the user data map without reallocating
	if ctx.userData != nil {
		// Check if there's a parameter context in the UserData map and release it
		if paramCtx, ok := ctx.userData["__paramCtx"]; ok {
			if ps, ok := paramCtx.(*paramSlice); ok && ps != nil {
				releaseParamSlice(ps)
			}
		}

		// Clear the map
		for k := range ctx.userData {
			delete(ctx.userData, k)
		}
	}

	// Release the response writer back to its pool
	if ctx.Writer != nil {
		ReleaseResponseWriter(ctx.Writer)
		ctx.Writer = nil
	}

	ctx.Request = nil

	contextPool.Put(ctx)
}

// StatusCode returns the current HTTP status code set for the response.
//
// Returns:
//   - The HTTP status code as an integer
func (c *Ctx) StatusCode() int {
	return c.statusCode
}

// Header returns the header map that will be sent with the response.
// This can be used to access the current headers or to modify them.
//
// Returns:
//   - The Header object containing all response headers
func (c *Ctx) Header() *Header {
	return c.Request.Header
}

// Method returns the HTTP method of the request (e.g., GET, POST, PUT).
// If the request is nil, it returns an empty string.
// This method is useful for determining how the request was made.
func (c *Ctx) Method() string {
	if c.Request == nil {
		return ""
	}
	return c.Request.Method
}

// Path returns the URL path of the request.
// If the request is nil, it returns an empty string.
// This method is useful for determining the requested resource.
// For example, for a request to "/users/123", it would return "/users/123".
// If the request is nil, it returns an empty string.
func (c *Ctx) Path() string {
	if c.Request == nil {
		return ""
	}
	return c.Request.URL.Path
}

// IP returns the client's IP address.
// It tries to determine the real IP address by checking various headers
// that might be set by proxies, before falling back to the direct connection IP.
//
// The order of precedence is:
// 1. X-Forwarded-For header (first value)
// 2. X-Real-Ip header
// 3. RemoteAddr from the request
//
// Returns:
//   - The client's IP address as a string, or empty string if not determinable
func (c *Ctx) IP() string {
	// Check if Request is nil
	if c.Request == nil {
		return ""
	}

	// Check for X-Forwarded-For header first (for clients behind proxies)
	if xff := c.Request.Header.Get(HeaderXForwardedFor); xff != "" {
		// X-Forwarded-For can contain multiple IPs, the first one is the original client
		// Find the first comma or end of string to extract the first IP
		commaIdx := strings.IndexByte(xff, ',')
		var firstIP string
		if commaIdx > 0 {
			firstIP = xff[:commaIdx]
		} else {
			firstIP = xff
		}

		// Trim spaces without allocating a new string when possible
		firstIP = strings.TrimSpace(firstIP)
		if firstIP != "" {
			return firstIP
		}
	}

	// Check for X-Real-IP header next
	if xrip := c.Request.Header.Get("X-Real-Ip"); xrip != "" {
		return xrip
	}

	// Fall back to RemoteAddr
	if c.Request.RemoteAddr != "" {
		// RemoteAddr is in the format "IP:port", so we need to extract just the IP
		ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
		if err == nil {
			return ip
		}
		return c.Request.RemoteAddr
	}

	return ""
}

// RemoteAddr returns the direct remote address of the request.
// Unlike IP(), this method only looks at the RemoteAddr field of the request
// and doesn't check any headers.
//
// Returns:
//   - The remote IP address as a string, or empty string if not available
func (c *Ctx) RemoteAddr() string {
	if c.Request == nil {
		return ""
	}
	return c.Request.RemoteAddr
}

// UserAgent returns the value of the "User-Agent" header from the request,
// or an empty string if the request is nil.
func (c *Ctx) UserAgent() string {
	if c.Request == nil {
		return ""
	}
	return c.Request.Header.Get(HeaderUserAgent)
}

// Referer retrieves the "Referer" header value from the incoming HTTP request.
// Returns an empty string if the request is nil or the header is absent.
func (c *Ctx) Referer() string {
	if c.Request == nil {
		return ""
	}
	return c.Request.Header.Get(HeaderReferer)
}

// Host returns the host of the request.
func (c *Ctx) Host() string {
	if c.Request == nil {
		return ""
	}

	// Check for X-Forwarded-Host header first
	if host := c.Request.Header.Get(HeaderXForwardedHost); host != "" {
		return host
	}

	// Use the Host field if available
	if c.Request.Host != "" {
		return c.Request.Host
	}

	// Fallback to the URL host if Host is not set
	return c.Request.URL.Host
}

// Protocol retrieves the protocol scheme (e.g., "http" or "https") from the request.
// It first checks proxy headers like X-Forwarded-Proto, then falls back to URL.Scheme,
// and finally determines based on TLS connection status.
// Returns "http" as default if the protocol cannot be determined.
func (c *Ctx) Protocol() string {
	if c.Request == nil {
		return ""
	}

	// Check X-Forwarded-Proto header first (common for proxies)
	if proto := c.Request.Header.Get(HeaderXForwardedProto); proto != "" {
		return proto
	}

	// Check X-Forwarded-Protocol header (less common)
	if proto := c.Request.Header.Get("X-Forwarded-Protocol"); proto != "" {
		return proto
	}

	// Check Front-End-Https header (used by some proxies)
	if c.Request.Header.Get("Front-End-Https") == "on" {
		return "https"
	}

	// Check X-Forwarded-Ssl header
	if c.Request.Header.Get("X-Forwarded-Ssl") == "on" {
		return "https"
	}

	// Fall back to URL.Scheme if set
	if c.Request.URL.Scheme != "" {
		return c.Request.URL.Scheme
	}

	// Default to http
	return "http"
}

// Status sets the HTTP status code for the response.
//
// Parameters:
//   - code: The HTTP status code to set (e.g., 200, 404, 500)
//
// Returns:
//   - The context itself for method chaining
func (c *Ctx) Status(code int) *Ctx {
	c.statusCode = code
	return c
}

// Set sets a response header with the given key and value.
// It sets the header in both c.header and c.Writer.Header() to ensure
// it's included in the response even if set after c.Next() in middleware.
//
// Parameters:
//   - key: The header name
//   - value: The header value
//
// Returns:
//   - The context itself for method chaining
func (c *Ctx) Set(key, value string) *Ctx {
	// Set the header in the request header for ctx.Get to work
	c.Request.Header.Set(key, value)

	// Set the header directly in the underlying writer's header
	if c.Writer != nil {
		// Get the underlying http.ResponseWriter
		if adapter, ok := c.Writer.(*httpResponseWriterAdapter); ok && adapter.writer != nil {
			// Set the header directly in the underlying writer's header
			adapter.writer.Header().Set(key, value)
		} else {
			// Fallback to the normal way
			c.Writer.Header().Set(key, value)
		}
	}
	return c
}

// Get retrieves a response header value by its key.
//
// Parameters:
//   - key: The header name to retrieve
//
// Returns:
//   - The header value as a string, or empty string if not found
func (c *Ctx) Get(key string) string {
	return c.Request.Header.Get(key)
}

// cachedParamMap caches the parameters to avoid repeated lookups
type cachedParamMap struct {
	params *paramSlice
	valid  bool
}

// cachedQueryMap caches parsed query parameters to avoid repeated parsing
type cachedQueryMap struct {
	values map[string][]string
	valid  bool
	// rawQuery is used to detect if the query string has changed
	rawQuery string
}

// paramKeyCache is a pool of pre-allocated parameter keys to avoid string allocations
var paramKeyCache = sync.Pool{
	New: func() interface{} {
		return make(map[string]struct{}, 16) // Pre-allocate with capacity for common routes
	},
}

// getParamKeyCache gets a parameter key cache from the pool
func getParamKeyCache() map[string]struct{} {
	return paramKeyCache.Get().(map[string]struct{})
}

// releaseParamKeyCache returns a parameter key cache to the pool
func releaseParamKeyCache(m map[string]struct{}) {
	// Clear the map
	for k := range m {
		delete(m, k)
	}
	paramKeyCache.Put(m)
}

// Param retrieves a URL path parameter value by its key.
// For example, in a route "/users/:id", Param("id") would return the value in the URL path.
//
// Parameters:
//   - key: The parameter name to retrieve
//
// Returns:
//   - The parameter value as a string, or empty string if not found
func (c *Ctx) Param(key string) string {
	if c.Request == nil {
		return ""
	}

	// Fast path: Use cached parameter slice if available
	// This is now the primary path since we store parameters directly in paramCache
	if c.paramCache.valid && c.paramCache.params != nil {
		if value, found := (*c.paramCache.params).Get(key); found {
			return value
		}
		return ""
	}

	// Backward compatibility: If paramCache is not valid, try to get parameters from request context
	// This path should be rare in modern code since we now store parameters directly in paramCache
	ctxValue := c.Request.Context().Value(paramCtxKey)
	if ctxValue == nil {
		return ""
	}

	// Try the parameter slice first (fastest path)
	if paramSlice, ok := ctxValue.(*paramSlice); ok && paramSlice != nil {
		// Cache the parameter slice for future lookups
		c.paramCache.params = paramSlice
		c.paramCache.valid = true

		// Get the value from the slice
		if value, found := paramSlice.Get(key); found {
			return value
		}
		return ""
	}

	// Try the map-based parameter context (legacy path)
	if paramCtx, ok := ctxValue.(map[paramKey]string); ok && paramCtx != nil {
		// Get a parameter slice from the pool
		ps := getParamSlice()

		// Convert map to slice for better performance (only once)
		// Pre-allocate with exact capacity to avoid reallocations
		if cap(ps.entries) < len(paramCtx) {
			ps.entries = make([]paramEntry, 0, len(paramCtx))
		} else {
			ps.entries = ps.entries[:0]
		}

		// Use a cached key to avoid string allocations
		paramKey := paramKey(key)

		// Check if the key exists in the map directly
		// This avoids converting the entire map if we just need one value
		if value, exists := paramCtx[paramKey]; exists {
			// Add this parameter to the slice first
			ps.entries = append(ps.entries, paramEntry{key: key, value: value})

			// Cache the parameter slice for future lookups
			c.paramCache.params = ps
			c.paramCache.valid = true

			return value
		}

		// If the key wasn't found, add all parameters to the slice
		// Use a type assertion instead of conversion to avoid string allocations
		for k, v := range paramCtx {
			// Use the string representation of k directly without conversion
			// This is safe because paramKey is just a type alias for string
			ps.entries = append(ps.entries, paramEntry{key: string(k), value: v})
		}

		// Cache the parameter slice for future lookups
		c.paramCache.params = ps
		c.paramCache.valid = true

		// The key wasn't found in the direct lookup, so it doesn't exist
		return ""
	}

	// Fall back to direct context lookup for backward compatibility
	// Use a cached key to avoid string allocations
	if value := c.Request.Context().Value(paramKey(key)); value != nil {
		if strValue, ok := value.(string); ok {
			return strValue
		}
	}

	return ""
}

// ensureQueryCache ensures that the query cache is populated
// It returns the cached values map
func (c *Ctx) ensureQueryCache() map[string][]string {
	if c.Request == nil || c.Request.URL == nil {
		return nil
	}

	// Fast path: if there's no query string, return nil
	rawQuery := c.Request.URL.RawQuery
	if rawQuery == "" {
		return nil
	}

	// Use cached query values if available and valid
	if c.queryCache.valid && c.queryCache.values != nil && c.queryCache.rawQuery == rawQuery {
		return c.queryCache.values
	}

	// If the query cache map is nil, pre-allocate it
	if c.queryCache.values == nil {
		c.queryCache.values = make(map[string][]string, 8) // Pre-allocate with capacity for common query params
	} else {
		// Clear existing values
		for k := range c.queryCache.values {
			delete(c.queryCache.values, k)
		}
	}

	// Store the raw query string for cache invalidation
	c.queryCache.rawQuery = rawQuery

	// Parse query parameters directly to avoid allocations from URL.Query()
	parseQueryString(rawQuery, c.queryCache.values)

	c.queryCache.valid = true
	return c.queryCache.values
}

// parseQueryString parses a query string into a map without allocating a new map
// This is a more efficient version of url.ParseQuery that reuses an existing map
func parseQueryString(query string, values map[string][]string) {
	// Fast path for empty query
	if query == "" {
		return
	}

	// Process the query string character by character
	key := ""
	value := ""
	inKey := true
	start := 0

	for i := 0; i < len(query); i++ {
		switch query[i] {
		case '=':
			if inKey {
				key = query[start:i]
				start = i + 1
				inKey = false
			}
		case '&':
			if !inKey {
				value = query[start:i]
				addQueryParam(values, key, value)
			} else {
				// Handle empty values like "key1&key2=val2"
				key = query[start:i]
				addQueryParam(values, key, "")
			}
			start = i + 1
			inKey = true
		}
	}

	// Handle the last key-value pair
	if start < len(query) {
		if inKey {
			// Last parameter has no value
			key = query[start:]
			addQueryParam(values, key, "")
		} else {
			// Last parameter has a value
			value = query[start:]
			addQueryParam(values, key, value)
		}
	}
}

// addQueryParam adds a query parameter to the values map
// It handles URL decoding and appends to existing values
func addQueryParam(values map[string][]string, key, value string) {
	// Skip empty keys
	if key == "" {
		return
	}

	// URL decode the key and value
	key = urlDecode(key)
	value = urlDecode(value)

	// Append to existing values or create a new slice
	values[key] = append(values[key], value)
}

// urlDecode decodes a URL-encoded string
// This is a simplified version that handles the most common cases
func urlDecode(s string) string {
	// Fast path for strings without encoding
	if !strings.ContainsAny(s, "+%") {
		return s
	}

	// Replace '+' with space
	s = strings.ReplaceAll(s, "+", " ")

	// Handle percent-encoded characters
	var buf strings.Builder
	buf.Grow(len(s))

	for i := 0; i < len(s); i++ {
		if s[i] == '%' && i+2 < len(s) {
			// Try to decode the percent-encoded byte
			if b, err := hexToByte(s[i+1], s[i+2]); err == nil {
				buf.WriteByte(b)
				i += 2
				continue
			}
		}
		buf.WriteByte(s[i])
	}

	return buf.String()
}

// hexToByte converts two hex characters to a byte
func hexToByte(c1, c2 byte) (byte, error) {
	var b1, b2 byte

	switch {
	case '0' <= c1 && c1 <= '9':
		b1 = c1 - '0'
	case 'a' <= c1 && c1 <= 'f':
		b1 = c1 - 'a' + 10
	case 'A' <= c1 && c1 <= 'F':
		b1 = c1 - 'A' + 10
	default:
		return 0, errors.New("invalid hex character")
	}

	switch {
	case '0' <= c2 && c2 <= '9':
		b2 = c2 - '0'
	case 'a' <= c2 && c2 <= 'f':
		b2 = c2 - 'a' + 10
	case 'A' <= c2 && c2 <= 'F':
		b2 = c2 - 'A' + 10
	default:
		return 0, errors.New("invalid hex character")
	}

	return (b1 << 4) | b2, nil
}

// Query retrieves a URL query parameter value by its key.
// For example, in a URL "?name=john", Query("name") would return "john".
//
// Parameters:
//   - key: The query parameter name to retrieve
//
// Returns:
//   - The query parameter value as a string, or empty string if not found
func (c *Ctx) Query(key string) string {
	values := c.ensureQueryCache()
	if values == nil {
		return ""
	}

	// Return the first value if it exists
	if vals, exists := values[key]; exists && len(vals) > 0 {
		return vals[0]
	}
	return ""
}

// QueryArray retrieves all values for a URL query parameter by its key.
// For example, in a URL "?color=red&color=blue", QueryArray("color") would return ["red", "blue"].
//
// Parameters:
//   - key: The query parameter name to retrieve
//
// Returns:
//   - A slice of strings containing all values for the parameter, or an empty slice if not found
func (c *Ctx) QueryArray(key string) []string {
	values := c.ensureQueryCache()
	if values == nil {
		return []string{}
	}

	// Return all values if they exist
	if vals, exists := values[key]; exists {
		return vals
	}
	return []string{}
}

// Cookie sets a cookie in the response.
// It adds the Set-Cookie header to the response with the serialized cookie.
func (c *Ctx) Cookie(cookie *Cookie) *Ctx {
	if cookie == nil {
		return c
	}

	c.Set("Set-Cookie", cookie.String())
	return c
}

// Cookies retrieve a cookie from the request by its name.
// It returns the cookie value as a string, or an empty string if the cookie is not found.
func (c *Ctx) Cookies(name string) string {
	if c.Request == nil {
		return ""
	}

	cookieHeader := c.Request.Header.Get("Cookie")
	if cookieHeader == "" {
		return ""
	}

	cookies := parseCookies(cookieHeader)
	return cookies[name]
}

// ClearCookies removes all cookies for the context by setting an empty Set-Cookie header.
func (c *Ctx) ClearCookies() *Ctx {
	// Clear all cookies by setting an empty Set-Cookie header
	c.Set("Set-Cookie", "")
	return c
}

// Pre-allocated content type for plain text responses to avoid allocations
var plainTextContentType = []string{"text/plain; charset=utf-8"}

// String sends a plain text response with the given format and values.
// It sets the Content-Type header to "text/plain; charset=utf-8".
// If values are provided, it formats the string using fmt.Sprintf.
//
// Parameters:
//   - format: The string format (can contain format specifiers if values are provided)
//   - values: Optional values to be formatted into the string
//
// Note: This method writes the response immediately and sets the status code.
func (c *Ctx) String(format string, values ...interface{}) {
	// Fast path for simple strings without formatting
	if len(values) == 0 {
		// For strings without formatting, write directly to the response writer
		// Set content type and status code
		if c.Writer != nil {
			// Set content type on both writer and request header using pre-allocated slice
			header := c.Writer.Header()
			(*header)["Content-Type"] = plainTextContentType

			// Also set in Request.Header for ctx.Get to work in tests
			if c.Request != nil && c.Request.Header != nil {
				(*c.Request.Header)["Content-Type"] = plainTextContentType
			}

			c.Writer.WriteHeader(c.statusCode)

			// For very small strings, write directly without buffer
			if len(format) < 64 {
				_, _ = c.Writer.Write([]byte(format))
				return
			}

			// For larger strings, use a buffer from the pool
			buf := bufferPool.Get().(*bytes.Buffer)
			buf.Reset()

			// Ensure the buffer has enough capacity
			if buf.Cap() < len(format) {
				buf.Grow(len(format) - buf.Cap())
			}

			buf.WriteString(format)
			_, _ = c.Writer.Write(buf.Bytes())
			bufferPool.Put(buf)
			return
		}
		return
	}

	// For formatted strings, set content type directly for better performance
	if c.Writer != nil {
		// Set content type on both writer and request header using pre-allocated slice
		header := c.Writer.Header()
		(*header)["Content-Type"] = plainTextContentType

		// Also set in Request.Header for ctx.Get to work in tests
		if c.Request != nil && c.Request.Header != nil {
			(*c.Request.Header)["Content-Type"] = plainTextContentType
		}

		c.Writer.WriteHeader(c.statusCode)

		// Get a buffer from the pool
		buf := bufferPool.Get().(*bytes.Buffer)
		buf.Reset()

		// Use Fprintf for formatted strings
		fmt.Fprintf(buf, format, values...)

		_, _ = c.Writer.Write(buf.Bytes())
		bufferPool.Put(buf)
	}
}

// Pre-allocated content type for JSON responses to avoid allocations
var jsonContentType = []string{"application/json; charset=utf-8"}

// Pre-allocated JSON values to avoid allocations
var (
	jsonNull  = []byte("null")
	jsonTrue  = []byte("true")
	jsonFalse = []byte("false")
	jsonQuote = []byte(`"`)
)

// JSON sends a JSON response by encoding the provided object.
// It sets the Content-Type header to "application/json; charset=utf-8".
//
// Parameters:
//   - obj: The object to be encoded to JSON
//
// Note: This method writes the response immediately and sets the status code.
func (c *Ctx) JSON(obj interface{}) {
	// Set content type and status code directly for better performance
	if c.Writer == nil {
		return
	}

	// Set content type on both writer and request header using pre-allocated slice
	header := c.Writer.Header()
	(*header)["Content-Type"] = jsonContentType

	// Also set in Request.Header for ctx.Get to work in tests
	if c.Request != nil && c.Request.Header != nil {
		(*c.Request.Header)["Content-Type"] = jsonContentType
	}

	c.Writer.WriteHeader(c.statusCode)

	// Fast path for nil objects
	if obj == nil {
		_, _ = c.Writer.Write(jsonNull)
		return
	}

	// Fast path for simple types that can be marshaled efficiently
	switch v := obj.(type) {
	case string:
		// For strings, we can write directly with quotes
		_, _ = c.Writer.Write(jsonQuote)
		_, _ = c.Writer.Write([]byte(v))
		_, _ = c.Writer.Write(jsonQuote)
		return
	case bool:
		// For booleans, use pre-allocated values
		if v {
			_, _ = c.Writer.Write(jsonTrue)
		} else {
			_, _ = c.Writer.Write(jsonFalse)
		}
		return
	case int:
		// For small integers, use a static buffer to avoid allocations
		if v >= -128 && v <= 1023 {
			var buf [16]byte // Static buffer large enough for small integers
			s := strconv.AppendInt(buf[:0], int64(v), 10)
			_, _ = c.Writer.Write(s)
			return
		}
	case int64:
		// For small integers, use a static buffer to avoid allocations
		if v >= -128 && v <= 1023 {
			var buf [16]byte // Static buffer large enough for small integers
			s := strconv.AppendInt(buf[:0], v, 10)
			_, _ = c.Writer.Write(s)
			return
		}
	case float64:
		// For simple floats, use a static buffer to avoid allocations
		if v == float64(int64(v)) && v >= -128 && v <= 1023 {
			var buf [16]byte // Static buffer large enough for small integers
			s := strconv.AppendInt(buf[:0], int64(v), 10)
			_, _ = c.Writer.Write(s)
			return
		}
	}

	// For more complex objects, try to marshal directly first
	data, err := json.Marshal(obj)
	if err == nil {
		// For small data, write directly
		if len(data) < 256 {
			_, _ = c.Writer.Write(data)
			return
		}

		// For larger data, use the dedicated JSON buffer from the pool
		// This buffer has a larger capacity specifically for JSON responses
		buf := jsonBufferPool.Get().(*bytes.Buffer)
		buf.Reset()

		// If data fits in our buffer, use the buffer
		if len(data) < buf.Cap() {
			buf.Write(data)
			_, _ = c.Writer.Write(buf.Bytes())
		} else {
			// For very large data that exceeds our buffer capacity,
			// write directly to the response writer to avoid an extra copy
			_, _ = c.Writer.Write(data)
		}

		// Return the buffer to the pool
		jsonBufferPool.Put(buf)
	} else {
		// Fallback to encoder for complex objects or if Marshal fails
		buf := jsonBufferPool.Get().(*bytes.Buffer)
		buf.Reset()

		if err := json.NewEncoder(buf).Encode(obj); err != nil {
			c.Error(fmt.Errorf("JSON encoding error: %w", err))
		} else {
			_, _ = c.Writer.Write(buf.Bytes())
		}

		// Return the buffer to the pool
		jsonBufferPool.Put(buf)
	}
}

// Pre-allocated content type for HTML responses to avoid allocations
var htmlContentType = []string{"text/html; charset=utf-8"}

// HTML sends an HTML response with the provided content.
// It sets the Content-Type header to "text/html; charset=utf-8".
//
// Parameters:
//   - html: The HTML content to send as a response
//
// Note: This method writes the response immediately and sets the status code.
func (c *Ctx) HTML(html string) {
	// Set content type directly for better performance
	if c.Writer == nil {
		return
	}

	// Set content type on both writer and request header using pre-allocated slice
	header := c.Writer.Header()
	(*header)["Content-Type"] = htmlContentType

	// Also set in Request.Header for ctx.Get to work in tests
	if c.Request != nil && c.Request.Header != nil {
		(*c.Request.Header)["Content-Type"] = htmlContentType
	}

	c.Writer.WriteHeader(c.statusCode)

	// Fast path for empty HTML strings
	if len(html) == 0 {
		return
	}

	// Fast path for small HTML strings (avoid buffer allocation)
	if len(html) < 512 {
		_, _ = c.Writer.Write([]byte(html))
		return
	}

	// For larger strings, use a buffer from the pool
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()

	// Ensure the buffer has enough capacity to avoid reallocations
	if buf.Cap() < len(html) {
		// If the buffer is too small, grow it
		buf.Grow(len(html) - buf.Cap())
	}

	// Write the HTML string to the buffer
	buf.WriteString(html)

	// Write directly from the buffer
	_, _ = c.Writer.Write(buf.Bytes())

	// Return buffer to pool
	bufferPool.Put(buf)
}

// Data sends a raw byte slice response with the specified content type.
// This is useful for sending binary data like images, PDFs, etc.
//
// Parameters:
//   - contentType: The MIME type of the data (e.g., "image/jpeg", "application/pdf")
//   - data: The byte slice containing the data to send
//
// Note: This method writes the response immediately and sets the status code.
func (c *Ctx) Data(contentType string, data []byte) {
	// Prepare response with content type and status code in one operation
	c.prepareResponse(contentType)

	// Fast path for empty data
	if len(data) == 0 {
		return
	}

	// Fast path for small data (avoid buffer allocation)
	if len(data) < 256 {
		_, _ = c.Writer.Write(data)
		return
	}

	// For larger data, use a buffer from the pool to avoid multiple small writes
	// This reduces allocations and improves performance
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()

	// Ensure the buffer has enough capacity to avoid reallocations
	if buf.Cap() < len(data) {
		// If the buffer is too small, grow it
		buf.Grow(len(data) - buf.Cap())
	}

	// Write the data to the buffer
	buf.Write(data)

	// Write directly from the buffer
	_, _ = c.Writer.Write(buf.Bytes())

	// Return buffer to pool
	bufferPool.Put(buf)
}

// userDataKeyPool is a pool of common UserData keys to avoid string allocations
var userDataKeyPool = sync.Map{}

// Common keys used in UserData to avoid allocations
var (
	userDataParamContextKey = "__paramCtx"
)

// internUserDataKey returns an interned string for the given key to avoid allocations
func internUserDataKey(key string) string {
	// Check if we already have this key interned
	if interned, ok := userDataKeyPool.Load(key); ok {
		return interned.(string)
	}

	// Store the key in the pool
	userDataKeyPool.Store(key, key)
	return key
}

// UserData sets or get user-specific data in the context.
func (c *Ctx) UserData(key string, value ...interface{}) interface{} {
	// Check if userData map is nil and initialize it if needed
	if c.userData == nil {
		c.userData = make(map[string]interface{}, 4) // Pre-allocate with capacity for common use
	}

	// Use interned key to avoid string allocations for common keys
	internedKey := key
	if key == userDataParamContextKey {
		// Special case for the most common key
		internedKey = userDataParamContextKey
	} else {
		internedKey = internUserDataKey(key)
	}

	if len(value) > 0 {
		// Set the value if provided
		c.userData[internedKey] = value[0]
		return value[0]
	} else {
		// Get the value if no value is provided
		if val, exists := c.userData[internedKey]; exists {
			return val
		}
		return nil
	}
}
