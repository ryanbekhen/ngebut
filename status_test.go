package ngebut

import (
	"testing"
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
		if got != tc.text {
			t.Errorf("StatusText(%d) = %q, want %q", tc.code, got, tc.text)
		}
	}
}

// TestStatusCodes tests that all status codes are defined correctly
func TestStatusCodes(t *testing.T) {
	// Test that status codes are defined with the correct values
	if StatusOK != 200 {
		t.Errorf("StatusOK = %d, want 200", StatusOK)
	}
	if StatusCreated != 201 {
		t.Errorf("StatusCreated = %d, want 201", StatusCreated)
	}
	if StatusBadRequest != 400 {
		t.Errorf("StatusBadRequest = %d, want 400", StatusBadRequest)
	}
	if StatusInternalServerError != 500 {
		t.Errorf("StatusInternalServerError = %d, want 500", StatusInternalServerError)
	}

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
		if text := StatusText(code); text == "" {
			t.Errorf("StatusText(%d) returned empty string, expected a description", code)
		}
	}
}

// TestHTTPMethods tests that all HTTP methods are defined correctly
func TestHTTPMethods(t *testing.T) {
	// Test that HTTP methods are defined with the correct values
	if MethodGet != "GET" {
		t.Errorf("MethodGet = %q, want \"GET\"", MethodGet)
	}
	if MethodPost != "POST" {
		t.Errorf("MethodPost = %q, want \"POST\"", MethodPost)
	}
	if MethodPut != "PUT" {
		t.Errorf("MethodPut = %q, want \"PUT\"", MethodPut)
	}
	if MethodDelete != "DELETE" {
		t.Errorf("MethodDelete = %q, want \"DELETE\"", MethodDelete)
	}
	if MethodPatch != "PATCH" {
		t.Errorf("MethodPatch = %q, want \"PATCH\"", MethodPatch)
	}
	if MethodHead != "HEAD" {
		t.Errorf("MethodHead = %q, want \"HEAD\"", MethodHead)
	}
	if MethodOptions != "OPTIONS" {
		t.Errorf("MethodOptions = %q, want \"OPTIONS\"", MethodOptions)
	}
	if MethodConnect != "CONNECT" {
		t.Errorf("MethodConnect = %q, want \"CONNECT\"", MethodConnect)
	}
	if MethodTrace != "TRACE" {
		t.Errorf("MethodTrace = %q, want \"TRACE\"", MethodTrace)
	}
}

// TestStatusTextEdgeCases tests edge cases for the StatusText function
func TestStatusTextEdgeCases(t *testing.T) {
	// Test negative status code
	if text := StatusText(-1); text != "" {
		t.Errorf("StatusText(-1) = %q, want \"\"", text)
	}

	// Test zero status code
	if text := StatusText(0); text != "" {
		t.Errorf("StatusText(0) = %q, want \"\"", text)
	}

	// Test status code 306 (unused)
	if text := StatusText(306); text != "" {
		t.Errorf("StatusText(306) = %q, want \"\"", text)
	}

	// Test very large status code
	if text := StatusText(9999); text != "" {
		t.Errorf("StatusText(9999) = %q, want \"\"", text)
	}
}
