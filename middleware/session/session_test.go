package session

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/ryanbekhen/ngebut"
	"github.com/ryanbekhen/ngebut/internal/memory"
)

// TestNew tests the New function
func TestNew(t *testing.T) {
	// Test with default config
	store := New()
	if store == nil {
		t.Fatal("New() returned nil")
	}
	if store.manager == nil {
		t.Fatal("New() returned a store with nil manager")
	}

	// Test with custom config
	customConfig := Config{
		CookieName: "custom_session",
		MaxAge:     3600,
		Path:       "/api",
		Secure:     true,
		HttpOnly:   false,
	}
	store = New(customConfig)
	if store == nil {
		t.Fatal("New(customConfig) returned nil")
	}
	if store.manager == nil {
		t.Fatal("New(customConfig) returned a store with nil manager")
	}
}

// TestNewMiddleware tests the NewMiddleware function
func TestNewMiddleware(t *testing.T) {
	// Test with default config
	middleware := NewMiddleware()
	if middleware == nil {
		t.Fatal("NewMiddleware() returned nil")
	}

	// Test with custom config
	customConfig := Config{
		CookieName: "custom_session",
		MaxAge:     3600,
		Path:       "/api",
		Secure:     true,
		HttpOnly:   false,
	}
	middleware = NewMiddleware(customConfig)
	if middleware == nil {
		t.Fatal("NewMiddleware(customConfig) returned nil")
	}
}

// TestDefaultConfig tests the DefaultConfig function
func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.CookieName != "ngebut_session" {
		t.Errorf("DefaultConfig() returned unexpected CookieName: %s", config.CookieName)
	}

	if config.MaxAge != 86400 {
		t.Errorf("DefaultConfig() returned unexpected MaxAge: %d", config.MaxAge)
	}

	if config.Path != "/" {
		t.Errorf("DefaultConfig() returned unexpected Path: %s", config.Path)
	}

	if config.Secure != false {
		t.Errorf("DefaultConfig() returned unexpected Secure: %v", config.Secure)
	}

	if config.HttpOnly != true {
		t.Errorf("DefaultConfig() returned unexpected HttpOnly: %v", config.HttpOnly)
	}
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
	if err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// Test Get
	retrievedSession, err := store.Get("test-session-id")
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}
	if retrievedSession == nil {
		t.Fatal("Retrieved session is nil")
	}
	if retrievedSession.ID != "test-session-id" {
		t.Errorf("Retrieved session has wrong ID: %s", retrievedSession.ID)
	}

	// Test Get with non-existent ID
	nonExistentSession, err := store.Get("non-existent-id")
	if err != nil {
		t.Fatalf("Get with non-existent ID returned error: %v", err)
	}
	if nonExistentSession != nil {
		t.Error("Get with non-existent ID should return nil")
	}

	// Test Delete
	err = store.Delete("test-session-id")
	if err != nil {
		t.Fatalf("Failed to delete session: %v", err)
	}

	// Verify session was deleted
	deletedSession, err := store.Get("test-session-id")
	if err != nil {
		t.Fatalf("Get after delete returned error: %v", err)
	}
	if deletedSession != nil {
		t.Error("Session should be nil after deletion")
	}

	// Test expired session
	expiredSession := &Session{
		ID:        "expired-session-id",
		Values:    make(map[string]interface{}),
		CreatedAt: time.Now().Add(-2 * time.Hour),
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}

	err = store.Save(expiredSession)
	if err != nil {
		t.Fatalf("Failed to save expired session: %v", err)
	}

	// Get should return nil for expired session
	retrievedExpiredSession, err := store.Get("expired-session-id")
	if err != nil {
		t.Fatalf("Get expired session returned error: %v", err)
	}
	if retrievedExpiredSession != nil {
		t.Error("Get should return nil for expired session")
	}
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

	if val := session.Get("key1"); val != "value1" {
		t.Errorf("session.Get(\"key1\") returned %v, expected \"value1\"", val)
	}

	if val := session.Get("key2"); val != 123 {
		t.Errorf("session.Get(\"key2\") returned %v, expected 123", val)
	}

	// Test Delete
	session.Delete("key1")
	if val := session.Get("key1"); val != nil {
		t.Errorf("session.Get(\"key1\") after Delete returned %v, expected nil", val)
	}

	// Test Clear
	session.Clear()
	if val := session.Get("key2"); val != nil {
		t.Errorf("session.Get(\"key2\") after Clear returned %v, expected nil", val)
	}
	if len(session.Values) != 0 {
		t.Errorf("session.Values has %d items after Clear, expected 0", len(session.Values))
	}
}

// TestParseCookies tests the parseCookies function
func TestParseCookies(t *testing.T) {
	cookieHeader := "name1=value1; name2=value2; name3=value3"
	cookies := parseCookies(cookieHeader)

	if len(cookies) != 3 {
		t.Errorf("parseCookies returned %d cookies, expected 3", len(cookies))
	}

	if cookies["name1"] != "value1" {
		t.Errorf("cookies[\"name1\"] = %s, expected \"value1\"", cookies["name1"])
	}

	if cookies["name2"] != "value2" {
		t.Errorf("cookies[\"name2\"] = %s, expected \"value2\"", cookies["name2"])
	}

	if cookies["name3"] != "value3" {
		t.Errorf("cookies[\"name3\"] = %s, expected \"value3\"", cookies["name3"])
	}

	// Test with empty cookie header
	emptyCookies := parseCookies("")
	if len(emptyCookies) != 0 {
		t.Errorf("parseCookies(\"\") returned %d cookies, expected 0", len(emptyCookies))
	}

	// Test with malformed cookie header
	malformedCookies := parseCookies("name1; name2=value2; =value3")
	if len(malformedCookies) != 2 {
		t.Errorf("parseCookies with malformed header returned %d cookies, expected 2", len(malformedCookies))
	}
	if malformedCookies["name2"] != "value2" {
		t.Errorf("malformedCookies[\"name2\"] = %s, expected \"value2\"", malformedCookies["name2"])
	}
	if malformedCookies[""] != "value3" {
		t.Errorf("malformedCookies[\"\"] = %s, expected \"value3\"", malformedCookies[""])
	}
}

// TestGenerateSessionID tests the generateSessionID function
func TestGenerateSessionID(t *testing.T) {
	id1, err := generateSessionID()
	if err != nil {
		t.Fatalf("generateSessionID() returned error: %v", err)
	}
	if id1 == "" {
		t.Fatal("generateSessionID() returned empty string")
	}

	id2, err := generateSessionID()
	if err != nil {
		t.Fatalf("generateSessionID() returned error: %v", err)
	}

	// IDs should be different
	if id1 == id2 {
		t.Error("generateSessionID() returned the same ID twice")
	}
}

// TestManager tests the Manager functionality
func TestManager(t *testing.T) {
	config := DefaultConfig()
	memoryStorage := memory.New(time.Second)
	store := NewStorageAdapter(memoryStorage)
	manager := NewManager(config, store)

	if manager == nil {
		t.Fatal("NewManager returned nil")
	}

	if manager.config.CookieName != config.CookieName {
		t.Errorf("manager.config.CookieName = %s, expected %s", manager.config.CookieName, config.CookieName)
	}

	if manager.store != store {
		t.Error("manager.store is not the same as the provided store")
	}
}

// TestGetSession tests the GetSession function
func TestGetSession(t *testing.T) {
	// Create a mock context with no session
	ctx := &ngebut.Ctx{
		Request: ngebut.NewRequest(nil),
	}

	// GetSession should return nil when no session is in context
	session := GetSession(ctx)
	if session != nil {
		t.Error("GetSession returned non-nil session when none was set")
	}

	// Test with nil request
	ctx.Request = nil
	session = GetSession(ctx)
	if session != nil {
		t.Error("GetSession returned non-nil session with nil request")
	}
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
	if session == nil {
		t.Error("No session was created by middleware")
		return
	}

	// Verify session properties
	if session.ID == "" {
		t.Error("Session ID is empty")
	}
	if session.Values == nil {
		t.Error("Session Values map is nil")
	}
	if session.CreatedAt.IsZero() {
		t.Error("Session CreatedAt is zero")
	}
	if session.ExpiresAt.IsZero() {
		t.Error("Session ExpiresAt is zero")
	}

	// Test session methods
	session.Set("testKey", "testValue")
	if val := session.Get("testKey"); val != "testValue" {
		t.Errorf("Session.Get returned %v, expected 'testValue'", val)
	}

	// Check that a session cookie was set
	resp := w.Result()
	cookies := resp.Cookies()
	if len(cookies) == 0 {
		t.Error("No cookies were set")
	} else {
		found := false
		for _, cookie := range cookies {
			if cookie.Name == "session_id" {
				found = true
				if cookie.Value == "" {
					t.Error("Session cookie has empty value")
				}
				if cookie.Path != "/" {
					t.Errorf("Session cookie has unexpected path: %s", cookie.Path)
				}
				if !cookie.HttpOnly {
					t.Error("Session cookie is not HttpOnly")
				}
				break
			}
		}
		if !found {
			t.Error("Session cookie was not set")
		}
	}
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
	if err != nil {
		t.Fatalf("Failed to save test session: %v", err)
	}

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
	if session == nil {
		t.Error("No session was retrieved by middleware")
		return
	}

	// Verify it's the correct session
	if session.ID != "test-session-id" {
		t.Errorf("Retrieved session has wrong ID: %s", session.ID)
	}

	// Check that the existing value is present
	if val := session.Get("existingKey"); val != "existingValue" {
		t.Errorf("Session.Get returned %v, expected 'existingValue'", val)
	}

	// Modify the session
	session.Set("newKey", "newValue")

	// Save the session to simulate what would happen after the request is processed
	err = store.Save(session)
	if err != nil {
		t.Fatalf("Failed to save updated session: %v", err)
	}

	// Verify the session was updated in the store
	updatedSession, err := store.Get("test-session-id")
	if err != nil {
		t.Fatalf("Failed to get updated session: %v", err)
	}
	if updatedSession == nil {
		t.Fatal("Updated session is nil")
	}
	if val := updatedSession.Get("newKey"); val != "newValue" {
		t.Errorf("Updated session has wrong value for newKey: %v", val)
	}
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
	if len(cookies) == 0 {
		t.Error("No cookies were set")
	} else {
		found := false
		for _, cookie := range cookies {
			if cookie.Name == "custom_session" {
				found = true
				if cookie.Value == "" {
					t.Error("Session cookie has empty value")
				}
				if cookie.Path != "/api" {
					t.Errorf("Session cookie has unexpected path: %s", cookie.Path)
				}
				if cookie.MaxAge != 3600 {
					t.Errorf("Session cookie has unexpected MaxAge: %d", cookie.MaxAge)
				}
				if !cookie.Secure {
					t.Error("Session cookie is not Secure")
				}
				if cookie.HttpOnly {
					t.Error("Session cookie is HttpOnly when it should not be")
				}
				break
			}
		}
		if !found {
			t.Error("Custom session cookie was not set")
		}
	}

	// Check that a session was created and stored in the context
	session := GetSession(ctx)
	if session == nil {
		t.Error("No session was created by middleware")
		return
	}
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
	if err != nil {
		t.Fatalf("Failed to save expired session: %v", err)
	}

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
	if session == nil {
		t.Error("No session was created by middleware")
		return
	}

	// Verify it's a new session
	if session.ID == "expired-session-id" {
		t.Error("Middleware did not create a new session for expired session")
	}

	// Store the session ID for later comparison
	sessionID := session.ID

	// Check that a new session cookie was set
	resp := w.Result()
	cookies := resp.Cookies()
	if len(cookies) == 0 {
		t.Error("No cookies were set")
	} else {
		found := false
		for _, cookie := range cookies {
			if cookie.Name == "ngebut_session" {
				found = true
				if cookie.Value == "expired-session-id" {
					t.Error("Session cookie still has the expired session ID")
				}
				if cookie.Value != sessionID {
					t.Errorf("Session cookie value (%s) doesn't match the new session ID (%s)", cookie.Value, sessionID)
				}
				break
			}
		}
		if !found {
			t.Error("New session cookie was not set")
		}
	}

	// Verify the expired session is no longer in the store
	retrievedExpiredSession, err := store.Get("expired-session-id")
	if err != nil {
		t.Fatalf("Get expired session returned error: %v", err)
	}
	if retrievedExpiredSession != nil {
		t.Error("Expired session should be nil after retrieval attempt")
	}
}
