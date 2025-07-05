package cors

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ryanbekhen/ngebut"
	"github.com/stretchr/testify/assert"
)

// TestCORSMiddlewareE2E tests the CORS middleware in an end-to-end scenario
func TestCORSMiddlewareE2E(t *testing.T) {
	// Test cases for different CORS scenarios
	testCases := []struct {
		name           string
		config         Config
		method         string
		origin         string
		expectedOrigin string
		expectedStatus int
	}{
		{
			name:           "Default config with allowed origin",
			config:         DefaultConfig(),
			method:         "GET",
			origin:         "http://example.com",
			expectedOrigin: "*",
			expectedStatus: http.StatusOK,
		},
		{
			name: "Custom config with specific allowed origin",
			config: Config{
				AllowOrigins:     "http://example.com",
				AllowMethods:     "GET,POST",
				AllowHeaders:     "Content-Type,Authorization",
				ExposeHeaders:    "X-Custom-Header",
				AllowCredentials: true,
				MaxAge:           3600,
			},
			method:         "GET",
			origin:         "http://example.com",
			expectedOrigin: "http://example.com",
			expectedStatus: http.StatusOK,
		},
		{
			name: "Custom config with disallowed origin",
			config: Config{
				AllowOrigins: "http://allowed.com",
			},
			method:         "GET",
			origin:         "http://disallowed.com",
			expectedOrigin: "",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test HTTP request
			req, _ := http.NewRequest(tc.method, "http://example.com/test", nil)
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}
			w := httptest.NewRecorder()

			// Create a test context
			ctx := ngebut.GetContext(w, req)

			// Create the middleware with the test case config
			middleware := New(tc.config)

			// Define a handler that will be called after the middleware
			handler := func(c *ngebut.Ctx) {
				c.Status(tc.expectedStatus)
				c.String("%s", "OK")
			}

			// Call the middleware followed by the handler
			middleware(ctx)
			handler(ctx)

			// Get the response
			resp := w.Result()
			defer resp.Body.Close()

			// Verify the response status
			assert.Equal(t, tc.expectedStatus, resp.StatusCode, "Unexpected status code")

			// Verify CORS headers
			actualOrigin := resp.Header.Get("Access-Control-Allow-Origin")
			assert.Equal(t, tc.expectedOrigin, actualOrigin, "Unexpected Access-Control-Allow-Origin header")

			if tc.config.AllowCredentials {
				assert.Equal(t, "true", resp.Header.Get("Access-Control-Allow-Credentials"), "Unexpected Access-Control-Allow-Credentials header")
			}

			if tc.config.ExposeHeaders != "" {
				assert.Equal(t, tc.config.ExposeHeaders, resp.Header.Get("Access-Control-Expose-Headers"), "Unexpected Access-Control-Expose-Headers header")
			}
		})
	}
}

// TestCORSPreflightE2E tests the CORS middleware with preflight requests
func TestCORSPreflightE2E(t *testing.T) {
	// Create a custom config
	config := Config{
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

	// Create the middleware with the custom config
	middleware := New(config)

	// Call the middleware (for preflight requests, the middleware will end the request)
	middleware(ctx)

	// Get the response
	resp := w.Result()
	defer resp.Body.Close()

	// Debug output
	t.Logf("Response status code: %d", resp.StatusCode)
	t.Logf("Response headers: %v", resp.Header)

	// Check if the status code was set in the context
	t.Logf("Context status code: %d", ctx.StatusCode())

	// Verify the response status for preflight
	assert.Equal(t, http.StatusNoContent, resp.StatusCode, "Unexpected status code for preflight request")

	// Verify CORS headers for preflight
	assert.Equal(t, "http://example.com", resp.Header.Get("Access-Control-Allow-Origin"), "Unexpected Access-Control-Allow-Origin header")
	assert.Equal(t, "GET,POST", resp.Header.Get("Access-Control-Allow-Methods"), "Unexpected Access-Control-Allow-Methods header")
	assert.Equal(t, "Content-Type,Authorization", resp.Header.Get("Access-Control-Allow-Headers"), "Unexpected Access-Control-Allow-Headers header")
	assert.Equal(t, "true", resp.Header.Get("Access-Control-Allow-Credentials"), "Unexpected Access-Control-Allow-Credentials header")
	assert.Equal(t, "3600", resp.Header.Get("Access-Control-Max-Age"), "Unexpected Access-Control-Max-Age header")
}
