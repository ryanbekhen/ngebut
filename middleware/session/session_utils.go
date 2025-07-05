package session

import (
	"time"

	"github.com/ryanbekhen/ngebut"
	"github.com/ryanbekhen/ngebut/internal/memory"
)

// getSessionIDFromCookie retrieves the "Cookie" header value from the given context.
func getSessionIDFromCookie(c *ngebut.Ctx, cfg *Config, sessionID *string) {
	cookieHeader := c.Request.Header.Get(ngebut.HeaderCookie)
	if cookieHeader != "" {
		cookies := parseCookies(cookieHeader)
		*sessionID = cookies[cfg.sessionName]
		// For backward compatibility
		if *sessionID == "" && cfg.CookieName != "" {
			*sessionID = cookies[cfg.CookieName]
		}
	}
}

// ExpireSession simulates session expiration by setting the session's expiry time to a past time.
// This is used in tests to verify that expired sessions are handled correctly.
func ExpireSession(sessionID string) {
	// Create a memory store
	memoryStorage := memory.New(time.Minute * 5)
	store := NewStorageAdapter(memoryStorage)

	// Try to get the session
	session, err := store.Get(sessionID)
	if err != nil || session == nil {
		return
	}

	// Set the expiry time to a past time
	session.ExpiresAt = time.Now().Add(-1 * time.Hour)

	// Save the session
	_ = store.Save(session)
}
