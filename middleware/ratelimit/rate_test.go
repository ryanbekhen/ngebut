package ratelimit

import (
	"testing"
	"time"

	"golang.org/x/time/rate"
)

// TestRateEvery tests the rate.Every function to understand how it works
func TestRateEvery(t *testing.T) {
	// Test different durations and requests
	testCases := []struct {
		duration time.Duration
		requests int
		desc     string
	}{
		{time.Minute, 60, "1 minute / 60 requests = 1 request per second"},
		{time.Minute, 1, "1 minute / 1 request = 1 request per minute"},
		{time.Second, 2, "1 second / 2 requests = 2 requests per second"},
		{500 * time.Millisecond, 1, "500ms / 1 request = 1 request per 500ms"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			// Calculate the rate limit
			rateLimit := rate.Every(tc.duration / time.Duration(tc.requests))

			// Create a limiter with the calculated rate and a burst of 1
			limiter := rate.NewLimiter(rateLimit, 1)

			// Check if the first request is allowed
			allowed1 := limiter.Allow()
			t.Logf("First request allowed: %v", allowed1)

			// Check if the second request is allowed immediately
			allowed2 := limiter.Allow()
			t.Logf("Second request allowed immediately: %v", allowed2)

			// Wait for half the expected interval
			halfInterval := tc.duration / time.Duration(tc.requests) / 2
			t.Logf("Waiting for half interval: %v", halfInterval)
			time.Sleep(halfInterval)

			// Check if a request is allowed after waiting for half the interval
			allowed3 := limiter.Allow()
			t.Logf("Request allowed after half interval: %v", allowed3)

			// Wait for the full expected interval
			fullInterval := tc.duration / time.Duration(tc.requests)
			t.Logf("Waiting for full interval: %v", fullInterval)
			time.Sleep(fullInterval)

			// Check if a request is allowed after waiting for the full interval
			allowed4 := limiter.Allow()
			t.Logf("Request allowed after full interval: %v", allowed4)
		})
	}
}

// TestRateLimiterBehavior tests the behavior of the rate limiter with different configurations
func TestRateLimiterBehavior(t *testing.T) {
	// Test the behavior of the rate limiter with the default config
	defaultConfig := DefaultConfig()
	t.Logf("Default config: %+v", defaultConfig)

	// Create a rate limiter with the default config
	defaultLimiter := NewVisitor(defaultConfig)

	// Check if the first request is allowed
	allowed1 := defaultLimiter.Allow()
	t.Logf("First request with default config allowed: %v", allowed1)

	// Check if the second request is allowed immediately
	allowed2 := defaultLimiter.Allow()
	t.Logf("Second request with default config allowed immediately: %v", allowed2)

	// Wait for the rate limit to reset
	t.Logf("Waiting for rate limit to reset: %v", defaultConfig.Duration)
	time.Sleep(defaultConfig.Duration)

	// Check if a request is allowed after waiting
	allowed3 := defaultLimiter.Allow()
	t.Logf("Request with default config allowed after waiting: %v", allowed3)

	// Test the behavior of the rate limiter with a custom config
	customConfig := Config{
		Requests:  3,
		Burst:     1,
		Duration:  time.Second,
		ExpiresIn: time.Minute,
	}
	t.Logf("Custom config: %+v", customConfig)

	// Create a rate limiter with the custom config
	customLimiter := NewVisitor(customConfig)

	// Check if the first request is allowed
	allowed4 := customLimiter.Allow()
	t.Logf("First request with custom config allowed: %v", allowed4)

	// Check if the second request is allowed immediately
	allowed5 := customLimiter.Allow()
	t.Logf("Second request with custom config allowed immediately: %v", allowed5)

	// Check if the third request is allowed immediately
	allowed6 := customLimiter.Allow()
	t.Logf("Third request with custom config allowed immediately: %v", allowed6)

	// Check if the fourth request is allowed immediately
	allowed7 := customLimiter.Allow()
	t.Logf("Fourth request with custom config allowed immediately: %v", allowed7)

	// Wait for the rate limit to reset
	t.Logf("Waiting for rate limit to reset: %v", customConfig.Duration)
	time.Sleep(customConfig.Duration)

	// Check if a request is allowed after waiting
	allowed8 := customLimiter.Allow()
	t.Logf("Request with custom config allowed after waiting: %v", allowed8)
}

// TestRateLimiterCorrection tests a corrected version of the NewVisitor function
func TestRateLimiterCorrection(t *testing.T) {
	// Define a corrected version of the NewVisitor function
	newVisitorCorrected := func(cfg Config) *rate.Limiter {
		// Calculate requests per second
		requestsPerSecond := float64(cfg.Requests) / cfg.Duration.Seconds()
		// Create a limiter with the calculated rate and the specified burst
		return rate.NewLimiter(rate.Limit(requestsPerSecond), cfg.Burst)
	}

	// Test the corrected function with different configs
	testCases := []struct {
		config Config
		desc   string
	}{
		{DefaultConfig(), "Default config"},
		{Config{Requests: 3, Burst: 1, Duration: time.Second, ExpiresIn: time.Minute}, "3 requests per second"},
		{Config{Requests: 1, Burst: 0, Duration: 500 * time.Millisecond, ExpiresIn: time.Minute}, "1 request per 500ms"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			// Create a limiter with the corrected function
			limiter := newVisitorCorrected(tc.config)

			// Check if the first request is allowed
			allowed1 := limiter.Allow()
			t.Logf("First request allowed: %v", allowed1)

			// Make multiple requests to test the rate limiting
			for i := 0; i < tc.config.Requests+tc.config.Burst; i++ {
				allowed := limiter.Allow()
				t.Logf("Request %d allowed: %v", i+2, allowed)
			}

			// Wait for the rate limit to reset
			t.Logf("Waiting for rate limit to reset: %v", tc.config.Duration)
			time.Sleep(tc.config.Duration)

			// Check if a request is allowed after waiting
			allowed2 := limiter.Allow()
			t.Logf("Request allowed after waiting: %v", allowed2)
		})
	}
}
