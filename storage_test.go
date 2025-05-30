package ngebut

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNewError tests the NewError function
func TestNewError(t *testing.T) {
	// Create a new error
	err := NewError("test error")

	// Check that the error is not nil
	assert.NotNil(t, err, "NewError should not return nil")

	// Check that the error is of the correct type
	var storageErr *StorageError
	ok := errors.As(err, &storageErr)
	assert.True(t, ok, "NewError should return a *StorageError")

	// Check that the error message is correct
	assert.Equal(t, "test error", storageErr.Error(), "Error message should match")
}

// TestStorageError tests the StorageError type
func TestStorageError(t *testing.T) {
	// Create a new StorageError
	err := &StorageError{message: "test error"}

	// Check that the error message is correct
	assert.Equal(t, "test error", err.Error(), "Error message should match")
}

// TestErrNotFound tests the ErrNotFound variable
func TestErrNotFound(t *testing.T) {
	// Check that ErrNotFound is not nil
	assert.NotNil(t, ErrNotFound, "ErrNotFound should not be nil")

	// Check that ErrNotFound is of the correct type
	storageErr, ok := ErrNotFound.(*StorageError)
	assert.True(t, ok, "ErrNotFound should be a *StorageError")

	// Check that the error message is correct
	assert.Equal(t, "key not found", storageErr.Error(), "Error message should match")
}

// TestStorageInterface tests that the Storage interface is defined correctly
func TestStorageInterface(t *testing.T) {
	// This test doesn't actually test any functionality, but it ensures that
	// the Storage interface is defined correctly and can be used as a type.
	var _ Storage = (*mockStorage)(nil)
	assert.NotPanics(t, func() {
		var _ Storage = &mockStorage{}
	}, "mockStorage should implement Storage interface")
}

// mockStorage is a mock implementation of the Storage interface for testing
type mockStorage struct{}

func (m *mockStorage) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, ErrNotFound
}

func (m *mockStorage) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return nil
}

func (m *mockStorage) Delete(ctx context.Context, key string) error {
	return nil
}

func (m *mockStorage) Clear(ctx context.Context) error {
	return nil
}

func (m *mockStorage) Has(ctx context.Context, key string) (bool, error) {
	return false, nil
}
