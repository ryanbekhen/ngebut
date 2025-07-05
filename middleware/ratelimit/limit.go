package ratelimit

import (
	"github.com/ryanbekhen/ngebut"
	"golang.org/x/time/rate"
	"sync"
	"time"
)

// Config holds the configuration settings for rate limiting, such as requests per duration, burst size, and expiration time.
type Config struct {
	Requests  int           // Max requests per duration
	Burst     int           // Burst size
	Duration  time.Duration // Duration window (e.g., 1 minute)
	ExpiresIn time.Duration // Visitor entry expiration
}

// Visitor represents a client with a rate limiter and the last recorded activity time.
type Visitor struct {
	limiter  *rate.Limiter // The rate limiter instance for the visitor
	lastSeen time.Time     // The last time this visitor was seen
}

// ErrLimiter is the default HTTP error returned when a client exceeds the rate limit.
var ErrLimiter = ngebut.NewHttpError(ngebut.StatusTooManyRequests, "limit reached")

var (
	// visitors stores the active visitors and their associated rate limiters.
	visitors = make(map[string]*Visitor)

	// mu is a Mutex used to synchronize access to the shared visitors map,
	// ensuring thread-safe operations.
	mu sync.Mutex
)

// NewVisitor creates and returns a new rate limiter instance
// based on the provided configuration.
func NewVisitor(cfg Config) *rate.Limiter {
	rateLimit := rate.Every(cfg.Duration / time.Duration(cfg.Requests))
	return rate.NewLimiter(rateLimit, cfg.Burst)
}

// CleanupVisitors periodically removes stale visitor entries
// from the visitors map after they exceed the specified expiration duration.
func CleanupVisitors(expiresIn time.Duration) {
	for {
		time.Sleep(time.Minute)
		mu.Lock()
		for ip, v := range visitors {
			if time.Since(v.lastSeen) > expiresIn {
				delete(visitors, ip)
			}
		}
		mu.Unlock()
	}
}

// GetVisitor retrieves the rate limiter for a given IP address.
// If the visitor does not exist, a new one is created using the provided config.
func GetVisitor(ip string, cfg Config) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	v, exists := visitors[ip]
	if !exists {
		limiter := NewVisitor(cfg)
		visitors[ip] = &Visitor{limiter, time.Now()}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

// DefaultConfig returns a Config object with default rate limiting settings:
// 1 request, burst of 5, a 1-minute duration window, and a 1-hour expiration time.
func DefaultConfig() Config {
	return Config{
		Requests:  1,
		Burst:     5,
		Duration:  time.Minute,
		ExpiresIn: time.Hour,
	}
}

// New creates and returns rate limiting middleware for the Ngebut framework.
// It accepts an optional custom Config; if none is provided, DefaultConfig is used.
func New(config ...Config) func(c *ngebut.Ctx) error {

	cfg := DefaultConfig()
	if len(config) > 0 {
		cfg = config[0]
		go CleanupVisitors(config[0].ExpiresIn)
	}

	return func(c *ngebut.Ctx) error {
		ip := c.IP()
		limiter := GetVisitor(ip, cfg)

		if !limiter.Allow() {
			message := map[string]interface{}{
				"Message": "rate limited reached",
			}
			c.Status(ngebut.StatusTooManyRequests).JSON(message)
			return ErrLimiter
		}

		c.Next()
		return nil
	}
}
