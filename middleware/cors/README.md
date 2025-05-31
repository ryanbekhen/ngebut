# CORS Middleware for Ngebut

This middleware implements Cross-Origin Resource Sharing (CORS) support for Ngebut applications.

## Usage

### Basic Usage with Default Configuration

```go
package main

import (
    "github.com/ryanbekhen/ngebut"
    "github.com/ryanbekhen/ngebut/middleware/cors"
)

func main() {
    app := ngebut.New()

    // Use CORS middleware with default configuration
    app.Use(cors.New())

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
    "github.com/ryanbekhen/ngebut/middleware/cors"
)

func main() {
    app := ngebut.New()

    // Use CORS middleware with custom configuration
    app.Use(cors.New(cors.Config{
        AllowOrigins:     "https://example.com,https://api.example.com",
        AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
        AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
        ExposeHeaders:    "Content-Length",
        AllowCredentials: true,
        MaxAge:           86400, // 24 hours
    }))

    app.GET("/", func(c *ngebut.Ctx) {
        c.String("Hello, World!")
    })

    app.Listen(":3000")
}
```

## Configuration Options

- `AllowOrigins`: A comma-separated list of origins a cross-domain request can be executed from. Default: `"*"`
- `AllowMethods`: A comma-separated list of methods the client is allowed to use with cross-domain requests. Default: `"GET,POST,PUT,DELETE,HEAD,OPTIONS,PATCH"`
- `AllowHeaders`: A comma-separated list of non-simple headers the client is allowed to use with cross-domain requests. Default: `""`
- `ExposeHeaders`: A comma-separated list of headers which are safe to expose to the API of a CORS API specification. Default: `""`
- `AllowCredentials`: Indicates whether the request can include user credentials like cookies, HTTP authentication or client side SSL certificates. Default: `false`
- `MaxAge`: Indicates how long (in seconds) the results of a preflight request can be cached. Default: `0` (no caching)

