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
