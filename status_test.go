package ngebut

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestStatusText tests the StatusText function
func TestStatusText(t *testing.T) {
	// Test some common status codes
	testCases := []struct {
		code int
		text string
	}{
		{StatusOK, "OK"},
		{StatusCreated, "Created"},
		{StatusNoContent, "No Content"},
		{StatusMovedPermanently, "Moved Permanently"},
		{StatusFound, "Found"},
		{StatusBadRequest, "Bad Request"},
		{StatusUnauthorized, "Unauthorized"},
		{StatusForbidden, "Forbidden"},
		{StatusNotFound, "Not Found"},
		{StatusMethodNotAllowed, "Method Not Allowed"},
		{StatusInternalServerError, "Internal Server Error"},
		{StatusNotImplemented, "Not Implemented"},
		{StatusBadGateway, "Bad Gateway"},
		{StatusServiceUnavailable, "Service Unavailable"},
		{StatusGatewayTimeout, "Gateway Timeout"},
		// Test a non-standard status code
		{999, ""},
	}

	for _, tc := range testCases {
		got := StatusText(tc.code)
		assert.Equal(t, tc.text, got, "StatusText(%d) returned incorrect value", tc.code)
	}
}

// TestStatusCodes tests that all status codes are defined correctly
func TestStatusCodes(t *testing.T) {
	// Test that status codes are defined with the correct values
	assert.Equal(t, 200, StatusOK, "StatusOK should be 200")
	assert.Equal(t, 201, StatusCreated, "StatusCreated should be 201")
	assert.Equal(t, 400, StatusBadRequest, "StatusBadRequest should be 400")
	assert.Equal(t, 500, StatusInternalServerError, "StatusInternalServerError should be 500")

	// Test that all status codes have a corresponding text
	// This ensures that StatusText handles all defined status codes
	statusCodes := []int{
		StatusContinue, StatusSwitchingProtocols, StatusProcessing, StatusEarlyHints,
		StatusOK, StatusCreated, StatusAccepted, StatusNonAuthoritativeInfo,
		StatusNoContent, StatusResetContent, StatusPartialContent, StatusMultiStatus,
		StatusAlreadyReported, StatusIMUsed,
		StatusMultipleChoices, StatusMovedPermanently, StatusFound, StatusSeeOther,
		StatusNotModified, StatusUseProxy, StatusTemporaryRedirect, StatusPermanentRedirect,
		StatusBadRequest, StatusUnauthorized, StatusPaymentRequired, StatusForbidden,
		StatusNotFound, StatusMethodNotAllowed, StatusNotAcceptable, StatusProxyAuthRequired,
		StatusRequestTimeout, StatusConflict, StatusGone, StatusLengthRequired,
		StatusPreconditionFailed, StatusRequestEntityTooLarge, StatusRequestURITooLong,
		StatusUnsupportedMediaType, StatusRequestedRangeNotSatisfiable, StatusExpectationFailed,
		StatusTeapot, StatusMisdirectedRequest, StatusUnprocessableEntity, StatusLocked,
		StatusFailedDependency, StatusTooEarly, StatusUpgradeRequired, StatusPreconditionRequired,
		StatusTooManyRequests, StatusRequestHeaderFieldsTooLarge, StatusUnavailableForLegalReasons,
		StatusInternalServerError, StatusNotImplemented, StatusBadGateway, StatusServiceUnavailable,
		StatusGatewayTimeout, StatusHTTPVersionNotSupported, StatusVariantAlsoNegotiates,
		StatusInsufficientStorage, StatusLoopDetected, StatusNotExtended, StatusNetworkAuthenticationRequired,
	}

	for _, code := range statusCodes {
		text := StatusText(code)
		assert.NotEmpty(t, text, "StatusText(%d) returned empty string, expected a description", code)
	}
}

// TestHTTPMethods tests that all HTTP methods are defined correctly
func TestHTTPMethods(t *testing.T) {
	// Test that HTTP methods are defined with the correct values
	assert.Equal(t, "GET", MethodGet, "MethodGet should be \"GET\"")
	assert.Equal(t, "POST", MethodPost, "MethodPost should be \"POST\"")
	assert.Equal(t, "PUT", MethodPut, "MethodPut should be \"PUT\"")
	assert.Equal(t, "DELETE", MethodDelete, "MethodDelete should be \"DELETE\"")
	assert.Equal(t, "PATCH", MethodPatch, "MethodPatch should be \"PATCH\"")
	assert.Equal(t, "HEAD", MethodHead, "MethodHead should be \"HEAD\"")
	assert.Equal(t, "OPTIONS", MethodOptions, "MethodOptions should be \"OPTIONS\"")
	assert.Equal(t, "CONNECT", MethodConnect, "MethodConnect should be \"CONNECT\"")
	assert.Equal(t, "TRACE", MethodTrace, "MethodTrace should be \"TRACE\"")
}

// TestStatusTextEdgeCases tests edge cases for the StatusText function
func TestStatusTextEdgeCases(t *testing.T) {
	// Test negative status code
	assert.Empty(t, StatusText(-1), "StatusText(-1) should return empty string")

	// Test zero status code
	assert.Empty(t, StatusText(0), "StatusText(0) should return empty string")

	// Test status code 306 (unused)
	assert.Empty(t, StatusText(306), "StatusText(306) should return empty string")

	// Test very large status code
	assert.Empty(t, StatusText(9999), "StatusText(9999) should return empty string")
}
