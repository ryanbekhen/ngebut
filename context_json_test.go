package ngebut

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

// Simple test struct for JSON serialization
type testJSONStruct struct {
	ID       int                    `json:"id"`
	Name     string                 `json:"name"`
	Email    string                 `json:"email"`
	Active   bool                   `json:"active"`
	Tags     []string               `json:"tags"`
	Score    float64                `json:"score"`
	Metadata map[string]interface{} `json:"metadata"`
}

// TestJSON tests the JSON method with different types of data
func TestJSON(t *testing.T) {
	// Create a test context
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	ctx := GetContext(w, req)
	defer ReleaseContext(ctx)

	// Test with a string
	t.Run("String", func(t *testing.T) {
		w.Body.Reset()
		ctx.JSON("hello world")
		expected := `"hello world"`
		if w.Body.String() != expected {
			t.Errorf("Expected %s, got %s", expected, w.Body.String())
		}
	})

	// Test with an integer
	t.Run("Integer", func(t *testing.T) {
		w.Body.Reset()
		ctx.JSON(123)
		expected := `123`
		if w.Body.String() != expected {
			t.Errorf("Expected %s, got %s", expected, w.Body.String())
		}
	})

	// Test with a boolean
	t.Run("Boolean", func(t *testing.T) {
		w.Body.Reset()
		ctx.JSON(true)
		expected := `true`
		if w.Body.String() != expected {
			t.Errorf("Expected %s, got %s", expected, w.Body.String())
		}
	})

	// Test with a simple struct
	t.Run("SimpleStruct", func(t *testing.T) {
		w.Body.Reset()
		simpleStruct := testJSONStruct{
			ID:     1,
			Name:   "John Doe",
			Email:  "john@example.com",
			Active: true,
			Score:  98.6,
		}
		ctx.JSON(simpleStruct)

		// Parse the JSON to verify it's valid
		var result testJSONStruct
		err := json.Unmarshal(w.Body.Bytes(), &result)
		if err != nil {
			t.Errorf("Failed to parse JSON: %v", err)
		}

		// Verify the values
		if result.ID != 1 || result.Name != "John Doe" || result.Email != "john@example.com" || !result.Active || result.Score != 98.6 {
			t.Errorf("JSON values don't match expected values")
		}
	})

	// Test with a complex struct
	t.Run("ComplexStruct", func(t *testing.T) {
		w.Body.Reset()
		complexStruct := testJSONStruct{
			ID:     1,
			Name:   "John Doe",
			Email:  "john@example.com",
			Active: true,
			Tags:   []string{"user", "admin", "member"},
			Score:  98.6,
			Metadata: map[string]interface{}{
				"lastLogin": "2023-01-01",
				"visits":    42,
				"preferences": map[string]interface{}{
					"theme":      "dark",
					"fontSize":   12,
					"showAvatar": true,
				},
			},
		}
		ctx.JSON(complexStruct)

		// Parse the JSON to verify it's valid
		var result testJSONStruct
		err := json.Unmarshal(w.Body.Bytes(), &result)
		if err != nil {
			t.Errorf("Failed to parse JSON: %v", err)
		}

		// Verify the values
		if result.ID != 1 || result.Name != "John Doe" || result.Email != "john@example.com" || !result.Active || result.Score != 98.6 {
			t.Errorf("JSON values don't match expected values")
		}

		// Verify the tags
		if len(result.Tags) != 3 || result.Tags[0] != "user" || result.Tags[1] != "admin" || result.Tags[2] != "member" {
			t.Errorf("JSON tags don't match expected values")
		}

		// Verify the metadata exists
		if result.Metadata == nil {
			t.Errorf("JSON metadata is nil")
		}
	})
}

// BenchmarkJSON benchmarks the JSON method with different types of data
func BenchmarkJSON(b *testing.B) {
	// Create a test context
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	ctx := GetContext(w, req)
	defer ReleaseContext(ctx)

	// Simple string
	b.Run("String", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			w.Body.Reset()
			ctx.JSON("hello world")
		}
	})

	// Simple integer
	b.Run("Integer", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			w.Body.Reset()
			ctx.JSON(123)
		}
	})

	// Simple boolean
	b.Run("Boolean", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			w.Body.Reset()
			ctx.JSON(true)
		}
	})

	// Simple struct
	simpleStruct := testJSONStruct{
		ID:     1,
		Name:   "John Doe",
		Email:  "john@example.com",
		Active: true,
		Score:  98.6,
	}
	b.Run("SimpleStruct", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			w.Body.Reset()
			ctx.JSON(simpleStruct)
		}
	})

	// Complex struct with nested data
	complexStruct := testJSONStruct{
		ID:     1,
		Name:   "John Doe",
		Email:  "john@example.com",
		Active: true,
		Tags:   []string{"user", "admin", "member"},
		Score:  98.6,
		Metadata: map[string]interface{}{
			"lastLogin": "2023-01-01",
			"visits":    42,
			"preferences": map[string]interface{}{
				"theme":      "dark",
				"fontSize":   12,
				"showAvatar": true,
			},
		},
	}
	b.Run("ComplexStruct", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			w.Body.Reset()
			ctx.JSON(complexStruct)
		}
	})

	// Array of structs
	arrayOfStructs := []testJSONStruct{
		simpleStruct,
		{
			ID:     2,
			Name:   "Jane Smith",
			Email:  "jane@example.com",
			Active: false,
			Score:  87.3,
		},
	}
	b.Run("ArrayOfStructs", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			w.Body.Reset()
			ctx.JSON(arrayOfStructs)
		}
	})
}
