package session

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/ryanbekhen/ngebut"
	"github.com/ryanbekhen/ngebut/internal/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNew tests the New function
func TestNew(t *testing.T) {
	// Test with default config
	store := New()
	assert.NotNil(t, store, "New() returned nil")
	assert.NotNil(t, store.manager, "New() returned a store with nil manager")

	// Test with custom config
	customConfig := Config{
		CookieName: "custom_session",
		MaxAge:     3600,
		Path:       "/api",
		Secure:     true,
		HttpOnly:   false,
	}
	store = New(customConfig)
	assert.NotNil(t, store, "New(customConfig) returned nil")
	assert.NotNil(t, store.manager, "New(customConfig) returned a store with nil manager")
}

// TestNewMiddleware tests the NewMiddleware function
func TestNewMiddleware(t *testing.T) {
	// Test with default config
	middleware := NewMiddleware()
	assert.NotNil(t, middleware, "NewMiddleware() returned nil")

	// Test with custom config
	customConfig := Config{
		CookieName: "custom_session",
		MaxAge:     3600,
		Path:       "/api",
		Secure:     true,
		HttpOnly:   false,
	}
	middleware = NewMiddleware(customConfig)
	assert.NotNil(t, middleware, "NewMiddleware(customConfig) returned nil")
}

// TestDefaultConfig tests the DefaultConfig function
func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, "ngebut_session", config.CookieName, "DefaultConfig() returned unexpected CookieName")
	assert.Equal(t, 86400, config.MaxAge, "DefaultConfig() returned unexpected MaxAge")
	assert.Equal(t, "/", config.Path, "DefaultConfig() returned unexpected Path")
	assert.False(t, config.Secure, "DefaultConfig() returned unexpected Secure value")
	assert.True(t, config.HttpOnly, "DefaultConfig() returned unexpected HttpOnly value")
}

// TestStorageAdapter tests the StorageAdapter implementation with internal/memory
func TestStorageAdapter(t *testing.T) {
	// Create a memory storage with a 1-second cleanup interval
	memoryStorage := memory.New(time.Second)
	store := NewStorageAdapter(memoryStorage)

	// Test creating a new session
	session := &Session{
		ID:        "test-session-id",
		Values:    make(map[string]interface{}),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}

	// Test Save
	err := store.Save(session)
	require.NoError(t, err, "Failed to save session")

	// Test Get
	retrievedSession, err := store.Get("test-session-id")
	require.NoError(t, err, "Failed to get session")
	assert.NotNil(t, retrievedSession, "Retrieved session is nil")
	assert.Equal(t, "test-session-id", retrievedSession.ID, "Retrieved session has wrong ID")

	// Test Get with non-existent ID
	nonExistentSession, err := store.Get("non-existent-id")
	assert.NoError(t, err, "Get with non-existent ID returned error")
	assert.Nil(t, nonExistentSession, "Get with non-existent ID should return nil")

	// Test Delete
	err = store.Delete("test-session-id")
	require.NoError(t, err, "Failed to delete session")

	// Verify session was deleted
	deletedSession, err := store.Get("test-session-id")
	assert.NoError(t, err, "Get after delete returned error")
	assert.Nil(t, deletedSession, "Session should be nil after deletion")

	// Test expired session
	expiredSession := &Session{
		ID:        "expired-session-id",
		Values:    make(map[string]interface{}),
		CreatedAt: time.Now().Add(-2 * time.Hour),
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}

	err = store.Save(expiredSession)
	require.NoError(t, err, "Failed to save expired session")

	// Get should return nil for expired session
	retrievedExpiredSession, err := store.Get("expired-session-id")
	assert.NoError(t, err, "Get expired session returned error")
	assert.Nil(t, retrievedExpiredSession, "Get should return nil for expired session")
}

// TestSessionMethods tests the Session methods
func TestSessionMethods(t *testing.T) {
	session := &Session{
		ID:        "test-session-id",
		Values:    make(map[string]interface{}),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}

	// Test Set and Get
	session.Set("key1", "value1")
	session.Set("key2", 123)

	assert.Equal(t, "value1", session.Get("key1"), "session.Get(\"key1\") returned unexpected value")
	assert.Equal(t, 123, session.Get("key2"), "session.Get(\"key2\") returned unexpected value")

	// Test Delete
	session.Delete("key1")
	assert.Nil(t, session.Get("key1"), "session.Get(\"key1\") after Delete should return nil")

	// Test Clear
	session.Clear()
	assert.Nil(t, session.Get("key2"), "session.Get(\"key2\") after Clear should return nil")
	assert.Equal(t, 0, len(session.Values), "session.Values should be empty after Clear")
}

// TestParseCookies tests the parseCookies function
func TestParseCookies(t *testing.T) {
	cookieHeader := "name1=value1; name2=value2; name3=value3"
	cookies := parseCookies(cookieHeader)

	assert.Equal(t, 3, len(cookies), "parseCookies returned unexpected number of cookies")
	assert.Equal(t, "value1", cookies["name1"], "cookies[\"name1\"] has unexpected value")
	assert.Equal(t, "value2", cookies["name2"], "cookies[\"name2\"] has unexpected value")
	assert.Equal(t, "value3", cookies["name3"], "cookies[\"name3\"] has unexpected value")

	// Test with empty cookie header
	emptyCookies := parseCookies("")
	assert.Equal(t, 0, len(emptyCookies), "parseCookies(\"\") returned non-empty map")

	// Test with malformed cookie header
	malformedCookies := parseCookies("name1; name2=value2; =value3")
	assert.Equal(t, 2, len(malformedCookies), "parseCookies with malformed header returned unexpected number of cookies")
	assert.Equal(t, "value2", malformedCookies["name2"], "malformedCookies[\"name2\"] has unexpected value")
	assert.Equal(t, "value3", malformedCookies[""], "malformedCookies[\"\"] has unexpected value")
}

// TestGenerateSessionID tests the generateSessionID function
func TestGenerateSessionID(t *testing.T) {
	id1, err := generateSessionID()
	require.NoError(t, err, "generateSessionID() returned error")
	assert.NotEmpty(t, id1, "generateSessionID() returned empty string")

	id2, err := generateSessionID()
	require.NoError(t, err, "generateSessionID() returned error")

	// IDs should be different
	assert.NotEqual(t, id1, id2, "generateSessionID() returned the same ID twice")
}

// TestManager tests the Manager functionality
func TestManager(t *testing.T) {
	config := DefaultConfig()
	memoryStorage := memory.New(time.Second)
	store := NewStorageAdapter(memoryStorage)
	manager := NewManager(config, store)

	assert.NotNil(t, manager, "NewManager returned nil")
	assert.Equal(t, config.CookieName, manager.config.CookieName, "manager.config.CookieName has unexpected value")
	assert.Equal(t, store, manager.store, "manager.store is not the same as the provided store")
}

// TestGetSession tests the GetSession function
func TestGetSession(t *testing.T) {
	// Create a mock context with no session
	ctx := &ngebut.Ctx{
		Request: ngebut.NewRequest(nil),
	}

	// GetSession should return nil when no session is in context
	session := GetSession(ctx)
	assert.Nil(t, session, "GetSession returned non-nil session when none was set")

	// Test with nil request
	ctx.Request = nil
	session = GetSession(ctx)
	assert.Nil(t, session, "GetSession returned non-nil session with nil request")
}

// TestMiddlewareSessionCreation tests that the middleware creates a new session when none exists
func TestMiddlewareSessionCreation(t *testing.T) {
	// Create a test HTTP request and response recorder
	req, _ := http.NewRequest("GET", "http://example.com/test", nil)
	w := httptest.NewRecorder()

	// Create a test context
	ctx := ngebut.GetContext(w, req)

	// Create the middleware with default config
	middleware := NewMiddleware().(func(*ngebut.Ctx))

	// Call the middleware directly
	middleware(ctx)

	// Check that a session was created and stored in the context
	session := GetSession(ctx)
	assert.NotNil(t, session, "No session was created by middleware")

	// Verify session properties
	assert.NotEmpty(t, session.ID, "Session ID is empty")
	assert.NotNil(t, session.Values, "Session Values map is nil")
	assert.False(t, session.CreatedAt.IsZero(), "Session CreatedAt is zero")
	assert.False(t, session.ExpiresAt.IsZero(), "Session ExpiresAt is zero")

	// Test session methods
	session.Set("testKey", "testValue")
	assert.Equal(t, "testValue", session.Get("testKey"), "Session.Get returned unexpected value")

	// Check that a session cookie was set
	resp := w.Result()
	cookies := resp.Cookies()
	assert.NotEmpty(t, cookies, "No cookies were set")

	found := false
	for _, cookie := range cookies {
		if cookie.Name == "session_id" {
			found = true
			assert.NotEmpty(t, cookie.Value, "Session cookie has empty value")
			assert.Equal(t, "/", cookie.Path, "Session cookie has unexpected path")
			assert.True(t, cookie.HttpOnly, "Session cookie is not HttpOnly")
			break
		}
	}
	assert.True(t, found, "Session cookie was not set")
}

// TestMiddlewareSessionRetrieval tests that the middleware retrieves an existing session
func TestMiddlewareSessionRetrieval(t *testing.T) {
	// Create a memory store and add a test session
	memoryStorage := memory.New(time.Second)
	store := NewStorageAdapter(memoryStorage)
	testSession := &Session{
		ID:        "test-session-id",
		Values:    map[string]interface{}{"existingKey": "existingValue"},
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}
	err := store.Save(testSession)
	require.NoError(t, err, "Failed to save test session")

	// Create a test HTTP request with a session cookie
	req, _ := http.NewRequest("GET", "http://example.com/test", nil)
	req.Header.Set("Cookie", "ngebut_session=test-session-id")
	w := httptest.NewRecorder()

	// Create a test context
	ctx := ngebut.GetContext(w, req)

	// Create a custom middleware that uses our test store
	config := DefaultConfig()
	manager := NewManager(config, store)
	middleware := func(c *ngebut.Ctx) {
		// Try to get the session ID from the cookie
		sessionID := c.Request.Header.Get("Cookie")
		// Parse the cookie header to find our session cookie
		if sessionID != "" {
			cookies := parseCookies(sessionID)
			sessionID = cookies[config.CookieName]
		}

		var session *Session
		var err error

		if sessionID != "" {
			// Try to get the session from the store
			session, err = manager.store.Get(sessionID)
			if err != nil {
				c.Error(err)
				return
			}
		}

		// If no session was found or it's expired, create a new one
		if session == nil {
			// Generate a new session ID
			newID, err := generateSessionID()
			if err != nil {
				c.Error(err)
				return
			}

			// Create a new session
			session = &Session{
				ID:        newID,
				Values:    make(map[string]interface{}),
				CreatedAt: time.Now(),
				ExpiresAt: time.Now().Add(time.Duration(config.MaxAge) * time.Second),
			}

			// Save the session to the store
			if err := manager.store.Save(session); err != nil {
				c.Error(err)
				return
			}

			// Set the session cookie
			httpOnlyStr := ""
			if config.HttpOnly {
				httpOnlyStr = "; HttpOnly"
			}

			secureStr := ""
			if config.Secure {
				secureStr = "; Secure"
			}

			c.Set("Set-Cookie", config.CookieName+"="+session.ID+
				"; Path="+config.Path+
				"; Max-Age="+strconv.Itoa(config.MaxAge)+
				httpOnlyStr+
				secureStr)
		}

		// Store the session in the request context for handlers to access
		// Create a new context with the session
		sessionCtx := context.WithValue(c.Request.Context(), sessionKey("session"), session)
		c.Request = c.Request.WithContext(sessionCtx)
	}

	// Call the middleware directly
	middleware(ctx)

	// Check that the session was retrieved
	session := GetSession(ctx)
	assert.NotNil(t, session, "No session was retrieved by middleware")

	// Verify it's the correct session
	assert.Equal(t, "test-session-id", session.ID, "Retrieved session has wrong ID")

	// Check that the existing value is present
	assert.Equal(t, "existingValue", session.Get("existingKey"), "Session.Get returned unexpected value for existing key")

	// Modify the session
	session.Set("newKey", "newValue")

	// Save the session to simulate what would happen after the request is processed
	err = store.Save(session)
	require.NoError(t, err, "Failed to save updated session")

	// Verify the session was updated in the store
	updatedSession, err := store.Get("test-session-id")
	require.NoError(t, err, "Failed to get updated session")
	assert.NotNil(t, updatedSession, "Updated session is nil")
	assert.Equal(t, "newValue", updatedSession.Get("newKey"), "Updated session has wrong value for newKey")
}

// TestMiddlewareCustomConfig tests the middleware with custom configuration
func TestMiddlewareCustomConfig(t *testing.T) {
	// Create a test HTTP request and response recorder
	req, _ := http.NewRequest("GET", "http://example.com/test", nil)
	w := httptest.NewRecorder()

	// Create a test context
	ctx := ngebut.GetContext(w, req)

	// Create the middleware with custom config
	customConfig := Config{
		CookieName: "custom_session",
		MaxAge:     3600,
		Path:       "/api",
		Secure:     true,
		HttpOnly:   false,
	}
	middleware := NewMiddleware(customConfig).(func(*ngebut.Ctx))

	// Call the middleware directly
	middleware(ctx)

	// Check that a session cookie was set with the custom configuration
	resp := w.Result()
	cookies := resp.Cookies()
	assert.NotEmpty(t, cookies, "No cookies were set")

	found := false
	for _, cookie := range cookies {
		if cookie.Name == "custom_session" {
			found = true
			assert.NotEmpty(t, cookie.Value, "Session cookie has empty value")
			assert.Equal(t, "/api", cookie.Path, "Session cookie has unexpected path")
			assert.Equal(t, 3600, cookie.MaxAge, "Session cookie has unexpected MaxAge")
			assert.True(t, cookie.Secure, "Session cookie is not Secure")
			assert.False(t, cookie.HttpOnly, "Session cookie is HttpOnly when it should not be")
			break
		}
	}
	assert.True(t, found, "Custom session cookie was not set")

	// Check that a session was created and stored in the context
	session := GetSession(ctx)
	assert.NotNil(t, session, "No session was created by middleware")
}

// TestMiddlewareExpiredSession tests that the middleware creates a new session when the existing one is expired
func TestMiddlewareExpiredSession(t *testing.T) {
	// Create a memory store and add an expired test session
	memoryStorage := memory.New(time.Second)
	store := NewStorageAdapter(memoryStorage)
	expiredSession := &Session{
		ID:        "expired-session-id",
		Values:    map[string]interface{}{"key": "value"},
		CreatedAt: time.Now().Add(-2 * time.Hour),
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
	}
	err := store.Save(expiredSession)
	require.NoError(t, err, "Failed to save expired session")

	// Create a test HTTP request with the expired session cookie
	req, _ := http.NewRequest("GET", "http://example.com/test", nil)
	req.Header.Set("Cookie", "ngebut_session=expired-session-id")
	w := httptest.NewRecorder()

	// Create a test context
	ctx := ngebut.GetContext(w, req)

	// Create the middleware with default config but using our store
	config := DefaultConfig()
	manager := NewManager(config, store)
	middleware := func(c *ngebut.Ctx) {
		// Try to get the session ID from the cookie
		sessionID := c.Request.Header.Get("Cookie")
		// Parse the cookie header to find our session cookie
		if sessionID != "" {
			cookies := parseCookies(sessionID)
			sessionID = cookies[config.CookieName]
		}

		var session *Session
		var err error

		if sessionID != "" {
			// Try to get the session from the store
			session, err = manager.store.Get(sessionID)
			if err != nil {
				c.Error(err)
				return
			}
		}

		// If no session was found or it's expired, create a new one
		if session == nil {
			// Generate a new session ID
			newID, err := generateSessionID()
			if err != nil {
				c.Error(err)
				return
			}

			// Create a new session
			session = &Session{
				ID:        newID,
				Values:    make(map[string]interface{}),
				CreatedAt: time.Now(),
				ExpiresAt: time.Now().Add(time.Duration(config.MaxAge) * time.Second),
			}

			// Save the session to the store
			if err := manager.store.Save(session); err != nil {
				c.Error(err)
				return
			}

			// Set the session cookie
			httpOnlyStr := ""
			if config.HttpOnly {
				httpOnlyStr = "; HttpOnly"
			}

			secureStr := ""
			if config.Secure {
				secureStr = "; Secure"
			}

			c.Set("Set-Cookie", config.CookieName+"="+session.ID+
				"; Path="+config.Path+
				"; Max-Age="+strconv.Itoa(config.MaxAge)+
				httpOnlyStr+
				secureStr)
		}

		// Store the session in the request context for handlers to access
		// Create a new context with the session
		sessionCtx := context.WithValue(c.Request.Context(), sessionKey("session"), session)
		c.Request = c.Request.WithContext(sessionCtx)
	}

	// Call the middleware directly
	middleware(ctx)

	// Check that a new session was created
	session := GetSession(ctx)
	assert.NotNil(t, session, "No session was created by middleware")

	// Verify it's a new session
	assert.NotEqual(t, "expired-session-id", session.ID, "Middleware did not create a new session for expired session")

	// Store the session ID for later comparison
	sessionID := session.ID

	// Check that a new session cookie was set
	resp := w.Result()
	cookies := resp.Cookies()
	assert.NotEmpty(t, cookies, "No cookies were set")

	found := false
	for _, cookie := range cookies {
		if cookie.Name == "ngebut_session" {
			found = true
			assert.NotEqual(t, "expired-session-id", cookie.Value, "Session cookie still has the expired session ID")
			assert.Equal(t, sessionID, cookie.Value, "Session cookie value doesn't match the new session ID")
			break
		}
	}
	assert.True(t, found, "New session cookie was not set")

	// Verify the expired session is no longer in the store
	retrievedExpiredSession, err := store.Get("expired-session-id")
	assert.NoError(t, err, "Get expired session returned error")
	assert.Nil(t, retrievedExpiredSession, "Expired session should be nil after retrieval attempt")
}

// TestMiddlewareSessionIDFromCookie tests that the middleware retrieves the session ID from the cookie
func TestMiddlewareSessionIDFromCookie(t *testing.T) {
	// Create a memory store and add a test session
	memoryStorage := memory.New(time.Second)
	store := NewStorageAdapter(memoryStorage)
	testSession := &Session{
		ID:        "test-session-id",
		Values:    map[string]interface{}{"key": "value"},
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}
	err := store.Save(testSession)
	require.NoError(t, err, "Failed to save test session")

	// Create a test HTTP request with the session cookie
	req, _ := http.NewRequest("GET", "http://example.com/test", nil)
	req.Header.Set("Cookie", "ngebut_session=test-session-id")
	w := httptest.NewRecorder()

	// Create a test context
	ctx := ngebut.GetContext(w, req)

	// Create the middleware with default config but using our store
	config := DefaultConfig()
	manager := NewManager(config, store)
	middleware := func(c *ngebut.Ctx) {
		// Try to get the session ID from the cookie
		sessionID := c.Request.Header.Get("Cookie")
		if sessionID != "" {
			cookies := parseCookies(sessionID)
			sessionID = cookies[config.CookieName]
		}

		var session *Session
		var err error

		if sessionID != "" {
			// Try to get the session from the store
			session, err = manager.store.Get(sessionID)
			if err != nil {
				c.Error(err)
				return
			}
		}

		if session == nil {
			c.Error(errors.New("No valid session found"))
			return
		}

		// Store the session in the request context for handlers to access
		sessionCtx := context.WithValue(c.Request.Context(), sessionKey("session"), session)
		c.Request = c.Request.WithContext(sessionCtx)
	}

	// Call the middleware directly
	middleware(ctx)

	// Check that the session was retrieved correctly
	session := GetSession(ctx)
	assert.NotNil(t, session, "No session was retrieved by middleware")
	assert.Equal(t, "test-session-id", session.ID, "Retrieved session has wrong ID")
	assert.Equal(t, "value", session.Get("key"), "Session.Get returned unexpected value for key")
}
