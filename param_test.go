package ngebut

import (
	"testing"
)

// TestParamKey tests the paramKey type
func TestParamKey(t *testing.T) {
	// Create a paramKey
	key := paramKey("id")

	// Check that it can be used as a context key
	if key != "id" {
		t.Errorf("paramKey(\"id\") = %q, want %q", key, "id")
	}

	// Check that it can be converted to a string
	if string(key) != "id" {
		t.Errorf("string(paramKey(\"id\")) = %q, want %q", string(key), "id")
	}
}

// TestParamMapPool tests the paramMapPool
func TestParamMapPool(t *testing.T) {
	// Get a map from the pool
	m1 := paramMapPool.Get()
	if m1 == nil {
		t.Fatal("paramMapPool.Get() returned nil")
	}

	// Check that it's a map[string]string
	paramMap, ok := m1.(map[string]string)
	if !ok {
		t.Fatalf("paramMapPool.Get() returned %T, want map[string]string", m1)
	}

	// Check that the map is empty
	if len(paramMap) != 0 {
		t.Errorf("len(paramMap) = %d, want 0", len(paramMap))
	}

	// Put the map back in the pool
	paramMapPool.Put(paramMap)

	// Get another map from the pool
	m2 := paramMapPool.Get()
	if m2 == nil {
		t.Fatal("paramMapPool.Get() returned nil on second call")
	}

	// Check that it's a map[string]string
	paramMap2, ok := m2.(map[string]string)
	if !ok {
		t.Fatalf("paramMapPool.Get() returned %T, want map[string]string", m2)
	}

	// Check that the map is empty
	if len(paramMap2) != 0 {
		t.Errorf("len(paramMap2) = %d, want 0", len(paramMap2))
	}

	// Put the map back in the pool
	paramMapPool.Put(paramMap2)
}

// TestGetParamMap tests the getParamMap function
func TestGetParamMap(t *testing.T) {
	// Get a map from the pool
	paramMap := getParamMap()
	if paramMap == nil {
		t.Fatal("getParamMap() returned nil")
	}

	// Check that the map is empty
	if len(paramMap) != 0 {
		t.Errorf("len(paramMap) = %d, want 0", len(paramMap))
	}

	// Add some values to the map
	paramMap["id"] = "123"
	paramMap["name"] = "John"

	// Check that the values were added
	if paramMap["id"] != "123" {
		t.Errorf("paramMap[\"id\"] = %q, want %q", paramMap["id"], "123")
	}
	if paramMap["name"] != "John" {
		t.Errorf("paramMap[\"name\"] = %q, want %q", paramMap["name"], "John")
	}

	// Put the map back in the pool
	releaseParamMap(paramMap)
}

// TestReleaseParamMap tests the releaseParamMap function
func TestReleaseParamMap(t *testing.T) {
	// Get a map from the pool
	paramMap := getParamMap()
	if paramMap == nil {
		t.Fatal("getParamMap() returned nil")
	}

	// Add some values to the map
	paramMap["id"] = "123"
	paramMap["name"] = "John"

	// Release the map
	releaseParamMap(paramMap)

	// Get another map from the pool (might be the same one)
	paramMap2 := getParamMap()
	if paramMap2 == nil {
		t.Fatal("getParamMap() returned nil on second call")
	}

	// Check that the map is empty
	if len(paramMap2) != 0 {
		t.Errorf("len(paramMap2) = %d, want 0", len(paramMap2))
	}

	// Check that the previous values are gone
	if paramMap2["id"] != "" {
		t.Errorf("paramMap2[\"id\"] = %q, want \"\"", paramMap2["id"])
	}
	if paramMap2["name"] != "" {
		t.Errorf("paramMap2[\"name\"] = %q, want \"\"", paramMap2["name"])
	}

	// Put the map back in the pool
	releaseParamMap(paramMap2)
}

// TestParamMapCapacity tests the capacity of maps from the pool
func TestParamMapCapacity(t *testing.T) {
	// Get a map from the pool
	paramMap := getParamMap()
	if paramMap == nil {
		t.Fatal("getParamMap() returned nil")
	}

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
		if paramMap[key] != "value" {
			t.Errorf("paramMap[%q] = %q, want %q", key, paramMap[key], "value")
		}
	}

	// Put the map back in the pool
	releaseParamMap(paramMap)
}
