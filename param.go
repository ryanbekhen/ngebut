package ngebut

import (
	"sync"
)

// paramKey is a type for URL parameter keys to avoid string allocations in context
type paramKey string

// paramContextKey is the key used to store the parameter context in the request context
type paramContextKey struct{}

// predefined instance of paramContextKey to avoid creating a new one for each lookup
var paramCtxKey = paramContextKey{}

// paramMap is a reusable map for URL parameters
var paramMapPool = sync.Pool{
	New: func() interface{} {
		return make(map[string]string, 8) // Pre-allocate with capacity for 8 params
	},
}

// getParamMap gets a parameter map from the pool
func getParamMap() map[string]string {
	return paramMapPool.Get().(map[string]string)
}

// releaseParamMap returns a parameter map to the pool
func releaseParamMap(m map[string]string) {
	// Clear the map
	for k := range m {
		delete(m, k)
	}
	paramMapPool.Put(m)
}

// GetParam retrieves a URL parameter from the request context
func (c *Ctx) GetParam(key string) string {
	// Get the parameter context from the request context
	if c.Request == nil {
		return ""
	}

	paramCtx, ok := c.Request.Context().Value(paramCtxKey).(map[paramKey]string)
	if !ok || paramCtx == nil {
		return ""
	}

	// Return the parameter value
	return paramCtx[paramKey(key)]
}

// paramEntry represents a single parameter entry in a route
type paramEntry struct {
	key   string
	value string
}

// paramSlice is a slice of parameter entries with optimized access
// It's more efficient than a map for small numbers of parameters
type paramSlice struct {
	entries []paramEntry
}

// Get retrieves a parameter value by key
// Returns the value and a boolean indicating if the key was found
func (ps *paramSlice) Get(key string) (string, bool) {
	// Linear search for parameters
	for _, entry := range ps.entries {
		if entry.key == key {
			return entry.value, true
		}
	}
	return "", false
}

// Set sets a parameter value by key
// If the key already exists, its value is updated
// If the key doesn't exist, a new entry is added
func (ps *paramSlice) Set(key, value string) {
	// First check if the key already exists
	for i, entry := range ps.entries {
		if entry.key == key {
			ps.entries[i].value = value
			return
		}
	}

	// Key doesn't exist, add a new entry
	ps.entries = append(ps.entries, paramEntry{key: key, value: value})
}

// paramSlicePool is a pool of parameter slices for reuse
var paramSlicePool = sync.Pool{
	New: func() interface{} {
		// Create a new paramSlice with pre-allocated entries
		// Most routes have fewer than 8 parameters, so this is a good balance
		// Pre-allocate the entries slice with capacity for common routes
		return &paramSlice{
			entries: make([]paramEntry, 0, 16), // Increased capacity to reduce reallocations
		}
	},
}

// getParamSlice gets a parameter slice from the pool
func getParamSlice() *paramSlice {
	return paramSlicePool.Get().(*paramSlice)
}

// releaseParamSlice returns a parameter slice to the pool
func releaseParamSlice(ps *paramSlice) {
	// Clear the slice
	ps.entries = ps.entries[:0]
	paramSlicePool.Put(ps)
}
