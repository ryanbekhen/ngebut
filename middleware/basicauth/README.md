# 📦 BasicAuth Middleware for ngebut

A simple, secure, and reusable **Basic Authentication** middleware for the custom `ngebut` Go framework.

---

## 🚀 Features

✅ Lightweight and fast  
✅ Uses `crypto/subtle` for constant-time credential comparison (prevents timing attacks)  
✅ Easy to configure with default or custom credentials  
✅ Standards-compliant: `Authorization` header with `Basic` scheme and Base64 encoding

---

## 📌 How It Works

- The client must send the `Authorization` header in this format:

- This middleware decodes the Base64 string, splits it by `:`, and compares both username and password securely.
- If the credentials are valid, the request proceeds to the next handler.

---

## ⚙️ Configuration

### ✅ Using Custom Config

```go
package main

import (
	"github.com/ryanbekhen/ngebut"
	"github.com/ryanbekhen/ngebut/middleware/basicauth"
)

func main() {
	app := ngebut.New()

	// Apply BasicAuth middleware
	app.Use(basicauth.New(basicauth.Config{
		Username: "admin",
		Password: "secret",
	}))

	app.GET("/hello", func(c *ngebut.Ctx) error {
		return c.String("Hello, authenticated user!")
	})

	app.Listen(":3000")
}

```

## ⚙️ Configuration

### ✅ Using Default Config

```go
package main

import (
	"github.com/ryanbekhen/ngebut"
	"github.com/ryanbekhen/ngebut/middleware/basicauth"
)

func main() {
	app := ngebut.New()

	// Apply BasicAuth middleware
	app.Use(basicauth.New())

	app.GET("/hello", func(c *ngebut.Ctx) error {
		return c.String("Hello, authenticated user!")
	})

	app.Listen(":3000")
}

```
### Default config credential
```json
{
  "username": "example",
  "password": "example"
}
```