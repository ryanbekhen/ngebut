package ngebut

import (
	"net/textproto"
	"strings"
	"sync"
)

// headerMutex protects Header operations from concurrent access
var headerMutex sync.RWMutex

// Header represents the key-value pairs in an HTTP header.
// The keys should be in canonical form, as returned by
// textproto.CanonicalMIMEHeaderKey.
type Header map[string][]string

// Add adds the key, value pair to the header.
// It appends to any existing values associated with key.
// The key is case insensitive; it is canonicalized by
// textproto.CanonicalMIMEHeaderKey.
// This optimized version reduces allocations by appending directly when possible.
func (h Header) Add(key, value string) {
	key = textproto.CanonicalMIMEHeaderKey(key)

	// Use a single lock for the entire operation to avoid race conditions
	// This is simpler and often more efficient than using multiple locks
	headerMutex.Lock()
	defer headerMutex.Unlock()

	// Check if the key exists
	values, exists := h[key]

	if !exists || values == nil {
		// Need to create a new entry
		h[key] = []string{value}
		return
	}

	// Append to existing values
	// This will only allocate a new backing array if the capacity is exceeded
	h[key] = append(values, value)
}

// Set sets the header entries associated with key to the
// single element value. It replaces any existing values
// associated with key. The key is case insensitive; it is
// canonicalized by textproto.CanonicalMIMEHeaderKey.
// To use non-canonical keys, assign to the map directly.
func (h Header) Set(key, value string) {
	key = textproto.CanonicalMIMEHeaderKey(key)

	// Create the slice outside the lock
	values := []string{value}

	// Shorter critical section
	headerMutex.Lock()
	h[key] = values
	headerMutex.Unlock()
}

// Get gets the first value associated with the given key.
// If there are no values associated with the key, Get returns "".
// It is case insensitive; textproto.CanonicalMIMEHeaderKey is used
// to canonicalize the provided key.
// To use non-canonical keys, access the map directly.
func (h Header) Get(key string) string {
	key = textproto.CanonicalMIMEHeaderKey(key)

	headerMutex.RLock()
	values := h[key]
	headerMutex.RUnlock()

	if len(values) == 0 {
		return ""
	}
	return values[0]
}

// Values returns all values associated with the given key.
// It is case insensitive; textproto.CanonicalMIMEHeaderKey is
// used to canonicalize the provided key. To use non-canonical
// keys, access the map directly.
// The returned slice is a copy to avoid concurrent modification issues.
// This optimized version avoids unnecessary copying for single-value headers.
func (h Header) Values(key string) []string {
	key = textproto.CanonicalMIMEHeaderKey(key)

	headerMutex.RLock()
	values := h[key]
	headerMutex.RUnlock()

	// Fast path for empty values
	if len(values) == 0 {
		return nil
	}

	// Fast path for single-value headers (common case)
	// Return a new slice with the same backing array
	if len(values) == 1 {
		return values[:1:1] // Create a slice with capacity=1 to prevent appends
	}

	// For multi-value headers, create a copy to avoid concurrent modification
	result := make([]string, len(values))
	copy(result, values)
	return result
}

// Del deletes the values associated with key.
// The key is case insensitive; it is canonicalized by
// textproto.CanonicalMIMEHeaderKey.
func (h Header) Del(key string) {
	key = textproto.CanonicalMIMEHeaderKey(key)

	// Shorter critical section
	headerMutex.Lock()
	delete(h, key)
	headerMutex.Unlock()
}

// Clone returns a copy of h or nil if h is nil.
func (h Header) Clone() Header {
	if h == nil {
		return nil
	}

	// First, get a snapshot of the keys and count values
	// This reduces the time we hold the read lock
	headerMutex.RLock()
	keys := make([]string, 0, len(h))
	valuesCounts := make(map[string]int, len(h))
	totalValues := 0

	for k, vv := range h {
		keys = append(keys, k)
		count := len(vv)
		valuesCounts[k] = count
		totalValues += count
	}
	headerMutex.RUnlock()

	// Create a new header
	h2 := make(Header, len(keys))

	// If there are no values, return the empty header
	if totalValues == 0 {
		return h2
	}

	// Create a shared backing array for all values
	sv := make([]string, totalValues)

	// Copy values for each key with minimal locking
	svIndex := 0
	for _, k := range keys {
		headerMutex.RLock()
		vv, exists := h[k]
		if !exists {
			headerMutex.RUnlock()
			continue
		}

		// Copy the values while holding the lock
		n := copy(sv[svIndex:], vv)
		headerMutex.RUnlock()

		// Set up the slice in the new header
		h2[k] = sv[svIndex : svIndex+n : svIndex+n]
		svIndex += n
	}

	return h2
}

// WriteSubset writes a header in wire format.
// If exclude is not nil, keys where exclude[key] == true are not written.
// This optimized version reduces allocations by avoiding unnecessary copying.
func (h Header) WriteSubset(w stringWriter, exclude map[string]bool) error {
	// First, get a snapshot of the keys to process
	// This reduces the time we hold the read lock
	headerMutex.RLock()

	// Pre-allocate keys slice to avoid resizing
	keys := make([]string, 0, len(h))
	for key := range h {
		if exclude == nil || !exclude[key] {
			keys = append(keys, key)
		}
	}
	headerMutex.RUnlock()

	// Process each key individually with minimal locking
	for _, key := range keys {
		// Get the values for this key
		headerMutex.RLock()
		values, exists := h[key]
		if !exists || len(values) == 0 {
			headerMutex.RUnlock()
			continue
		}

		// Create a reference to the values slice to use outside the lock
		// This avoids copying the entire slice
		valuesCopy := values
		headerMutex.RUnlock()

		// Write each value
		for _, v := range valuesCopy {
			// Clean the value (trim spaces, replace newlines)
			// Only allocate a new string if necessary
			cleaned := v
			if strings.ContainsAny(v, "\r\n ") {
				cleaned = strings.TrimSpace(v)
				cleaned = strings.ReplaceAll(cleaned, "\n", " ")
				cleaned = strings.ReplaceAll(cleaned, "\r", " ")
			}

			// Write the header line
			if _, err := w.WriteString(key + ": " + cleaned + "\r\n"); err != nil {
				return err
			}
		}
	}

	return nil
}

// Write writes a header in wire format.
func (h Header) Write(w stringWriter) error {
	return h.WriteSubset(w, nil)
}

// NewHeader creates a new empty Header with pre-allocated capacity.
func NewHeader() *Header {
	h := make(Header, 8) // Pre-allocate with capacity for common headers
	return &h
}

// NewHeaderFromMap creates a new Header from a map[string][]string.
// This optimized version avoids unnecessary copying of values when possible.
func NewHeaderFromMap(m map[string][]string) *Header {
	// Fast path for empty map
	if len(m) == 0 {
		h := make(Header, 0)
		return &h
	}

	// Pre-allocate with exact size
	h := make(Header, len(m))

	// Copy only non-empty values
	for k, v := range m {
		if len(v) == 0 {
			continue
		}
		h[k] = v // Direct reference, no copy
	}

	return &h
}

// UpdateHeaderFromMap updates an existing Header with values from a map[string][]string.
// This function avoids allocating a new Header map, reducing memory allocations.
// It returns the updated Header.
func UpdateHeaderFromMap(h *Header, m map[string][]string) *Header {
	// Clear the existing header
	for k := range *h {
		delete(*h, k)
	}

	// Fast path for empty map
	if len(m) == 0 {
		return h
	}

	// Copy only non-empty values
	for k, v := range m {
		if len(v) == 0 {
			continue
		}
		(*h)[k] = v // Direct reference, no copy
	}

	return h
}

// stringWriter is the interface that wraps the WriteString method.
// It is used by Header.Write and Header.WriteSubset to write headers in wire format.
type stringWriter interface {
	// WriteString writes a string and returns the number of bytes written and any error encountered.
	WriteString(s string) (n int, err error)
}
