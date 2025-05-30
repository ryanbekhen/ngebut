package ngebut

import (
	"errors"
	"testing"
)

// TestNewHttpError tests the NewHttpError function
func TestNewHttpError(t *testing.T) {
	// Create a new HttpError
	code := StatusBadRequest
	message := "Bad request"
	err := NewHttpError(code, message)

	// Check that the fields are set correctly
	if err.Code != code {
		t.Errorf("err.Code = %d, want %d", err.Code, code)
	}
	if err.Message != message {
		t.Errorf("err.Message = %q, want %q", err.Message, message)
	}
	if err.Err != nil {
		t.Errorf("err.Err = %v, want nil", err.Err)
	}
}

// TestNewHttpErrorWithError tests the NewHttpErrorWithError function
func TestNewHttpErrorWithError(t *testing.T) {
	// Create a new HttpError with an underlying error
	code := StatusInternalServerError
	message := "Internal server error"
	originalErr := errors.New("database connection failed")
	err := NewHttpErrorWithError(code, message, originalErr)

	// Check that the fields are set correctly
	if err.Code != code {
		t.Errorf("err.Code = %d, want %d", err.Code, code)
	}
	if err.Message != message {
		t.Errorf("err.Message = %q, want %q", err.Message, message)
	}
	if err.Err != originalErr {
		t.Errorf("err.Err = %v, want %v", err.Err, originalErr)
	}
}

// TestHttpErrorError tests the Error method of HttpError
func TestHttpErrorError(t *testing.T) {
	// Test Error() with no underlying error
	err1 := NewHttpError(StatusBadRequest, "Bad request")
	expected1 := "Bad request"
	if got := err1.Error(); got != expected1 {
		t.Errorf("err1.Error() = %q, want %q", got, expected1)
	}

	// Test Error() with an underlying error
	originalErr := errors.New("database connection failed")
	err2 := NewHttpErrorWithError(StatusInternalServerError, "Internal server error", originalErr)
	expected2 := "Internal server error: database connection failed"
	if got := err2.Error(); got != expected2 {
		t.Errorf("err2.Error() = %q, want %q", got, expected2)
	}
}

// TestHttpErrorUnwrap tests the Unwrap method of HttpError
func TestHttpErrorUnwrap(t *testing.T) {
	// Test Unwrap() with no underlying error
	err1 := NewHttpError(StatusBadRequest, "Bad request")
	if got := err1.Unwrap(); got != nil {
		t.Errorf("err1.Unwrap() = %v, want nil", got)
	}

	// Test Unwrap() with an underlying error
	originalErr := errors.New("database connection failed")
	err2 := NewHttpErrorWithError(StatusInternalServerError, "Internal server error", originalErr)
	if got := err2.Unwrap(); got != originalErr {
		t.Errorf("err2.Unwrap() = %v, want %v", got, originalErr)
	}
}

// TestHttpErrorWithStandardErrors tests that HttpError works with standard error handling
func TestHttpErrorWithStandardErrors(t *testing.T) {
	// Create a new HttpError with an underlying error
	originalErr := errors.New("database connection failed")
	err := NewHttpErrorWithError(StatusInternalServerError, "Internal server error", originalErr)

	// Test errors.Is
	if !errors.Is(err, originalErr) {
		t.Errorf("errors.Is(err, originalErr) = false, want true")
	}

	// Test errors.As
	var httpErr *HttpError
	if !errors.As(err, &httpErr) {
		t.Errorf("errors.As(err, &httpErr) = false, want true")
	}
	if httpErr != err {
		t.Errorf("httpErr = %v, want %v", httpErr, err)
	}
}
