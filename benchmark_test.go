package ngebut

import (
	"net/http/httptest"
	"sync"
	"testing"
)

// Simple response struct for JSON benchmarks
type benchResponse struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
	Data    any    `json:"data,omitempty"`
}

// BenchmarkRouting benchmarks the routing performance
func BenchmarkRouting(b *testing.B) {
	// Setup
	server := New(DefaultConfig())
	server.GET("/", func(c *Ctx) {
		c.String("Hello, World!")
	})
	server.GET("/users", func(c *Ctx) {
		c.String("Users")
	})
	server.GET("/users/:id", func(c *Ctx) {
		c.String("User: %s", c.Param("id"))
	})
	server.GET("/users/:id/profile", func(c *Ctx) {
		c.String("Profile for user: %s", c.Param("id"))
	})
	server.GET("/api/v1/products", func(c *Ctx) {
		c.String("Products")
	})

	// Benchmark
	b.ResetTimer()
	b.Run("Static Route", func(b *testing.B) {
		req := httptest.NewRequest(MethodGet, "http://example.com/users", nil)
		w := getTestWriter()
		ctx := GetContext(w, req)

		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			server.router.ServeHTTP(ctx, ctx.Request)
			ReleaseContext(ctx)
			ctx = GetContext(w, req)
		}
	})

	b.Run("Param Route", func(b *testing.B) {
		req := httptest.NewRequest(MethodGet, "http://example.com/users/123", nil)
		w := getTestWriter()
		ctx := GetContext(w, req)

		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			server.router.ServeHTTP(ctx, ctx.Request)
			ReleaseContext(ctx)
			ctx = GetContext(w, req)
		}
	})

	b.Run("Nested Param Route", func(b *testing.B) {
		req := httptest.NewRequest(MethodGet, "http://example.com/users/123/profile", nil)
		w := getTestWriter()
		ctx := GetContext(w, req)

		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			server.router.ServeHTTP(ctx, ctx.Request)
			ReleaseContext(ctx)
			ctx = GetContext(w, req)
		}
	})

	b.Run("Deep Nested Route", func(b *testing.B) {
		req := httptest.NewRequest(MethodGet, "http://example.com/api/v1/products", nil)
		w := getTestWriter()
		ctx := GetContext(w, req)

		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			server.router.ServeHTTP(ctx, ctx.Request)
			ReleaseContext(ctx)
			ctx = GetContext(w, req)
		}
	})
}

// Pool for large JSON response objects
var largeResponsePool = sync.Pool{
	New: func() interface{} {
		// Pre-allocate a slice of 100 benchResponse objects
		data := make([]benchResponse, 100)
		for i := 0; i < 100; i++ {
			data[i] = benchResponse{
				Message: "Item " + string(rune(i)),
				Status:  200,
				Data: map[string]string{
					"field1": "value1",
					"field2": "value2",
					"field3": "value3",
				},
			}
		}
		return &data
	},
}

// BenchmarkResponses benchmarks different response types
func BenchmarkResponses(b *testing.B) {
	// Setup
	server := New(DefaultConfig())
	server.GET("/string", func(c *Ctx) {
		c.String("Hello, World!")
	})
	server.GET("/json", func(c *Ctx) {
		c.JSON(benchResponse{
			Message: "Hello, World!",
			Status:  200,
		})
	})
	server.GET("/json-large", func(c *Ctx) {
		// Get a pre-allocated response object from the pool
		data := largeResponsePool.Get().(*[]benchResponse)

		// Use the object
		c.JSON(*data)

		// Return the object to the pool for reuse
		largeResponsePool.Put(data)
	})
	server.GET("/html", func(c *Ctx) {
		c.HTML("<html><body><h1>Hello, World!</h1></body></html>")
	})

	// Benchmark
	b.ResetTimer()
	b.Run("String Response", func(b *testing.B) {
		req := httptest.NewRequest(MethodGet, "http://example.com/string", nil)
		w := getTestWriter()
		ctx := GetContext(w, req)

		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			server.router.ServeHTTP(ctx, ctx.Request)
			ReleaseContext(ctx)
			ctx = GetContext(w, req)
		}
	})

	b.Run("JSON Response", func(b *testing.B) {
		req := httptest.NewRequest(MethodGet, "http://example.com/json", nil)
		w := getTestWriter()
		ctx := GetContext(w, req)

		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			server.router.ServeHTTP(ctx, ctx.Request)
			ReleaseContext(ctx)
			ctx = GetContext(w, req)
		}
	})

	b.Run("Large JSON Response", func(b *testing.B) {
		req := httptest.NewRequest(MethodGet, "http://example.com/json-large", nil)
		w := getTestWriter()
		ctx := GetContext(w, req)

		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			server.router.ServeHTTP(ctx, ctx.Request)
			ReleaseContext(ctx)
			ctx = GetContext(w, req)
		}
	})

	b.Run("HTML Response", func(b *testing.B) {
		req := httptest.NewRequest(MethodGet, "http://example.com/html", nil)
		w := getTestWriter()
		ctx := GetContext(w, req)

		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			server.router.ServeHTTP(ctx, ctx.Request)
			ReleaseContext(ctx)
			ctx = GetContext(w, req)
		}
	})
}

// BenchmarkMiddleware benchmarks middleware performance
func BenchmarkMiddleware(b *testing.B) {
	// Setup server with no middleware
	noMiddleware := New(DefaultConfig())
	noMiddleware.GET("/", func(c *Ctx) {
		c.String("Hello, World!")
	})

	// Setup server with one middleware
	oneMiddleware := New(DefaultConfig())
	oneMiddleware.Use(func(c *Ctx) {
		c.Next()
	})
	oneMiddleware.GET("/", func(c *Ctx) {
		c.String("Hello, World!")
	})

	// Setup server with multiple middleware
	multiMiddleware := New(DefaultConfig())
	multiMiddleware.Use(func(c *Ctx) {
		c.Next()
	})
	multiMiddleware.Use(func(c *Ctx) {
		c.Next()
	})
	multiMiddleware.Use(func(c *Ctx) {
		c.Next()
	})
	multiMiddleware.GET("/", func(c *Ctx) {
		c.String("Hello, World!")
	})

	// Benchmark
	b.ResetTimer()
	b.Run("No Middleware", func(b *testing.B) {
		req := httptest.NewRequest(MethodGet, "http://example.com/", nil)
		w := getTestWriter()
		ctx := GetContext(w, req)

		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			noMiddleware.router.ServeHTTP(ctx, ctx.Request)
			ReleaseContext(ctx)
			ctx = GetContext(w, req)
		}
	})

	b.Run("One Middleware", func(b *testing.B) {
		req := httptest.NewRequest(MethodGet, "http://example.com/", nil)
		w := getTestWriter()
		ctx := GetContext(w, req)

		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			oneMiddleware.router.ServeHTTP(ctx, ctx.Request)
			ReleaseContext(ctx)
			ctx = GetContext(w, req)
		}
	})

	b.Run("Multiple Middleware", func(b *testing.B) {
		req := httptest.NewRequest(MethodGet, "http://example.com/", nil)
		w := getTestWriter()
		ctx := GetContext(w, req)

		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			multiMiddleware.router.ServeHTTP(ctx, ctx.Request)
			ReleaseContext(ctx)
			ctx = GetContext(w, req)
		}
	})
}

// BenchmarkGroupRouting benchmarks group routing performance
func BenchmarkGroupRouting(b *testing.B) {
	// Setup
	server := New(DefaultConfig())

	// Create a group
	api := server.Group("/api")

	// Add routes to the group
	api.GET("/users", func(c *Ctx) {
		c.String("Users")
	})

	// Create a nested group
	v1 := api.Group("/v1")

	// Add routes to the nested group
	v1.GET("/products", func(c *Ctx) {
		c.String("Products")
	})

	// Create another nested group with middleware
	v2 := api.Group("/v2")
	v2.Use(func(c *Ctx) {
		c.Next()
	})

	// Add routes to the second nested group
	v2.GET("/products", func(c *Ctx) {
		c.String("Products V2")
	})

	// Benchmark
	b.ResetTimer()
	b.Run("Group Route", func(b *testing.B) {
		req := httptest.NewRequest(MethodGet, "http://example.com/api/users", nil)
		w := getTestWriter()
		ctx := GetContext(w, req)

		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			server.router.ServeHTTP(ctx, ctx.Request)
			ReleaseContext(ctx)
			ctx = GetContext(w, req)
		}
	})

	b.Run("Nested Group Route", func(b *testing.B) {
		req := httptest.NewRequest(MethodGet, "http://example.com/api/v1/products", nil)
		w := getTestWriter()
		ctx := GetContext(w, req)

		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			server.router.ServeHTTP(ctx, ctx.Request)
			ReleaseContext(ctx)
			ctx = GetContext(w, req)
		}
	})

	b.Run("Nested Group with Middleware", func(b *testing.B) {
		req := httptest.NewRequest(MethodGet, "http://example.com/api/v2/products", nil)
		w := getTestWriter()
		ctx := GetContext(w, req)

		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			server.router.ServeHTTP(ctx, ctx.Request)
			ReleaseContext(ctx)
			ctx = GetContext(w, req)
		}
	})
}

// BenchmarkContextOperations benchmarks context operations
func BenchmarkContextOperations(b *testing.B) {
	// Setup
	server := New(DefaultConfig())

	// Route with param access
	server.GET("/users/:id", func(c *Ctx) {
		id := c.Param("id")
		c.String("User: %s", id)
	})

	// Route with query param access
	server.GET("/search", func(c *Ctx) {
		query := c.Query("q")
		c.String("Search: %s", query)
	})

	// Route with header access
	server.GET("/headers", func(c *Ctx) {
		userAgent := c.Get(HeaderUserAgent)
		c.String("User-Agent: %s", userAgent)
	})

	// Route with JSON binding
	server.POST("/json", func(c *Ctx) {
		var data map[string]interface{}
		if err := c.BindJSON(&data); err != nil {
			c.Status(StatusBadRequest).String("Bad Request")
			return
		}
		c.JSON(data)
	})

	// Benchmark
	b.ResetTimer()
	b.Run("Param Access", func(b *testing.B) {
		req := httptest.NewRequest(MethodGet, "http://example.com/users/123", nil)
		w := getTestWriter()
		ctx := GetContext(w, req)

		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			server.router.ServeHTTP(ctx, ctx.Request)
			ReleaseContext(ctx)
			ctx = GetContext(w, req)
		}
	})

	b.Run("Query Param Access", func(b *testing.B) {
		req := httptest.NewRequest(MethodGet, "http://example.com/search?q=test", nil)
		w := getTestWriter()
		ctx := GetContext(w, req)

		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			server.router.ServeHTTP(ctx, ctx.Request)
			ReleaseContext(ctx)
			ctx = GetContext(w, req)
		}
	})

	b.Run("Header Access", func(b *testing.B) {
		req := httptest.NewRequest(MethodGet, "http://example.com/headers", nil)
		req.Header.Set(HeaderUserAgent, "Benchmark-Agent")
		w := getTestWriter()
		ctx := GetContext(w, req)

		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			server.router.ServeHTTP(ctx, ctx.Request)
			ReleaseContext(ctx)
			ctx = GetContext(w, req)
		}
	})
}

// BenchmarkHTTPMethods benchmarks different HTTP methods
func BenchmarkHTTPMethods(b *testing.B) {
	// Setup
	server := New(DefaultConfig())

	handler := func(c *Ctx) {
		c.String("OK")
	}

	server.GET("/resource", handler)
	server.POST("/resource", handler)
	server.PUT("/resource", handler)
	server.DELETE("/resource", handler)
	server.PATCH("/resource", handler)

	// Benchmark
	b.ResetTimer()
	b.Run("GET", func(b *testing.B) {
		req := httptest.NewRequest(MethodGet, "http://example.com/resource", nil)
		w := getTestWriter()
		ctx := GetContext(w, req)

		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			server.router.ServeHTTP(ctx, ctx.Request)
			ReleaseContext(ctx)
			ctx = GetContext(w, req)
		}
	})

	b.Run("POST", func(b *testing.B) {
		req := httptest.NewRequest(MethodPost, "http://example.com/resource", nil)
		w := getTestWriter()
		ctx := GetContext(w, req)

		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			server.router.ServeHTTP(ctx, ctx.Request)
			ReleaseContext(ctx)
			ctx = GetContext(w, req)
		}
	})

	b.Run("PUT", func(b *testing.B) {
		req := httptest.NewRequest(MethodPut, "http://example.com/resource", nil)
		w := getTestWriter()
		ctx := GetContext(w, req)

		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			server.router.ServeHTTP(ctx, ctx.Request)
			ReleaseContext(ctx)
			ctx = GetContext(w, req)
		}
	})

	b.Run("DELETE", func(b *testing.B) {
		req := httptest.NewRequest(MethodDelete, "http://example.com/resource", nil)
		w := getTestWriter()
		ctx := GetContext(w, req)

		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			server.router.ServeHTTP(ctx, ctx.Request)
			ReleaseContext(ctx)
			ctx = GetContext(w, req)
		}
	})

	b.Run("PATCH", func(b *testing.B) {
		req := httptest.NewRequest(MethodPatch, "http://example.com/resource", nil)
		w := getTestWriter()
		ctx := GetContext(w, req)

		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			server.router.ServeHTTP(ctx, ctx.Request)
			ReleaseContext(ctx)
			ctx = GetContext(w, req)
		}
	})
}

// BenchmarkStaticFileServing benchmarks static file serving
func BenchmarkStaticFileServing(b *testing.B) {
	// Setup
	server := New(DefaultConfig())
	server.STATIC("/assets", "examples/static/assets")

	// Benchmark
	b.ResetTimer()
	b.Run("HTML File", func(b *testing.B) {
		req := httptest.NewRequest(MethodGet, "http://example.com/assets/index.html", nil)
		w := getTestWriter()
		ctx := GetContext(w, req)

		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			server.router.ServeHTTP(ctx, ctx.Request)
			ReleaseContext(ctx)
			ctx = GetContext(w, req)
		}
	})

	b.Run("CSS File", func(b *testing.B) {
		req := httptest.NewRequest(MethodGet, "http://example.com/assets/css/style.css", nil)
		w := getTestWriter()
		ctx := GetContext(w, req)

		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			server.router.ServeHTTP(ctx, ctx.Request)
			ReleaseContext(ctx)
			ctx = GetContext(w, req)
		}
	})

	b.Run("Text File", func(b *testing.B) {
		req := httptest.NewRequest(MethodGet, "http://example.com/assets/sample.txt", nil)
		w := getTestWriter()
		ctx := GetContext(w, req)

		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			server.router.ServeHTTP(ctx, ctx.Request)
			ReleaseContext(ctx)
			ctx = GetContext(w, req)
		}
	})
}
