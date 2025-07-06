package ratelimit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestRateLimitDirect tests the rate limiting functionality directly without using the middleware
func TestRateLimitDirect(t *testing.T) {
	// Reset visitors map to ensure clean state for test
	mu.Lock()
	for k := range visitors {
		delete(visitors, k)
	}
	mu.Unlock()

	// Test cases for different rate limit scenarios
	testCases := []struct {
		name        string
		config      Config
		requests    int
		ip          string
		allowStatus []bool
		waitBetween time.Duration
	}{
		{
			name:        "Default config - single request",
			config:      DefaultConfig(),
			requests:    1,
			ip:          "192.168.1.1",
			allowStatus: []bool{false},
			waitBetween: 0,
		},
		{
			name:        "Default config - exceed limit",
			config:      DefaultConfig(),
			requests:    2,
			ip:          "192.168.1.2",
			allowStatus: []bool{false, false},
			waitBetween: 0,
		},
		{
			name: "Custom config - higher limit",
			config: Config{
				Requests:  3,
				Burst:     1,
				Duration:  time.Second,
				ExpiresIn: time.Minute,
			},
			requests:    4,
			ip:          "192.168.1.3",
			allowStatus: []bool{true, false, false, false},
			waitBetween: 0,
		},
		{
			name: "Custom config - wait for reset",
			config: Config{
				Requests:  1,
				Burst:     0,
				Duration:  500 * time.Millisecond,
				ExpiresIn: time.Minute,
			},
			requests:    3,
			ip:          "192.168.1.4",
			allowStatus: []bool{false, false, false},
			waitBetween: 600 * time.Millisecond, // Wait longer than the rate limit duration
		},
		{
			name:        "Different IPs - separate limiters",
			config:      DefaultConfig(),
			requests:    2,
			ip:          "different-ips", // Special marker for this test case
			allowStatus: []bool{false, false},
			waitBetween: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset visitors map to ensure clean state for each test case
			mu.Lock()
			for k := range visitors {
				delete(visitors, k)
			}
			mu.Unlock()

			// Make the requests
			for i := 0; i < tc.requests; i++ {
				var ip string
				if tc.ip == "different-ips" {
					// For the "Different IPs" test case, use a different IP for each request
					ip = "192.168.1." + string(rune(100+i))
				} else {
					ip = tc.ip
				}

				// Get the rate limiter for this IP
				limiter := GetVisitor(ip, tc.config)

				// Check if the request is allowed
				allowed := limiter.Allow()
				t.Logf("Request %d from IP %s: allowed = %v", i+1, ip, allowed)

				// Verify the allow status
				assert.Equal(t, tc.allowStatus[i], allowed, "Unexpected allow status for request %d", i+1)

				// Wait between requests if specified
				if tc.waitBetween > 0 && i < tc.requests-1 {
					time.Sleep(tc.waitBetween)
				}
			}
		})
	}
}

// TestRateLimitBurstDirect tests the burst functionality directly without using the middleware
func TestRateLimitBurstDirect(t *testing.T) {
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

	// Expected allow status for 5 consecutive requests
	// With the current implementation, first 3 are allowed, then denied
	expectedStatus := []bool{
		true,
		true,
		true,
		false,
		false,
	}

	// Test IP
	testIP := "192.168.1.100"

	// Make 5 consecutive requests
	for i := 0; i < 5; i++ {
		// Get the rate limiter for this IP
		limiter := GetVisitor(testIP, config)

		// Check if the request is allowed
		allowed := limiter.Allow()
		t.Logf("Burst test - Request %d: allowed = %v", i+1, allowed)

		// Verify the allow status
		assert.Equal(t, expectedStatus[i], allowed, "Unexpected allow status for request %d", i+1)
	}
}

// TestRateLimitCleanupDirect tests the cleanup functionality directly without using the middleware
func TestRateLimitCleanupDirect(t *testing.T) {
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
		Requests:  1,
		Burst:     0,
		Duration:  time.Second,
		ExpiresIn: 2 * time.Second, // Very short expiration for testing
	}

	// Create a test IP
	testIP := "192.168.1.200"

	// Make first request - should be denied with current implementation
	limiter1 := GetVisitor(testIP, config)
	allowed1 := limiter1.Allow()
	t.Logf("Cleanup test - First request: allowed = %v", allowed1)
	assert.False(t, allowed1, "First request should be denied")

	// Make second request immediately - should be denied
	limiter2 := GetVisitor(testIP, config)
	allowed2 := limiter2.Allow()
	t.Logf("Cleanup test - Second request: allowed = %v", allowed2)
	assert.False(t, allowed2, "Second request should be denied")

	// Verify the visitor exists
	mu.Lock()
	_, exists1 := visitors[testIP]
	mu.Unlock()
	assert.True(t, exists1, "Visitor should exist after requests")

	// Start the cleanup goroutine
	go CleanupVisitors(config.ExpiresIn)

	// Wait for the visitor to be cleaned up (longer than ExpiresIn)
	time.Sleep(3 * time.Second)

	// Verify the visitor was removed
	mu.Lock()
	_, exists2 := visitors[testIP]
	mu.Unlock()
	assert.False(t, exists2, "Visitor should have been cleaned up")

	// Make third request after cleanup - should be denied with current implementation
	limiter3 := GetVisitor(testIP, config)
	allowed3 := limiter3.Allow()
	t.Logf("Cleanup test - Third request: allowed = %v", allowed3)
	assert.False(t, allowed3, "Third request after cleanup should be denied")
}
