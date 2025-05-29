package ngebut

import (
	"fmt"
)

// HttpError represents an HTTP error with a status code and message.
type HttpError struct {
	Code    int    // HTTP status code
	Message string // Error message
	Err     error  // Original error, if any
}

// Error implements the error interface.
func (e *HttpError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the wrapped error, if any.
func (e *HttpError) Unwrap() error {
	return e.Err
}

// NewHttpError creates a new HttpError with the given status code and message.
func NewHttpError(code int, message string) *HttpError {
	return &HttpError{
		Code:    code,
		Message: message,
	}
}

// NewHttpErrorWithError creates a new HttpError with the given status code, message, and error.
func NewHttpErrorWithError(code int, message string, err error) *HttpError {
	return &HttpError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
