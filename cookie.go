package ngebut

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Cookie represents an HTTP cookie as sent in the Set-Cookie header of an HTTP response.
type Cookie struct {
	Name        string    `json:"name"`         // The name of the cookie
	Value       string    `json:"value"`        // The value of the cookie
	Path        string    `json:"path"`         // Specifies a URL path which is allowed to receive the cookie
	Domain      string    `json:"domain"`       // Specifies the domain which is allowed to receive the cookie
	MaxAge      int       `json:"max_age"`      // The maximum age (in seconds) of the cookie
	Expires     time.Time `json:"expires"`      // The expiration date of the cookie
	Secure      bool      `json:"secure"`       // Indicates that the cookie should only be transmitted over a secure HTTPS connection
	HTTPOnly    bool      `json:"http_only"`    // Indicates that the cookie is accessible only through the HTTP protocol
	SameSite    string    `json:"same_site"`    // Controls whether or not a cookie is sent with cross-site requests
	Partitioned bool      `json:"partitioned"`  // Indicates if the cookie is stored in a partitioned cookie jar
	SessionOnly bool      `json:"session_only"` // Indicates if the cookie is a session-only cookie
}

// String returns the serialized cookie as it would appear in the Set-Cookie header.
func (c *Cookie) String() string {
	var b strings.Builder

	// Write the required name=value pair
	b.WriteString(c.Name)
	b.WriteString("=")
	b.WriteString(c.Value)

	// Add optional attributes
	if c.Path != "" {
		b.WriteString("; Path=")
		b.WriteString(c.Path)
	}

	if c.Domain != "" {
		b.WriteString("; Domain=")
		b.WriteString(c.Domain)
	}

	if !c.Expires.IsZero() && !c.SessionOnly {
		b.WriteString("; Expires=")
		b.WriteString(c.Expires.UTC().Format(http.TimeFormat))
	}

	if c.MaxAge > 0 && !c.SessionOnly {
		b.WriteString("; Max-Age=")
		b.WriteString(strconv.Itoa(c.MaxAge))
	}

	if c.Secure {
		b.WriteString("; Secure")
	}

	if c.HTTPOnly {
		b.WriteString("; HttpOnly")
	}

	if c.SameSite != "" {
		b.WriteString("; SameSite=")
		b.WriteString(c.SameSite)
	}

	if c.Partitioned {
		b.WriteString("; Partitioned")
	}

	return b.String()
}

// parseCookies parses the cookie header and returns a map of cookie name to value.
// It splits the cookie header by semicolons, then splits each part by equals sign
// to extract the cookie name and value pairs.
// Empty parts and malformed cookies are skipped.
func parseCookies(cookieHeader string) map[string]string {
	cookies := make(map[string]string)
	parts := strings.Split(cookieHeader, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 && kv[0] != "" {
			cookies[kv[0]] = kv[1]
		}
	}
	return cookies
}
