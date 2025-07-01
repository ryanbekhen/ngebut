package ngebut

import (
	"github.com/ryanbekhen/ngebut/internal/unsafe"
)

// b2s converts a byte slice to a string without memory allocation.
// Note: The returned string must not be modified, as it points to the same
// memory as the byte slice.
func b2s(b []byte) string {
	return unsafe.B2S(b)
}

// s2b converts a string to a byte slice without memory allocation.
// Note: The returned byte slice must not be modified, as it points to the same
// memory as the string.
func s2b(s string) []byte {
	return unsafe.S2B(s)
}

// equalBytesUnsafe compares a byte slice with a string without memory allocation.
// This is more efficient than bytes.Equal(a, []byte(b)) as it avoids the allocation
// of a new byte slice.
func equalBytesUnsafe(a []byte, b string) bool {
	return unsafe.EqualBytes(a, b)
}
