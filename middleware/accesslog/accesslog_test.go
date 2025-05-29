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
)

// TestNew tests the New function
func TestNew(t *testing.T) {
	// Test with default config
	middleware := New()
	if middleware == nil {
		t.Fatal("New() returned nil")
	}

	// Test with custom config
	customConfig := Config{
		Format: "${method} ${path}",
	}
	middleware = New(customConfig)
	if middleware == nil {
		t.Fatal("New(customConfig) returned nil")
	}
}

// TestDefaultConfig tests the DefaultConfig function
func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	if config.Format == "" {
		t.Error("DefaultConfig() returned empty Format")
	}
	if config.Format != "${time} | ${status} | ${latency_human} | ${method} ${path} | ${error}" {
		t.Errorf("DefaultConfig() returned unexpected Format: %s", config.Format)
	}
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
	if logger != testLogger {
		t.Error("Logger was not set correctly")
	}
}

// TestHelperFunctions tests the helper functions
func TestHelperFunctions(t *testing.T) {
	// Test replaceTag
	msg := "Hello ${name}!"
	result := replaceTag(msg, "${name}", "World")
	if result != "Hello World!" {
		t.Errorf("replaceTag returned %q, expected %q", result, "Hello World!")
	}

	// Test intToString
	result = intToString(123)
	if result != "123" {
		t.Errorf("intToString returned %q, expected %q", result, "123")
	}

	// Test int64ToString
	result = int64ToString(int64(9223372036854775807))
	if result != "9223372036854775807" {
		t.Errorf("int64ToString returned %q, expected %q", result, "9223372036854775807")
	}
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
	if logOutput == "" {
		t.Error("No log output was produced")
	}

	// Check that the log contains expected information
	if !strings.Contains(logOutput, "GET") {
		t.Error("Log output doesn't contain HTTP method")
	}
	if !strings.Contains(logOutput, "/test") {
		t.Error("Log output doesn't contain request path")
	}
	if !strings.Contains(logOutput, "200") {
		t.Error("Log output doesn't contain status code")
	}
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
	if !strings.Contains(logOutput, "test error") {
		t.Error("Log output doesn't contain the error message")
	}
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
			if !strings.Contains(logOutput, statusStr) {
				t.Errorf("Log output doesn't contain status code %s", statusStr)
			}

			// Check the log level
			if !strings.Contains(logOutput, tc.logLevel) {
				t.Errorf("Log output doesn't contain expected log level %s", tc.logLevel)
			}
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
		if !strings.Contains(logOutput, val) {
			t.Errorf("Log output doesn't contain expected value: %s", val)
		}
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
	if !handlerCalled {
		t.Error("Handler was not called")
	}

	// Check that something was logged
	logOutput := buf.String()
	if logOutput == "" {
		t.Error("No log output was produced")
	}

	// Print the log output for debugging
	t.Logf("Log output: %s", logOutput)

	// Check that the log contains latency information
	if !strings.Contains(logOutput, "ns") && !strings.Contains(logOutput, "µs") && !strings.Contains(logOutput, "ms") {
		t.Error("Log output doesn't contain latency information (ns, µs, or ms)")
	}
}
