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
	limiter  *rate.Limiter
	lastSeen time.Time
}

var ErrLimiter = ngebut.NewHttpError(ngebut.StatusTooManyRequests, "limit reached")

var (
	visitors = make(map[string]*Visitor)

	// mu is a Mutex used to synchronize access to the shared visitors map, ensuring thread-safe operations.
	mu sync.Mutex
)

// NewVisitor creates a new limiter for a given config
func NewVisitor(cfg Config) *rate.Limiter {
	rateLimit := rate.Every(cfg.Duration / time.Duration(cfg.Requests))
	return rate.NewLimiter(rateLimit, cfg.Burst)
}

// CleanupVisitors removes stale visitors
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

// GetVisitor gets or creates a visitor limiter
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

func DefaultConfig() Config {
	return Config{
		Requests:  1,
		Burst:     5,
		Duration:  time.Minute,
		ExpiresIn: time.Hour,
	}
}

// New with custom config
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
