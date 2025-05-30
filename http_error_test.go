package ngebut

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewHttpError tests the NewHttpError function
func TestNewHttpError(t *testing.T) {
	// Create a new HttpError
	code := StatusBadRequest
	message := "Bad request"
	err := NewHttpError(code, message)

	// Check that the fields are set correctly using testify
	assert.Equal(t, code, err.Code, "err.Code should match expected value")
	assert.Equal(t, message, err.Message, "err.Message should match expected value")
	assert.Nil(t, err.Err, "err.Err should be nil")
}

// TestNewHttpErrorWithError tests the NewHttpErrorWithError function
func TestNewHttpErrorWithError(t *testing.T) {
	// Create a new HttpError with an underlying error
	code := StatusInternalServerError
	message := "Internal server error"
	originalErr := errors.New("database connection failed")
	err := NewHttpErrorWithError(code, message, originalErr)

	// Check that the fields are set correctly using testify
	assert.Equal(t, code, err.Code, "err.Code should match expected value")
	assert.Equal(t, message, err.Message, "err.Message should match expected value")
	assert.Equal(t, originalErr, err.Err, "err.Err should match original error")
}

// TestHttpErrorError tests the Error method of HttpError
func TestHttpErrorError(t *testing.T) {
	// Test Error() with no underlying error
	err1 := NewHttpError(StatusBadRequest, "Bad request")
	expected1 := "Bad request"
	assert.Equal(t, expected1, err1.Error(), "err1.Error() should return expected message")

	// Test Error() with an underlying error
	originalErr := errors.New("database connection failed")
	err2 := NewHttpErrorWithError(StatusInternalServerError, "Internal server error", originalErr)
	expected2 := "Internal server error: database connection failed"
	assert.Equal(t, expected2, err2.Error(), "err2.Error() should return expected message with original error")
}

// TestHttpErrorUnwrap tests the Unwrap method of HttpError
func TestHttpErrorUnwrap(t *testing.T) {
	// Test Unwrap() with no underlying error
	err1 := NewHttpError(StatusBadRequest, "Bad request")
	assert.Nil(t, err1.Unwrap(), "err1.Unwrap() should return nil")

	// Test Unwrap() with an underlying error
	originalErr := errors.New("database connection failed")
	err2 := NewHttpErrorWithError(StatusInternalServerError, "Internal server error", originalErr)
	assert.Equal(t, originalErr, err2.Unwrap(), "err2.Unwrap() should return original error")
}

// TestHttpErrorWithStandardErrors tests that HttpError works with standard error handling
func TestHttpErrorWithStandardErrors(t *testing.T) {
	// Create a new HttpError with an underlying error
	originalErr := errors.New("database connection failed")
	err := NewHttpErrorWithError(StatusInternalServerError, "Internal server error", originalErr)

	// Test errors.Is
	assert.True(t, errors.Is(err, originalErr), "errors.Is(err, originalErr) should be true")

	// Test errors.As
	var httpErr *HttpError
	assert.True(t, errors.As(err, &httpErr), "errors.As(err, &httpErr) should be true")
	assert.Equal(t, err, httpErr, "httpErr should equal err")
}
