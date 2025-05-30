package memory

import (
	"context"
	"testing"
	"time"

	"github.com/ryanbekhen/ngebut"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStorageImplementsInterface verifies that Storage implements the ngebut.Storage interface
func TestStorageImplementsInterface(t *testing.T) {
	// This is a compile-time check
	var _ ngebut.Storage = (*Storage)(nil)
}

// TestNew verifies that New creates a Storage instance with the expected properties
func TestNew(t *testing.T) {
	// Test with cleanup disabled
	s1 := New(0)
	require.NotNil(t, s1, "New(0) returned nil")
	assert.Nil(t, s1.cleanupTicker, "New(0) should not create a cleanup ticker")
	assert.Nil(t, s1.stopCleanup, "New(0) should not create a stop channel")

	// Test with cleanup enabled
	s2 := New(time.Second)
	require.NotNil(t, s2, "New(time.Second) returned nil")
	assert.NotNil(t, s2.cleanupTicker, "New(time.Second) should create a cleanup ticker")
	assert.NotNil(t, s2.stopCleanup, "New(time.Second) should create a stop channel")

	// Clean up
	_ = s2.Close()
}

// TestSet tests the Set method
func TestSet(t *testing.T) {
	s := New(0)
	ctx := context.Background()

	// Test setting a value
	err := s.Set(ctx, "key1", []byte("value1"), 0)
	assert.NoError(t, err, "Set returned an error")

	// Test setting a value with TTL
	err = s.Set(ctx, "key2", []byte("value2"), time.Minute)
	assert.NoError(t, err, "Set with TTL returned an error")

	// Verify the values were set correctly
	s.mu.RLock()
	item1, exists1 := s.items["key1"]
	item2, exists2 := s.items["key2"]
	s.mu.RUnlock()

	assert.True(t, exists1, "key1 was not set")
	assert.Equal(t, "value1", string(item1.value), "key1 value is incorrect")
	assert.True(t, item1.expireAt.IsZero(), "key1 should not have an expiration time")

	assert.True(t, exists2, "key2 was not set")
	assert.Equal(t, "value2", string(item2.value), "key2 value is incorrect")
	assert.False(t, item2.expireAt.IsZero(), "key2 should have an expiration time")
}

// TestGet tests the Get method
func TestGet(t *testing.T) {
	s := New(0)
	ctx := context.Background()

	// Set up test data
	_ = s.Set(ctx, "key1", []byte("value1"), 0)
	_ = s.Set(ctx, "key2", []byte("value2"), time.Minute)
	_ = s.Set(ctx, "expired", []byte("expired"), time.Nanosecond)

	// Wait for the item to expire
	time.Sleep(time.Millisecond * 10)

	// Test getting an existing value
	value1, err := s.Get(ctx, "key1")
	assert.NoError(t, err, "Get for key1 returned an error")
	assert.Equal(t, "value1", string(value1), "Get for key1 returned incorrect value")

	// Test getting a value with TTL
	value2, err := s.Get(ctx, "key2")
	assert.NoError(t, err, "Get for key2 returned an error")
	assert.Equal(t, "value2", string(value2), "Get for key2 returned incorrect value")

	// Test getting a non-existent key
	_, err = s.Get(ctx, "nonexistent")
	assert.Equal(t, ngebut.ErrNotFound, err, "Get for nonexistent key should return ErrNotFound")

	// Test getting an expired key
	_, err = s.Get(ctx, "expired")
	assert.Equal(t, ngebut.ErrNotFound, err, "Get for expired key should return ErrNotFound")
}

// TestDelete tests the Delete method
func TestDelete(t *testing.T) {
	s := New(0)
	ctx := context.Background()

	// Set up test data
	_ = s.Set(ctx, "key1", []byte("value1"), 0)

	// Test deleting an existing key
	err := s.Delete(ctx, "key1")
	assert.NoError(t, err, "Delete returned an error")

	// Verify the key was deleted
	s.mu.RLock()
	_, exists := s.items["key1"]
	s.mu.RUnlock()

	assert.False(t, exists, "key1 was not deleted")

	// Test deleting a non-existent key (should not error)
	err = s.Delete(ctx, "nonexistent")
	assert.NoError(t, err, "Delete for nonexistent key returned an error")
}

// TestClear tests the Clear method
func TestClear(t *testing.T) {
	s := New(0)
	ctx := context.Background()

	// Set up test data
	_ = s.Set(ctx, "key1", []byte("value1"), 0)
	_ = s.Set(ctx, "key2", []byte("value2"), 0)

	// Test clearing all keys
	err := s.Clear(ctx)
	assert.NoError(t, err, "Clear returned an error")

	// Verify all keys were cleared
	s.mu.RLock()
	count := len(s.items)
	s.mu.RUnlock()

	assert.Equal(t, 0, count, "Clear did not remove all items")
}

// TestHas tests the Has method
func TestHas(t *testing.T) {
	s := New(0)
	ctx := context.Background()

	// Set up test data
	_ = s.Set(ctx, "key1", []byte("value1"), 0)
	_ = s.Set(ctx, "key2", []byte("value2"), time.Minute)
	_ = s.Set(ctx, "expired", []byte("expired"), time.Nanosecond)

	// Wait for the item to expire
	time.Sleep(time.Millisecond * 10)

	// Test checking an existing key
	exists, err := s.Has(ctx, "key1")
	assert.NoError(t, err, "Has for key1 returned an error")
	assert.True(t, exists, "Has for key1 returned false, expected true")

	// Test checking a key with TTL
	exists, err = s.Has(ctx, "key2")
	assert.NoError(t, err, "Has for key2 returned an error")
	assert.True(t, exists, "Has for key2 returned false, expected true")

	// Test checking a non-existent key
	exists, err = s.Has(ctx, "nonexistent")
	assert.NoError(t, err, "Has for nonexistent key returned an error")
	assert.False(t, exists, "Has for nonexistent key returned true, expected false")

	// Test checking an expired key
	exists, err = s.Has(ctx, "expired")
	assert.NoError(t, err, "Has for expired key returned an error")
	assert.False(t, exists, "Has for expired key returned true, expected false")
}

// TestClose tests the Close method
func TestClose(t *testing.T) {
	// Create a storage with cleanup enabled
	s := New(time.Second)

	// Test closing the storage
	err := s.Close()
	assert.NoError(t, err, "Close returned an error")

	// Verify the cleanup goroutine was stopped
	// This is a bit tricky to test directly, but we can check that the ticker was stopped
	// by trying to send to the stopCleanup channel, which should not block if Close worked correctly
	select {
	case s.stopCleanup <- struct{}{}:
		assert.Fail(t, "stopCleanup channel is still open after Close")
	default:
		// This is expected, the channel should be closed
	}
}

// TestCleanup tests the cleanup method indirectly
func TestCleanup(t *testing.T) {
	// Create a storage with a short cleanup interval
	s := New(time.Millisecond * 50)
	ctx := context.Background()

	// Set up test data with short TTLs
	_ = s.Set(ctx, "key1", []byte("value1"), time.Millisecond*10)
	_ = s.Set(ctx, "key2", []byte("value2"), time.Millisecond*10)
	_ = s.Set(ctx, "key3", []byte("value3"), time.Second) // This one should not expire

	// Wait for cleanup to run
	time.Sleep(time.Millisecond * 100)

	// Verify expired items were removed
	s.mu.RLock()
	_, exists1 := s.items["key1"]
	_, exists2 := s.items["key2"]
	_, exists3 := s.items["key3"]
	s.mu.RUnlock()

	assert.False(t, exists1, "key1 was not cleaned up")
	assert.False(t, exists2, "key2 was not cleaned up")
	assert.True(t, exists3, "key3 was cleaned up but should not have been")

	// Clean up
	_ = s.Close()
}
