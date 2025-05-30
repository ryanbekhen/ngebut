package ngebut

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestBindJSON_Success tests successful JSON binding
func TestBindJSON_Success(t *testing.T) {
	// Create JSON data
	jsonData := `{"name":"test-value","age":30,"active":true}`

	// Create a test request with the JSON data
	req, err := http.NewRequest("POST", "/test", strings.NewReader(jsonData))
	require.NoError(t, err, "Failed to create request")
	assert.NotNil(t, req, "Request should not be nil")

	req.Header.Set("Content-Type", "application/json")

	// Create a test response recorder
	res := httptest.NewRecorder()

	// Create a context
	ctx := GetContext(res, req)

	// Define a struct to bind to
	type TestStruct struct {
		Name   string `json:"name"`
		Age    int    `json:"age"`
		Active bool   `json:"active"`
	}

	// Bind the JSON data
	var data TestStruct
	err = ctx.BindJSON(&data)
	assert.NoError(t, err, "BindJSON should not return an error")
	assert.Equal(t, "test-value", data.Name, "Name should match the expected value")
	assert.Equal(t, 30, data.Age, "Age should match the expected value")
	assert.True(t, data.Active, "Active should be true")
}

// TestBindJSON_NilBody tests BindJSON with a nil request body
func TestBindJSON_NilBody(t *testing.T) {
	// Create a test request with a nil body
	req, err := http.NewRequest("POST", "/test", nil)
	require.NoError(t, err, "Failed to create request")

	// Create a test response recorder
	res := httptest.NewRecorder()

	// Create a context
	ctx := GetContext(res, req)

	// Define a struct to bind to
	type TestStruct struct {
		Name string `json:"name"`
	}

	// Bind the JSON data
	var data TestStruct
	err = ctx.BindJSON(&data)

	// Check for errors
	assert.Error(t, err, "Expected BindJSON to return an error for nil body")
	assert.Equal(t, "request body is nil", err.Error(), "Unexpected error message")
}

// TestBindJSON_InvalidJSON tests BindJSON with invalid JSON data
func TestBindJSON_InvalidJSON(t *testing.T) {
	// Create invalid JSON data
	jsonData := `{"name":"test-value",invalid}`

	// Create a test request with the invalid JSON data
	req, err := http.NewRequest("POST", "/test", strings.NewReader(jsonData))
	require.NoError(t, err, "Failed to create request")
	req.Header.Set("Content-Type", "application/json")

	// Create a test response recorder
	res := httptest.NewRecorder()

	// Create a context
	ctx := GetContext(res, req)

	// Define a struct to bind to
	type TestStruct struct {
		Name string `json:"name"`
	}

	// Bind the JSON data
	var data TestStruct
	err = ctx.BindJSON(&data)

	// Check for errors
	assert.Error(t, err, "Expected BindJSON to return an error for invalid JSON")
	assert.Contains(t, err.Error(), "failed to unmarshal JSON", "Unexpected error message")
}

// TestBindForm_NilBody tests BindForm with a nil request body
func TestBindForm_NilBody(t *testing.T) {
	// Create a test request with a nil body
	req, err := http.NewRequest("POST", "/test", nil)
	require.NoError(t, err, "Failed to create request")

	// Create a test response recorder
	res := httptest.NewRecorder()

	// Create a context
	ctx := GetContext(res, req)

	// Define a struct to bind to
	type TestStruct struct {
		Name string `form:"name"`
	}

	// Bind the form data
	var data TestStruct
	err = ctx.BindForm(&data)

	// Check for errors
	assert.Error(t, err, "Expected BindForm to return an error for nil body")
	assert.Equal(t, "request body is nil", err.Error(), "Unexpected error message")
}

// TestBindForm_NotPointerToStruct tests BindForm with a non-pointer to struct
func TestBindForm_NotPointerToStruct(t *testing.T) {
	// Create form data
	formData := "name=test-value"

	// Create a test request with the form data
	req, err := http.NewRequest("POST", "/test", strings.NewReader(formData))
	require.NoError(t, err, "Failed to create request")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Create a test response recorder
	res := httptest.NewRecorder()

	// Create a context
	ctx := GetContext(res, req)

	// Try to bind to a non-pointer
	var data string
	err = ctx.BindForm(data)

	// Check for errors
	assert.Error(t, err, "Expected BindForm to return an error for non-pointer")
	assert.Equal(t, "obj must be a pointer to a struct", err.Error(), "Unexpected error message")

	// Try to bind to a pointer to non-struct
	dataPtr := "test"
	err = ctx.BindForm(&dataPtr)

	// Check for errors
	assert.Error(t, err, "Expected BindForm to return an error for pointer to non-struct")
	assert.Equal(t, "obj must be a pointer to a struct", err.Error(), "Unexpected error message")
}

// TestBindForm_UnsupportedContentType tests BindForm with an unsupported content type
func TestBindForm_UnsupportedContentType(t *testing.T) {
	// Create form data
	formData := "name=test-value"

	// Create a test request with the form data
	req, err := http.NewRequest("POST", "/test", strings.NewReader(formData))
	require.NoError(t, err, "Failed to create request")
	req.Header.Set("Content-Type", "application/json")

	// Create a test response recorder
	res := httptest.NewRecorder()

	// Create a context
	ctx := GetContext(res, req)

	// Define a struct to bind to
	type TestStruct struct {
		Name string `form:"name"`
	}

	// Bind the form data
	var data TestStruct
	err = ctx.BindForm(&data)

	// Check for errors
	assert.Error(t, err, "Expected BindForm to return an error for unsupported content type")
	assert.Contains(t, err.Error(), "unsupported Content-Type for form binding", "Unexpected error message")
}

// TestBindForm_DifferentTypes tests BindForm with different data types
func TestBindForm_DifferentTypes(t *testing.T) {
	// Create form data with different types
	formData := "name=test-value&age=30&height=1.85&active=true&count=100"

	// Create a test request with the form data
	req, err := http.NewRequest("POST", "/test", strings.NewReader(formData))
	require.NoError(t, err, "Failed to create request")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Create a test response recorder
	res := httptest.NewRecorder()

	// Create a context
	ctx := GetContext(res, req)

	// Define a struct with different types
	type TestStruct struct {
		Name   string  `form:"name"`
		Age    int     `form:"age"`
		Height float64 `form:"height"`
		Active bool    `form:"active"`
		Count  uint    `form:"count"`
	}

	// Bind the form data
	var data TestStruct
	err = ctx.BindForm(&data)

	// Check for errors
	assert.NoError(t, err, "BindForm should not return an error")

	// Check the bound data
	assert.Equal(t, "test-value", data.Name, "Name should match the expected value")
	assert.Equal(t, 30, data.Age, "Age should match the expected value")
	assert.Equal(t, 1.85, data.Height, "Height should match the expected value")
	assert.True(t, data.Active, "Active should be true")
	assert.Equal(t, uint(100), data.Count, "Count should match the expected value")
}

// TestBindForm_InvalidTypes tests BindForm with invalid type conversions
func TestBindForm_InvalidTypes(t *testing.T) {
	testCases := []struct {
		name        string
		formData    string
		fieldName   string
		fieldType   string
		expectedErr string
	}{
		{
			name:        "Invalid int",
			formData:    "age=not-a-number",
			fieldName:   "age",
			fieldType:   "int",
			expectedErr: "failed to parse age as int",
		},
		{
			name:        "Invalid uint",
			formData:    "count=-10",
			fieldName:   "count",
			fieldType:   "uint",
			expectedErr: "failed to parse count as uint",
		},
		{
			name:        "Invalid float",
			formData:    "height=not-a-float",
			fieldName:   "height",
			fieldType:   "float",
			expectedErr: "failed to parse height as float",
		},
		{
			name:        "Invalid bool",
			formData:    "active=not-a-bool",
			fieldName:   "active",
			fieldType:   "bool",
			expectedErr: "failed to parse active as bool",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test request with the form data
			req, err := http.NewRequest("POST", "/test", strings.NewReader(tc.formData))
			require.NoError(t, err, "Failed to create request")
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			// Create a test response recorder
			res := httptest.NewRecorder()

			// Create a context
			ctx := GetContext(res, req)

			// Define a struct based on the field type
			var data interface{}
			switch tc.fieldType {
			case "int":
				data = &struct {
					Age int `form:"age"`
				}{}
			case "uint":
				data = &struct {
					Count uint `form:"count"`
				}{}
			case "float":
				data = &struct {
					Height float64 `form:"height"`
				}{}
			case "bool":
				data = &struct {
					Active bool `form:"active"`
				}{}
			}

			// Bind the form data
			err = ctx.BindForm(data)

			// Check for errors
			assert.Error(t, err, "Expected BindForm to return an error for invalid %s", tc.fieldType)
			assert.Contains(t, err.Error(), tc.expectedErr, "Unexpected error message")
		})
	}
}
