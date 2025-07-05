package basicauth

import (
	"github.com/stretchr/testify/assert"
	"encoding/base64"
	"net/http/httptest"
	"testing"
	"github.com/ryanbekhen/ngebut"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	assert.Equal(t, "example", config.Username, "DefaultConfig() returned unexpected Username value")
	assert.Equal(t, "example", config.Password, "DefaultConfig() returned unexpected Password value")
}

func TestCustomConfig(t *testing.T) {
	config := DefaultConfig()
	config.Username = "admin"
	config.Password = "password"

	assert.Contains(t, config.Username, "admin")
	assert.Contains(t, config.Password, "password")
}

func TestNew(t *testing.T) {
	customConfig := Config{
		Username: "myuser",
		Password: "mypassword",
	}
	middleware := New(customConfig)
	assert.NotNil(t, middleware, "New() returned nil")
	assert.Equal(t, "myuser", customConfig.Username, "New() returned unexpected Username value")
}

func newTestCtxWithAuthHeader(authHeader string) *ngebut.Ctx {
	   req := httptest.NewRequest("GET", "/", nil)
	   if authHeader != "" {
			   req.Header.Set("Authorization", authHeader)
	   }
	   rw := httptest.NewRecorder()
	   ctx := ngebut.GetContext(rw, req)
	   return ctx
}


func TestBasicAuth_Success(t *testing.T) {
	   cfg := Config{Username: "user", Password: "pass"}
	   mw := New(cfg)
	   creds := base64.StdEncoding.EncodeToString([]byte("user:pass"))
	   ctx := newTestCtxWithAuthHeader("Basic " + creds)
	   err := mw(ctx)
	   assert.Nil(t, err, "Expected no error for valid credentials")
}

func TestBasicAuth_Failure_InvalidPassword(t *testing.T) {
	   cfg := Config{Username: "user", Password: "pass"}
	   mw := New(cfg)
	   creds := base64.StdEncoding.EncodeToString([]byte("user:wrong"))
	   ctx := newTestCtxWithAuthHeader("Basic " + creds)
	   err := mw(ctx)
	   assert.Error(t, err)
	   httpErr, ok := err.(*ngebut.HttpError)
	   assert.True(t, ok, "Error should be of type *HttpError")
	   assert.Equal(t, 401, httpErr.Code)
	   assert.Equal(t, "Unauthorized", httpErr.Message)
}

func TestBasicAuth_Failure_NoHeader(t *testing.T) {
	   cfg := Config{Username: "user", Password: "pass"}
	   mw := New(cfg)
	   ctx := newTestCtxWithAuthHeader("")
	   err := mw(ctx)
	   assert.Error(t, err)
	   httpErr, ok := err.(*ngebut.HttpError)
	   assert.True(t, ok, "Error should be of type *HttpError")
	   assert.Equal(t, 401, httpErr.Code)
	   assert.Equal(t, "Unauthorized", httpErr.Message)
}

func TestBasicAuth_Failure_MalformedBase64(t *testing.T) {
	   cfg := Config{Username: "user", Password: "pass"}
	   mw := New(cfg)
	   ctx := newTestCtxWithAuthHeader("Basic invalid-base64")
	   err := mw(ctx)
	   assert.Error(t, err)
	   httpErr, ok := err.(*ngebut.HttpError)
	   assert.True(t, ok, "Error should be of type *HttpError")
	   assert.Equal(t, 401, httpErr.Code)
	   assert.Equal(t, "Unauthorized", httpErr.Message)
}

func TestBasicAuth_Failure_NoColon(t *testing.T) {
	   cfg := Config{Username: "user", Password: "pass"}
	   mw := New(cfg)
	   creds := base64.StdEncoding.EncodeToString([]byte("userpass"))
	   ctx := newTestCtxWithAuthHeader("Basic " + creds)
	   err := mw(ctx)
	   assert.Error(t, err)
	   httpErr, ok := err.(*ngebut.HttpError)
	   assert.True(t, ok, "Error should be of type *HttpError")
	   assert.Equal(t, 401, httpErr.Code)
	   assert.Equal(t, "Unauthorized", httpErr.Message)
}

func TestBasicAuth_Failure(t *testing.T) {
	cfg := Config{Username: "user", Password: "pass"}
	mw := New(cfg)
	creds := base64.StdEncoding.EncodeToString([]byte("user:wrong"))
	ctx := newTestCtxWithAuthHeader("Basic " + creds)
	err := mw(ctx)
	assert.Equal(t, ErrUnauthorized, err)
}
