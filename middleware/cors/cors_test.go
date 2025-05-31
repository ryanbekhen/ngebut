package cors

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ryanbekhen/ngebut"
	"github.com/stretchr/testify/assert"
)

// TestDefaultConfig tests the DefaultConfig function
func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, "*", config.AllowOrigins, "DefaultConfig() returned unexpected AllowOrigins")
	assert.Equal(t, "GET,POST,PUT,DELETE,HEAD,OPTIONS,PATCH", config.AllowMethods, "DefaultConfig() returned unexpected AllowMethods")
	assert.Equal(t, "", config.AllowHeaders, "DefaultConfig() returned unexpected AllowHeaders")
	assert.Equal(t, "", config.ExposeHeaders, "DefaultConfig() returned unexpected ExposeHeaders")
	assert.False(t, config.AllowCredentials, "DefaultConfig() returned unexpected AllowCredentials value")
	assert.Equal(t, 0, config.MaxAge, "DefaultConfig() returned unexpected MaxAge")
}

// TestNew tests the New function
func TestNew(t *testing.T) {
	// Test with default config
	middleware := New()
	assert.NotNil(t, middleware, "New() returned nil")

	// Test with custom config
	customConfig := Config{
		AllowOrigins:     "http://example.com",
		AllowMethods:     "GET,POST",
		AllowHeaders:     "Content-Type,Authorization",
		ExposeHeaders:    "X-Custom-Header",
		AllowCredentials: true,
		MaxAge:           3600,
	}
	middleware = New(customConfig)
	assert.NotNil(t, middleware, "New(customConfig) returned nil")
}

// TestCORSMiddlewareWithDefaultConfig tests the CORS middleware with default configuration
func TestCORSMiddlewareWithDefaultConfig(t *testing.T) {
	// Create a test HTTP request with Origin header
	req, _ := http.NewRequest("GET", "http://example.com/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()

	// Create a test context
	ctx := ngebut.GetContext(w, req)

	// Create the middleware with default config
	middleware := New()

	// Call the middleware
	middleware(ctx)

	// Check that CORS headers were set correctly
	assert.Equal(t, "*", ctx.Get("Access-Control-Allow-Origin"), "Unexpected Access-Control-Allow-Origin header")
}

// TestCORSMiddlewareWithCustomConfig tests the CORS middleware with custom configuration
func TestCORSMiddlewareWithCustomConfig(t *testing.T) {
	// Create a custom config
	customConfig := Config{
		AllowOrigins:     "http://example.com",
		AllowMethods:     "GET,POST",
		AllowHeaders:     "Content-Type,Authorization",
		ExposeHeaders:    "X-Custom-Header",
		AllowCredentials: true,
		MaxAge:           3600,
	}

	// Create a test HTTP request with Origin header
	req, _ := http.NewRequest("GET", "http://example.com/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()

	// Create a test context
	ctx := ngebut.GetContext(w, req)

	// Create the middleware with custom config
	middleware := New(customConfig)

	// Call the middleware
	middleware(ctx)

	// Check that CORS headers were set correctly
	assert.Equal(t, "http://example.com", ctx.Get("Access-Control-Allow-Origin"), "Unexpected Access-Control-Allow-Origin header")
	assert.Equal(t, "Origin", ctx.Get("Vary"), "Unexpected Vary header")
	assert.Equal(t, "X-Custom-Header", ctx.Get("Access-Control-Expose-Headers"), "Unexpected Access-Control-Expose-Headers header")
	assert.Equal(t, "true", ctx.Get("Access-Control-Allow-Credentials"), "Unexpected Access-Control-Allow-Credentials header")
}

// TestCORSMiddlewareWithDisallowedOrigin tests the CORS middleware with a disallowed origin
func TestCORSMiddlewareWithDisallowedOrigin(t *testing.T) {
	// Create a custom config with specific allowed origins
	customConfig := Config{
		AllowOrigins: "http://allowed.com",
	}

	// Create a test HTTP request with a disallowed Origin header
	req, _ := http.NewRequest("GET", "http://example.com/test", nil)
	req.Header.Set("Origin", "http://disallowed.com")
	w := httptest.NewRecorder()

	// Create a test context
	ctx := ngebut.GetContext(w, req)

	// Create the middleware with custom config
	middleware := New(customConfig)

	// Call the middleware
	middleware(ctx)

	// Check that CORS headers were set correctly (should be empty for disallowed origin)
	assert.Equal(t, "", ctx.Get("Access-Control-Allow-Origin"), "Unexpected Access-Control-Allow-Origin header")
	assert.Equal(t, "Origin", ctx.Get("Vary"), "Unexpected Vary header")
}

// TestCORSMiddlewareWithNoOrigin tests the CORS middleware with no Origin header
func TestCORSMiddlewareWithNoOrigin(t *testing.T) {
	// Create a test HTTP request with no Origin header
	req, _ := http.NewRequest("GET", "http://example.com/test", nil)
	w := httptest.NewRecorder()

	// Create a test context
	ctx := ngebut.GetContext(w, req)

	// Create the middleware with default config
	middleware := New()

	// Call the middleware
	middleware(ctx)

	// Check that no CORS headers were set
	assert.Equal(t, "", ctx.Get("Access-Control-Allow-Origin"), "Unexpected Access-Control-Allow-Origin header")
}

// TestCORSMiddlewareWithPreflightRequest tests the CORS middleware with a preflight OPTIONS request
func TestCORSMiddlewareWithPreflightRequest(t *testing.T) {
	// Create a custom config
	customConfig := Config{
		AllowOrigins:     "http://example.com",
		AllowMethods:     "GET,POST",
		AllowHeaders:     "Content-Type,Authorization",
		ExposeHeaders:    "X-Custom-Header",
		AllowCredentials: true,
		MaxAge:           3600,
	}

	// Create a test HTTP preflight request
	req, _ := http.NewRequest("OPTIONS", "http://example.com/test", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")
	w := httptest.NewRecorder()

	// Create a test context
	ctx := ngebut.GetContext(w, req)

	// Create the middleware with custom config
	middleware := New(customConfig)

	// Call the middleware
	middleware(ctx)

	// Check that preflight CORS headers were set correctly
	assert.Equal(t, "http://example.com", ctx.Get("Access-Control-Allow-Origin"), "Unexpected Access-Control-Allow-Origin header")
	assert.Equal(t, "GET,POST", ctx.Get("Access-Control-Allow-Methods"), "Unexpected Access-Control-Allow-Methods header")
	assert.Equal(t, "Content-Type,Authorization", ctx.Get("Access-Control-Allow-Headers"), "Unexpected Access-Control-Allow-Headers header")
	assert.Equal(t, "true", ctx.Get("Access-Control-Allow-Credentials"), "Unexpected Access-Control-Allow-Credentials header")
	assert.Equal(t, "3600", ctx.Get("Access-Control-Max-Age"), "Unexpected Access-Control-Max-Age header")

	// Note: In a real application, the status code would be 204 for preflight requests,
	// but in the test environment, we're not checking the status code directly
	// because it requires additional response writing that would complicate the tests.
}

// TestCORSMiddlewareWithPreflightRequestNoAllowHeaders tests the CORS middleware with a preflight OPTIONS request when no AllowHeaders are specified
func TestCORSMiddlewareWithPreflightRequestNoAllowHeaders(t *testing.T) {
	// Create a custom config with no AllowHeaders
	customConfig := Config{
		AllowOrigins:     "http://example.com",
		AllowMethods:     "GET,POST",
		AllowCredentials: true,
		MaxAge:           3600,
	}

	// Create a test HTTP preflight request
	req, _ := http.NewRequest("OPTIONS", "http://example.com/test", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type, Authorization")
	w := httptest.NewRecorder()

	// Create a test context
	ctx := ngebut.GetContext(w, req)

	// Create the middleware with custom config
	middleware := New(customConfig)

	// Call the middleware
	middleware(ctx)

	// Check that preflight CORS headers were set correctly
	// When no AllowHeaders are specified, the middleware should mirror the requested headers
	assert.Equal(t, "Content-Type, Authorization", ctx.Get("Access-Control-Allow-Headers"), "Unexpected Access-Control-Allow-Headers header")
}

// TestCORSMiddlewareWithWildcardOrigin tests the CORS middleware with wildcard origin
func TestCORSMiddlewareWithWildcardOrigin(t *testing.T) {
	// Create a test HTTP request with Origin header
	req, _ := http.NewRequest("GET", "http://example.com/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()

	// Create a test context
	ctx := ngebut.GetContext(w, req)

	// Create the middleware with default config (which has wildcard origin)
	middleware := New()

	// Call the middleware
	middleware(ctx)

	// Check that CORS headers were set correctly
	assert.Equal(t, "*", ctx.Get("Access-Control-Allow-Origin"), "Unexpected Access-Control-Allow-Origin header")
	// No Vary header should be set with wildcard origin
	assert.Equal(t, "", ctx.Get("Vary"), "Unexpected Vary header")
}

// TestCORSMiddlewareWithMultipleAllowedOrigins tests the CORS middleware with multiple allowed origins
func TestCORSMiddlewareWithMultipleAllowedOrigins(t *testing.T) {
	// Create a custom config with multiple allowed origins
	customConfig := Config{
		AllowOrigins: "http://example1.com,http://example2.com",
	}

	// Test cases for different origins
	testCases := []struct {
		name           string
		origin         string
		expectedOrigin string
		expectVary     bool
	}{
		{"AllowedOrigin1", "http://example1.com", "http://example1.com", true},
		{"AllowedOrigin2", "http://example2.com", "http://example2.com", true},
		{"DisallowedOrigin", "http://example3.com", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test HTTP request with Origin header
			req, _ := http.NewRequest("GET", "http://example.com/test", nil)
			req.Header.Set("Origin", tc.origin)
			w := httptest.NewRecorder()

			// Create a test context
			ctx := ngebut.GetContext(w, req)

			// Create the middleware with custom config
			middleware := New(customConfig)

			// Call the middleware
			middleware(ctx)

			// Check that CORS headers were set correctly
			assert.Equal(t, tc.expectedOrigin, ctx.Get("Access-Control-Allow-Origin"), "Unexpected Access-Control-Allow-Origin header")
			if tc.expectVary {
				assert.Equal(t, "Origin", ctx.Get("Vary"), "Unexpected Vary header")
			}
		})
	}
}

// TestCORSMiddlewareWithAllowCredentials tests the CORS middleware with AllowCredentials
func TestCORSMiddlewareWithAllowCredentials(t *testing.T) {
	// Create a custom config with AllowCredentials
	customConfig := Config{
		AllowOrigins:     "http://example.com",
		AllowCredentials: true,
	}

	// Create a test HTTP request with Origin header
	req, _ := http.NewRequest("GET", "http://example.com/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()

	// Create a test context
	ctx := ngebut.GetContext(w, req)

	// Create the middleware with custom config
	middleware := New(customConfig)

	// Call the middleware
	middleware(ctx)

	// Check that CORS headers were set correctly
	assert.Equal(t, "true", ctx.Get("Access-Control-Allow-Credentials"), "Unexpected Access-Control-Allow-Credentials header")
}

// TestCORSMiddlewareWithExposeHeaders tests the CORS middleware with ExposeHeaders
func TestCORSMiddlewareWithExposeHeaders(t *testing.T) {
	// Create a custom config with ExposeHeaders
	customConfig := Config{
		AllowOrigins:  "http://example.com",
		ExposeHeaders: "X-Custom-Header1,X-Custom-Header2",
	}

	// Create a test HTTP request with Origin header
	req, _ := http.NewRequest("GET", "http://example.com/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()

	// Create a test context
	ctx := ngebut.GetContext(w, req)

	// Create the middleware with custom config
	middleware := New(customConfig)

	// Call the middleware
	middleware(ctx)

	// Check that CORS headers were set correctly
	assert.Equal(t, "X-Custom-Header1,X-Custom-Header2", ctx.Get("Access-Control-Expose-Headers"), "Unexpected Access-Control-Expose-Headers header")
}

// TestCORSMiddlewareWithMaxAge tests the CORS middleware with MaxAge
func TestCORSMiddlewareWithMaxAge(t *testing.T) {
	// Create a custom config with MaxAge
	customConfig := Config{
		AllowOrigins: "http://example.com",
		MaxAge:       3600,
	}

	// Create a test HTTP preflight request
	req, _ := http.NewRequest("OPTIONS", "http://example.com/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()

	// Create a test context
	ctx := ngebut.GetContext(w, req)

	// Create the middleware with custom config
	middleware := New(customConfig)

	// Call the middleware
	middleware(ctx)

	// Check that CORS headers were set correctly
	assert.Equal(t, "3600", ctx.Get("Access-Control-Max-Age"), "Unexpected Access-Control-Max-Age header")

	// Note: In a real application, the status code would be 204 for preflight requests,
	// but in the test environment, we're not checking the status code directly
	// because it requires additional response writing that would complicate the tests.
}

// TestCORSMiddlewareWithAllowHeadersWildcard tests the CORS middleware with wildcard in AllowHeaders
func TestCORSMiddlewareWithAllowHeadersWildcard(t *testing.T) {
	// Create a custom config with wildcard in AllowMethods
	customConfig := Config{
		AllowOrigins: "http://example.com",
		AllowHeaders: "*",
	}

	// Create a test HTTP preflight request
	req, _ := http.NewRequest("OPTIONS", "http://example.com/test", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Headers", "X-Custom-Header")
	w := httptest.NewRecorder()

	// Create a test context
	ctx := ngebut.GetContext(w, req)

	// Create the middleware with custom config
	middleware := New(customConfig)

	// Call the middleware
	middleware(ctx)

	// Check that CORS headers were set correctly
	assert.Equal(t, "*", ctx.Get("Access-Control-Allow-Headers"), "Unexpected Access-Control-Allow-Headers header")
}

// TestCORSMiddlewareWithAllowMethodsWildcard tests the CORS middleware with wildcard in AllowMethods
func TestCORSMiddlewareWithAllowMethodsWildcard(t *testing.T) {
	// Create a custom config with wildcard in AllowMethods
	customConfig := Config{
		AllowOrigins: "http://example.com",
		AllowMethods: "*",
	}

	// Create a test HTTP preflight request
	req, _ := http.NewRequest("OPTIONS", "http://example.com/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()

	// Create a test context
	ctx := ngebut.GetContext(w, req)

	// Create the middleware with custom config
	middleware := New(customConfig)

	// Call the middleware
	middleware(ctx)

	// Check that CORS headers were set correctly
	assert.Equal(t, "*", ctx.Get("Access-Control-Allow-Methods"), "Unexpected Access-Control-Allow-Methods header")
}
