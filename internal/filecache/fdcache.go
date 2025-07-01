package filecache

import (
	"os"
	"sync"
	"time"
)

// FileDescriptor represents a cached file descriptor
type FileDescriptor struct {
	File       *os.File
	ModTime    time.Time
	Size       int64
	LastAccess time.Time
}

// FDCache is a cache for file descriptors
type FDCache struct {
	descriptors map[string]*FileDescriptor
	mutex       sync.RWMutex
	maxSize     int
	expiration  time.Duration
}

// NewFDCache creates a new file descriptor cache
func NewFDCache(maxSize int, expiration time.Duration) *FDCache {
	cache := &FDCache{
		descriptors: make(map[string]*FileDescriptor, maxSize),
		maxSize:     maxSize,
		expiration:  expiration,
	}

	// Start a goroutine to periodically clean up expired file descriptors
	go cache.cleanupLoop()

	return cache
}

// Get retrieves a file descriptor from the cache
func (c *FDCache) Get(path string) (*FileDescriptor, bool) {
	c.mutex.RLock()
	fd, exists := c.descriptors[path]
	c.mutex.RUnlock()

	if !exists {
		return nil, false
	}

	// Update last access time
	c.mutex.Lock()
	fd.LastAccess = time.Now()
	c.mutex.Unlock()

	return fd, true
}

// Set adds a file descriptor to the cache
func (c *FDCache) Set(path string, file *os.File, modTime time.Time, size int64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Check if we need to make room in the cache
	if len(c.descriptors) >= c.maxSize {
		c.evictLRU()
	}

	// Add the file descriptor to the cache
	c.descriptors[path] = &FileDescriptor{
		File:       file,
		ModTime:    modTime,
		Size:       size,
		LastAccess: time.Now(),
	}
}

// Remove removes a file descriptor from the cache
func (c *FDCache) Remove(path string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if fd, exists := c.descriptors[path]; exists {
		// Close the file before removing it from the cache
		fd.File.Close()
		delete(c.descriptors, path)
	}
}

// IsModified checks if a file has been modified since it was cached
func (c *FDCache) IsModified(path string, fileInfo os.FileInfo) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	fd, exists := c.descriptors[path]
	if !exists {
		return true
	}

	return fileInfo.ModTime().After(fd.ModTime)
}

// evictLRU evicts the least recently used file descriptor
func (c *FDCache) evictLRU() {
	var oldestPath string
	var oldestTime time.Time

	// Find the least recently used file descriptor
	for path, fd := range c.descriptors {
		if oldestPath == "" || fd.LastAccess.Before(oldestTime) {
			oldestPath = path
			oldestTime = fd.LastAccess
		}
	}

	// Remove the least recently used file descriptor
	if oldestPath != "" {
		if fd := c.descriptors[oldestPath]; fd != nil {
			fd.File.Close()
		}
		delete(c.descriptors, oldestPath)
	}
}

// cleanupLoop periodically cleans up expired file descriptors
func (c *FDCache) cleanupLoop() {
	ticker := time.NewTicker(c.expiration / 2)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}
}

// cleanup removes expired file descriptors
func (c *FDCache) cleanup() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	for path, fd := range c.descriptors {
		// If the file descriptor hasn't been accessed in the expiration period, remove it
		if now.Sub(fd.LastAccess) > c.expiration {
			fd.File.Close()
			delete(c.descriptors, path)
		}
	}
}

// Clear removes all file descriptors from the cache
func (c *FDCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Close all file descriptors
	for _, fd := range c.descriptors {
		fd.File.Close()
	}

	// Clear the map
	c.descriptors = make(map[string]*FileDescriptor, c.maxSize)
}

// Count returns the number of file descriptors in the cache
func (c *FDCache) Count() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.descriptors)
}

// DefaultFDCache is the default file descriptor cache
var DefaultFDCache = NewFDCache(100, 5*time.Minute)
