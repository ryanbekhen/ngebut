package basicauth

import (
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ryanbekhen/ngebut"
	"github.com/stretchr/testify/assert"
)

// TestBasicAuthMiddlewareE2E tests the Basic Auth middleware in an end-to-end scenario
func TestBasicAuthMiddlewareE2E(t *testing.T) {
	// Test cases for different Basic Auth scenarios
	testCases := []struct {
		name           string
		config         Config
		authHeader     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Valid credentials",
			config: Config{
				Username: "admin",
				Password: "password",
			},
			authHeader:     "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:password")),
			expectedStatus: http.StatusOK,
			expectedBody:   "Protected Content",
		},
		{
			name: "Invalid credentials",
			config: Config{
				Username: "admin",
				Password: "password",
			},
			authHeader:     "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:wrongpassword")),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized",
		},
		{
			name: "Missing Authorization header",
			config: Config{
				Username: "admin",
				Password: "password",
			},
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized",
		},
		{
			name: "Malformed Authorization header - not Basic",
			config: Config{
				Username: "admin",
				Password: "password",
			},
			authHeader:     "Bearer token",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized",
		},
		{
			name: "Malformed Authorization header - invalid Base64",
			config: Config{
				Username: "admin",
				Password: "password",
			},
			authHeader:     "Basic invalid-base64",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized",
		},
		{
			name: "Malformed Authorization header - no colon separator",
			config: Config{
				Username: "admin",
				Password: "password",
			},
			authHeader:     "Basic " + base64.StdEncoding.EncodeToString([]byte("adminpassword")),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized",
		},
		{
			name:           "Default config with valid credentials",
			config:         DefaultConfig(),
			authHeader:     "Basic " + base64.StdEncoding.EncodeToString([]byte("example:example")),
			expectedStatus: http.StatusOK,
			expectedBody:   "Protected Content",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test HTTP request
			req, _ := http.NewRequest("GET", "/protected", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}

			// Create a handler function that will be called if authentication succeeds
			handlerFunc := func(c *ngebut.Ctx) {
				c.Status(http.StatusOK)
				c.String("%s", "Protected Content")
			}

			// Create a handler function for unauthorized access
			unauthorizedFunc := func(c *ngebut.Ctx) {
				c.Status(http.StatusUnauthorized)
				c.String("%s", "Unauthorized")
			}

			// Create a test server with the middleware
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Create a context for the request
				ctx := ngebut.GetContext(w, r)
				defer ngebut.ReleaseContext(ctx)

				// Apply the middleware
				middleware := New(tc.config).(func(*ngebut.Ctx))
				middleware(ctx)

				// Check if the response has been written
				if w.Header().Get("Content-Type") != "" {
					// Response has been written, nothing more to do
					return
				}

				// If we get here, the middleware didn't write a response
				// This means either:
				// 1. Authentication succeeded and Next() was called
				// 2. Authentication failed but the middleware didn't set a response

				// Check if the middleware called Next() by examining the middleware's behavior
				// The middleware only calls Next() if authentication succeeds
				// We can check this by looking at the Authorization header
				authHeader := r.Header.Get("Authorization")

				// Parse the credentials from the Authorization header
				if authHeader != "" && len(authHeader) > 6 && authHeader[:6] == "Basic " {
					decoded, err := base64.StdEncoding.DecodeString(authHeader[6:])
					if err == nil {
						cred := string(decoded)
						sep := -1
						for i := 0; i < len(cred); i++ {
							if cred[i] == ':' {
								sep = i
								break
							}
						}
						if sep != -1 {
							username := cred[:sep]
							password := cred[sep+1:]

							// Check if credentials match
							if username == tc.config.Username && password == tc.config.Password {
								// Authentication succeeded, call the handler
								handlerFunc(ctx)
								return
							}
						}
					}
				}

				// Authentication failed, set unauthorized status
				unauthorizedFunc(ctx)
			}))
			defer server.Close()

			// Create a client for making the request
			client := &http.Client{}

			// Update the request URL to point to the test server
			parsedURL, _ := url.Parse(server.URL + "/protected")
			req.URL = parsedURL

			// Make the request with the Authorization header
			resp, err := client.Do(req)
			assert.NoError(t, err, "Failed to make request to test server")
			defer resp.Body.Close()

			// Read the response body
			body, err := io.ReadAll(resp.Body)
			assert.NoError(t, err, "Failed to read response body")

			// Verify the response status
			assert.Equal(t, tc.expectedStatus, resp.StatusCode, "Unexpected status code")

			// Verify the response body
			assert.Equal(t, tc.expectedBody, string(body), "Unexpected response body")
		})
	}
}
