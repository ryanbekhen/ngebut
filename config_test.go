package ngebut

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// TestDefaultConfig tests the DefaultConfig function
func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	assert.Equal(t, config.ReadTimeout, 5*time.Second, "DefaultConfig().ReadTimeout should be 5 seconds")
	assert.Equal(t, config.WriteTimeout, 10*time.Second, "DefaultConfig().WriteTimeout should be 10 seconds")
	assert.Equal(t, config.IdleTimeout, 15*time.Second, "DefaultConfig().IdleTimeout should be 15 seconds")
	assert.Equal(t, config.DisableStartupMessage, false, "DefaultConfig().DisableStartupMessage should be false")
	assert.NotNil(t, config.ErrorHandler, "DefaultConfig().ErrorHandler should not be nil")
}

// TestConfigZeroValues tests that a zero-value Config has zero values for all fields
func TestConfigZeroValues(t *testing.T) {
	var config Config
	assert.Equal(t, config.ReadTimeout, 0*time.Second, "Zero-value Config.ReadTimeout should be 0 seconds")
	assert.Equal(t, config.WriteTimeout, 0*time.Second, "Zero-value Config.WriteTimeout should be 0 seconds")
	assert.Equal(t, config.IdleTimeout, 0*time.Second, "Zero-value Config.IdleTimeout should be 0 seconds")
	assert.Equal(t, config.DisableStartupMessage, false, "Zero-value Config.DisableStartupMessage should be false")
	assert.Nil(t, config.ErrorHandler, "Zero-value Config.ErrorHandler should be nil")
}

// TestConfigCustomValues tests setting custom values for Config fields
func TestConfigCustomValues(t *testing.T) {
	customHandler := func(c *Ctx) {
		c.Status(StatusInternalServerError).String("Custom error handler")
	}

	config := Config{
		ReadTimeout:           30 * time.Second,
		WriteTimeout:          45 * time.Second,
		IdleTimeout:           60 * time.Second,
		DisableStartupMessage: true,
		ErrorHandler:          customHandler,
	}

	assert.Equal(t, config.ReadTimeout, 30*time.Second, "Custom Config.ReadTimeout should be 30 seconds")
	assert.Equal(t, config.WriteTimeout, 45*time.Second, "Custom Config.WriteTimeout should be 45 seconds")
	assert.Equal(t, config.IdleTimeout, 60*time.Second, "Custom Config.IdleTimeout should be 60 seconds")
	assert.Equal(t, config.DisableStartupMessage, true, "Custom Config.DisableStartupMessage should be true")
	assert.NotNil(t, config.ErrorHandler, "Custom Config.ErrorHandler should not be nil")
}
