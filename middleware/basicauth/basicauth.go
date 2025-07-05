package basicauth

import (
	"crypto/subtle"
	"encoding/base64"
	"github.com/ryanbekhen/ngebut"
)

// Config represents the configuration structure for username and password authentication.
type Config struct {
	// Username represents the username required for basic authentication in the configuration.
	Username string

	// Password represents the password required for basic authentication in the configuration.
	Password string
}

// DefaultConfig returns a Config instance with default empty values for username and password.
func DefaultConfig() Config {
	return Config{
		Username: "example",
		Password: "example",
	}
}

// New creates and returns a middleware function for Basic Authentication using the provided configuration or defaults.
// The returned middleware returns an error if authentication fails, or nil if successful.
func New(config ...Config) func(c *ngebut.Ctx) error {
	// Determine which config to use
	cfg := DefaultConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	// Return the middleware function
	return func(c *ngebut.Ctx) error {
		// Get Basic Authentication value
		authHeader := c.Get(ngebut.HeaderAuthorization)

		// Standard prefix of Basic Authentication
		const prefix = "Basic "
		if len(authHeader) <= len(prefix) || authHeader[:len(prefix)] != prefix {
			return ErrUnauthorized
		}

		// Attempt to decode the Base64-encoded credentials from the Authorization header.
		// The header format must be: "Basic <base64(username:password)>".
		// If decoding fails, it means the client sent an invalid Base64 string.
		// In that case, we stop processing and treat it as unauthorized.
		decoded, err := base64.StdEncoding.DecodeString(authHeader[len(prefix):])
		if err != nil {
			return ErrUnauthorized
		}

		// Convert the decoded Base64 bytes into a string representation
		// in the expected format: "username:password".
		cred := string(decoded)

		// Find the position of the colon separator in the credentials.
		// According to the Basic Auth specification, the username and password
		// must be separated by exactly one ':' character.
		// If no ':' is found, the credentials are considered malformed.
		sep := -1
		for i := 0; i < len(cred); i++ {
			if cred[i] == ':' {
				sep = i
				break
			}
		}

		// If no colon ':' was found, the credential format is invalid.
		// According to the Basic Auth standard, the credentials must be in the format "username:password".
		// Returning early ensures unauthorized requests are rejected.
		if sep == -1 {
			return ErrUnauthorized
		}

		// Extract the username and password from the credential string
		// based on the position of the ':' separator.
		// Example: For "admin:secret", username = "admin", password = "secret".
		username := cred[:sep]
		password := cred[sep+1:]

		// Perform a constant-time comparison between the provided credentials
		// and the expected credentials from the config.
		// Using crypto/subtle avoids timing attacks by ensuring the comparison time
		// is independent of how similar the strings are.
		if subtle.ConstantTimeCompare([]byte(username), []byte(cfg.Username)) == 1 &&
			subtle.ConstantTimeCompare([]byte(password), []byte(cfg.Password)) == 1 {
			// Credentials are valid; proceed to the next handler in the chain.
			c.Next()
			return nil
		}
		return ErrUnauthorized
	}
}

// ErrUnauthorized is returned when basic authentication fails.
var ErrUnauthorized = ngebut.NewHttpError(ngebut.StatusUnauthorized, "Unauthorized")
