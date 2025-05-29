package memory

import (
	"context"
	"testing"
	"time"

	"github.com/ryanbekhen/ngebut"
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
	if s1 == nil {
		t.Fatal("New(0) returned nil")
	}
	if s1.cleanupTicker != nil {
		t.Error("New(0) should not create a cleanup ticker")
	}
	if s1.stopCleanup != nil {
		t.Error("New(0) should not create a stop channel")
	}

	// Test with cleanup enabled
	s2 := New(time.Second)
	if s2 == nil {
		t.Fatal("New(time.Second) returned nil")
	}
	if s2.cleanupTicker == nil {
		t.Error("New(time.Second) should create a cleanup ticker")
	}
	if s2.stopCleanup == nil {
		t.Error("New(time.Second) should create a stop channel")
	}

	// Clean up
	_ = s2.Close()
}

// TestSet tests the Set method
func TestSet(t *testing.T) {
	s := New(0)
	ctx := context.Background()

	// Test setting a value
	err := s.Set(ctx, "key1", []byte("value1"), 0)
	if err != nil {
		t.Errorf("Set returned an error: %v", err)
	}

	// Test setting a value with TTL
	err = s.Set(ctx, "key2", []byte("value2"), time.Minute)
	if err != nil {
		t.Errorf("Set with TTL returned an error: %v", err)
	}

	// Verify the values were set correctly
	s.mu.RLock()
	item1, exists1 := s.items["key1"]
	item2, exists2 := s.items["key2"]
	s.mu.RUnlock()

	if !exists1 {
		t.Error("key1 was not set")
	}
	if string(item1.value) != "value1" {
		t.Errorf("key1 value is %s, expected value1", string(item1.value))
	}
	if !item1.expireAt.IsZero() {
		t.Error("key1 should not have an expiration time")
	}

	if !exists2 {
		t.Error("key2 was not set")
	}
	if string(item2.value) != "value2" {
		t.Errorf("key2 value is %s, expected value2", string(item2.value))
	}
	if item2.expireAt.IsZero() {
		t.Error("key2 should have an expiration time")
	}
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
	if err != nil {
		t.Errorf("Get for key1 returned an error: %v", err)
	}
	if string(value1) != "value1" {
		t.Errorf("Get for key1 returned %s, expected value1", string(value1))
	}

	// Test getting a value with TTL
	value2, err := s.Get(ctx, "key2")
	if err != nil {
		t.Errorf("Get for key2 returned an error: %v", err)
	}
	if string(value2) != "value2" {
		t.Errorf("Get for key2 returned %s, expected value2", string(value2))
	}

	// Test getting a non-existent key
	_, err = s.Get(ctx, "nonexistent")
	if err != ngebut.ErrNotFound {
		t.Errorf("Get for nonexistent key returned %v, expected ErrNotFound", err)
	}

	// Test getting an expired key
	_, err = s.Get(ctx, "expired")
	if err != ngebut.ErrNotFound {
		t.Errorf("Get for expired key returned %v, expected ErrNotFound", err)
	}
}

// TestDelete tests the Delete method
func TestDelete(t *testing.T) {
	s := New(0)
	ctx := context.Background()

	// Set up test data
	_ = s.Set(ctx, "key1", []byte("value1"), 0)

	// Test deleting an existing key
	err := s.Delete(ctx, "key1")
	if err != nil {
		t.Errorf("Delete returned an error: %v", err)
	}

	// Verify the key was deleted
	s.mu.RLock()
	_, exists := s.items["key1"]
	s.mu.RUnlock()

	if exists {
		t.Error("key1 was not deleted")
	}

	// Test deleting a non-existent key (should not error)
	err = s.Delete(ctx, "nonexistent")
	if err != nil {
		t.Errorf("Delete for nonexistent key returned an error: %v", err)
	}
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
	if err != nil {
		t.Errorf("Clear returned an error: %v", err)
	}

	// Verify all keys were cleared
	s.mu.RLock()
	count := len(s.items)
	s.mu.RUnlock()

	if count != 0 {
		t.Errorf("Clear did not remove all items, %d items remain", count)
	}
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
	if err != nil {
		t.Errorf("Has for key1 returned an error: %v", err)
	}
	if !exists {
		t.Error("Has for key1 returned false, expected true")
	}

	// Test checking a key with TTL
	exists, err = s.Has(ctx, "key2")
	if err != nil {
		t.Errorf("Has for key2 returned an error: %v", err)
	}
	if !exists {
		t.Error("Has for key2 returned false, expected true")
	}

	// Test checking a non-existent key
	exists, err = s.Has(ctx, "nonexistent")
	if err != nil {
		t.Errorf("Has for nonexistent key returned an error: %v", err)
	}
	if exists {
		t.Error("Has for nonexistent key returned true, expected false")
	}

	// Test checking an expired key
	exists, err = s.Has(ctx, "expired")
	if err != nil {
		t.Errorf("Has for expired key returned an error: %v", err)
	}
	if exists {
		t.Error("Has for expired key returned true, expected false")
	}
}

// TestClose tests the Close method
func TestClose(t *testing.T) {
	// Create a storage with cleanup enabled
	s := New(time.Second)

	// Test closing the storage
	err := s.Close()
	if err != nil {
		t.Errorf("Close returned an error: %v", err)
	}

	// Verify the cleanup goroutine was stopped
	// This is a bit tricky to test directly, but we can check that the ticker was stopped
	// by trying to send to the stopCleanup channel, which should not block if Close worked correctly
	select {
	case s.stopCleanup <- struct{}{}:
		t.Error("stopCleanup channel is still open after Close")
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

	if exists1 {
		t.Error("key1 was not cleaned up")
	}
	if exists2 {
		t.Error("key2 was not cleaned up")
	}
	if !exists3 {
		t.Error("key3 was cleaned up but should not have been")
	}

	// Clean up
	_ = s.Close()
}
