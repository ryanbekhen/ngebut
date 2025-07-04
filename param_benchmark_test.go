package ngebut

import (
	"net/http/httptest"
	"testing"
)

// BenchmarkParamAccess benchmarks the performance of parameter access
func BenchmarkParamAccess(b *testing.B) {
	// Setup
	server := New(DefaultConfig())

	// Route with one parameter
	server.GET("/users/:id", func(c *Ctx) {
		id := c.Param("id")
		c.String("User: %s", id)
	})

	// Route with multiple parameters
	server.GET("/users/:id/posts/:postId", func(c *Ctx) {
		id := c.Param("id")
		postId := c.Param("postId")
		c.String("User: %s, Post: %s", id, postId)
	})

	// Route with many parameters (more than fixed array size)
	server.GET("/a/:p1/b/:p2/c/:p3/d/:p4/e/:p5/f/:p6/g/:p7/h/:p8/i/:p9", func(c *Ctx) {
		// Access all parameters to ensure they're all processed
		p1 := c.Param("p1")
		p9 := c.Param("p9")
		c.String("First: %s, Last: %s", p1, p9)
	})

	// Benchmark
	b.ResetTimer()

	b.Run("Single Parameter", func(b *testing.B) {
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

	b.Run("Multiple Parameters", func(b *testing.B) {
		req := httptest.NewRequest(MethodGet, "http://example.com/users/123/posts/456", nil)
		w := getTestWriter()
		ctx := GetContext(w, req)

		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			server.router.ServeHTTP(ctx, ctx.Request)
			ReleaseContext(ctx)
			ctx = GetContext(w, req)
		}
	})

	b.Run("Many Parameters", func(b *testing.B) {
		req := httptest.NewRequest(MethodGet, "http://example.com/a/1/b/2/c/3/d/4/e/5/f/6/g/7/h/8/i/9", nil)
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
