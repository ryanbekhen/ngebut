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

// Params is a fixed-size struct for URL parameters to avoid allocations
type Params struct {
	keys   [4]string
	values [4]string
	len    int
}

// Get retrieves a parameter value by key
// Returns the value and a boolean indicating if the key was found
func (p *Params) Get(key string) (string, bool) {
	// Linear search for parameters
	for i := 0; i < p.len; i++ {
		if p.keys[i] == key {
			return p.values[i], true
		}
	}
	return "", false
}

// Set sets a parameter value by key
// If the key already exists, its value is updated
// If the key doesn't exist, a new entry is added
func (p *Params) Set(key, value string) {
	// First check if the key already exists
	for i := 0; i < p.len; i++ {
		if p.keys[i] == key {
			p.values[i] = value
			return
		}
	}

	// Key doesn't exist, add a new entry if there's space
	if p.len < len(p.keys) {
		p.keys[p.len] = key
		p.values[p.len] = value
		p.len++
	}
}

// Reset clears all parameters
func (p *Params) Reset() {
	p.len = 0
}

// paramsPool is a pool of Params structs for reuse
var paramsPool = sync.Pool{
	New: func() interface{} {
		return &Params{}
	},
}

// getParams gets a Params struct from the pool
func getParams() *Params {
	return paramsPool.Get().(*Params)
}

// releaseParams returns a Params struct to the pool
func releaseParams(p *Params) {
	p.Reset()
	paramsPool.Put(p)
}

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
			entries: make([]paramEntry, 0, 32), // Increased capacity to further reduce reallocations
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

// stringHash computes a simple hash code for a string
// This is a fast, non-cryptographic hash function suitable for string lookups
// It's based on the FNV-1a hash algorithm
func stringHash(s string) uint32 {
	h := uint32(2166136261)
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return h
}

// routeParams is a more efficient structure for storing route parameters
// It uses separate slices for keys and values instead of a map or slice of structs
type routeParams struct {
	// Fixed-size array for small number of parameters (most common case)
	// This avoids allocations for routes with few parameters
	// Placed first for better cache locality in the common case
	fixedKeys   [16]string // Increased size to handle more parameters without allocations
	fixedValues [16]string // Increased size to handle more parameters without allocations
	fixedHashes [16]uint32 // Hash codes for fixed keys to avoid string comparisons
	count       int        // Number of parameters in fixed arrays

	// Dynamic slices for routes with many parameters (rare case)
	keys   []string
	values []string
	hashes []uint32 // Hash codes for dynamic keys
}

// Get retrieves a parameter value by key
// Returns the value and a boolean indicating if the key was found
func (rp *routeParams) Get(key string) (string, bool) {
	// Compute hash code once
	hash := stringHash(key)

	// First check fixed-size arrays (fastest, zero allocation)
	// Use hash comparison first, then string comparison only if hash matches
	for i := 0; i < rp.count; i++ {
		// Hash comparison is much faster than string comparison
		if rp.fixedHashes[i] == hash && rp.fixedKeys[i] == key {
			return rp.fixedValues[i], true
		}
	}

	// If not found in fixed arrays, check dynamic slices
	// Use hash comparison first, then string comparison only if hash matches
	for i := 0; i < len(rp.keys); i++ {
		// Hash comparison is much faster than string comparison
		if rp.hashes[i] == hash && rp.keys[i] == key {
			return rp.values[i], true
		}
	}
	return "", false
}

// Set sets a parameter value by key
// If the key already exists, its value is updated
// If the key doesn't exist, a new entry is added
func (rp *routeParams) Set(key, value string) {
	// Compute hash code once
	hash := stringHash(key)

	// First check if the key already exists in fixed arrays
	for i := 0; i < rp.count; i++ {
		// Hash comparison is much faster than string comparison
		if rp.fixedHashes[i] == hash && rp.fixedKeys[i] == key {
			rp.fixedValues[i] = value
			return
		}
	}

	// Then check dynamic slices
	for i := 0; i < len(rp.keys); i++ {
		// Hash comparison is much faster than string comparison
		if rp.hashes[i] == hash && rp.keys[i] == key {
			rp.values[i] = value
			return
		}
	}

	// Key doesn't exist, try to add to fixed arrays first
	if rp.count < len(rp.fixedKeys) {
		rp.fixedKeys[rp.count] = key
		rp.fixedValues[rp.count] = value
		rp.fixedHashes[rp.count] = hash
		rp.count++
		return
	}

	// If fixed arrays are full, add to dynamic slices
	rp.keys = append(rp.keys, key)
	rp.values = append(rp.values, value)
	rp.hashes = append(rp.hashes, hash)
}

// Reset clears all parameters
func (rp *routeParams) Reset() {
	rp.count = 0
	rp.keys = rp.keys[:0]
	rp.values = rp.values[:0]
	rp.hashes = rp.hashes[:0]
}

// routeParamsPool is a pool of routeParams structs for reuse
var routeParamsPool = sync.Pool{
	New: func() interface{} {
		return &routeParams{
			keys:   make([]string, 0, 32), // Increased capacity to reduce reallocations
			values: make([]string, 0, 32), // Increased capacity to reduce reallocations
			hashes: make([]uint32, 0, 32), // Pre-allocate hash codes slice with same capacity
			count:  0,
		}
	},
}

// getRouteParams gets a routeParams struct from the pool
func getRouteParams() *routeParams {
	return routeParamsPool.Get().(*routeParams)
}

// releaseRouteParams returns a routeParams struct to the pool
func releaseRouteParams(rp *routeParams) {
	rp.Reset()
	routeParamsPool.Put(rp)
}
