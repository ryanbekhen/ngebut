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
	   cfg := DefaultConfig()
	   if len(config) > 0 {
			   cfg = config[0]
	   }

	   return func(c *ngebut.Ctx) error {
			   authHeader := c.Get(ngebut.HeaderAuthorization)
			   const prefix = "Basic "
			   if len(authHeader) <= len(prefix) || authHeader[:len(prefix)] != prefix {
					   return ErrUnauthorized
			   }

			   decoded, err := base64.StdEncoding.DecodeString(authHeader[len(prefix):])
			   if err != nil {
					   return ErrUnauthorized
			   }
			   cred := string(decoded)
			   sep := -1
			   for i := 0; i < len(cred); i++ {
					   if cred[i] == ':' {
							   sep = i
							   break
					   }
			   }
			   if sep == -1 {
					   return ErrUnauthorized
			   }
			   username := cred[:sep]
			   password := cred[sep+1:]
			   if subtle.ConstantTimeCompare([]byte(username), []byte(cfg.Username)) == 1 &&
					   subtle.ConstantTimeCompare([]byte(password), []byte(cfg.Password)) == 1 {
					   c.Next()
					   return nil
			   }
			   return ErrUnauthorized
	   }
}

// ErrUnauthorized is returned when basic authentication fails.
var ErrUnauthorized = ngebut.NewHttpError(ngebut.StatusUnauthorized, "Unauthorized")
