package filecache

import (
	"os"
	"testing"
	"time"
)

func TestNewFDCache(t *testing.T) {
	cache := NewFDCache(100, 5*time.Minute)
	if cache == nil {
		t.Fatal("NewFDCache() returned nil")
	}
	if cache.descriptors == nil {
		t.Error("FDCache.descriptors is nil")
	}
	if cache.maxSize != 100 {
		t.Errorf("Expected maxSize 100, got %d", cache.maxSize)
	}
	if cache.expiration != 5*time.Minute {
		t.Errorf("Expected expiration 5m0s, got %v", cache.expiration)
	}
}

func TestFDCacheSetAndGet(t *testing.T) {
	cache := NewFDCache(100, 5*time.Minute)

	// Create a temporary file for testing
	tmpfile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Get the file info
	fileInfo, err := tmpfile.Stat()
	if err != nil {
		t.Fatal(err)
	}

	// Add the file descriptor to the cache
	cache.Set(tmpfile.Name(), tmpfile, fileInfo.ModTime(), fileInfo.Size())

	// Test getting the file descriptor from the cache
	fd, exists := cache.Get(tmpfile.Name())
	if !exists {
		t.Error("Failed to get file descriptor from cache")
	}
	if fd == nil {
		t.Fatal("Get() returned nil file descriptor")
	}
	if fd.File != tmpfile {
		t.Error("File descriptor does not match the original file")
	}
	if fd.ModTime != fileInfo.ModTime() {
		t.Errorf("Expected modTime %v, got %v", fileInfo.ModTime(), fd.ModTime)
	}
	if fd.Size != fileInfo.Size() {
		t.Errorf("Expected size %d, got %d", fileInfo.Size(), fd.Size)
	}

	// Test getting a non-existent file descriptor
	_, exists = cache.Get("nonexistent.txt")
	if exists {
		t.Error("Get() returned true for non-existent file descriptor")
	}

	// Close the file to avoid "too many open files" errors
	tmpfile.Close()
}

func TestFDCacheEviction(t *testing.T) {
	// Create a cache with a small max size
	cache := NewFDCache(2, 5*time.Minute)

	// Create temporary files for testing
	files := make([]*os.File, 3)
	for i := 0; i < 3; i++ {
		tmpfile, err := os.CreateTemp("", "test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpfile.Name())
		files[i] = tmpfile

		// Get the file info
		fileInfo, err := tmpfile.Stat()
		if err != nil {
			t.Fatal(err)
		}

		// Add the file descriptor to the cache
		cache.Set(tmpfile.Name(), tmpfile, fileInfo.ModTime(), fileInfo.Size())

		// Sleep a bit to ensure different last access times
		time.Sleep(10 * time.Millisecond)
	}

	// Check that the cache size is within limits
	if cache.Count() > cache.maxSize {
		t.Errorf("Cache size %d exceeds max size %d", cache.Count(), cache.maxSize)
	}

	// The first file should have been evicted
	_, exists := cache.Get(files[0].Name())
	if exists {
		t.Error("First file was not evicted from the cache")
	}

	// The last two files should still be in the cache
	_, exists = cache.Get(files[1].Name())
	if !exists {
		t.Error("Second file was unexpectedly evicted from the cache")
	}
	_, exists = cache.Get(files[2].Name())
	if !exists {
		t.Error("Third file was unexpectedly evicted from the cache")
	}

	// Close the files to avoid "too many open files" errors
	for _, file := range files {
		file.Close()
	}
}

func TestFDCacheRemove(t *testing.T) {
	cache := NewFDCache(100, 5*time.Minute)

	// Create a temporary file for testing
	tmpfile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Get the file info
	fileInfo, err := tmpfile.Stat()
	if err != nil {
		t.Fatal(err)
	}

	// Add the file descriptor to the cache
	cache.Set(tmpfile.Name(), tmpfile, fileInfo.ModTime(), fileInfo.Size())

	// Verify the file descriptor is in the cache
	_, exists := cache.Get(tmpfile.Name())
	if !exists {
		t.Error("Failed to get file descriptor from cache")
	}

	// Remove the file descriptor
	cache.Remove(tmpfile.Name())

	// Verify the file descriptor is no longer in the cache
	_, exists = cache.Get(tmpfile.Name())
	if exists {
		t.Error("File descriptor still in cache after Remove()")
	}

	// Create a new file since the original was closed by Remove()
	tmpfile, err = os.CreateTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()
}

func TestFDCacheClear(t *testing.T) {
	cache := NewFDCache(100, 5*time.Minute)

	// Create temporary files for testing
	files := make([]*os.File, 5)
	for i := 0; i < 5; i++ {
		tmpfile, err := os.CreateTemp("", "test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpfile.Name())
		files[i] = tmpfile

		// Get the file info
		fileInfo, err := tmpfile.Stat()
		if err != nil {
			t.Fatal(err)
		}

		// Add the file descriptor to the cache
		cache.Set(tmpfile.Name(), tmpfile, fileInfo.ModTime(), fileInfo.Size())
	}

	// Verify the cache has file descriptors
	if cache.Count() != 5 {
		t.Errorf("Expected 5 file descriptors in cache, got %d", cache.Count())
	}

	// Clear the cache
	cache.Clear()

	// Verify the cache is empty
	if cache.Count() != 0 {
		t.Errorf("Expected 0 file descriptors in cache after Clear(), got %d", cache.Count())
	}

	// Create new files since the originals were closed by Clear()
	for i := 0; i < 5; i++ {
		tmpfile, err := os.CreateTemp("", "test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpfile.Name())
		tmpfile.Close()
	}
}

func TestFDCacheIsModified(t *testing.T) {
	cache := NewFDCache(100, 5*time.Minute)

	// Create a temporary file for testing
	tmpfile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Write some data to the file
	if _, err := tmpfile.Write([]byte("test data")); err != nil {
		t.Fatal(err)
	}

	// Get the file info
	fileInfo, err := tmpfile.Stat()
	if err != nil {
		t.Fatal(err)
	}

	// Add the file descriptor to the cache with an older modification time
	oldModTime := fileInfo.ModTime().Add(-1 * time.Hour)
	cache.Set(tmpfile.Name(), tmpfile, oldModTime, fileInfo.Size())

	// Check if the file is modified (it should be)
	if !cache.IsModified(tmpfile.Name(), fileInfo) {
		t.Error("IsModified() returned false for a modified file")
	}

	// Update the cache with the current modification time
	cache.Set(tmpfile.Name(), tmpfile, fileInfo.ModTime(), fileInfo.Size())

	// Check if the file is modified (it should not be)
	if cache.IsModified(tmpfile.Name(), fileInfo) {
		t.Error("IsModified() returned true for an unmodified file")
	}

	// Close the file to avoid "too many open files" errors
	tmpfile.Close()
}

func TestFDCacheCount(t *testing.T) {
	cache := NewFDCache(100, 5*time.Minute)

	// Create temporary files for testing
	files := make([]*os.File, 3)
	for i := 0; i < 3; i++ {
		tmpfile, err := os.CreateTemp("", "test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpfile.Name())
		files[i] = tmpfile

		// Get the file info
		fileInfo, err := tmpfile.Stat()
		if err != nil {
			t.Fatal(err)
		}

		// Add the file descriptor to the cache
		cache.Set(tmpfile.Name(), tmpfile, fileInfo.ModTime(), fileInfo.Size())
	}

	// Check the cache count
	if cache.Count() != 3 {
		t.Errorf("Expected cache count 3, got %d", cache.Count())
	}

	// Close the files to avoid "too many open files" errors
	for _, file := range files {
		file.Close()
	}
}
