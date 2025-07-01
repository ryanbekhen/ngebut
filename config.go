package ngebut

import "time"

// Config represents server configuration options.
type Config struct {
	// ReadTimeout is the maximum duration for reading the entire request, including the body.
	ReadTimeout time.Duration

	// WriteTimeout is the maximum duration before timing out writes of the response.
	WriteTimeout time.Duration

	// IdleTimeout is the maximum amount of time to wait for the next request when keep-alives are enabled.
	IdleTimeout time.Duration

	// DisableStartupMessage determines whether to print the startup message when the server starts.
	DisableStartupMessage bool

	// ErrorHandler is called when an error occurs during request processing.
	ErrorHandler Handler
}

// DefaultConfig returns a default server configuration with pre-configured timeouts
// and other settings suitable for most applications.
// The default configuration includes:
// - ReadTimeout: 5 seconds
// - WriteTimeout: 10 seconds
// - IdleTimeout: 15 seconds
// - DisableStartupMessage: false
// - ErrorHandler: default error handler
func DefaultConfig() Config {
	return Config{
		ReadTimeout:           5 * time.Second,
		WriteTimeout:          10 * time.Second,
		IdleTimeout:           15 * time.Second,
		DisableStartupMessage: false,
		ErrorHandler:          defaultErrorHandler,
	}
}

// Static defines configuration options when defining static assets.
type Static struct {
	// When set to true, the server tries minimizing CPU usage by caching compressed files.
	// Optional. Default value false
	Compress bool `json:"compress"`

	// When set to true, enables byte range requests.
	// Optional. Default value false
	ByteRange bool `json:"byte_range"`

	// When set to true, enables directory browsing.
	// Optional. Default value false.
	Browse bool `json:"browse"`

	// When set to true, enables direct download.
	// Optional. Default value false.
	Download bool `json:"download"`

	// The name of the index file for serving a directory.
	// Optional. Default value "index.html".
	Index string `json:"index"`

	// Expiration duration for inactive file handlers.
	// Use a negative time.Duration to disable it.
	//
	// Optional. Default value 10 * time.Second.
	CacheDuration time.Duration `json:"cache_duration"`

	// The value for the Cache-Control HTTP-header
	// that is set on the file response. MaxAge is defined in seconds.
	//
	// Optional. Default value 0.
	MaxAge int `json:"max_age"`

	// When set to true, enables in-memory caching of file contents.
	// This can significantly improve performance for frequently accessed files.
	// Optional. Default value false.
	InMemoryCache bool `json:"in_memory_cache"`

	// Maximum size of the in-memory cache in bytes.
	// Optional. Default value 100MB.
	MaxCacheSize int64 `json:"max_cache_size"`

	// Maximum number of files to store in the in-memory cache.
	// Optional. Default value 1000.
	MaxCacheItems int `json:"max_cache_items"`

	// ModifyResponse defines a function that allows you to alter the response.
	//
	// Optional. Default: nil
	ModifyResponse Handler

	// Next defines a function to skip this middleware when returned true.
	//
	// Optional. Default: nil
	Next func(c *Ctx) bool
}

// DefaultStaticConfig is the default static configuration.
var DefaultStaticConfig = Static{
	Compress:       false,
	ByteRange:      false,
	Browse:         false,
	Download:       false,
	Index:          "index.html",
	CacheDuration:  10 * time.Second,
	MaxAge:         0,
	InMemoryCache:  true,              // Enable in-memory caching by default for better performance
	MaxCacheSize:   100 * 1024 * 1024, // 100MB
	MaxCacheItems:  1000,              // 1000 files
	ModifyResponse: nil,
}
