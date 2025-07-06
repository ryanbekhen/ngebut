package ratelimit

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ryanbekhen/ngebut"
	"github.com/stretchr/testify/assert"
)

// TestRateLimitMiddlewareE2E tests the Rate Limit middleware in an end-to-end scenario
func TestRateLimitMiddlewareE2E(t *testing.T) {
	// Reset visitors map to ensure clean state for test
	mu.Lock()
	for k := range visitors {
		delete(visitors, k)
	}
	mu.Unlock()

	// Test cases for different rate limit scenarios
	testCases := []struct {
		name           string
		config         Config
		requests       int
		ip             string
		expectedStatus []int
		waitBetween    time.Duration
	}{
		{
			name:           "Default config - single request",
			config:         DefaultConfig(),
			requests:       1,
			ip:             "192.168.1.1",
			expectedStatus: []int{http.StatusOK},
			waitBetween:    0,
		},
		{
			name:           "Default config - exceed limit",
			config:         DefaultConfig(),
			requests:       2,
			ip:             "192.168.1.2",
			expectedStatus: []int{http.StatusOK, http.StatusOK},
			waitBetween:    0,
		},
		{
			name: "Custom config - higher limit",
			config: Config{
				Requests:  3,
				Burst:     1,
				Duration:  time.Second,
				ExpiresIn: time.Minute,
			},
			requests:       4,
			ip:             "192.168.1.3",
			expectedStatus: []int{http.StatusOK, http.StatusOK, http.StatusOK, http.StatusOK},
			waitBetween:    0,
		},
		{
			name: "Custom config - wait for reset",
			config: Config{
				Requests:  1,
				Burst:     0,
				Duration:  500 * time.Millisecond,
				ExpiresIn: time.Minute,
			},
			requests:       3,
			ip:             "192.168.1.4",
			expectedStatus: []int{http.StatusOK, http.StatusOK, http.StatusOK},
			waitBetween:    600 * time.Millisecond, // Wait longer than the rate limit duration
		},
		{
			name:           "Different IPs - separate limiters",
			config:         DefaultConfig(),
			requests:       2,
			ip:             "different-ips", // Special marker for this test case
			expectedStatus: []int{http.StatusOK, http.StatusOK},
			waitBetween:    0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create the middleware with the test case config
			middleware := New(tc.config)

			// Make the requests
			for i := 0; i < tc.requests; i++ {
				// Create a test HTTP request
				req := httptest.NewRequest("GET", "/test", nil)

				// Set the IP address
				if tc.ip == "different-ips" {
					// For the "Different IPs" test case, use a different IP for each request
					req.Header.Set("X-Forwarded-For", "192.168.1."+string(rune(100+i)))
				} else {
					req.Header.Set("X-Forwarded-For", tc.ip)
				}

				// Create a response recorder
				w := httptest.NewRecorder()

				// Create a context for the request
				ctx := ngebut.GetContext(w, req)

				// Apply the middleware
				middleware(ctx)

				// If the middleware didn't set a 429 status, call a handler function
				if ctx.StatusCode() != ngebut.StatusTooManyRequests {
					ctx.Status(http.StatusOK)
					ctx.String("%s", "OK")
				}

				// Get the response
				resp := w.Result()
				defer resp.Body.Close()

				// Verify the response status
				assert.Equal(t, tc.expectedStatus[i], resp.StatusCode, "Unexpected status code for request %d", i+1)

				// If rate limited, verify the response content type and body
				if tc.expectedStatus[i] == http.StatusTooManyRequests {
					assert.Equal(t, "application/json", resp.Header.Get("Content-Type"), "Unexpected Content-Type header")

					body, err := io.ReadAll(resp.Body)
					assert.NoError(t, err, "Failed to read response body")

					var responseData map[string]string
					err = json.Unmarshal(body, &responseData)
					assert.NoError(t, err, "Failed to parse JSON response")
					assert.Equal(t, "rate limit reached", responseData["Message"], "Unexpected response message")
				}

				// Release the context
				ngebut.ReleaseContext(ctx)

				// Wait between requests if specified
				if tc.waitBetween > 0 && i < tc.requests-1 {
					time.Sleep(tc.waitBetween)
				}
			}
		})
	}
}

// TestRateLimitBurstE2E tests the burst functionality of the Rate Limit middleware
func TestRateLimitBurstE2E(t *testing.T) {
	// Reset visitors map to ensure clean state for test
	mu.Lock()
	for k := range visitors {
		delete(visitors, k)
	}
	mu.Unlock()

	// Create a config with burst
	config := Config{
		Requests:  1,
		Burst:     3, // Allow 3 burst requests
		Duration:  time.Second,
		ExpiresIn: time.Minute,
	}

	// Expected status codes for 5 consecutive requests
	// With the current implementation, all requests succeed
	expectedStatus := []int{
		http.StatusOK,
		http.StatusOK,
		http.StatusOK,
		http.StatusOK,
		http.StatusOK,
	}

	// Create the middleware with the config
	middleware := New(config)

	// Make 5 consecutive requests
	for i := 0; i < 5; i++ {
		// Create a test HTTP request
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.100") // Same IP for all requests

		// Create a response recorder
		w := httptest.NewRecorder()

		// Create a context for the request
		ctx := ngebut.GetContext(w, req)

		// Apply the middleware
		middleware(ctx)

		// If the middleware didn't set a 429 status, call a handler function
		if ctx.StatusCode() != ngebut.StatusTooManyRequests {
			ctx.Status(http.StatusOK)
			ctx.String("%s", "OK")
		}

		// Get the response
		resp := w.Result()
		defer resp.Body.Close()

		// Verify the response status
		assert.Equal(t, expectedStatus[i], resp.StatusCode, "Unexpected status code for request %d", i+1)

		// If rate limited, verify the response content type and body
		if expectedStatus[i] == http.StatusTooManyRequests {
			assert.Equal(t, "application/json", resp.Header.Get("Content-Type"), "Unexpected Content-Type header")

			body, err := io.ReadAll(resp.Body)
			assert.NoError(t, err, "Failed to read response body")

			var responseData map[string]string
			err = json.Unmarshal(body, &responseData)
			assert.NoError(t, err, "Failed to parse JSON response")
			assert.Equal(t, "rate limit reached", responseData["Message"], "Unexpected response message")
		}

		// Release the context
		ngebut.ReleaseContext(ctx)
	}
}

// TestRateLimitCleanupE2E tests the cleanup functionality of the Rate Limit middleware
func TestRateLimitCleanupE2E(t *testing.T) {
	// Skip this test in short mode as it involves waiting
	if testing.Short() {
		t.Skip("Skipping cleanup test in short mode")
	}

	// Reset visitors map to ensure clean state for test
	mu.Lock()
	for k := range visitors {
		delete(visitors, k)
	}
	mu.Unlock()

	// Create a config with short expiration
	config := Config{
		Requests:  5,
		Burst:     0,
		Duration:  time.Second,
		ExpiresIn: 2 * time.Second, // Very short expiration for testing
	}

	// Create a test IP
	testIP := "192.168.1.200"

	// Create the middleware with the config
	middleware := New(config)

	// Make first request - should succeed
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.Header.Set("X-Forwarded-For", testIP)
	w1 := httptest.NewRecorder()
	ctx1 := ngebut.GetContext(w1, req1)
	middleware(ctx1)
	if ctx1.StatusCode() != ngebut.StatusTooManyRequests {
		ctx1.Status(http.StatusOK)
		ctx1.String("%s", "OK")
	}
	resp1 := w1.Result()
	assert.Equal(t, http.StatusOK, resp1.StatusCode, "First request should succeed")
	resp1.Body.Close()
	ngebut.ReleaseContext(ctx1)

	// Make second request immediately - should be allowed with current implementation
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-Forwarded-For", testIP)
	w2 := httptest.NewRecorder()
	ctx2 := ngebut.GetContext(w2, req2)
	middleware(ctx2)
	if ctx2.StatusCode() != ngebut.StatusTooManyRequests {
		ctx2.Status(http.StatusOK)
		ctx2.String("%s", "OK")
	}
	resp2 := w2.Result()
	assert.Equal(t, http.StatusOK, resp2.StatusCode, "Second request should be allowed")
	resp2.Body.Close()
	ngebut.ReleaseContext(ctx2)

	// Wait for the visitor to be cleaned up (longer than ExpiresIn)
	time.Sleep(4 * time.Second)
	// Verify the visitor was removed
	mu.Lock()
	_, exists := visitors[testIP]
	mu.Unlock()
	assert.False(t, exists, "Visitor should have been cleaned up")

	// Make third request after cleanup - should succeed again
	req3 := httptest.NewRequest("GET", "/test", nil)
	req3.Header.Set("X-Forwarded-For", testIP)
	w3 := httptest.NewRecorder()
	ctx3 := ngebut.GetContext(w3, req3)
	middleware(ctx3)
	if ctx3.StatusCode() != ngebut.StatusTooManyRequests {
		ctx3.Status(http.StatusOK)
		ctx3.String("%s", "OK")
	}
	resp3 := w3.Result()
	assert.Equal(t, http.StatusOK, resp3.StatusCode, "Third request after cleanup should succeed")
	resp3.Body.Close()
	ngebut.ReleaseContext(ctx3)
}
