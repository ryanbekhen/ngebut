# Access Log Middleware for Ngebut

This middleware implements HTTP request logging for Ngebut applications, providing detailed access logs with customizable formats.

## Usage

### Basic Usage with Default Configuration

```go
package main

import (
    "github.com/ryanbekhen/ngebut"
    "github.com/ryanbekhen/ngebut/middleware/accesslog"
)

func main() {
    app := ngebut.New()

    // Use AccessLog middleware with default configuration
    app.Use(accesslog.New())

    app.GET("/", func(c *ngebut.Ctx) {
        c.String("Hello, World!")
    })

    app.Listen(":3000")
}
```

### Custom Configuration

```go
package main

import (
    "github.com/ryanbekhen/ngebut"
    "github.com/ryanbekhen/ngebut/middleware/accesslog"
)

func main() {
    app := ngebut.New()

    // Use AccessLog middleware with custom format
    app.Use(accesslog.New(accesslog.Config{
        Format: "${time} | ${remote_ip} | ${method} ${path}${query} | ${status} | ${latency_human} | ${user_agent}",
    }))

    app.GET("/", func(c *ngebut.Ctx) {
        c.String("Hello, World!")
    })

    app.Listen(":3000")
}
```

## Configuration Options

- `Format`: The format string for the access log. Default: `"${time} | ${status} | ${latency_human} | ${method} ${path} | ${error}"`

## Available Format Placeholders

The following placeholders can be used in the format string:

- `${remote_ip}` - The client's IP address
- `${method}` - The HTTP method (GET, POST, etc.)
- `${path}` - The request path
- `${status}` - The HTTP status code
- `${latency}` - The request latency (raw duration)
- `${latency_human}` - The request latency in human-readable format (e.g., "1.20ms")
- `${bytes_in}` - The number of bytes received
- `${user_agent}` - The User-Agent header
- `${referer}` - The Referer header
- `${time}` - The current time in the format "2006-01-02 15:04:05"
- `${query}` - The URL query string
- `${error}` - The error message if an error occurred during request processing

## Log Levels

The middleware automatically uses appropriate log levels based on the response status code:

- Info level for 2xx and 3xx status codes
- Warn level for 4xx status codes
- Error level for 5xx status codes