package basicauth

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
