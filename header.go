package ngebut

import (
	"net/textproto"
	"strings"
)

// Header represents the key-value pairs in an HTTP header.
// The keys should be in canonical form, as returned by
// textproto.CanonicalMIMEHeaderKey.
type Header map[string][]string

// Add adds the key, value pair to the header.
// It appends to any existing values associated with key.
// The key is case insensitive; it is canonicalized by
// textproto.CanonicalMIMEHeaderKey.
func (h Header) Add(key, value string) {
	textproto.MIMEHeader(h).Add(key, value)
}

// Set sets the header entries associated with key to the
// single element value. It replaces any existing values
// associated with key. The key is case insensitive; it is
// canonicalized by textproto.CanonicalMIMEHeaderKey.
// To use non-canonical keys, assign to the map directly.
func (h Header) Set(key, value string) {
	textproto.MIMEHeader(h).Set(key, value)
}

// Get gets the first value associated with the given key.
// If there are no values associated with the key, Get returns "".
// It is case insensitive; textproto.CanonicalMIMEHeaderKey is used
// to canonicalize the provided key.
// To use non-canonical keys, access the map directly.
func (h Header) Get(key string) string {
	return textproto.MIMEHeader(h).Get(key)
}

// Values returns all values associated with the given key.
// It is case insensitive; textproto.CanonicalMIMEHeaderKey is
// used to canonicalize the provided key. To use non-canonical
// keys, access the map directly.
// The returned slice is not a copy.
func (h Header) Values(key string) []string {
	return textproto.MIMEHeader(h).Values(key)
}

// Del deletes the values associated with key.
// The key is case insensitive; it is canonicalized by
// textproto.CanonicalMIMEHeaderKey.
func (h Header) Del(key string) {
	textproto.MIMEHeader(h).Del(key)
}

// Clone returns a copy of h or nil if h is nil.
func (h Header) Clone() Header {
	if h == nil {
		return nil
	}

	// Find total number of values.
	nv := 0
	for _, vv := range h {
		nv += len(vv)
	}
	sv := make([]string, nv) // shared backing array for headers' values
	h2 := make(Header, len(h))
	for k, vv := range h {
		n := copy(sv, vv)
		h2[k] = sv[:n:n]
		sv = sv[n:]
	}
	return h2
}

// WriteSubset writes a header in wire format.
// If exclude is not nil, keys where exclude[key] == true are not written.
func (h Header) WriteSubset(w stringWriter, exclude map[string]bool) error {
	for key, values := range h {
		if exclude != nil && exclude[key] {
			continue
		}
		for _, v := range values {
			v = strings.TrimSpace(v)
			v = strings.ReplaceAll(v, "\n", " ")
			v = strings.ReplaceAll(v, "\r", " ")
			if _, err := w.WriteString(key + ": " + v + "\r\n"); err != nil {
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

// stringWriter is the interface that wraps the WriteString method.
// It is used by Header.Write and Header.WriteSubset to write headers in wire format.
type stringWriter interface {
	// WriteString writes a string and returns the number of bytes written and any error encountered.
	WriteString(s string) (n int, err error)
}
