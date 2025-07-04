package basicauth

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	assert.Contains(t, "myuser", config.Username, "DefaultConfig() returned unexpected Username value")
	assert.Contains(t, "mypassword", config.Password, "DefaultConfig() returned unexpected Password value")
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
