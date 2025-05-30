package ngebut

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestProtocolWithXForwardedProto tests the Protocol method with X-Forwarded-Proto header
func TestProtocolWithXForwardedProto(t *testing.T) {
	// Create a request with X-Forwarded-Proto header
	req, _ := http.NewRequest("GET", "http://example.com/test", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	res := httptest.NewRecorder()
	ctx := GetContext(res, req)

	// Check that Protocol returns the X-Forwarded-Proto header
	assert.Equal(t, "https", ctx.Protocol(), "Protocol should return X-Forwarded-Proto header value")
}

// TestProtocolWithXForwardedProtocol tests the Protocol method with X-Forwarded-Protocol header
func TestProtocolWithXForwardedProtocol(t *testing.T) {
	// Create a request with X-Forwarded-Protocol header
	req, _ := http.NewRequest("GET", "http://example.com/test", nil)
	req.Header.Set("X-Forwarded-Protocol", "https")
	res := httptest.NewRecorder()
	ctx := GetContext(res, req)

	// Check that Protocol returns the X-Forwarded-Protocol header
	assert.Equal(t, "https", ctx.Protocol(), "Protocol should return X-Forwarded-Protocol header value")
}

// TestProtocolWithFrontEndHttps tests the Protocol method with Front-End-Https header
func TestProtocolWithFrontEndHttps(t *testing.T) {
	// Create a request with Front-End-Https header
	req, _ := http.NewRequest("GET", "http://example.com/test", nil)
	req.Header.Set("Front-End-Https", "on")
	res := httptest.NewRecorder()
	ctx := GetContext(res, req)

	// Check that Protocol returns https
	assert.Equal(t, "https", ctx.Protocol(), "Protocol should return https when Front-End-Https is on")
}

// TestProtocolWithXForwardedSsl tests the Protocol method with X-Forwarded-Ssl header
func TestProtocolWithXForwardedSsl(t *testing.T) {
	// Create a request with X-Forwarded-Ssl header
	req, _ := http.NewRequest("GET", "http://example.com/test", nil)
	req.Header.Set("X-Forwarded-Ssl", "on")
	res := httptest.NewRecorder()
	ctx := GetContext(res, req)

	// Check that Protocol returns https
	assert.Equal(t, "https", ctx.Protocol(), "Protocol should return https when X-Forwarded-Ssl is on")
}

// TestProtocolWithURLScheme tests the Protocol method with URL.Scheme
func TestProtocolWithURLScheme(t *testing.T) {
	// Create a request with URL.Scheme set to https
	req, _ := http.NewRequest("GET", "https://example.com/test", nil)
	res := httptest.NewRecorder()
	ctx := GetContext(res, req)

	// Check that Protocol returns the URL.Scheme
	assert.Equal(t, "https", ctx.Protocol(), "Protocol should return URL.Scheme")
}

// TestProtocolDefault tests the Protocol method with no protocol information
func TestProtocolDefault(t *testing.T) {
	// Create a request with no protocol information
	req, _ := http.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := GetContext(res, req)

	// Check that Protocol returns the default http
	assert.Equal(t, "http", ctx.Protocol(), "Protocol should return default http")
}

// TestProtocolNilRequest tests the Protocol method with nil request
func TestProtocolNilRequest(t *testing.T) {
	// Create a context with nil request
	res := httptest.NewRecorder()
	ctx := GetContext(res, nil)
	ctx.Request = nil

	// Check that Protocol returns empty string
	assert.Equal(t, "", ctx.Protocol(), "Protocol should return empty string for nil request")
}

// TestProtocolPriority tests the Protocol method prioritizes headers over URL.Scheme
func TestProtocolPriority(t *testing.T) {
	// Create a request with both X-Forwarded-Proto and URL.Scheme
	req, _ := http.NewRequest("GET", "http://example.com/test", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	res := httptest.NewRecorder()
	ctx := GetContext(res, req)

	// Check that Protocol returns the X-Forwarded-Proto header (prioritized over URL.Scheme)
	assert.Equal(t, "https", ctx.Protocol(), "Protocol should prioritize X-Forwarded-Proto over URL.Scheme")
}
