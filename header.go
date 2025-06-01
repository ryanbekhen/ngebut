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
func (h Header) Add(key, value string) {
	key = textproto.CanonicalMIMEHeaderKey(key)

	// Use a shorter critical section
	headerMutex.RLock()
	values, exists := h[key]
	headerMutex.RUnlock()

	if !exists {
		// Need to create a new entry
		headerMutex.Lock()
		// Check again in case another goroutine added it
		if h[key] == nil {
			h[key] = []string{value}
		} else {
			h[key] = append(h[key], value)
		}
		headerMutex.Unlock()
	} else {
		// Append to existing values
		newValues := make([]string, len(values), len(values)+1)
		copy(newValues, values)
		newValues = append(newValues, value)

		headerMutex.Lock()
		h[key] = newValues
		headerMutex.Unlock()
	}
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
func (h Header) Values(key string) []string {
	key = textproto.CanonicalMIMEHeaderKey(key)

	headerMutex.RLock()
	values := h[key]
	headerMutex.RUnlock()

	// Return a copy to avoid concurrent modification issues
	if len(values) == 0 {
		return nil
	}

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
func (h Header) WriteSubset(w stringWriter, exclude map[string]bool) error {
	// First, get a snapshot of the keys and values
	// This reduces the time we hold the read lock
	headerMutex.RLock()
	// Pre-allocate to avoid resizing
	keyValuePairs := make([]struct {
		key    string
		values []string
	}, 0, len(h))

	for key, values := range h {
		if exclude != nil && exclude[key] {
			continue
		}
		// Make a copy of values to avoid concurrent modification
		valuesCopy := make([]string, len(values))
		copy(valuesCopy, values)
		keyValuePairs = append(keyValuePairs, struct {
			key    string
			values []string
		}{key, valuesCopy})
	}
	headerMutex.RUnlock()

	// Now write the headers without holding the lock
	for _, pair := range keyValuePairs {
		for _, v := range pair.values {
			v = strings.TrimSpace(v)
			v = strings.ReplaceAll(v, "\n", " ")
			v = strings.ReplaceAll(v, "\r", " ")
			if _, err := w.WriteString(pair.key + ": " + v + "\r\n"); err != nil {
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

// NewHeader creates a new empty Header.
func NewHeader() *Header {
	h := make(Header)
	return &h
}

// NewHeaderFromMap creates a new Header from a map[string][]string.
func NewHeaderFromMap(m map[string][]string) *Header {
	h := make(Header, len(m))
	for k, v := range m {
		h[k] = v
	}
	return &h
}

// stringWriter is the interface that wraps the WriteString method.
// It is used by Header.Write and Header.WriteSubset to write headers in wire format.
type stringWriter interface {
	// WriteString writes a string and returns the number of bytes written and any error encountered.
	WriteString(s string) (n int, err error)
}
