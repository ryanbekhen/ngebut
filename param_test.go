package ngebut

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestParamKey tests the paramKey type
func TestParamKey(t *testing.T) {
	// Create a paramKey
	key := paramKey("id")

	// Check that it can be used as a context key
	assert.Equal(t, paramKey("id"), key, "paramKey should maintain its value")

	// Check that it can be converted to a string
	assert.Equal(t, "id", string(key), "paramKey should convert to string correctly")
}

// TestParamMapPool tests the paramMapPool
func TestParamMapPool(t *testing.T) {
	// Get a map from the pool
	paramMap := paramMapPool.Get()
	assert.NotNil(t, paramMap, "paramMapPool.Get() should not return nil")

	// Check that the map is empty
	assert.Empty(t, paramMap, "map from pool should be empty")

	// Put the map back in the pool
	paramMapPool.Put(paramMap)

	// Get another map from the pool
	paramMap2 := paramMapPool.Get()
	assert.NotNil(t, paramMap2, "paramMapPool.Get() should not return nil on second call")

	// Check that the map is empty
	assert.Empty(t, paramMap2, "map from pool should be empty")

	// Put the map back in the pool
	paramMapPool.Put(paramMap2)
}

// TestGetParamMap tests the getParamMap function
func TestGetParamMap(t *testing.T) {
	// Get a map from the pool
	paramMap := getParamMap()
	assert.NotNil(t, paramMap, "getParamMap() should not return nil")

	// Check that the map is empty
	assert.Empty(t, paramMap, "map from getParamMap() should be empty")

	// Add some values to the map
	paramMap["id"] = "123"
	paramMap["name"] = "John"

	// Check that the values were added
	assert.Equal(t, "123", paramMap["id"], "paramMap should store values correctly")
	assert.Equal(t, "John", paramMap["name"], "paramMap should store values correctly")

	// Put the map back in the pool
	releaseParamMap(paramMap)
}

// TestReleaseParamMap tests the releaseParamMap function
func TestReleaseParamMap(t *testing.T) {
	// Get a map from the pool
	paramMap := getParamMap()
	assert.NotNil(t, paramMap, "getParamMap() should not return nil")

	// Add some values to the map
	paramMap["id"] = "123"
	paramMap["name"] = "John"

	// Release the map
	releaseParamMap(paramMap)

	// Get another map from the pool (might be the same one)
	paramMap2 := getParamMap()
	assert.NotNil(t, paramMap2, "getParamMap() should not return nil on second call")

	// Check that the map is empty
	assert.Empty(t, paramMap2, "map should be cleared after release")

	// Check that the previous values are gone
	assert.Equal(t, "", paramMap2["id"], "map should be cleared after release")
	assert.Equal(t, "", paramMap2["name"], "map should be cleared after release")

	// Put the map back in the pool
	releaseParamMap(paramMap2)
}

// TestParamMapCapacity tests the capacity of maps from the pool
func TestParamMapCapacity(t *testing.T) {
	// Get a map from the pool
	paramMap := getParamMap()
	assert.NotNil(t, paramMap, "getParamMap() should not return nil")

	// Check that the map has the expected capacity
	// Note: We can't directly check the capacity of a map,
	// but we can add elements and check that it doesn't need to resize
	for i := 0; i < 8; i++ {
		key := "key" + string(rune('0'+i))
		paramMap[key] = "value"
	}

	// Check that all values were added
	for i := 0; i < 8; i++ {
		key := "key" + string(rune('0'+i))
		assert.Equal(t, "value", paramMap[key], "map should store all test values")
	}

	// Put the map back in the pool
	releaseParamMap(paramMap)
}
