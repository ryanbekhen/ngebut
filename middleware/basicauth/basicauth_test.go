package basicauth

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	assert.Contains(t, "example", config.Username, "DefaultConfig() returned unexpected Username value")
	assert.Contains(t, "example", config.Password, "DefaultConfig() returned unexpected Password value")
}

func TestCustomeConfig(t *testing.T) {
	config := DefaultConfig()
	config.Username = "admin"
	config.Password = "password"

	assert.Contains(t, config.Username, "admin")
	assert.Contains(t, config.Password, "password")
}

func TestNew(t *testing.T) {
	middleware := New()
	assert.NotNil(t, middleware, "New() returned nil")

	customConfig := Config{
		Username: "myuser",
		Password: "mypassword",
	}
	assert.Equal(t, "myuser", customConfig.Username, "New() returned unexpected Username value")
}
