package filecache

import (
	"os"
	"sync"
	"time"
)

// CachedFile represents a cached file in memory
type CachedFile struct {
	Data         []byte
	ModTime      time.Time
	Size         int64
	ContentType  string
	LastAccessed time.Time
}

// Cache is an in-memory cache for static files
type Cache struct {
	files       map[string]*CachedFile
	mutex       sync.RWMutex
	maxSize     int64
	currentSize int64
	maxItems    int
}

// NewCache creates a new file cache with the specified maximum size and items
func NewCache(maxSize int64, maxItems int) *Cache {
	return &Cache{
		files:    make(map[string]*CachedFile),
		maxSize:  maxSize,
		maxItems: maxItems,
	}
}

// Get retrieves a file from the cache
func (c *Cache) Get(path string) (*CachedFile, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	file, exists := c.files[path]
	if exists {
		// Update last accessed time
		file.LastAccessed = time.Now()
	}
	return file, exists
}

// Set adds a file to the cache
func (c *Cache) Set(path string, data []byte, modTime time.Time, size int64, contentType string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// If the file is larger than the max size, don't add it to the cache
	if size > c.maxSize {
		return
	}

	// Check if we need to make room in the cache
	if c.currentSize+size > c.maxSize || len(c.files) >= c.maxItems {
		c.evict(size)
	}

	// Always make a copy of the data to avoid issues with buffer reuse
	// This ensures that the data in the cache is not modified when the buffer is reused
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)

	// Add the file to the cache
	c.files[path] = &CachedFile{
		Data:         dataCopy,
		ModTime:      modTime,
		Size:         size,
		ContentType:  contentType,
		LastAccessed: time.Now(),
	}

	c.currentSize += size
}

// evict removes files from the cache to make room for a new file
func (c *Cache) evict(neededSize int64) {
	// If the cache is empty, nothing to evict
	if len(c.files) == 0 {
		return
	}

	// If the needed size is larger than the max size, we can't cache it
	if neededSize > c.maxSize {
		return
	}

	// Find and remove the oldest files directly without sorting
	// This is more efficient for large caches
	for c.currentSize+neededSize > c.maxSize || len(c.files) >= c.maxItems {
		var oldestPath string
		var oldestTime time.Time
		var oldestSize int64

		// Find the oldest file with a single pass
		for path, file := range c.files {
			if oldestPath == "" || file.LastAccessed.Before(oldestTime) {
				oldestPath = path
				oldestTime = file.LastAccessed
				oldestSize = file.Size
			}
		}

		// If we found an oldest file, remove it
		if oldestPath != "" {
			delete(c.files, oldestPath)
			c.currentSize -= oldestSize
		} else {
			// No files left to evict
			break
		}
	}
}

// Clear removes all files from the cache
func (c *Cache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.files = make(map[string]*CachedFile)
	c.currentSize = 0
}

// Remove removes a file from the cache
func (c *Cache) Remove(path string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if file, exists := c.files[path]; exists {
		c.currentSize -= file.Size
		delete(c.files, path)
	}
}

// IsModified checks if a file has been modified since it was cached
func (c *Cache) IsModified(path string, fileInfo os.FileInfo) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	file, exists := c.files[path]
	if !exists {
		return true
	}

	return fileInfo.ModTime().After(file.ModTime)
}

// Size returns the current size of the cache in bytes
func (c *Cache) Size() int64 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.currentSize
}

// Count returns the number of files in the cache
func (c *Cache) Count() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.files)
}

// DefaultCache is the default file cache instance
var DefaultCache = NewCache(100*1024*1024, 1000) // 100MB, 1000 files
