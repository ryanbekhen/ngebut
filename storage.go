package ngebut

import (
	"context"
	"time"
)

// Storage defines the interface for storage implementations.
// This interface can be used for various storage needs including session management.
type Storage interface {
	// Get retrieves a value for the given key.
	// Returns ErrNotFound if the key doesn't exist.
	Get(ctx context.Context, key string) ([]byte, error)

	// Set stores a value for the given key.
	// If ttl is positive, the key will expire after the specified duration.
	// If ttl is zero or negative, the key will not expire.
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Delete removes a key from the storage.
	// It's not an error to delete a non-existent key.
	Delete(ctx context.Context, key string) error

	// Clear removes all keys from the storage.
	Clear(ctx context.Context) error

	// Has checks if a key exists in the storage.
	Has(ctx context.Context, key string) (bool, error)
}

// ErrNotFound is returned when a key is not found in the storage.
var ErrNotFound = NewError("key not found")

// NewError creates a new error with the given message.
func NewError(message string) error {
	return &StorageError{message: message}
}

// StorageError represents a storage error.
type StorageError struct {
	message string
}

// Error returns the error message.
func (e *StorageError) Error() string {
	return e.message
}
