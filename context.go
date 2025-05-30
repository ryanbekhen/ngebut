package ngebut

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/goccy/go-json"
	"net"
	"net/http"
	"net/url"
	"reflect"
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
	header     Header
	body       []byte
	err        error

	// Fields for the middleware pattern
	middlewareStack []MiddlewareFunc
	middlewareIndex int
	handler         Handler
}

// contextPool is a pool of Ctx objects for reuse
var contextPool = sync.Pool{
	New: func() interface{} {
		return &Ctx{
			statusCode:      StatusOK,
			header:          make(Header),
			body:            make([]byte, 0, 512),
			err:             nil,
			middlewareStack: make([]MiddlewareFunc, 0),
			middlewareIndex: -1,
		}
	},
}

// bufferPool is a pool of bytes.Buffer objects for reuse
var bufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 512))
	},
}

// copyHeadersToWriter copies all headers from c.header to c.Writer.Header()
// and ensures headers set after c.Next() in middleware are included
// This optimized version reduces allocations and improves performance
func (c *Ctx) copyHeadersToWriter() {
	if c.Writer == nil {
		return
	}

	// Get the writer's header map
	writerHeader := c.Writer.Header()

	// First, copy any headers from the writer that aren't in the context
	// This is typically needed for headers set by middleware after c.Next()
	for key, values := range writerHeader {
		if len(values) == 0 {
			continue
		}

		// Check if this header is already in the context
		if _, exists := c.header[key]; !exists && len(values) > 0 {
			// Copy the values directly to avoid allocations
			c.header[key] = values
		}
	}

	// Now copy all headers from the context to the writer
	// This will include both original context headers and those we just copied from the writer
	for key, values := range c.header {
		if len(values) == 0 {
			continue
		}

		// Set all values at once to avoid multiple map lookups
		writerHeader[key] = values
	}
}

// write implements the http.ResponseWriter interface.
// It stores the data in the context's body buffer for later use.
//
// Parameters:
//   - data: The byte slice to write
//
// Returns:
//   - The number of bytes written
//   - Any error that occurred during writing (always nil in this implementation)
func (c *Ctx) write(data []byte) (int, error) {
	// Use append directly with a pre-calculated capacity to avoid allocations
	// when possible. If the capacity is sufficient, no allocation will occur.
	c.body = append(c.body, data...)
	return len(data), nil
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
func GetContextFromRequest(w http.ResponseWriter, r *Request) *Ctx {
	ctx := contextPool.Get().(*Ctx)
	ctx.Writer = NewResponseWriter(w)
	ctx.Request = r

	// Use headers directly without copying for zero-allocation
	if w != nil && w.Header() != nil {
		ctx.header = Header(w.Header())
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

	for k := range ctx.header {
		delete(ctx.header, k)
	}

	ctx.body = ctx.body[:0]
	ctx.middlewareStack = ctx.middlewareStack[:0]
	ctx.middlewareIndex = -1
	ctx.handler = nil

	// Release the response writer back to its pool
	if ctx.Writer != nil {
		ReleaseResponseWriter(ctx.Writer)
		ctx.Writer = nil
	}

	ctx.Request = nil

	contextPool.Put(ctx)
}

// NewContext creates a new Ctx without using the pool.
// This is mainly for backward compatibility and testing.
// Unlike GetContext, this function allocates a new Ctx every time it's called.
//
// Parameters:
//   - w: The http.ResponseWriter to use for the response
//   - r: The http.Request to process
//
// Returns:
//   - A newly allocated *Ctx object ready for request processing
func NewContext(w http.ResponseWriter, r *http.Request) *Ctx {
	return &Ctx{
		Writer:          NewResponseWriter(w),
		Request:         NewRequest(r),
		statusCode:      StatusOK,
		header:          make(Header),
		body:            make([]byte, 0, 512),
		err:             nil,
		middlewareStack: make([]MiddlewareFunc, 0),
		middlewareIndex: -1,
	}
}

// NewContextFromRequest creates a new Ctx without using the pool.
// This is mainly for backward compatibility and testing.
// Similar to NewContext but accepts a *Request instead of http.Request.
// Unlike GetContextFromRequest, this function allocates a new Ctx every time it's called.
//
// Parameters:
//   - w: The http.ResponseWriter to use for the response
//   - r: The *Request to process (already wrapped ngebut Request)
//
// Returns:
//   - A newly allocated *Ctx object ready for request processing
func NewContextFromRequest(w http.ResponseWriter, r *Request) *Ctx {
	return &Ctx{
		Writer:          NewResponseWriter(w),
		Request:         r,
		statusCode:      StatusOK,
		header:          make(Header),
		body:            make([]byte, 0, 512),
		err:             nil,
		middlewareStack: make([]MiddlewareFunc, 0),
		middlewareIndex: -1,
	}
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
func (c *Ctx) Header() Header {
	return c.header
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
	c.header.Set(key, value)
	if c.Writer != nil {
		c.Writer.Header().Set(key, value)
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
	return c.header.Get(key)
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

	// Try the new parameter context first
	if paramCtx, ok := c.Request.Context().Value(paramContextKey{}).(map[paramKey]string); ok && paramCtx != nil {
		return paramCtx[paramKey(key)]
	}

	// Fall back to the old method for backward compatibility
	if value := c.Request.Context().Value(paramKey(key)); value != nil {
		if strValue, ok := value.(string); ok {
			return strValue
		}
	}

	return ""
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
	return c.Request.URL.Query().Get(key)
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
	return c.Request.URL.Query()[key]
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

// textContentTypeBytes is the byte slice for the text/plain content type header
var textContentTypeBytes = []byte("text/plain; charset=utf-8")

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
	c.Set("Content-Type", "text/plain; charset=utf-8")
	c.copyHeadersToWriter()
	c.Writer.WriteHeader(c.statusCode)

	var responseBytes []byte

	// Fast path for simple strings without formatting
	if len(values) == 0 {
		// For short strings, avoid buffer allocation completely
		if len(format) < 1024 {
			// Store directly in context body
			c.body = append(c.body[:0], format...)
			// Write directly to response writer
			_, _ = c.Writer.Write(c.body)
			return
		}

		// For longer strings, use a buffer from the pool
		buf := bufferPool.Get().(*bytes.Buffer)
		buf.Reset()
		buf.WriteString(format)
		responseBytes = buf.Bytes()
	} else {
		// For formatted strings, use a buffer from the pool
		buf := bufferPool.Get().(*bytes.Buffer)
		buf.Reset()
		fmt.Fprintf(buf, format, values...)
		responseBytes = buf.Bytes()
		defer bufferPool.Put(buf)
	}

	// Store in context body and write to response writer
	c.body = append(c.body[:0], responseBytes...)
	_, _ = c.Writer.Write(responseBytes)
}

// jsonContentType is the content type for JSON responses
const jsonContentType = "application/json; charset=utf-8"

// JSON sends a JSON response by encoding the provided object.
// It sets the Content-Type header to "application/json; charset=utf-8".
//
// Parameters:
//   - obj: The object to be encoded to JSON
//
// Note: This method writes the response immediately and sets the status code.
func (c *Ctx) JSON(obj interface{}) {
	c.Set("Content-Type", jsonContentType)
	c.copyHeadersToWriter()
	c.Writer.WriteHeader(c.statusCode)

	// Get buffer from pool
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	// Use faster JSON encoding
	if err := json.NewEncoder(buf).Encode(obj); err != nil {
		c.Error(fmt.Errorf("JSON encoding error: %w", err))
		return
	}

	if cap(c.body) < len(buf.Bytes()) {
		newCap := len(buf.Bytes()) * 2
		c.body = make([]byte, 0, newCap)
	}

	// Get bytes and reuse
	data := buf.Bytes()
	c.body = c.body[:0]
	c.body = append(c.body, data...)

	// Write response
	_, _ = c.Writer.Write(data)
}

// HTML sends an HTML response with the provided content.
// It sets the Content-Type header to "text/html; charset=utf-8".
//
// Parameters:
//   - html: The HTML content to send as a response
//
// Note: This method writes the response immediately and sets the status code.
func (c *Ctx) HTML(html string) {
	c.Set("Content-Type", "text/html; charset=utf-8")
	c.copyHeadersToWriter()
	c.Writer.WriteHeader(c.statusCode)

	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	buf.WriteString(html)

	// Get bytes once and reuse to avoid multiple calls to buf.Bytes()
	bytes := buf.Bytes()
	c.write(bytes)
	_, _ = c.Writer.Write(bytes)
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
	c.Set("Content-Type", contentType)
	c.copyHeadersToWriter()
	c.Writer.WriteHeader(c.statusCode)

	// Store data in context's body buffer and write to response writer
	// in a single operation to avoid duplicate writes
	c.write(data)
	_, _ = c.Writer.Write(data)
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
	// Check for X-Forwarded-For header first (for clients behind proxies)
	if xff := c.Get("X-Forwarded-For"); xff != "" {
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
	if xrip := c.Get("X-Real-Ip"); xrip != "" {
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
	if c.Request.RemoteAddr != "" {
		ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
		if err == nil {
			return ip
		}
		return c.Request.RemoteAddr
	}
	return ""
}

// BindJSON unmarshals the JSON data from the request body into the provided object.
// It reads the request body, decodes the JSON, and populates the object.
// If the request body is nil or if unmarshaling fails, it returns an error.
// This method is typically used in route handlers to bind incoming JSON data to a struct.
// Parameters:
//   - obj: The object to unmarshal the JSON data into
//
// Returns:
//   - An error if the request body is nil or if unmarshaling fails
//   - nil if successful
//
// Example usage in a route handler:
//
//	func MyHandler(c *ngebut.Ctx) {
//		   var data MyDataType
//		   if err := c.BindJSON(&data); err != nil {
//		       c.Error(err)
//		       return
//		   }
//		   // Now data is populated with the JSON from the request body
//		   c.JSON(data)
//	}
func (c *Ctx) BindJSON(obj interface{}) error {
	if c.Request.Body == nil {
		return errors.New("request body is nil")
	}

	// Unmarshal the JSON data into the provided object
	if err := json.Unmarshal(c.Request.Body, obj); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}

// BindForm parses form data from the request and binds it to the provided object.
// It supports the following Content-Types:
// - application/x-www-form-urlencoded
// - multipart/form-data
// - text/plain (treated as URL-encoded)
// - empty Content-Type (treated as URL-encoded)
// The struct fields should be tagged with `form:"field_name"` to specify the form field name.
// If a field doesn't have a form tag, it will be skipped.
// Parameters:
//   - obj: The object to bind the form data to
//
// Returns:
//   - An error if parsing the form data fails or if the provided object is not a pointer to a struct
//   - nil if successful
//
// Example usage in a route handler:
//
//	func MyHandler(c *ngebut.Ctx) {
//	    var data MyDataType
//	    if err := c.BindForm(&data); err != nil {
//	        c.Error(err)
//	        return
//	    }
//	    // Now data is populated with the form data from the request
//	    c.JSON(data)
//	}
func (c *Ctx) BindForm(obj interface{}) error {
	// Check if the request has a body
	if c.Request.Body == nil {
		return errors.New("request body is nil")
	}

	// Check if obj is a pointer to a struct
	objValue := reflect.ValueOf(obj)
	if objValue.Kind() != reflect.Ptr || objValue.Elem().Kind() != reflect.Struct {
		return errors.New("obj must be a pointer to a struct")
	}

	// Parse the form data based on the Content-Type header
	contentType := c.Request.Header.Get("Content-Type")
	var values url.Values

	if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
		// Parse URL-encoded form data
		body := string(c.Request.Body)
		var err error
		values, err = url.ParseQuery(body)
		if err != nil {
			return fmt.Errorf("failed to parse form data: %w", err)
		}
	} else if strings.HasPrefix(contentType, "multipart/form-data") {
		// Parse multipart form data
		// Create a new http.Request with the same body for parsing
		httpReq, err := http.NewRequest(c.Request.Method, c.Request.URL.String(), bytes.NewReader(c.Request.Body))
		if err != nil {
			return fmt.Errorf("failed to create request for multipart parsing: %w", err)
		}

		// Copy headers to ensure Content-Type with boundary is preserved
		for k, v := range c.Request.Header {
			httpReq.Header[k] = v
		}

		// Parse the multipart form
		err = httpReq.ParseMultipartForm(32 << 20) // 32MB max memory
		if err != nil {
			return fmt.Errorf("failed to parse multipart form: %w", err)
		}

		values = httpReq.Form
	} else if contentType == "" || strings.HasPrefix(contentType, "text/plain") {
		// Handle plain form data or no content type (treat as URL-encoded)
		body := string(c.Request.Body)
		var err error
		values, err = url.ParseQuery(body)
		if err != nil {
			return fmt.Errorf("failed to parse form data: %w", err)
		}
	} else {
		return fmt.Errorf("unsupported Content-Type for form binding: %s", contentType)
	}

	// Bind the form values to the struct fields
	objElem := objValue.Elem()
	objType := objElem.Type()

	for i := 0; i < objElem.NumField(); i++ {
		field := objType.Field(i)
		fieldValue := objElem.Field(i)

		// Skip unexported fields
		if !fieldValue.CanSet() {
			continue
		}

		// Get the form tag
		formTag := field.Tag.Get("form")
		if formTag == "" {
			// Skip fields without a form tag
			continue
		}

		// Get the form value
		formValue := values.Get(formTag)
		if formValue == "" {
			// Skip empty values
			continue
		}

		// Set the field value based on its type
		switch fieldValue.Kind() {
		case reflect.String:
			fieldValue.SetString(formValue)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			intValue, err := strconv.ParseInt(formValue, 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse %s as int: %w", formTag, err)
			}
			fieldValue.SetInt(intValue)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			uintValue, err := strconv.ParseUint(formValue, 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse %s as uint: %w", formTag, err)
			}
			fieldValue.SetUint(uintValue)
		case reflect.Float32, reflect.Float64:
			floatValue, err := strconv.ParseFloat(formValue, 64)
			if err != nil {
				return fmt.Errorf("failed to parse %s as float: %w", formTag, err)
			}
			fieldValue.SetFloat(floatValue)
		case reflect.Bool:
			boolValue, err := strconv.ParseBool(formValue)
			if err != nil {
				return fmt.Errorf("failed to parse %s as bool: %w", formTag, err)
			}
			fieldValue.SetBool(boolValue)
		default:
			// Skip unsupported types
			continue
		}
	}

	return nil
}
