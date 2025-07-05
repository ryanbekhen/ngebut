package session

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ryanbekhen/ngebut"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSessionMiddlewareE2E tests the session middleware in an end-to-end scenario
func TestSessionMiddlewareE2E(t *testing.T) {
	// Create a test HTTP request for setting a session
	reqSet, _ := http.NewRequest("GET", "http://example.com/set-session", nil)
	wSet := httptest.NewRecorder()

	// Create a test context for setting a session
	ctxSet := ngebut.GetContext(wSet, reqSet)

	// Create the middleware with default config
	middleware := NewMiddleware().(func(*ngebut.Ctx))

	// Define a handler that sets a session value
	setHandler := func(c *ngebut.Ctx) {
		// Get the session
		session := GetSession(c)
		require.NotNil(t, session, "Session should not be nil")

		// Set a value in the session
		session.Set("testKey", "testValue")
		err := session.Save()
		require.NoError(t, err, "Failed to save session")

		c.String("%s", "Session set")
	}

	// Call the middleware followed by the handler
	middleware(ctxSet)
	setHandler(ctxSet)

	// Get the response
	respSet := wSet.Result()
	defer respSet.Body.Close()

	// Verify that a session cookie was set
	cookies := respSet.Cookies()
	assert.NotEmpty(t, cookies, "No cookies were set")

	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "session_id" {
			sessionCookie = cookie
			break
		}
	}
	require.NotNil(t, sessionCookie, "Session cookie was not set")

	// Create a test HTTP request for getting the session
	reqGet, _ := http.NewRequest("GET", "http://example.com/get-session", nil)

	// Add the session cookie to the request
	reqGet.Header.Set("Cookie", sessionCookie.Name+"="+sessionCookie.Value)

	wGet := httptest.NewRecorder()

	// Create a test context for getting the session
	ctxGet := ngebut.GetContext(wGet, reqGet)

	// Define a handler that gets a session value
	getHandler := func(c *ngebut.Ctx) {
		// Get the session
		session := GetSession(c)
		require.NotNil(t, session, "Session should not be nil")

		// Get the value from the session
		value := session.Get("testKey")
		if value != nil {
			c.String("%s", value.(string))
		} else {
			c.Status(http.StatusNotFound)
			c.String("%s", "Value not found")
		}
	}

	// Call the middleware followed by the handler
	middleware(ctxGet)
	getHandler(ctxGet)

	// Get the response
	respGet := wGet.Result()
	defer respGet.Body.Close()

	// Read the response body
	body := make([]byte, 1024)
	n, _ := respGet.Body.Read(body)

	// Verify the session value was retrieved correctly
	assert.Equal(t, "testValue", string(body[:n]), "Unexpected session value")
}

// TestSessionExpireE2E tests session expiration in an end-to-end scenario
func TestSessionExpireE2E(t *testing.T) {
	// Create a test HTTP request for setting a session
	reqSet, _ := http.NewRequest("GET", "http://example.com/set-session", nil)
	wSet := httptest.NewRecorder()

	// Create a test context for setting a session
	ctxSet := ngebut.GetContext(wSet, reqSet)

	// Create the middleware with a short expiry time
	middleware := NewMiddleware(Config{
		MaxAge:     1,                   // 1 second expiry
		Expiration: 1 * time.Second,     // 1 second expiry
		KeyLookup:  "cookie:session_id", // Ensure the cookie name is set correctly
	}).(func(*ngebut.Ctx))

	// Define a handler that sets a session value
	setHandler := func(c *ngebut.Ctx) {
		// Get the session
		session := GetSession(c)
		require.NotNil(t, session, "Session should not be nil")

		// Set a value in the session
		session.Set("testKey", "testValue")
		err := session.Save()
		require.NoError(t, err, "Failed to save session")

		c.String("%s", "Session set")
	}

	// Call the middleware followed by the handler
	middleware(ctxSet)
	setHandler(ctxSet)

	// Get the response
	respSet := wSet.Result()
	defer respSet.Body.Close()

	// Try to get the session cookie from the response
	cookies := respSet.Cookies()

	// Validate cookies
	assert.NotEmpty(t, cookies, "No cookies were set")
	assert.Equal(t, 1, len(cookies), "Unexpected number of cookies")
	sessionCookie := cookies[0]
	assert.Equal(t, "session_id", sessionCookie.Name, "Unexpected cookie name")
	assert.NotEmpty(t, sessionCookie.Value, "Cookie value should not be empty")

	// Validate response headers
	setCookieHeader := respSet.Header.Get("Set-Cookie")
	assert.NotEmpty(t, setCookieHeader, "Set-Cookie header should not be empty")

	// We'll skip this assertion since we're now getting the cookie from the Set-Cookie header
	// assert.NotEmpty(t, cookies, "No cookies were set")

	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "session_id" {
			sessionCookie = cookie
			break
		}
	}

	// If no session_id cookie was found, try to parse it from the Set-Cookie header
	if sessionCookie == nil && setCookieHeader != "" {
		// Use http.ReadSetCookies to parse the Set-Cookie header
		parsedCookies := http.ReadSetCookies(respSet.Header)
		for _, cookie := range parsedCookies {
			if cookie.Name == "session_id" {
				sessionCookie = cookie
				t.Logf("Parsed cookie from header: Name=%s, Value=%s", cookie.Name, cookie.Value)
				break
			}
		}
	}

	require.NotNil(t, sessionCookie, "Session cookie was not set")

	// Simulate session expiration
	ExpireSession(sessionCookie.Value)

	// Create a test HTTP request for getting the session
	reqGet, _ := http.NewRequest("GET", "http://example.com/get-session", nil)

	// Add the session cookie to the request
	reqGet.Header.Set("Cookie", sessionCookie.Name+"="+sessionCookie.Value)

	wGet := httptest.NewRecorder()

	// Create a test context for getting the session
	ctxGet := ngebut.GetContext(wGet, reqGet)

	// Define a handler that gets a session value
	getHandler := func(c *ngebut.Ctx) {
		// Get the session
		session := GetSession(c)
		require.NotNil(t, session, "Session should not be nil")

		// Get the value from the session
		value := session.Get("testKey")
		if value != nil {
			c.String("%s", value.(string))
		} else {
			c.Status(http.StatusNotFound)
			c.String("%s", "Value not found")
		}
	}

	// Call the middleware followed by the handler
	middleware(ctxGet)
	getHandler(ctxGet)

	// Get the response
	respGet := wGet.Result()
	defer respGet.Body.Close()

	// Verify that the session has expired
	assert.Equal(t, http.StatusNotFound, respGet.StatusCode, "Expected NotFound status after session expiry")
}
