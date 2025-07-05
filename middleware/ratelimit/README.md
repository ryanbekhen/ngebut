# 📦 RateLimit Middleware for ngebut

A flexible and efficient **Rate Limiting** middleware for the `ngebut` Go framework that helps protect your API from excessive requests.

---

## 🚀 Features

✅ IP-based rate limiting  
✅ Configurable request limits and burst capacity  
✅ Automatic cleanup of stale visitor records  
✅ Thread-safe implementation with mutex protection  
✅ Customizable time windows for rate limiting

---

## 📌 How It Works

- The middleware identifies clients by their IP address
- Each client is assigned a rate limiter with configured limits
- When a client exceeds their rate limit, a 429 Too Many Requests response is returned
- Stale visitor records are automatically cleaned up after the configured expiration time
- Uses the efficient `golang.org/x/time/rate` package for the core rate limiting logic

---

## ⚙️ Configuration

### ✅ Using Default Config

```go
package main

import (
	"github.com/ryanbekhen/ngebut"
	"github.com/ryanbekhen/ngebut/middleware/ratelimit"
)

func main() {
	app := ngebut.New()

	// Apply RateLimit middleware with default configuration
	app.Use(ratelimit.New())

	app.GET("/hello", func(c *ngebut.Ctx) error {
		return c.String("Hello, World!")
	})

	app.Listen(":3000")
}
```

### ✅ Using Custom Config

```go
package main

import (
	"github.com/ryanbekhen/ngebut"
	"github.com/ryanbekhen/ngebut/middleware/ratelimit"
	"time"
)

func main() {
	app := ngebut.New()

	// Apply RateLimit middleware with custom configuration
	app.Use(ratelimit.New(ratelimit.Config{
		Requests:  100,        // 100 requests
		Burst:     10,         // Allow burst of 10 requests
		Duration:  time.Hour,  // Per hour
		ExpiresIn: time.Hour * 24, // Keep visitor records for 24 hours
	}))

	app.GET("/hello", func(c *ngebut.Ctx) error {
		return c.String("Hello, World!")
	})

	app.Listen(":3000")
}
```

## 📋 Configuration Options

- `Requests`: Maximum number of requests allowed per duration. Default: `1`
- `Burst`: Maximum burst size (number of requests that can exceed the rate temporarily). Default: `5`
- `Duration`: Time window for rate limiting. Default: `time.Minute` (1 minute)
- `ExpiresIn`: How long visitor records are kept before cleanup. Default: `time.Hour` (1 hour)

## 🔄 Default Configuration

```go
Config{
	Requests:  1,
	Burst:     5,
	Duration:  time.Minute,
	ExpiresIn: time.Hour,
}
```

This default configuration allows 1 request per minute with a burst capacity of 5 requests.