package ngebut

import (
	"sync"
)

// paramKey is a type for URL parameter keys to avoid string allocations in context
type paramKey string

// paramContextKey is the key used to store the parameter context in the request context
type paramContextKey struct{}

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

	paramCtx, ok := c.Request.Context().Value(paramContextKey{}).(map[paramKey]string)
	if !ok || paramCtx == nil {
		return ""
	}

	// Return the parameter value
	return paramCtx[paramKey(key)]
}
