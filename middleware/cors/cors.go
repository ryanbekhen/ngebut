package cors

import (
	"strconv"
	"strings"

	"github.com/ryanbekhen/ngebut"
)

// Config represents the configuration for the CORS middleware.
type Config struct {
	// AllowOrigins is a comma-separated list of origins a cross-domain request can be executed from.
	// If the special "*" value is present, all origins will be allowed.
	// Default value is "*"
	AllowOrigins string

	// AllowMethods is a comma-separated list of methods the client is allowed to use with
	// cross-domain requests. Default value is simple methods (GET, POST, PUT, DELETE, HEAD, OPTIONS)
	AllowMethods string

	// AllowHeaders is a comma-separated list of non-simple headers the client is allowed to use with
	// cross-domain requests. Default value is ""
	AllowHeaders string

	// ExposeHeaders indicates which headers are safe to expose to the API of a CORS
	// API specification as a comma-separated list. Default value is ""
	ExposeHeaders string

	// AllowCredentials indicates whether the request can include user credentials like
	// cookies, HTTP authentication or client side SSL certificates. Default value is false
	AllowCredentials bool

	// MaxAge indicates how long (in seconds) the results of a preflight request
	// can be cached. Default value is 0 which stands for no max age.
	MaxAge int
}

// DefaultConfig returns the default configuration for the CORS middleware.
func DefaultConfig() Config {
	return Config{
		AllowOrigins: "*",
		AllowMethods: strings.Join([]string{
			ngebut.MethodGet,
			ngebut.MethodPost,
			ngebut.MethodPut,
			ngebut.MethodDelete,
			ngebut.MethodHead,
			ngebut.MethodOptions,
			ngebut.MethodPatch,
		}, ","),
		AllowHeaders:     "",
		ExposeHeaders:    "",
		AllowCredentials: false,
		MaxAge:           0,
	}
}

// New returns a middleware that handles CORS.
// If no config is provided, it uses the default config.
// If multiple configs are provided, only the first one is used.
func New(config ...Config) ngebut.Middleware {
	// Determine which config to use
	cfg := DefaultConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	// Store the config values for easier access
	allowOrigins := cfg.AllowOrigins
	allowMethods := cfg.AllowMethods
	allowHeaders := cfg.AllowHeaders
	exposeHeaders := cfg.ExposeHeaders

	// Return the middleware function
	return func(c *ngebut.Ctx) {
		// Get origin from request
		origin := c.Get(ngebut.HeaderOrigin)

		// Skip if no Origin header is present
		if origin == "" {
			c.Next()
			return
		}

		// Check if the origin is allowed
		allowOrigin := ""
		if allowOrigins == "*" {
			allowOrigin = "*"
		} else {
			// Check if the request Origin is in the list of allowed origins
			// Split the comma-separated list of allowed origins
			origins := strings.Split(allowOrigins, ",")
			for _, o := range origins {
				o = strings.TrimSpace(o)
				if o == origin || o == "*" {
					allowOrigin = origin
					break
				}
			}
		}

		// Set CORS headers
		c.Set(ngebut.HeaderAccessControlAllowOrigin, allowOrigin)

		// Set Vary header if not using wildcard origin
		if allowOrigin != "*" {
			c.Set(ngebut.HeaderVary, "Origin")
		}

		// Handle preflight OPTIONS request
		if c.Request.Method == ngebut.MethodOptions {
			// Set preflight headers
			c.Set(ngebut.HeaderAccessControlAllowMethods, allowMethods)

			// Set Allow-Headers header if specified
			if cfg.AllowHeaders != "" {
				c.Set(ngebut.HeaderAccessControlAllowHeaders, allowHeaders)
			} else {
				// If no allowed headers are specified, mirror the requested headers
				requestHeaders := c.Get(ngebut.HeaderAccessControlRequestHeaders)
				if requestHeaders != "" {
					c.Set(ngebut.HeaderAccessControlAllowHeaders, requestHeaders)
				}
			}

			// Set Allow-Credentials header if specified
			if cfg.AllowCredentials {
				c.Set(ngebut.HeaderAccessControlAllowCredentials, "true")
			}

			// Set Max-Age header if specified
			if cfg.MaxAge > 0 {
				c.Set(ngebut.HeaderAccessControlMaxAge, strconv.Itoa(cfg.MaxAge))
			}

			// Respond with 204 No Content for preflight requests
			c.Status(ngebut.StatusNoContent)
			return
		}

		// For non-OPTIONS requests

		// Set Expose-Headers header if specified
		if cfg.ExposeHeaders != "" {
			c.Set(ngebut.HeaderAccessControlExposeHeaders, exposeHeaders)
		}

		// Set Allow-Credentials header if specified
		if cfg.AllowCredentials {
			c.Set(ngebut.HeaderAccessControlAllowCredentials, "true")
		}

		// Continue processing the request
		c.Next()
	}
}
