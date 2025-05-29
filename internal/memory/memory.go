package memory

import (
	"context"
	"sync"
	"time"

	"github.com/ryanbekhen/ngebut"
)

// item represents a stored item with its value and expiration time.
type item struct {
	value    []byte
	expireAt time.Time
}

// Storage implements the ngebut.Storage interface using an in-memory map.
// It provides thread-safe operations for storing and retrieving data with optional TTL support.
// The storage includes an automatic cleanup mechanism to remove expired items.
type Storage struct {
	// items stores the key-value pairs with their expiration times
	items map[string]item

	// mu provides thread-safety for concurrent access to the items map
	mu sync.RWMutex

	// cleanupTicker triggers periodic cleanup of expired items
	cleanupTicker *time.Ticker

	// stopCleanup signals the cleanup goroutine to stop
	stopCleanup chan struct{}
}

// New creates a new memory storage instance.
// The cleanupInterval parameter specifies how often to check for and remove expired items.
// If cleanupInterval is zero or negative, automatic cleanup is disabled.
func New(cleanupInterval time.Duration) *Storage {
	s := &Storage{
		items: make(map[string]item),
	}

	// Start cleanup goroutine if interval is positive
	if cleanupInterval > 0 {
		s.cleanupTicker = time.NewTicker(cleanupInterval)
		s.stopCleanup = make(chan struct{})

		go func() {
			for {
				select {
				case <-s.cleanupTicker.C:
					s.cleanup()
				case <-s.stopCleanup:
					s.cleanupTicker.Stop()
					return
				}
			}
		}()
	}

	return s
}

// valuePool is a pool of byte slices for reuse to reduce memory allocations.
// It pre-allocates byte slices with a capacity of 512 bytes, which is a reasonable
// size for most values stored in the cache.
var valuePool = sync.Pool{
	New: func() interface{} {
		// Pre-allocate a reasonable size buffer
		return make([]byte, 0, 512)
	},
}

// Get retrieves a value for the given key.
// It returns the value as a byte slice if the key exists and has not expired.
// If the key doesn't exist or has expired, it returns ngebut.ErrNotFound.
// The context parameter is currently unused but included for interface compatibility.
func (s *Storage) Get(_ context.Context, key string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, exists := s.items[key]
	if !exists {
		return nil, ngebut.ErrNotFound
	}

	// Check if the item has expired
	if !item.expireAt.IsZero() && time.Now().After(item.expireAt) {
		// Item has expired, remove it
		// We need to unlock and relock with a write lock
		s.mu.RUnlock()
		s.mu.Lock()
		delete(s.items, key)
		s.mu.Unlock()
		s.mu.RLock()
		return nil, ngebut.ErrNotFound
	}

	// Return a copy of the value to prevent modification of the stored value
	// Use a pooled buffer to reduce allocations
	buf := valuePool.Get().([]byte)
	// Ensure the buffer has enough capacity
	if cap(buf) < len(item.value) {
		// If not, create a new one with sufficient capacity
		buf = make([]byte, 0, len(item.value))
	}
	// Reset the buffer and copy the value
	buf = buf[:0]
	buf = append(buf, item.value...)

	return buf, nil
}

// Set stores a value for the given key.
// It takes a key, a value as byte slice, and an optional TTL (time-to-live) duration.
// If ttl is positive, the item will expire after the specified duration.
// If ttl is zero or negative, the item will never expire.
// The context parameter is currently unused but included for interface compatibility.
// It returns nil on success or an error if the operation fails.
func (s *Storage) Set(_ context.Context, key string, value []byte, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a copy of the value to prevent modification of the stored value
	// Use a pooled buffer to reduce allocations
	buf := valuePool.Get().([]byte)
	// Ensure the buffer has enough capacity
	if cap(buf) < len(value) {
		// If not, create a new one with sufficient capacity
		buf = make([]byte, 0, len(value))
	}
	// Reset the buffer and copy the value
	buf = buf[:0]
	buf = append(buf, value...)

	var expireAt time.Time
	if ttl > 0 {
		expireAt = time.Now().Add(ttl)
	}

	s.items[key] = item{
		value:    buf,
		expireAt: expireAt,
	}

	return nil
}

// Delete removes a key from the storage.
// It deletes the key and its associated value regardless of whether the item has expired.
// The context parameter is currently unused but included for interface compatibility.
// It always returns nil, even if the key doesn't exist.
func (s *Storage) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.items, key)
	return nil
}

// Clear removes all keys from the storage.
// It deletes all items from the storage, effectively resetting it to an empty state.
// The context parameter is currently unused but included for interface compatibility.
// It always returns nil.
func (s *Storage) Clear(_ context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items = make(map[string]item)
	return nil
}

// Has checks if a key exists in the storage.
// It returns true if the key exists and has not expired, false otherwise.
// The context parameter is currently unused but included for interface compatibility.
// It always returns nil as the error value unless an internal error occurs.
func (s *Storage) Has(_ context.Context, key string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, exists := s.items[key]
	if !exists {
		return false, nil
	}

	// Check if the item has expired
	if !item.expireAt.IsZero() && time.Now().After(item.expireAt) {
		// Item has expired, remove it
		// We need to unlock and relock with a write lock
		s.mu.RUnlock()
		s.mu.Lock()
		delete(s.items, key)
		s.mu.Unlock()
		s.mu.RLock()
		return false, nil
	}

	return true, nil
}

// Close stops the cleanup goroutine if it's running.
// This method should be called when the storage is no longer needed to prevent resource leaks.
// It always returns nil.
func (s *Storage) Close() error {
	if s.cleanupTicker != nil {
		s.stopCleanup <- struct{}{}
	}
	return nil
}

// cleanup removes expired items from the storage.
// This method is called periodically by the cleanup goroutine if a cleanup interval was specified.
// It acquires a write lock on the storage to safely remove expired items.
func (s *Storage) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for key, item := range s.items {
		if !item.expireAt.IsZero() && now.After(item.expireAt) {
			delete(s.items, key)
		}
	}
}
