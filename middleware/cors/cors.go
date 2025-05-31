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

// Pre-defined constants to avoid string allocations
const (
	wildcard       = "*"
	originHeader   = "Origin"
	trueValue      = "true"
	defaultMethods = "GET,POST,PUT,DELETE,HEAD,OPTIONS,PATCH"
	emptyString    = ""
)

// DefaultConfig returns the default configuration for the CORS middleware.
func DefaultConfig() Config {
	return Config{
		AllowOrigins:     wildcard,
		AllowMethods:     defaultMethods,
		AllowHeaders:     emptyString,
		ExposeHeaders:    emptyString,
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

	// Pre-compute and store config values
	allowOrigins := cfg.AllowOrigins
	allowMethods := cfg.AllowMethods
	allowHeaders := cfg.AllowHeaders
	exposeHeaders := cfg.ExposeHeaders
	allowCredentials := cfg.AllowCredentials
	maxAge := cfg.MaxAge

	// Pre-compute max age string if needed
	var maxAgeStr string
	if maxAge > 0 {
		maxAgeStr = strconv.Itoa(maxAge)
	}

	// Pre-compute credentials string if needed
	var credentialsStr string
	if allowCredentials {
		credentialsStr = trueValue
	}

	// Pre-process origins for faster lookup
	isWildcardOrigin := allowOrigins == wildcard
	var originsMap map[string]struct{}

	// Only create the map if we're not using wildcard origins
	if !isWildcardOrigin {
		originsMap = make(map[string]struct{})
		for _, origin := range strings.Split(allowOrigins, ",") {
			originsMap[strings.TrimSpace(origin)] = struct{}{}
		}
	}

	// Return the middleware function
	return func(c *ngebut.Ctx) {
		// Get origin from request
		origin := c.Get(ngebut.HeaderOrigin)

		// Skip if no Origin header is present
		if origin == emptyString {
			c.Next()
			return
		}

		// Fast path for wildcard origin
		if isWildcardOrigin {
			c.Set(ngebut.HeaderAccessControlAllowOrigin, wildcard)
		} else {
			// Check if the origin is allowed using map lookup (O(1) operation)
			_, originAllowed := originsMap[origin]
			_, wildcardAllowed := originsMap[wildcard]

			if originAllowed || wildcardAllowed {
				c.Set(ngebut.HeaderAccessControlAllowOrigin, origin)
				c.Set(ngebut.HeaderVary, originHeader)
			} else {
				// Origin not allowed, but still set Vary header
				c.Set(ngebut.HeaderVary, originHeader)
			}
		}

		// Handle preflight OPTIONS request
		if c.Request.Method == ngebut.MethodOptions {
			// Set preflight headers
			c.Set(ngebut.HeaderAccessControlAllowMethods, allowMethods)

			// Set Allow-Headers header
			if allowHeaders != emptyString {
				c.Set(ngebut.HeaderAccessControlAllowHeaders, allowHeaders)
			} else {
				// Mirror the requested headers if no allowed headers are specified
				requestHeaders := c.Get(ngebut.HeaderAccessControlRequestHeaders)
				if requestHeaders != emptyString {
					c.Set(ngebut.HeaderAccessControlAllowHeaders, requestHeaders)
				}
			}

			// Set Allow-Credentials header if specified
			if allowCredentials {
				c.Set(ngebut.HeaderAccessControlAllowCredentials, credentialsStr)
			}

			// Set Max-Age header if specified
			if maxAge > 0 {
				c.Set(ngebut.HeaderAccessControlMaxAge, maxAgeStr)
			}

			// Respond with 204 No Content for preflight requests
			c.Status(ngebut.StatusNoContent)
			return
		}

		// For non-OPTIONS requests

		// Set Expose-Headers header if specified
		if exposeHeaders != emptyString {
			c.Set(ngebut.HeaderAccessControlExposeHeaders, exposeHeaders)
		}

		// Set Allow-Credentials header if specified
		if allowCredentials {
			c.Set(ngebut.HeaderAccessControlAllowCredentials, credentialsStr)
		}

		// Continue processing the request
		c.Next()
	}
}
