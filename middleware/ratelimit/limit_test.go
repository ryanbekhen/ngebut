package ratelimit

import (
	"github.com/ryanbekhen/ngebut"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
	"time"
)

// newTestCtx creates a new test context with a specific IP
func newTestCtx(ip string) *ngebut.Ctx {
	req := httptest.NewRequest("GET", "/", nil)
	// Set X-Forwarded-For header to simulate the IP
	req.Header.Set("X-Forwarded-For", ip)
	rw := httptest.NewRecorder()
	ctx := ngebut.GetContext(rw, req)
	return ctx
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, 1, cfg.Requests)
	assert.Equal(t, 5, cfg.Burst)
	assert.Equal(t, time.Minute, cfg.Duration)
	assert.Equal(t, time.Hour, cfg.ExpiresIn)
}

func TestRateLimit(t *testing.T) {
	// Reset visitors map to ensure clean state for test
	mu.Lock()
	for k := range visitors {
		delete(visitors, k)
	}
	mu.Unlock()

	cfg := Config{
		Requests:  5,
		Burst:     1,
		Duration:  time.Second,
		ExpiresIn: time.Minute,
	}

	middleware := New(cfg)

	// Test with first request from IP "127.0.0.1" - should allow
	ctx1 := newTestCtx("127.0.0.1")
	err1 := middleware(ctx1)
	assert.Nil(t, err1, "First request should be allowed")

	// Test with second request from same IP - should be rate limited
	ctx2 := newTestCtx("127.0.0.1")
	err2 := middleware(ctx2)
	assert.Equal(t, ErrLimiter, err2, "Second request should be rate limited")
	assert.Equal(t, ngebut.StatusTooManyRequests, ctx2.StatusCode(), "Status code should be 429")

	// Wait for rate limit window to reset
	time.Sleep(1100 * time.Millisecond)

	// Test with third request after window reset - should allow again
	ctx3 := newTestCtx("127.0.0.1")
	err3 := middleware(ctx3)
	assert.Nil(t, err3, "Request after window reset should be allowed")

	// Test with different IP - should allow regardless of previous requests
	ctx4 := newTestCtx("192.168.1.1")
	err4 := middleware(ctx4)
	assert.Nil(t, err4, "Request from different IP should be allowed")
}
