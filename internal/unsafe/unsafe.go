package unsafe

import (
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
	return unsafe.Slice((*byte)(unsafe.StringData(s)), len(s))
}
