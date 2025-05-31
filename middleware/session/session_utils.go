package session

import "github.com/ryanbekhen/ngebut"

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
