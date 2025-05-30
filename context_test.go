package ngebut

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBindForm_MultipartFormData(t *testing.T) {
	// Create a multipart form
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// Add a field
	err := w.WriteField("name", "test-value")
	assert.NoError(t, err, "Failed to write field")

	// Close the writer
	err = w.Close()
	assert.NoError(t, err, "Failed to close writer")

	// Create a test request with the multipart form
	req, err := http.NewRequest("POST", "/test", &b)
	assert.NoError(t, err, "Failed to create request")
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Create a test response recorder
	res := httptest.NewRecorder()

	// Create a context
	ctx := GetContext(res, req)

	// Define a struct to bind to
	type TestStruct struct {
		Name string `form:"name"`
	}

	// Bind the form data
	var data TestStruct
	err = ctx.BindForm(&data)

	// Check for errors
	assert.NoError(t, err, "BindForm returned an error")

	// Check the bound data
	assert.Equal(t, "test-value", data.Name, "Name value should match")
}

func TestBindForm_URLEncodedForm(t *testing.T) {
	// Create a URL-encoded form
	formData := "name=test-value"

	// Create a test request with the URL-encoded form
	req, err := http.NewRequest("POST", "/test", strings.NewReader(formData))
	assert.NoError(t, err, "Failed to create request")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Create a test response recorder
	res := httptest.NewRecorder()

	// Create a context
	ctx := GetContext(res, req)

	// Define a struct to bind to
	type TestStruct struct {
		Name string `form:"name"`
	}

	// Bind the form data
	var data TestStruct
	err = ctx.BindForm(&data)

	// Check for errors
	assert.NoError(t, err, "BindForm returned an error")

	// Check the bound data
	assert.Equal(t, "test-value", data.Name, "Name value should match")
}

func TestBindForm_TextPlainForm(t *testing.T) {
	// Create a plain text form
	formData := "name=test-value"

	// Create a test request with the plain text form
	req, err := http.NewRequest("POST", "/test", strings.NewReader(formData))
	assert.NoError(t, err, "Failed to create request")
	req.Header.Set("Content-Type", "text/plain")

	// Create a test response recorder
	res := httptest.NewRecorder()

	// Create a context
	ctx := GetContext(res, req)

	// Define a struct to bind to
	type TestStruct struct {
		Name string `form:"name"`
	}

	// Bind the form data
	var data TestStruct
	err = ctx.BindForm(&data)

	// Check for errors
	assert.NoError(t, err, "BindForm returned an error")

	// Check the bound data
	assert.Equal(t, "test-value", data.Name, "Name value should match")
}

// TestError tests the Error method
func TestError(t *testing.T) {
	// Create a context
	req, _ := http.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := GetContext(res, req)

	// Set an error
	testErr := errors.New("test error")
	ctx.Error(testErr)

	// Check that the error was set
	assert.Equal(t, testErr, ctx.err, "Error should be set correctly")

	// Check that the status code was set to 500
	assert.Equal(t, StatusInternalServerError, ctx.statusCode, "Status code should be set to 500")

	// Test with a status code already set
	ctx.statusCode = StatusBadRequest
	ctx.Error(testErr)

	// Check that the status code wasn't changed
	assert.Equal(t, StatusBadRequest, ctx.statusCode, "Status code should not change when already set")
}

// TestGetError tests the GetError method
func TestGetError(t *testing.T) {
	// Create a context
	req, _ := http.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := GetContext(res, req)

	// Check that GetError returns nil when no error is set
	assert.Nil(t, ctx.GetError(), "GetError should return nil when no error is set")

	// Set an error
	testErr := errors.New("test error")
	ctx.err = testErr

	// Check that GetError returns the error
	assert.Equal(t, testErr, ctx.GetError(), "GetError should return the set error")
}

// TestStatusCode tests the StatusCode method
func TestStatusCode(t *testing.T) {
	// Create a context
	req, _ := http.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := GetContext(res, req)

	// Check the default status code
	assert.Equal(t, StatusOK, ctx.StatusCode(), "Default status code should be StatusOK")

	// Set a status code
	ctx.statusCode = StatusNotFound

	// Check that StatusCode returns the set code
	assert.Equal(t, StatusNotFound, ctx.StatusCode(), "StatusCode should return the set status code")
}

// TestHeader tests the Header method
func TestHeader(t *testing.T) {
	// Create a context
	req, _ := http.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := GetContext(res, req)

	// Check that Header returns a non-nil map
	assert.NotNil(t, ctx.Header(), "Header should return a non-nil map")

	// Set a header value
	ctx.Set("X-Test", "test-value")

	// Check that Header returns the set value
	assert.Equal(t, "test-value", ctx.Get("X-Test"), "Header should return the set value")
}

// TestMethod tests the Method method
func TestMethod(t *testing.T) {
	// Create a context with a GET request
	req, _ := http.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := GetContext(res, req)

	// Check that Method returns "GET"
	assert.Equal(t, "GET", ctx.Method(), "Method should return GET")

	// Create a context with a POST request
	req, _ = http.NewRequest("POST", "/test", nil)
	res = httptest.NewRecorder()
	ctx = GetContext(res, req)

	// Check that Method returns "POST"
	assert.Equal(t, "POST", ctx.Method(), "Method should return POST")

	// Test with nil Request
	ctx.Request = nil
	assert.Equal(t, "", ctx.Method(), "Method should return empty string when Request is nil")
}

// TestPath tests the Path method
func TestPath(t *testing.T) {
	// Create a context with a request to /test
	req, _ := http.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := GetContext(res, req)

	// Check that Path returns "/test"
	assert.Equal(t, "/test", ctx.Path(), "Path should return the request path")

	// Create a context with a request to /users/123
	req, _ = http.NewRequest("GET", "/users/123", nil)
	res = httptest.NewRecorder()
	ctx = GetContext(res, req)

	// Check that Path returns "/users/123"
	assert.Equal(t, "/users/123", ctx.Path(), "Path should return the request path")

	// Test with nil Request
	ctx.Request = nil
	assert.Equal(t, "", ctx.Path(), "Path should return empty string when Request is nil")
}

// TestIP tests the IP method
func TestIP(t *testing.T) {
	// Create a context with X-Forwarded-For header
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1, 10.0.0.1")
	res := httptest.NewRecorder()
	ctx := GetContext(res, req)

	// Check that IP returns the first IP in X-Forwarded-For
	assert.Equal(t, "192.168.1.1", ctx.IP(), "IP should return the first IP in X-Forwarded-For")

	// Create a context with X-Real-IP header
	req, _ = http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Real-IP", "192.168.1.2")
	res = httptest.NewRecorder()
	ctx = GetContext(res, req)

	// Check that IP returns the X-Real-IP
	assert.Equal(t, "192.168.1.2", ctx.IP(), "IP should return the X-Real-IP")

	// Create a context with RemoteAddr
	req, _ = http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.3:1234"
	res = httptest.NewRecorder()
	ctx = GetContext(res, req)

	// Check that IP returns the RemoteAddr without port
	// This tests the net.SplitHostPort functionality
	ip, port, err := net.SplitHostPort(req.RemoteAddr)
	assert.NoError(t, err, "Failed to split host port")
	assert.Equal(t, "1234", port, "Port should be correctly split")
	assert.Equal(t, "192.168.1.3", ip, "IP should be correctly split")

	assert.Equal(t, "192.168.1.3", ctx.IP(), "IP should return the RemoteAddr without port")

	// Test with nil Request
	ctx.Request = nil
	assert.Equal(t, "", ctx.IP(), "IP should return empty string when Request is nil")
}

// TestRemoteAddr tests the RemoteAddr method
func TestRemoteAddr(t *testing.T) {
	// Create a context with RemoteAddr
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	res := httptest.NewRecorder()
	ctx := GetContext(res, req)

	// Check that RemoteAddr returns the full RemoteAddr
	assert.Equal(t, "192.168.1.1:1234", ctx.RemoteAddr(), "RemoteAddr should return the full RemoteAddr")

	// Test with nil Request
	ctx.Request = nil
	assert.Equal(t, "", ctx.RemoteAddr(), "RemoteAddr should return empty string when Request is nil")
}

// TestUserAgent tests the UserAgent method
func TestUserAgent(t *testing.T) {
	// Create a context with User-Agent header
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	res := httptest.NewRecorder()
	ctx := GetContext(res, req)

	// Check that UserAgent returns the User-Agent header
	assert.Equal(t, "Mozilla/5.0", ctx.UserAgent(), "UserAgent should return the User-Agent header")

	// Test with nil Request
	ctx.Request = nil
	assert.Equal(t, "", ctx.UserAgent(), "UserAgent should return empty string when Request is nil")
}

// TestReferer tests the Referer method
func TestReferer(t *testing.T) {
	// Create a context with Referer header
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Referer", "http://example.com")
	res := httptest.NewRecorder()
	ctx := GetContext(res, req)

	// Check that Referer returns the Referer header
	assert.Equal(t, "http://example.com", ctx.Referer(), "Referer should return the Referer header")

	// Test with nil Request
	ctx.Request = nil
	assert.Equal(t, "", ctx.Referer(), "Referer should return empty string when Request is nil")
}

// TestHost tests the Host method
func TestHost(t *testing.T) {
	// Create a context with Host header
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Host = "example.com"
	res := httptest.NewRecorder()
	ctx := GetContext(res, req)

	// Check that Host returns the Host header
	assert.Equal(t, "example.com", ctx.Host(), "Host should return the Host header")

	// Create a context with X-Forwarded-Host header
	req, _ = http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-Host", "forwarded.example.com")
	res = httptest.NewRecorder()
	ctx = GetContext(res, req)

	// Check that Host returns the X-Forwarded-Host header
	assert.Equal(t, "forwarded.example.com", ctx.Host(), "Host should return the X-Forwarded-Host header")

	// Test with nil Request
	ctx.Request = nil
	assert.Equal(t, "", ctx.Host(), "Host should return empty string when Request is nil")
}

// TestProtocol tests the Protocol method
func TestProtocol(t *testing.T) {
	// Create a context with HTTPS scheme
	req, _ := http.NewRequest("GET", "https://example.com/test", nil)
	res := httptest.NewRecorder()
	ctx := GetContext(res, req)

	// Check that Protocol returns "https"
	assert.Equal(t, "https", ctx.Protocol(), "Protocol should return https")

	// Create a context with HTTP scheme
	req, _ = http.NewRequest("GET", "http://example.com/test", nil)
	res = httptest.NewRecorder()
	ctx = GetContext(res, req)

	// Check that Protocol returns "http"
	assert.Equal(t, "http", ctx.Protocol(), "Protocol should return http")

	// Create a context with X-Forwarded-Proto header
	req, _ = http.NewRequest("GET", "http://example.com/test", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	res = httptest.NewRecorder()
	ctx = GetContext(res, req)

	// Check that Protocol returns the X-Forwarded-Proto header
	assert.Equal(t, "https", ctx.Protocol(), "Protocol should return the X-Forwarded-Proto header")

	// Test with nil Request
	ctx.Request = nil
	assert.Equal(t, "", ctx.Protocol(), "Protocol should return empty string when Request is nil")
}

// TestStatus tests the Status method
func TestStatus(t *testing.T) {
	// Create a context
	req, _ := http.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := GetContext(res, req)

	// Set a status code
	returnedCtx := ctx.Status(StatusNotFound)

	// Check that the status code was set
	assert.Equal(t, StatusNotFound, ctx.statusCode, "Status should set the status code")

	// Check that Status returns the context for chaining
	assert.Equal(t, ctx, returnedCtx, "Status should return the context for chaining")
}

// TestSetGet tests the Set and Get methods
func TestSetGet(t *testing.T) {
	// Create a context
	req, _ := http.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := GetContext(res, req)

	// Set a header value
	returnedCtx := ctx.Set("X-Test", "test-value")

	// Check that the header was set
	assert.Equal(t, "test-value", ctx.Get("X-Test"), "Set should set the header value")

	// Check that Set returns the context for chaining
	assert.Equal(t, ctx, returnedCtx, "Set should return the context for chaining")

	// Check that Get returns the set value
	assert.Equal(t, "test-value", ctx.Get("X-Test"), "Get should return the set value")

	// Check that Get returns an empty string for non-existent keys
	assert.Equal(t, "", ctx.Get("Non-Existent"), "Get should return empty string for non-existent keys")
}

// TestQuery tests the Query method
func TestQuery(t *testing.T) {
	// Create a context with query parameters
	req, _ := http.NewRequest("GET", "/test?name=John&age=30", nil)
	res := httptest.NewRecorder()
	ctx := GetContext(res, req)

	// Check that Query returns the correct values
	assert.Equal(t, "John", ctx.Query("name"), "Query should return the correct value for name")
	assert.Equal(t, "30", ctx.Query("age"), "Query should return the correct value for age")

	// Check that Query returns an empty string for non-existent keys
	assert.Equal(t, "", ctx.Query("non-existent"), "Query should return empty string for non-existent keys")

	// Test with nil Request
	ctx.Request = nil
	assert.Equal(t, "", ctx.Query("name"), "Query should return empty string when Request is nil")
}

// TestQueryArray tests the QueryArray method
func TestQueryArray(t *testing.T) {
	// Create a context with query parameters
	req, _ := http.NewRequest("GET", "/test?name=John&name=Jane&age=30", nil)
	res := httptest.NewRecorder()
	ctx := GetContext(res, req)

	// Check that QueryArray returns the correct values
	nameValues := ctx.QueryArray("name")
	assert.Equal(t, 2, len(nameValues), "QueryArray should return the correct number of values")
	assert.Equal(t, "John", nameValues[0], "QueryArray should return the correct first value")
	assert.Equal(t, "Jane", nameValues[1], "QueryArray should return the correct second value")

	// Check that QueryArray returns an empty slice for non-existent keys
	nonExistentValues := ctx.QueryArray("non-existent")
	assert.Empty(t, nonExistentValues, "QueryArray should return empty slice for non-existent keys")

	// Test with nil Request
	ctx.Request = nil
	nilValues := ctx.QueryArray("name")
	assert.Empty(t, nilValues, "QueryArray should return empty slice when Request is nil")
}

// TestString tests the String method
func TestString(t *testing.T) {
	// Create a context
	req, _ := http.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := GetContext(res, req)

	// Set a string response
	ctx.String("Hello, %s!", "World")

	// Check that the response body was set
	assert.Equal(t, "Hello, World!", string(ctx.body), "String should set the response body")

	// Check that the Content-Type header was set
	assert.Equal(t, "text/plain; charset=utf-8", ctx.Get("Content-Type"), "String should set the Content-Type header")
}

// TestJSON tests the JSON method
func TestJSON(t *testing.T) {
	// Create a context
	req, _ := http.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := GetContext(res, req)

	// Set a JSON response
	type TestStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	data := TestStruct{Name: "John", Age: 30}
	ctx.JSON(data)

	// Check that the response body was set
	expectedJSON := `{"name":"John","age":30}`
	assert.Equal(t, expectedJSON, string(ctx.body), "JSON should set the response body")

	// Check that the Content-Type header was set
	assert.Equal(t, "application/json; charset=utf-8", ctx.Get("Content-Type"), "JSON should set the Content-Type header")
}

// TestHTML tests the HTML method
func TestHTML(t *testing.T) {
	// Create a context
	req, _ := http.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := GetContext(res, req)

	// Set an HTML response
	ctx.HTML("<h1>Hello, World!</h1>")

	// Check that the response body was set
	assert.Equal(t, "<h1>Hello, World!</h1>", string(ctx.body), "HTML should set the response body")

	// Check that the Content-Type header was set
	assert.Equal(t, "text/html; charset=utf-8", ctx.Get("Content-Type"), "HTML should set the Content-Type header")
}

// TestData tests the Data method
func TestData(t *testing.T) {
	// Create a context
	req, _ := http.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := GetContext(res, req)

	// Set a binary response
	data := []byte{0x01, 0x02, 0x03, 0x04}
	ctx.Data("application/octet-stream", data)

	// Check that the response body was set
	assert.Equal(t, data, ctx.body, "Data should set the response body")

	// Check that the Content-Type header was set
	assert.Equal(t, "application/octet-stream", ctx.Get("Content-Type"), "Data should set the Content-Type header")
}

// TestUserData tests the UserData method
func TestUserData(t *testing.T) {
	// Create a context
	req, _ := http.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := GetContext(res, req)

	// Set user data
	ctx.UserData("key", "value")

	// Check that the user data was set
	assert.Equal(t, "value", ctx.UserData("key"), "UserData should set and return the user data")

	// Check that UserData returns nil for non-existent keys
	assert.Nil(t, ctx.UserData("non-existent"), "UserData should return nil for non-existent keys")

	// Test with nil userData map
	ctx.userData = nil
	assert.Nil(t, ctx.UserData("key"), "UserData should return nil when userData map is nil")

	// Test setting a value with nil userData map
	ctx.UserData("new-key", "new-value")
	assert.Equal(t, "new-value", ctx.UserData("new-key"), "UserData should create userData map when nil")
}

// TestNext tests the Next method
func TestNext(t *testing.T) {
	// Create a context
	req, _ := http.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := GetContext(res, req)

	// Set up middleware stack
	middlewareCalled := false
	handlerCalled := false

	middleware := func(c *Ctx) {
		middlewareCalled = true
		c.Next()
	}

	handler := func(c *Ctx) {
		handlerCalled = true
	}

	ctx.middlewareStack = []MiddlewareFunc{middleware}
	ctx.handler = handler
	ctx.middlewareIndex = -1

	// Call Next
	ctx.Next()

	// Check that middleware and handler were called
	assert.True(t, middlewareCalled, "Middleware should be called")
	assert.True(t, handlerCalled, "Handler should be called")

	// Test with empty middleware stack
	ctx = GetContext(res, req)
	handlerCalled = false
	ctx.middlewareStack = []MiddlewareFunc{}
	ctx.handler = handler
	ctx.middlewareIndex = -1

	// Call Next
	ctx.Next()

	// Check that handler was called
	assert.True(t, handlerCalled, "Handler should be called with empty middleware stack")

	// Test with nil handler
	ctx = GetContext(res, req)
	ctx.middlewareStack = []MiddlewareFunc{}
	ctx.handler = nil
	ctx.middlewareIndex = -1

	// Call Next (should not panic)
	assert.NotPanics(t, func() { ctx.Next() }, "Next should not panic with nil handler")
}

// TestGetContextReleaseContext tests the GetContext and ReleaseContext functions
func TestGetContextReleaseContext(t *testing.T) {
	// Create a request and response
	req, _ := http.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()

	// Get a context
	ctx := GetContext(res, req)

	// Check that the context is properly initialized
	assert.NotNil(t, ctx, "GetContext should not return nil")
	assert.NotNil(t, ctx.Writer, "ctx.Writer should not be nil")
	assert.NotNil(t, ctx.Request, "ctx.Request should not be nil")
	assert.Equal(t, StatusOK, ctx.statusCode, "ctx.statusCode should be StatusOK")
	assert.Nil(t, ctx.err, "ctx.err should be nil")
	assert.Equal(t, -1, ctx.middlewareIndex, "ctx.middlewareIndex should be -1")

	// Set some values on the context
	ctx.statusCode = StatusNotFound
	ctx.err = errors.New("test error")
	ctx.Set("X-Test", "test-value")
	ctx.body = append(ctx.body, []byte("test body")...)
	ctx.middlewareStack = append(ctx.middlewareStack, func(c *Ctx) {})
	ctx.middlewareIndex = 0
	ctx.handler = func(c *Ctx) {}

	// Release the context
	ReleaseContext(ctx)

	// Get another context (should be the same one from the pool)
	ctx2 := GetContext(res, req)

	// Check that the context was reset
	assert.Equal(t, StatusOK, ctx2.statusCode, "ctx2.statusCode should be StatusOK")
	assert.Nil(t, ctx2.err, "ctx2.err should be nil")
	assert.Equal(t, 0, len(ctx2.Request.Header), "ctx2.header should be empty")
	assert.Empty(t, ctx2.body, "ctx2.body should be empty")
	assert.Empty(t, ctx2.middlewareStack, "ctx2.middlewareStack should be empty")
	assert.Equal(t, -1, ctx2.middlewareIndex, "ctx2.middlewareIndex should be -1")
	assert.Nil(t, ctx2.handler, "ctx2.handler should be nil")

	// Release the second context
	ReleaseContext(ctx2)
}

// TestCopyHeadersToWriter tests the copyHeadersToWriter method
func TestCopyHeadersToWriter(t *testing.T) {
	// Create a context
	req, _ := http.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := GetContext(res, req)

	// Set some headers on the context
	ctx.Set("X-Test-1", "test-value-1")
	ctx.Set("X-Test-2", "test-value-2")

	// Set some headers on the writer
	ctx.Writer.Header().Set("X-Test-3", "test-value-3")

	// Copy headers to writer
	ctx.copyHeadersToWriter()

	// Check that all headers were copied to the writer
	assert.Equal(t, "test-value-1", ctx.Writer.Header().Get("X-Test-1"), "Header X-Test-1 should be copied to writer")
	assert.Equal(t, "test-value-2", ctx.Writer.Header().Get("X-Test-2"), "Header X-Test-2 should be copied to writer")
	assert.Equal(t, "test-value-3", ctx.Writer.Header().Get("X-Test-3"), "Header X-Test-3 should remain in writer")

	// Test with nil Writer
	ctx.Writer = nil
	// This should not panic
	assert.NotPanics(t, func() { ctx.copyHeadersToWriter() }, "copyHeadersToWriter should not panic with nil Writer")
}

// TestWrite tests the write method
func TestWrite(t *testing.T) {
	// Create a context
	req, _ := http.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := GetContext(res, req)

	// Write some data
	data := []byte("test data")
	n, err := ctx.write(data)

	// Check that the write was successful
	assert.NoError(t, err, "write should not return an error")
	assert.Equal(t, len(data), n, "write should return the correct number of bytes written")

	// Check that the data was written to the body
	assert.Equal(t, data, ctx.body, "data should be written to the body")

	// Write more data
	moreData := []byte(" more data")
	n, err = ctx.write(moreData)

	// Check that the write was successful
	assert.NoError(t, err, "write should not return an error")
	assert.Equal(t, len(moreData), n, "write should return the correct number of bytes written")

	// Check that the data was appended to the body
	expectedData := append(data, moreData...)
	assert.Equal(t, expectedData, ctx.body, "data should be appended to the body")
}

// TestParam tests the Param method
func TestParam(t *testing.T) {
	// Create a context
	req, _ := http.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := GetContext(res, req)

	// Test with no parameters
	assert.Equal(t, "", ctx.Param("id"), "Param should return empty string when no parameters")

	// Test with nil Request
	ctx.Request = nil
	assert.Equal(t, "", ctx.Param("id"), "Param should return empty string when Request is nil")

	// Create a request with parameters
	req, _ = http.NewRequest("GET", "/users/123", nil)
	res = httptest.NewRecorder()
	ctx = GetContext(res, req)

	// Manually set up the parameter context
	paramCtx := make(map[paramKey]string)
	paramCtx[paramKey("id")] = "123"
	ctx.Request.SetContext(context.WithValue(ctx.Request.Context(), paramContextKey{}, paramCtx))

	// Test getting a parameter
	assert.Equal(t, "123", ctx.Param("id"), "Param should return the parameter value")

	// Test getting a non-existent parameter
	assert.Equal(t, "", ctx.Param("name"), "Param should return empty string for non-existent parameters")
}

// TestGetParam tests the GetParam method
func TestGetParam(t *testing.T) {
	// Create a context
	req, _ := http.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := GetContext(res, req)

	// Test with no parameters
	assert.Equal(t, "", ctx.GetParam("id"), "GetParam should return empty string when no parameters")

	// Test with nil Request
	ctx.Request = nil
	assert.Equal(t, "", ctx.GetParam("id"), "GetParam should return empty string when Request is nil")

	// Create a request with parameters
	req, _ = http.NewRequest("GET", "/users/123", nil)
	res = httptest.NewRecorder()
	ctx = GetContext(res, req)

	// Manually set up the parameter context
	paramCtx := make(map[paramKey]string)
	paramCtx[paramKey("id")] = "123"
	ctx.Request.SetContext(context.WithValue(ctx.Request.Context(), paramContextKey{}, paramCtx))

	// Test getting a parameter
	assert.Equal(t, "123", ctx.GetParam("id"), "GetParam should return the parameter value")

	// Test getting a non-existent parameter
	assert.Equal(t, "", ctx.GetParam("name"), "GetParam should return empty string for non-existent parameters")
}
