// ngebut static server example

package main

import (
	"log"
	"path/filepath"

	"github.com/ryanbekhen/ngebut"
)

func main() {
	// Create a new ngebut app
	app := ngebut.New()

	// IMPORTANT: Register API routes BEFORE the catch-all static route
	// to prevent the static handler from intercepting API requests

	// Add some API routes for comparison
	app.GET("/api/info", func(c *ngebut.Ctx) {
		c.JSON(map[string]interface{}{
			"message": "ngebut Static File Server Example",
			"endpoints": map[string]string{
				"/":             "Basic static files (index.html)",
				"/public/":      "Static files with /public prefix",
				"/assets/":      "Advanced static with directory browsing",
				"/conditional/": "Conditional static serving (skips .private files)",
				"/downloads/":   "Force download mode",
				"/api/info":     "This API endpoint",
			},
		})
	})

	// Debug route to see registered routes
	app.GET("/debug/routes", func(c *ngebut.Ctx) {
		routes := []map[string]string{}
		for _, route := range app.Router().Routes {
			routes = append(routes, map[string]string{
				"pattern": route.Pattern,
				"method":  route.Method,
			})
		}
		c.JSON(map[string]interface{}{
			"routes": routes,
		})
	})

	// Simple test route
	app.GET("/test", func(c *ngebut.Ctx) {
		c.String("Test route is working!")
	})

	// Example 2: Static files with custom prefix
	// Serves files from ./static directory under /public path
	app.STATIC("/public/", "./examples/static/assets")

	// Example 3: Advanced static file serving with all features enabled
	app.STATIC("/assets/", "./examples/static/assets", ngebut.Static{
		Browse:    true,  // Enable directory browsing
		Download:  false, // Don't force downloads
		Index:     "",    // No default index file - force directory browsing
		ByteRange: true,  // Enable byte range requests (for video/audio)
		Compress:  false, // File compression (not implemented yet)
		ModifyResponse: func(c *ngebut.Ctx) {
			// Add a custom header to all static files served from /assets/
			c.Set("X-Static-Server", "ngebut-example")
		},
	})

	// Example 4: Conditional static serving with Next function
	app.STATIC("/conditional/", "./examples/static/assets", ngebut.Static{
		Browse: true,
		Next: func(c *ngebut.Ctx) bool {
			// Skip static serving for files with .private extension
			if filepath.Ext(c.Path()) == ".private" {
				return true // Skip this middleware, continue to next handler
			}
			return false // Process with static file serving
		},
	})

	// Example 5: Download mode - forces file downloads
	app.STATIC("/downloads/", "./examples/static/assets", ngebut.Static{
		Download: true, // All files will be served with Content-Disposition: attachment
		Browse:   true,
	})

	// Example 1: Basic static file serving (MOVED TO LAST)
	// Serves files from ./static directory at the root path
	// This MUST be registered last because it matches ALL paths (/*)
	app.STATIC("/", "./examples/static/assets")

	// Custom 404 handler
	app.NotFound(func(c *ngebut.Ctx) {
		c.Status(ngebut.StatusNotFound)
		c.HTML(`
			<html>
			<body style="font-family: Arial, sans-serif; text-align: center; margin-top: 50px;">
				<h1>404 - Page Not Found</h1>
				<p>The file or page you're looking for doesn't exist.</p>
				<p><a href="/">Go back to home</a></p>
			</body>
			</html>
		`)
	})

	log.Println("🚀 Static File Server Example")
	log.Println("Server starting on http://localhost:3000")
	log.Println("")
	log.Println("📁 Available endpoints:")
	log.Println("  • http://localhost:3000/                     - Basic static files")
	log.Println("  • http://localhost:3000/public/              - Static with /public prefix")
	log.Println("  • http://localhost:3000/assets/              - Advanced static with browsing")
	log.Println("  • http://localhost:3000/conditional/         - Conditional static serving")
	log.Println("  • http://localhost:3000/downloads/           - Force download mode")
	log.Println("  • http://localhost:3000/api/info             - API info endpoint")
	log.Println("")

	// Start the server
	if err := app.Listen(":3000"); err != nil {
		log.Fatal("❌ Server failed to start:", err)
	}
}
