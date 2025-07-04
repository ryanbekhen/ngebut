package ngebut

import (
	"net/http/httptest"
	"testing"
)

// BenchmarkRouterStatic benchmarks the router with static routes
func BenchmarkRouterStatic(b *testing.B) {
	router := NewRouter()

	// Register static routes
	router.GET("/", func(c *Ctx) {
		c.String("Hello, World!")
	})

	router.GET("/users", func(c *Ctx) {
		c.String("Users")
	})

	router.GET("/users/settings", func(c *Ctx) {
		c.String("User Settings")
	})

	router.GET("/about", func(c *Ctx) {
		c.String("About")
	})

	router.GET("/contact", func(c *Ctx) {
		c.String("Contact")
	})

	// Create a test request
	req := httptest.NewRequest(MethodGet, "http://example.com/users", nil)
	w := httptest.NewRecorder()

	// Reset the timer
	b.ResetTimer()
	b.ReportAllocs()

	// Run the benchmark
	for i := 0; i < b.N; i++ {
		ctx := GetContext(w, req)
		router.ServeHTTP(ctx, ctx.Request)
		ReleaseContext(ctx)
	}
}

// BenchmarkRouterParam benchmarks the router with parameterized routes
func BenchmarkRouterParam(b *testing.B) {
	router := NewRouter()

	// Register parameterized routes
	router.GET("/users/:id", func(c *Ctx) {
		id := c.Param("id")
		c.String("User: %s", id)
	})

	router.GET("/users/:id/posts/:postId", func(c *Ctx) {
		id := c.Param("id")
		postId := c.Param("postId")
		c.String("User: %s, Post: %s", id, postId)
	})

	router.GET("/categories/:category/products/:productId", func(c *Ctx) {
		category := c.Param("category")
		productId := c.Param("productId")
		c.String("Category: %s, Product: %s", category, productId)
	})

	// Create a test request
	req := httptest.NewRequest(MethodGet, "http://example.com/users/123", nil)
	w := httptest.NewRecorder()

	// Reset the timer
	b.ResetTimer()
	b.ReportAllocs()

	// Run the benchmark
	for i := 0; i < b.N; i++ {
		ctx := GetContext(w, req)
		router.ServeHTTP(ctx, ctx.Request)
		ReleaseContext(ctx)
	}
}

// BenchmarkRouterWildcard benchmarks the router with wildcard routes
func BenchmarkRouterWildcard(b *testing.B) {
	router := NewRouter()

	// Register wildcard routes
	router.GET("/files/*", func(c *Ctx) {
		c.String("Files")
	})

	router.GET("/static/*", func(c *Ctx) {
		c.String("Static")
	})

	router.GET("/api/*", func(c *Ctx) {
		c.String("API")
	})

	// Create a test request
	req := httptest.NewRequest(MethodGet, "http://example.com/files/images/logo.png", nil)
	w := httptest.NewRecorder()

	// Reset the timer
	b.ResetTimer()
	b.ReportAllocs()

	// Run the benchmark
	for i := 0; i < b.N; i++ {
		ctx := GetContext(w, req)
		router.ServeHTTP(ctx, ctx.Request)
		ReleaseContext(ctx)
	}
}

// BenchmarkRouterMixed benchmarks the router with a mix of static, parameterized, and wildcard routes
func BenchmarkRouterMixed(b *testing.B) {
	router := NewRouter()

	// Register a mix of routes
	router.GET("/", func(c *Ctx) {
		c.String("Hello, World!")
	})

	router.GET("/users", func(c *Ctx) {
		c.String("Users")
	})

	router.GET("/users/:id", func(c *Ctx) {
		id := c.Param("id")
		c.String("User: %s", id)
	})

	router.GET("/users/:id/posts/:postId", func(c *Ctx) {
		id := c.Param("id")
		postId := c.Param("postId")
		c.String("User: %s, Post: %s", id, postId)
	})

	router.GET("/files/*", func(c *Ctx) {
		c.String("Files")
	})

	// Create a test request
	req := httptest.NewRequest(MethodGet, "http://example.com/users/123", nil)
	w := httptest.NewRecorder()

	// Reset the timer
	b.ResetTimer()
	b.ReportAllocs()

	// Run the benchmark
	for i := 0; i < b.N; i++ {
		ctx := GetContext(w, req)
		router.ServeHTTP(ctx, ctx.Request)
		ReleaseContext(ctx)
	}
}

// BenchmarkRouterLongPath benchmarks the router with a long path
func BenchmarkRouterLongPath(b *testing.B) {
	router := NewRouter()

	// Register a route with a long path
	router.GET("/api/v1/users/:userId/accounts/:accountId/transactions/:transactionId/details", func(c *Ctx) {
		userId := c.Param("userId")
		accountId := c.Param("accountId")
		transactionId := c.Param("transactionId")
		c.String("User: %s, Account: %s, Transaction: %s", userId, accountId, transactionId)
	})

	// Create a test request
	req := httptest.NewRequest(MethodGet, "http://example.com/api/v1/users/123/accounts/456/transactions/789/details", nil)
	w := httptest.NewRecorder()

	// Reset the timer
	b.ResetTimer()
	b.ReportAllocs()

	// Run the benchmark
	for i := 0; i < b.N; i++ {
		ctx := GetContext(w, req)
		router.ServeHTTP(ctx, ctx.Request)
		ReleaseContext(ctx)
	}
}

// BenchmarkRouterNotFound benchmarks the router with a route that doesn't exist
func BenchmarkRouterNotFound(b *testing.B) {
	router := NewRouter()

	// Register some routes
	router.GET("/", func(c *Ctx) {
		c.String("Hello, World!")
	})

	router.GET("/users", func(c *Ctx) {
		c.String("Users")
	})

	// Create a test request for a route that doesn't exist
	req := httptest.NewRequest(MethodGet, "http://example.com/not-found", nil)
	w := httptest.NewRecorder()

	// Reset the timer
	b.ResetTimer()
	b.ReportAllocs()

	// Run the benchmark
	for i := 0; i < b.N; i++ {
		ctx := GetContext(w, req)
		router.ServeHTTP(ctx, ctx.Request)
		ReleaseContext(ctx)
	}
}
