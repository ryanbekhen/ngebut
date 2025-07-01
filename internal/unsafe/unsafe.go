package unsafe

import (
	"bytes"
	"reflect"
	"unsafe"
)

// B2S converts a byte slice to a string without memory allocation.
// Note: The returned string must not be modified, as it points to the same
// memory as the byte slice.
func B2S(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// S2B converts a string to a byte slice without memory allocation.
// Note: The returned byte slice must not be modified, as it points to the same
// memory as the string.
func S2B(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&bh))
}

// EqualBytes compares a byte slice with a string without memory allocation.
// This is more efficient than bytes.Equal(a, []byte(b)) as it avoids the allocation
// of a new byte slice.
func EqualBytes(a []byte, b string) bool {
	bBytes := S2B(b)
	if len(a) != len(bBytes) {
		return false
	}

	// Fast path for empty strings
	if len(a) == 0 {
		return true
	}

	// Fast path for single character comparison
	if len(a) == 1 {
		return a[0] == bBytes[0]
	}

	// Use bytes.Equal for the comparison
	return bytes.Equal(a, bBytes)
}

// AsInternal is a generic function to cast a struct pointer to another type
// with the same memory layout.
// Note: Use with extreme caution, as it bypasses Go's type system.
func AsInternal[T any, U any](ptr *T) *U {
	return (*U)(unsafe.Pointer(ptr))
}
