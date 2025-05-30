package accesslog

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ryanbekhen/ngebut"
	"github.com/ryanbekhen/ngebut/log"
	"github.com/stretchr/testify/assert"
)

// TestNew tests the New function
func TestNew(t *testing.T) {
	// Test with default config
	middleware := New()
	assert.NotNil(t, middleware, "New() returned nil")

	// Test with custom config
	customConfig := Config{
		Format: "${method} ${path}",
	}
	middleware = New(customConfig)
	assert.NotNil(t, middleware, "New(customConfig) returned nil")
}

// TestDefaultConfig tests the DefaultConfig function
func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	assert.NotEmpty(t, config.Format, "DefaultConfig() returned empty Format")
	assert.Equal(t, "${time} | ${status} | ${latency_human} | ${method} ${path} | ${error}", config.Format, "DefaultConfig() returned unexpected Format")
}

// TestLogger tests the logger initialization
func TestLogger(t *testing.T) {
	// Save the original logger to restore it later
	originalLogger := logger
	defer func() {
		logger = originalLogger
	}()

	// Create a custom logger with a buffer for testing
	buf := &bytes.Buffer{}
	testLogger := log.New(buf, log.InfoLevel)
	logger = testLogger

	// Verify that the logger is set correctly
	assert.Equal(t, testLogger, logger, "Logger was not set correctly")
}

// TestHelperFunctions tests the helper functions
func TestHelperFunctions(t *testing.T) {
	// Test replaceTag
	msg := "Hello ${name}!"
	result := replaceTag(msg, "${name}", "World")
	assert.Equal(t, "Hello World!", result, "replaceTag returned incorrect result")

	// Test intToString
	result = intToString(123)
	assert.Equal(t, "123", result, "intToString returned incorrect result")

	// Test int64ToString
	result = int64ToString(int64(9223372036854775807))
	assert.Equal(t, "9223372036854775807", result, "int64ToString returned incorrect result")
}

// TestMiddlewareBasic tests the basic functionality of the middleware
func TestMiddlewareBasic(t *testing.T) {
	// Save the original logger to restore it later
	originalLogger := logger
	defer func() {
		logger = originalLogger
	}()

	// Create a buffer to capture log output
	buf := &bytes.Buffer{}
	testLogger := log.New(buf, log.InfoLevel)
	logger = testLogger

	// Create a test HTTP request and response writer
	req, _ := http.NewRequest("GET", "http://example.com/test?query=value", nil)
	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set("Referer", "http://example.com")
	w := httptest.NewRecorder()

	// Create a test context
	ctx := ngebut.GetContext(w, req)

	// Create the middleware
	middleware := New().(func(*ngebut.Ctx))

	// Call the middleware
	middleware(ctx)

	// Flush the response
	ctx.Writer.Flush()

	// Check that something was logged
	logOutput := buf.String()
	assert.NotEmpty(t, logOutput, "No log output was produced")

	// Check that the log contains expected information
	assert.Contains(t, logOutput, "GET", "Log output doesn't contain HTTP method")
	assert.Contains(t, logOutput, "/test", "Log output doesn't contain request path")
	assert.Contains(t, logOutput, "200", "Log output doesn't contain status code")
}

// TestMiddlewareWithError tests the middleware with an error
func TestMiddlewareWithError(t *testing.T) {
	// Save the original logger to restore it later
	originalLogger := logger
	defer func() {
		logger = originalLogger
	}()

	// Create a buffer to capture log output
	buf := &bytes.Buffer{}
	testLogger := log.New(buf, log.InfoLevel)
	logger = testLogger

	// Create a test HTTP request and response writer
	req, _ := http.NewRequest("GET", "http://example.com/test", nil)
	w := httptest.NewRecorder()

	// Create a test context
	ctx := ngebut.GetContext(w, req)

	// Set an error in the context
	testError := errors.New("test error")
	ctx.Error(testError)

	// Create the middleware
	middleware := New().(func(*ngebut.Ctx))

	// Call the middleware
	middleware(ctx)

	// Flush the response
	ctx.Writer.Flush()

	// Check that the log contains the error
	logOutput := buf.String()
	assert.Contains(t, logOutput, "test error", "Log output doesn't contain the error message")
}

// TestMiddlewareStatusCodes tests the middleware with different status codes
func TestMiddlewareStatusCodes(t *testing.T) {
	testCases := []struct {
		name       string
		statusCode int
		logLevel   string
	}{
		{"Success", ngebut.StatusOK, "INFO"},
		{"Redirection", ngebut.StatusFound, "INFO"},
		{"ClientError", ngebut.StatusBadRequest, "WARN"},
		{"ServerError", ngebut.StatusInternalServerError, "ERROR"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Save the original logger to restore it later
			originalLogger := logger
			defer func() {
				logger = originalLogger
			}()

			// Create a buffer to capture log output
			buf := &bytes.Buffer{}
			testLogger := log.New(buf, log.DebugLevel)
			logger = testLogger

			// Create a test HTTP request and response writer
			req, _ := http.NewRequest("GET", "http://example.com/test", nil)
			w := httptest.NewRecorder()

			// Create a test context
			ctx := ngebut.GetContext(w, req)

			// Set the status code
			ctx.Status(tc.statusCode)

			// Create the middleware
			middleware := New().(func(*ngebut.Ctx))

			// Call the middleware
			middleware(ctx)

			// Flush the response
			ctx.Writer.Flush()

			// Check that the log contains the status code
			logOutput := buf.String()
			statusStr := strconv.Itoa(tc.statusCode)
			assert.Contains(t, logOutput, statusStr, "Log output doesn't contain status code "+statusStr)

			// Check the log level
			assert.Contains(t, logOutput, tc.logLevel, "Log output doesn't contain expected log level "+tc.logLevel)
		})
	}
}

// TestMiddlewareCustomFormat tests the middleware with a custom format
func TestMiddlewareCustomFormat(t *testing.T) {
	// Save the original logger to restore it later
	originalLogger := logger
	defer func() {
		logger = originalLogger
	}()

	// Create a buffer to capture log output
	buf := &bytes.Buffer{}
	testLogger := log.New(buf, log.InfoLevel)
	logger = testLogger

	// Create a test HTTP request and response writer
	req, _ := http.NewRequest("GET", "http://example.com/test?param=value", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	req.ContentLength = 100
	w := httptest.NewRecorder()

	// Create a test context
	ctx := ngebut.GetContext(w, req)

	// Set headers on the context
	ctx.Set("User-Agent", "test-agent")
	ctx.Set("Referer", "http://example.com/referer")

	// Create the middleware with custom format
	customFormat := "${remote_ip} ${method} ${path} ${query} ${bytes_in} ${user_agent} ${referer}"
	middleware := New(Config{Format: customFormat}).(func(*ngebut.Ctx))

	// Call the middleware
	middleware(ctx)

	// Flush the response
	ctx.Writer.Flush()

	// Check that the log contains all the expected placeholders
	logOutput := buf.String()
	expectedValues := []string{
		"192.168.1.1",                // remote_ip
		"GET",                        // method
		"/test",                      // path
		"param=value",                // query
		"100",                        // bytes_in
		"test-agent",                 // user_agent
		"http://example.com/referer", // referer
	}

	for _, val := range expectedValues {
		assert.Contains(t, logOutput, val, "Log output doesn't contain expected value: "+val)
	}
}

// TestMiddlewareLatency tests the latency reporting in the middleware
func TestMiddlewareLatency(t *testing.T) {
	// Save the original logger to restore it later
	originalLogger := logger
	defer func() {
		logger = originalLogger
	}()

	// Create a buffer to capture log output
	buf := &bytes.Buffer{}
	testLogger := log.New(buf, log.InfoLevel)
	logger = testLogger

	// Create a test HTTP request
	req, _ := http.NewRequest("GET", "http://example.com/test", nil)

	// Create a test response recorder
	w := httptest.NewRecorder()

	// Create a middleware with a format that includes latency
	middleware := New(Config{Format: "${latency} ${latency_human}"}).(func(*ngebut.Ctx))

	// Create a handler that simulates processing time
	handlerCalled := false
	handler := func(c *ngebut.Ctx) {
		handlerCalled = true
		// Simulate processing time
		time.Sleep(10 * time.Millisecond)
		c.Status(ngebut.StatusOK).String("OK")
	}

	// Create a test context
	ctx := ngebut.GetContext(w, req)

	// Call the middleware
	middleware(ctx)

	// Call the handler directly
	handler(ctx)

	// Flush the response
	ctx.Writer.Flush()

	// Check that the handler was called
	assert.True(t, handlerCalled, "Handler was not called")

	// Check that something was logged
	logOutput := buf.String()
	assert.NotEmpty(t, logOutput, "No log output was produced")

	// Check that the log contains latency information
	assert.True(t,
		strings.Contains(logOutput, "ns") ||
			strings.Contains(logOutput, "µs") ||
			strings.Contains(logOutput, "ms"),
		"Log output doesn't contain latency information (ns, µs, or ms)")
}
