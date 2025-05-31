package session

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/ryanbekhen/ngebut/internal/memory"
	"strconv"
	"strings"
	"time"

	"github.com/ryanbekhen/ngebut"
)

// Session represents a user session with identification, data storage, and expiration information.
type Session struct {
	// ID is the unique identifier for the session
	ID string

	// Values stores session data as key-value pairs
	Values map[string]interface{}

	// CreatedAt is the timestamp when the session was created
	CreatedAt time.Time

	// ExpiresAt is the timestamp when the session will expire
	ExpiresAt time.Time

	// store is the storage backend for this session
	store Store

	// cookieName is the name of the cookie that stores the session ID
	cookieName string

	// cookiePath is the path for the cookie
	cookiePath string
}

// Config represents the configuration for the Session middleware.
type Config struct {
	// Expiration is the duration after which the session will expire
	Expiration time.Duration
	// KeyLookup is the format of where to look for the session ID
	// Format: "source:name" where source can be "cookie", "header", or "query"
	// Example: "cookie:session_id"
	KeyLookup string
	// KeyGenerator is a function that generates a new session ID
	// If nil, a default UUID generator will be used
	KeyGenerator func() string
	// CookieName is the name of the cookie that will store the session ID (deprecated, use KeyLookup instead)
	CookieName string
	// MaxAge is the maximum age of the session in seconds (deprecated, use Expiration instead)
	MaxAge int
	// Path is the cookie path
	Path string
	// Domain is the cookie domain
	Domain string
	// Secure indicates if the cookie should only be sent over HTTPS
	Secure bool
	// HttpOnly indicates if the cookie should only be accessible via HTTP(S) requests
	HttpOnly bool
	// Storage is the storage backend for sessions
	// If nil, an in-memory storage will be used
	Storage ngebut.Storage

	// source is the source of the session ID (cookie, header, or query)
	// This is derived from KeyLookup
	source string
	// sessionName is the name of the session ID in the source
	// This is derived from KeyLookup
	sessionName string
}

// DefaultConfig returns the default configuration for the Session middleware.
func DefaultConfig() Config {
	cfg := Config{
		Expiration:   24 * time.Hour,
		KeyLookup:    "cookie:session_id",
		KeyGenerator: UUIDv4,           // Use UUIDv4 as the default generator
		CookieName:   "ngebut_session", // For backward compatibility
		MaxAge:       86400,            // 24 hours, for backward compatibility
		Path:         "/",
		Secure:       false,
		HttpOnly:     true,
		Storage:      nil, // Will use internal/memory by default
	}

	// Parse the KeyLookup string
	parts := strings.Split(cfg.KeyLookup, ":")
	if len(parts) == 2 {
		cfg.source = parts[0]
		cfg.sessionName = parts[1]
	} else {
		// Default to cookie if KeyLookup is invalid
		cfg.source = "cookie"
		cfg.sessionName = "session_id"
	}

	return cfg
}

// Store is the interface that session stores must implement
type Store interface {
	// Get retrieves a session by ID
	Get(id string) (*Session, error)
	// Save saves a session
	Save(session *Session) error
	// Delete removes a session
	Delete(id string) error
}

// StorageAdapter adapts the ngebut.Storage interface to the Store interface
type StorageAdapter struct {
	// storage is the underlying storage implementation
	storage ngebut.Storage
	// ctx is the context used for storage operations
	ctx context.Context
}

// NewStorageAdapter creates a new storage adapter with the specified storage implementation
func NewStorageAdapter(storage ngebut.Storage) *StorageAdapter {
	return &StorageAdapter{
		storage: storage,
		ctx:     context.Background(),
	}
}

// Get retrieves a session from the storage by its ID
func (a *StorageAdapter) Get(id string) (*Session, error) {
	// Try to get the session from the storage
	data, err := a.storage.Get(a.ctx, id)
	if err != nil {
		// If the key doesn't exist, return nil without an error
		if errors.Is(err, ngebut.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// Unmarshal the session data
	session := &Session{}
	if err := unmarshalSession(data, session); err != nil {
		return nil, err
	}

	// Check if session has expired
	if time.Now().After(session.ExpiresAt) {
		// Delete the expired session
		_ = a.storage.Delete(a.ctx, id)
		return nil, nil
	}

	// Set the store field so the session can save itself
	session.store = a

	return session, nil
}

// Save saves a session to the storage
func (a *StorageAdapter) Save(session *Session) error {
	// Marshal the session data
	data, err := marshalSession(session)
	if err != nil {
		return err
	}

	// Calculate TTL
	var ttl time.Duration
	if !session.ExpiresAt.IsZero() {
		ttl = time.Until(session.ExpiresAt)
		if ttl <= 0 {
			// Session has expired, delete it
			return a.Delete(session.ID)
		}
	}

	// Save the session to the storage
	return a.storage.Set(a.ctx, session.ID, data, ttl)
}

// Delete removes a session from the storage by its ID
func (a *StorageAdapter) Delete(id string) error {
	return a.storage.Delete(a.ctx, id)
}

// marshalSession marshals a session to a byte slice
func marshalSession(session *Session) ([]byte, error) {
	// For simplicity, we'll use a simple string representation
	// In a real implementation, you would use a more efficient serialization format like JSON or gob
	data := fmt.Sprintf("%s|%d|%d|", session.ID, session.CreatedAt.Unix(), session.ExpiresAt.Unix())

	// Add the values
	for k, v := range session.Values {
		// Special handling for nil values
		if v == nil {
			data += fmt.Sprintf("%s=__NIL_VALUE__;", k)
		} else {
			// Include type information along with the value
			// This will help with proper unmarshaling
			switch v.(type) {
			case string:
				data += fmt.Sprintf("%s=string:%v;", k, v)
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
				data += fmt.Sprintf("%s=number:%v;", k, v)
			case float32, float64:
				data += fmt.Sprintf("%s=float:%v;", k, v)
			case bool:
				data += fmt.Sprintf("%s=bool:%v;", k, v)
			default:
				// For other types, just convert to string
				data += fmt.Sprintf("%s=other:%v;", k, v)
			}
		}
	}

	return []byte(data), nil
}

// unmarshalSession unmarshals a byte slice to a session
func unmarshalSession(data []byte, session *Session) error {
	// Convert the byte slice to a string
	dataStr := string(data)

	// Split the string by the separator
	parts := strings.Split(dataStr, "|")
	if len(parts) < 3 {
		return fmt.Errorf("invalid session data format")
	}

	// Parse the session ID
	session.ID = parts[0]

	// Parse the created at timestamp
	createdAt, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return err
	}
	session.CreatedAt = time.Unix(createdAt, 0)

	// Parse the expires at timestamp
	expiresAt, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return err
	}
	session.ExpiresAt = time.Unix(expiresAt, 0)

	// Initialize the values map
	session.Values = make(map[string]interface{})

	// Parse the values
	if len(parts) > 3 && parts[3] != "" {
		valuePairs := strings.Split(parts[3], ";")
		for _, pair := range valuePairs {
			if pair == "" {
				continue
			}
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) == 2 {
				// Special handling for nil values
				if kv[1] == "__NIL_VALUE__" {
					session.Values[kv[0]] = nil
				} else {
					// Check if the value has type information
					typeValue := strings.SplitN(kv[1], ":", 2)
					if len(typeValue) == 2 {
						// Parse the value based on its type
						switch typeValue[0] {
						case "string":
							session.Values[kv[0]] = typeValue[1]
						case "number":
							// Try to parse as int first
							if intVal, err := strconv.ParseInt(typeValue[1], 10, 64); err == nil {
								session.Values[kv[0]] = intVal
							} else if uintVal, err := strconv.ParseUint(typeValue[1], 10, 64); err == nil {
								session.Values[kv[0]] = uintVal
							} else {
								// If parsing fails, keep as string
								session.Values[kv[0]] = typeValue[1]
							}
						case "float":
							if floatVal, err := strconv.ParseFloat(typeValue[1], 64); err == nil {
								session.Values[kv[0]] = floatVal
							} else {
								// If parsing fails, keep as string
								session.Values[kv[0]] = typeValue[1]
							}
						case "bool":
							if boolVal, err := strconv.ParseBool(typeValue[1]); err == nil {
								session.Values[kv[0]] = boolVal
							} else {
								// If parsing fails, keep as string
								session.Values[kv[0]] = typeValue[1]
							}
						default:
							// For other types, keep as string
							session.Values[kv[0]] = typeValue[1]
						}
					} else {
						// Backward compatibility: if no type information, treat as string
						session.Values[kv[0]] = kv[1]
					}
				}
			}
		}
	}

	return nil
}

// sessionKey is used as a key for storing session in context
type sessionKey string

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
		if len(kv) == 2 {
			cookies[kv[0]] = kv[1]
		}
	}
	return cookies
}

// Manager handles session creation, retrieval, and management.
// It uses a configured Store implementation for session persistence.
type Manager struct {
	// config contains the session configuration options
	config Config

	// store is the storage backend for sessions
	store Store
}

// Get retrieves a session from the store using the context.
// It returns the session and an error if one occurred.
// If no session is found and there was no session ID in the request, a new session is created
// but no cookie is set by default. Use GetOrCreate to create a new session and set a cookie.
// If there was a session ID in the request but no session was found,
// a new session is created but no cookie is set to avoid setting cookies on every request.
func (m *Manager) Get(c *ngebut.Ctx) (*Session, error) {
	// Get the session ID from the specified source
	var sessionID string

	switch m.config.source {
	case "cookie":
		// Try to get the session ID from the cookie
		getSessionIDFromCookie(c, &m.config, &sessionID)
	case "header":
		// Try to get the session ID from the header
		sessionID = c.Request.Header.Get(m.config.sessionName)
	case "query":
		// Try to get the session ID from the query parameters
		sessionID = c.Request.URL.Query().Get(m.config.sessionName)
	default:
		// Default to cookie if source is invalid
		getSessionIDFromCookie(c, &m.config, &sessionID)
	}

	var session *Session
	var err error

	if sessionID != "" {
		// Try to get the session from the store
		session, err = m.store.Get(sessionID)
		if err != nil {
			return nil, err
		}

		// Set the store field if session was found
		if session != nil {
			session.store = m.store
		}
	}

	// If no session was found or it's expired, create a new one
	if session == nil {
		// Generate a new session ID
		var newID string
		if m.config.KeyGenerator != nil {
			// Use the custom key generator
			newID = m.config.KeyGenerator()
		} else {
			// Use the default generator
			var err error
			newID, err = generateSessionID()
			if err != nil {
				return nil, err
			}
		}

		// Create a new session
		session = &Session{
			ID:         newID,
			Values:     make(map[string]interface{}),
			CreatedAt:  time.Now(),
			ExpiresAt:  time.Now().Add(m.config.Expiration),
			store:      m.store,
			cookieName: m.config.sessionName,
			cookiePath: m.config.Path,
		}

		// In Get method, we don't set cookies by default
		// Use GetOrCreate to set cookies when creating a new session
	}

	return session, nil
}

// GetOrCreate retrieves a session from the store using the context.
// It returns the session and an error if one occurred.
// If no session is found and there was no session ID in the request, a new session is created
// and a cookie is set. If there was a session ID in the request but no session was found,
// a new session is created but no cookie is set to avoid setting cookies on every request.
func (m *Manager) GetOrCreate(c *ngebut.Ctx) (*Session, error) {
	// Get the session ID from the specified source
	var sessionID string

	switch m.config.source {
	case "cookie":
		// Try to get the session ID from the cookie
		getSessionIDFromCookie(c, &m.config, &sessionID)
	case "header":
		// Try to get the session ID from the header
		sessionID = c.Request.Header.Get(m.config.sessionName)
	case "query":
		// Try to get the session ID from the query parameters
		sessionID = c.Request.URL.Query().Get(m.config.sessionName)
	default:
		// Default to cookie if source is invalid
		getSessionIDFromCookie(c, &m.config, &sessionID)
	}

	var session *Session
	var err error

	if sessionID != "" {
		// Try to get the session from the store
		session, err = m.store.Get(sessionID)
		if err != nil {
			return nil, err
		}

		// Set the store field if session was found
		if session != nil {
			session.store = m.store
		}
	}

	// If no session was found or it's expired, create a new one
	if session == nil {
		// Generate a new session ID
		var newID string
		if m.config.KeyGenerator != nil {
			// Use the custom key generator
			newID = m.config.KeyGenerator()
		} else {
			// Use the default generator
			var err error
			newID, err = generateSessionID()
			if err != nil {
				return nil, err
			}
		}

		// Create a new session
		session = &Session{
			ID:         newID,
			Values:     make(map[string]interface{}),
			CreatedAt:  time.Now(),
			ExpiresAt:  time.Now().Add(m.config.Expiration),
			store:      m.store,
			cookieName: m.config.sessionName,
			cookiePath: m.config.Path,
		}

		// Only set a cookie if there was no session ID in the request
		// This prevents setting a cookie on every request
		if sessionID == "" && (m.config.source == "cookie" || m.config.source == "") {
			// Set the session cookie
			httpOnlyStr := ""
			if m.config.HttpOnly {
				httpOnlyStr = "; HttpOnly"
			}

			secureStr := ""
			if m.config.Secure {
				secureStr = "; Secure"
			}

			// Calculate MaxAge from Expiration
			maxAge := int(m.config.Expiration.Seconds())
			if maxAge <= 0 {
				maxAge = m.config.MaxAge // Fallback to MaxAge for backward compatibility
			}

			cookieName := m.config.sessionName
			if cookieName == "" {
				cookieName = m.config.CookieName // Fallback to CookieName for backward compatibility
			}

			c.Set("Set-Cookie", cookieName+"="+session.ID+
				"; Path="+m.config.Path+
				"; Max-Age="+strconv.Itoa(maxAge)+
				httpOnlyStr+
				secureStr)
		}
	}

	return session, nil
}

// NewManager creates a new session manager with the specified configuration and storage backend.
// It returns a pointer to the new Manager instance.
func NewManager(config Config, store Store) *Manager {
	return &Manager{
		config: config,
		store:  store,
	}
}

// generateSessionID generates a random session ID.
// It creates a 32-byte random value and encodes it as a URL-safe base64 string.
// Returns an error if the random number generator fails.
func generateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// UUIDv4 generates a random UUID v4 string.
// This function is compatible with the KeyGenerator type in Config.
func UUIDv4() string {
	// Implementation based on RFC 4122
	u := make([]byte, 16)
	_, err := rand.Read(u)
	if err != nil {
		// In case of error, return a default string
		return "00000000-0000-0000-0000-000000000000"
	}

	// Set version (4) and variant (2)
	u[6] = (u[6] & 0x0f) | 0x40 // Version 4
	u[8] = (u[8] & 0x3f) | 0x80 // Variant 2

	// Format as UUID string
	return fmt.Sprintf("%x-%x-%x-%x-%x", u[0:4], u[4:6], u[6:8], u[8:10], u[10:])
}

// SessionStore is a wrapper around Manager that provides a simpler API for session management.
// It is designed to be compatible with the gofiber session API.
type SessionStore struct {
	manager *Manager
}

// Get retrieves a session from the store using the context.
// It returns the session and an error if one occurred.
// This method does not set cookies when creating a new session.
func (s *SessionStore) Get(c *ngebut.Ctx) (*Session, error) {
	return s.manager.Get(c)
}

// GetOrCreate retrieves a session from the store using the context.
// It returns the session and an error if one occurred.
// If no session is found, a new session is created and a cookie is set.
func (s *SessionStore) GetOrCreate(c *ngebut.Ctx) (*Session, error) {
	return s.manager.GetOrCreate(c)
}

// NewStore creates a new session store.
// It accepts an optional configuration. If no configuration is provided, it uses the default configuration.
// If multiple configurations are provided, only the first one is used.
// It returns a Manager that can be used to get sessions.
func NewStore(config ...Config) *Manager {
	// Determine which config to use
	cfg := DefaultConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	var store Store
	if cfg.Storage != nil {
		// Use the provided storage
		store = NewStorageAdapter(cfg.Storage)
	} else {
		// Create a memory store by default using internal/memory
		memoryStorage := memory.New(time.Minute * 5) // Cleanup every 5 minutes
		store = NewStorageAdapter(memoryStorage)
	}

	return NewManager(cfg, store)
}

// New creates a new session store.
// It accepts an optional configuration. If no configuration is provided, it uses the default configuration.
// If multiple configurations are provided, only the first one is used.
// It returns a SessionStore that can be used to get sessions.
func New(config ...Config) *SessionStore {
	manager := NewStore(config...)
	return &SessionStore{
		manager: manager,
	}
}

// NewMiddleware creates a new session middleware.
// It accepts an optional configuration. If no configuration is provided, it uses the default configuration.
// If multiple configurations are provided, only the first one is used.
// The middleware handles session creation, retrieval, and persistence throughout the request lifecycle.
// It returns a middleware function compatible with the ngebut framework.
// If no session is found and there was no session ID in the request, a new session is created
// and a cookie is set. If there was a session ID in the request but no session was found,
// a new session is created but no cookie is set to avoid setting cookies on every request.
func NewMiddleware(config ...Config) interface{} {
	// Determine which config to use
	cfg := DefaultConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	var store Store
	if cfg.Storage != nil {
		// Use the provided storage
		store = NewStorageAdapter(cfg.Storage)
	} else {
		// Create a memory store by default using internal/memory
		memoryStorage := memory.New(time.Minute * 5) // Cleanup every 5 minutes
		store = NewStorageAdapter(memoryStorage)
	}
	manager := NewManager(cfg, store)

	// Return the middleware function
	return func(c *ngebut.Ctx) {
		// Get the session using GetOrCreate to ensure a cookie is set if a new session is created
		session, err := manager.GetOrCreate(c)
		if err != nil {
			c.Error(err)
			return
		}

		// Store the session in the request context for handlers to access
		// Create a new context with the session
		sessionCtx := context.WithValue(c.Request.Context(), sessionKey("session"), session)
		c.Request = c.Request.WithContext(sessionCtx)

		// Process the request
		c.Next()

		// Save any changes to the session after the request is processed
		if err := manager.store.Save(session); err != nil {
			c.Error(err)
			return
		}
	}
}

// GetSession retrieves the session from the context.
// It extracts the session object that was previously stored in the request context by the session middleware.
// Returns nil if the context doesn't contain a session, which can happen if the session middleware
// wasn't used or if there was an error during session processing.
func GetSession(c *ngebut.Ctx) *Session {
	if c.Request == nil {
		return nil
	}

	ctx := c.Request.Context()
	if ctx == nil {
		return nil
	}

	session, ok := ctx.Value(sessionKey("session")).(*Session)
	if !ok {
		return nil
	}

	return session
}

// Set stores a value in the session with the specified key.
// The value can be of any type that can be stored as an interface{}.
// If a value with the same key already exists, it will be overwritten.
func (s *Session) Set(key string, value interface{}) {
	s.Values[key] = value
}

// Get retrieves a value from the session by its key.
// It returns the value as an interface{} which may need to be type-asserted by the caller.
// If the key doesn't exist, it returns nil.
func (s *Session) Get(key string) interface{} {
	return s.Values[key]
}

// Delete removes a value from the session by its key.
// If the key doesn't exist, the operation is a no-op.
func (s *Session) Delete(key string) {
	delete(s.Values, key)
}

// Clear removes all values from the session.
// It resets the Values map to an empty map, effectively removing all stored key-value pairs.
func (s *Session) Clear() {
	s.Values = make(map[string]interface{})
}

// Keys returns all keys in the session.
// It returns a slice of strings containing all the keys in the session's Values map.
func (s *Session) Keys() []string {
	keys := make([]string, 0, len(s.Values))
	for k := range s.Values {
		keys = append(keys, k)
	}
	return keys
}

// Destroy destroys the session.
// It clears all values and marks the session for deletion by setting its expiry to a past time.
// The session will be removed from the store when Save is called.
// If a context is provided, it will also remove the session cookie from the client.
func (s *Session) Destroy(c ...*ngebut.Ctx) error {
	s.Clear()
	s.ExpiresAt = time.Now().Add(-1 * time.Hour) // Set expiry to the past

	// If a context is provided, remove the session cookie from the client
	if len(c) > 0 && c[0] != nil {
		// Get the cookie name and path from the session
		cookieName := s.cookieName
		cookiePath := s.cookiePath

		// Use defaults if not set (for backward compatibility)
		if cookieName == "" {
			cookieName = "session_id" // Default cookie name
		}
		if cookiePath == "" {
			cookiePath = "/" // Default cookie path
		}

		// Set an expired cookie to remove it from the client
		c[0].Cookie(&ngebut.Cookie{
			Name:     cookieName,
			Value:    "",
			Path:     cookiePath,
			MaxAge:   -1,                             // Negative MaxAge means delete the cookie
			Expires:  time.Now().Add(-1 * time.Hour), // Set expiry to the past
			HTTPOnly: true,
		})
	}

	return nil
}

// SetExpiry sets a specific expiration time for the session.
// It updates the ExpiresAt field of the session.
func (s *Session) SetExpiry(expiry time.Duration) {
	s.ExpiresAt = time.Now().Add(expiry)
}

// Save saves the session to the store.
// It returns an error if the save operation fails.
// This method requires the session to be associated with a store.
func (s *Session) Save() error {
	// Get the session from the context
	if s.store == nil {
		return fmt.Errorf("session has no associated store")
	}

	return s.store.Save(s)
}
