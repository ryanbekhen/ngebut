package filecache

import (
	"os"
	"strconv"
	"testing"
	"time"
)

func TestNewCache(t *testing.T) {
	cache := NewCache(100*1024*1024, 1000)
	if cache == nil {
		t.Fatal("NewCache() returned nil")
	}
	if cache.files == nil {
		t.Error("Cache.files is nil")
	}
	if cache.maxSize != 100*1024*1024 {
		t.Errorf("Expected maxSize 100MB, got %d bytes", cache.maxSize)
	}
	if cache.maxItems != 1000 {
		t.Errorf("Expected maxItems 1000, got %d", cache.maxItems)
	}
	if cache.currentSize != 0 {
		t.Errorf("Expected currentSize 0, got %d", cache.currentSize)
	}
}

func TestCacheSetAndGet(t *testing.T) {
	cache := NewCache(100*1024*1024, 1000)

	// Test setting a file in the cache
	data := []byte("test data")
	modTime := time.Now()
	size := int64(len(data))
	contentType := "text/plain"

	cache.Set("test.txt", data, modTime, size, contentType)

	// Test getting the file from the cache
	file, exists := cache.Get("test.txt")
	if !exists {
		t.Error("Failed to get file from cache")
	}
	if file == nil {
		t.Fatal("Get() returned nil file")
	}
	if string(file.Data) != string(data) {
		t.Errorf("Expected data %q, got %q", string(data), string(file.Data))
	}
	if file.ModTime != modTime {
		t.Errorf("Expected modTime %v, got %v", modTime, file.ModTime)
	}
	if file.Size != size {
		t.Errorf("Expected size %d, got %d", size, file.Size)
	}
	if file.ContentType != contentType {
		t.Errorf("Expected contentType %s, got %s", contentType, file.ContentType)
	}

	// Test getting a non-existent file
	_, exists = cache.Get("nonexistent.txt")
	if exists {
		t.Error("Get() returned true for non-existent file")
	}
}

func TestCacheEviction(t *testing.T) {
	// Create a cache with a small max size
	cache := NewCache(100, 10)

	// Add a file that's larger than the max size
	data := make([]byte, 200)
	cache.Set("large.txt", data, time.Now(), int64(len(data)), "text/plain")

	// The file should not be in the cache because it's too large
	_, exists := cache.Get("large.txt")
	if exists {
		t.Error("File larger than max size was added to cache")
	}

	// Add several small files to trigger eviction
	for i := 0; i < 20; i++ {
		data := []byte("small file")
		cache.Set(strconv.Itoa(i), data, time.Now(), int64(len(data)), "text/plain")

		// Sleep a bit to ensure different last accessed times
		time.Sleep(10 * time.Millisecond)
	}

	// Check that the cache size is within limits
	if cache.currentSize > cache.maxSize {
		t.Errorf("Cache size %d exceeds max size %d", cache.currentSize, cache.maxSize)
	}
	if cache.Count() > cache.maxItems {
		t.Errorf("Cache item count %d exceeds max items %d", cache.Count(), cache.maxItems)
	}
}

func TestCacheRemove(t *testing.T) {
	cache := NewCache(100*1024*1024, 1000)

	// Add a file to the cache
	data := []byte("test data")
	cache.Set("test.txt", data, time.Now(), int64(len(data)), "text/plain")

	// Verify the file is in the cache
	_, exists := cache.Get("test.txt")
	if !exists {
		t.Error("Failed to get file from cache")
	}

	// Remove the file
	cache.Remove("test.txt")

	// Verify the file is no longer in the cache
	_, exists = cache.Get("test.txt")
	if exists {
		t.Error("File still in cache after Remove()")
	}

	// Verify the cache size was updated
	if cache.currentSize != 0 {
		t.Errorf("Expected currentSize 0 after Remove(), got %d", cache.currentSize)
	}
}

func TestCacheClear(t *testing.T) {
	cache := NewCache(100*1024*1024, 1000)

	// Add some files to the cache
	for i := 0; i < 10; i++ {
		data := []byte("test data")
		cache.Set(strconv.Itoa(i), data, time.Now(), int64(len(data)), "text/plain")
	}

	// Verify the cache has files
	if cache.Count() != 10 {
		t.Errorf("Expected 10 files in cache, got %d", cache.Count())
	}

	// Clear the cache
	cache.Clear()

	// Verify the cache is empty
	if cache.Count() != 0 {
		t.Errorf("Expected 0 files in cache after Clear(), got %d", cache.Count())
	}
	if cache.currentSize != 0 {
		t.Errorf("Expected currentSize 0 after Clear(), got %d", cache.currentSize)
	}
}

func TestCacheIsModified(t *testing.T) {
	cache := NewCache(100*1024*1024, 1000)

	// Create a temporary file for testing
	tmpfile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	// Write some data to the file
	if _, err := tmpfile.Write([]byte("test data")); err != nil {
		t.Fatal(err)
	}

	// Get the file info
	fileInfo, err := tmpfile.Stat()
	if err != nil {
		t.Fatal(err)
	}

	// Add the file to the cache with an older modification time
	oldModTime := fileInfo.ModTime().Add(-1 * time.Hour)
	cache.Set(tmpfile.Name(), []byte("test data"), oldModTime, fileInfo.Size(), "text/plain")

	// Check if the file is modified (it should be)
	if !cache.IsModified(tmpfile.Name(), fileInfo) {
		t.Error("IsModified() returned false for a modified file")
	}

	// Update the cache with the current modification time
	cache.Set(tmpfile.Name(), []byte("test data"), fileInfo.ModTime(), fileInfo.Size(), "text/plain")

	// Check if the file is modified (it should not be)
	if cache.IsModified(tmpfile.Name(), fileInfo) {
		t.Error("IsModified() returned true for an unmodified file")
	}
}

func TestCacheSize(t *testing.T) {
	cache := NewCache(100*1024*1024, 1000)

	// Add some files to the cache
	data1 := []byte("test data 1")
	data2 := []byte("test data 2")

	cache.Set("file1.txt", data1, time.Now(), int64(len(data1)), "text/plain")
	cache.Set("file2.txt", data2, time.Now(), int64(len(data2)), "text/plain")

	// Check the cache size
	expectedSize := int64(len(data1) + len(data2))
	if cache.Size() != expectedSize {
		t.Errorf("Expected cache size %d, got %d", expectedSize, cache.Size())
	}
}

func TestCacheCount(t *testing.T) {
	cache := NewCache(100*1024*1024, 1000)

	// Add some files to the cache
	for i := 0; i < 5; i++ {
		data := []byte("test data")
		cache.Set(strconv.Itoa(i), data, time.Now(), int64(len(data)), "text/plain")
	}

	// Check the cache count
	if cache.Count() != 5 {
		t.Errorf("Expected cache count 5, got %d", cache.Count())
	}
}
