package session

import "github.com/ryanbekhen/ngebut"

// getSessionIDFromCookie retrieves the "Cookie" header value from the given context.
func getSessionIDFromCookie(c *ngebut.Ctx) string {
	return c.Request.Header.Get(ngebut.HeaderCookie)
}
