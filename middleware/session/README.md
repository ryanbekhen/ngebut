# Session Middleware for Ngebut

This middleware implements session management for Ngebut applications, providing a way to store and retrieve user data across multiple requests.

## Usage

### Basic Usage with Default Configuration

```go
package main

import (
    "github.com/ryanbekhen/ngebut"
    "github.com/ryanbekhen/ngebut/middleware/session"
)

func main() {
    app := ngebut.New()

    // Use Session middleware with default configuration
    app.Use(session.NewMiddleware())

    app.GET("/", func(c *ngebut.Ctx) {
        // Get the session
        sess := session.GetSession(c)
        
        // Use the session
        sess.Set("visited", true)
        
        c.String("Hello, World!")
    })

    app.Listen(":3000")
}
```

### Custom Configuration

```go
package main

import (
    "time"
    "github.com/ryanbekhen/ngebut"
    "github.com/ryanbekhen/ngebut/middleware/session"
)

func main() {
    app := ngebut.New()

    // Use Session middleware with custom configuration
    app.Use(session.NewMiddleware(session.Config{
        Expiration: 12 * time.Hour,
        KeyLookup:  "cookie:custom_session_id",
        Path:       "/api",
        Secure:     true,
        HttpOnly:   true,
    }))

    app.GET("/", func(c *ngebut.Ctx) {
        // Get the session
        sess := session.GetSession(c)
        
        // Store data in the session
        sess.Set("user_id", 123)
        
        c.String("Hello, World!")
    })

    app.Listen(":3000")
}
```

## Configuration Options

- `Expiration`: Duration after which the session will expire. Default: `24 * time.Hour`
- `KeyLookup`: Format of where to look for the session ID. Format: "source:name" where source can be "cookie", "header", or "query". Default: `"cookie:session_id"`
- `KeyGenerator`: Function that generates a new session ID. Default: `UUIDv4`
- `Path`: The cookie path. Default: `"/"`
- `Domain`: The cookie domain. Default: `""`
- `Secure`: Indicates if the cookie should only be sent over HTTPS. Default: `false`
- `HttpOnly`: Indicates if the cookie should only be accessible via HTTP(S) requests. Default: `true`
- `Storage`: The storage backend for sessions. Default: In-memory storage

## Session Methods

- `Set(key string, value interface{})`: Stores a value in the session
- `Get(key string) interface{}`: Retrieves a value from the session
- `Delete(key string)`: Removes a value from the session
- `Clear()`: Removes all values from the session
- `Keys() []string`: Returns all keys in the session
- `Destroy(c ...*ngebut.Ctx) error`: Destroys the session
- `SetExpiry(expiry time.Duration)`: Sets a specific expiration time for the session
- `Save() error`: Saves the session to the store