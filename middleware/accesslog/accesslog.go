package accesslog

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ryanbekhen/ngebut"
	"github.com/ryanbekhen/ngebut/log"
)

// Config represents the configuration for the AccessLog middleware.
type Config struct {
	// Format is the format string for the access log.
	// Available placeholders:
	// - ${remote_ip} - the client's IP address
	// - ${method} - the HTTP method
	// - ${path} - the request path
	// - ${status} - the HTTP status code
	// - ${latency} - the request latency
	// - ${latency_human} - the request latency in human-readable format
	// - ${bytes_in} - the number of bytes received
	// - ${user_agent} - the User-Agent header
	// - ${referer} - the Referer header
	// - ${time} - the current time in the format "2006-01-02 15:04:05"
	// - ${query} - the URL query string
	// - ${error} - the error message if an error occurred during request processing
	Format string
}

// DefaultConfig returns the default configuration for the AccessLog middleware.
func DefaultConfig() Config {
	return Config{
		Format: "${time} | ${status} | ${latency_human} | ${method} ${path} | ${error}",
	}
}

// New returns a middleware that logs HTTP requests.
// If no config is provided, it uses the default config.
// If multiple configs are provided, only the first one is used.
func New(config ...Config) interface{} {
	// Determine which config to use
	cfg := DefaultConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	// Return the simple middleware pattern (without error return)
	return func(c *ngebut.Ctx) {
		// Record start time
		start := time.Now()

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get request details
		method := c.Request.Method
		path := c.Request.URL.Path
		status := c.StatusCode()
		ip := c.IP()
		bytesIn := c.Request.ContentLength
		userAgent := c.Get("User-Agent")
		referer := c.Get("Referer")
		query := c.Request.URL.RawQuery

		// Format the log message
		msg := cfg.Format
		msg = replaceTag(msg, "${remote_ip}", ip)
		msg = replaceTag(msg, "${method}", method)
		msg = replaceTag(msg, "${path}", path)
		msg = replaceTag(msg, "${status}", intToString(status))
		msg = replaceTag(msg, "${latency}", latency.String())
		msg = replaceTag(msg, "${latency_human}", formatLatency(latency))
		msg = replaceTag(msg, "${bytes_in}", int64ToString(bytesIn))
		msg = replaceTag(msg, "${user_agent}", userAgent)
		msg = replaceTag(msg, "${referer}", referer)
		msg = replaceTag(msg, "${time}", time.Now().Format("2006-01-02 15:04:05"))
		msg = replaceTag(msg, "${query}", query)

		// Check for errors in the context
		if err := c.GetError(); err != nil {
			// Add "error: " prefix to make it recognizable for the console writer's coloring logic
			msg = replaceTag(msg, "${error}", "error: "+err.Error())
		} else {
			msg = replaceTag(msg, "${error}", "")
		}

		// Get error from context if any
		err := c.GetError()

		// Log the message using our own logger with appropriate level based on status code
		if status >= 500 {
			// Server error (5xx)
			if err != nil {
				logger.Error().Err(err).Msg(msg)
			} else {
				logger.Error().Msg(msg)
			}
		} else if status >= 400 {
			// Client error (4xx)
			if err != nil {
				logger.Warn().Err(err).Msg(msg)
			} else {
				logger.Warn().Msg(msg)
			}
		} else {
			// Success (2xx) or Redirection (3xx)
			if err != nil {
				logger.Info().Err(err).Msg(msg)
			} else {
				logger.Info().Msg(msg)
			}
		}
	}
}

// Initialize a logger for the accesslog package
var logger *log.Logger

func init() {
	// Set up pretty logging for development
	console := log.DefaultConsoleWriter()
	console.Out = os.Stdout
	console.NoColor = false // Enable color output

	// Create a new logger with the console writer
	logger = log.New(console, log.InfoLevel)

	// Check if there's a global logger set
	if globalLogger := log.GetLogger(); globalLogger != nil {
		// Use the global logger if available, but only if it's a *log.Logger
		if loggerImpl, ok := globalLogger.(*log.Logger); ok {
			logger = loggerImpl
		}
	}
}

// Helper functions for string replacements and conversions

// replaceTag replaces all occurrences of a tag in a message with a value.
// It takes the original message, the tag to replace, and the value to replace it with.
// It returns the modified message with all occurrences of the tag replaced.
func replaceTag(msg, tag, value string) string {
	return strings.Replace(msg, tag, value, -1)
}

// intToString converts an integer to its string representation.
// It's a wrapper around strconv.Itoa for consistent naming with other conversion functions.
func intToString(n int) string {
	return strconv.Itoa(n)
}

// int64ToString converts a 64-bit integer to its string representation.
// It uses base 10 for the conversion.
func int64ToString(n int64) string {
	return strconv.FormatInt(n, 10)
}

// formatLatency formats a duration in a human-readable way with appropriate units (ns, µs, ms, s)
func formatLatency(d time.Duration) string {
	if d < time.Microsecond {
		return strconv.FormatInt(d.Nanoseconds(), 10) + "ns"
	}
	if d < time.Millisecond {
		return strconv.FormatFloat(float64(d.Nanoseconds())/float64(time.Microsecond), 'f', 2, 64) + "µs"
	}
	if d < time.Second {
		return strconv.FormatFloat(float64(d.Nanoseconds())/float64(time.Millisecond), 'f', 2, 64) + "ms"
	}
	return strconv.FormatFloat(float64(d.Nanoseconds())/float64(time.Second), 'f', 2, 64) + "s"
}
