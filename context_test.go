package ngebut

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBindForm_MultipartFormData(t *testing.T) {
	// Create a multipart form
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// Add a field
	err := w.WriteField("name", "test-value")
	if err != nil {
		t.Fatalf("Failed to write field: %v", err)
	}

	// Close the writer
	err = w.Close()
	if err != nil {
		t.Fatalf("Failed to close writer: %v", err)
	}

	// Create a test request with the multipart form
	req, err := http.NewRequest("POST", "/test", &b)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

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
	if err != nil {
		t.Errorf("BindForm returned an error: %v", err)
	}

	// Check the bound data
	if data.Name != "test-value" {
		t.Errorf("Expected Name to be 'test-value', got '%s'", data.Name)
	}
}

func TestBindForm_URLEncodedForm(t *testing.T) {
	// Create a URL-encoded form
	formData := "name=test-value"

	// Create a test request with the URL-encoded form
	req, err := http.NewRequest("POST", "/test", strings.NewReader(formData))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
	if err != nil {
		t.Errorf("BindForm returned an error: %v", err)
	}

	// Check the bound data
	if data.Name != "test-value" {
		t.Errorf("Expected Name to be 'test-value', got '%s'", data.Name)
	}
}

func TestBindForm_TextPlainForm(t *testing.T) {
	// Create a plain text form
	formData := "name=test-value"

	// Create a test request with the plain text form
	req, err := http.NewRequest("POST", "/test", strings.NewReader(formData))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "text/plain")

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
	if err != nil {
		t.Errorf("BindForm returned an error: %v", err)
	}

	// Check the bound data
	if data.Name != "test-value" {
		t.Errorf("Expected Name to be 'test-value', got '%s'", data.Name)
	}
}
