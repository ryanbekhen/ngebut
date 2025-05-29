package ngebut

import (
	"testing"
	"time"
)

// TestDefaultConfig tests the DefaultConfig function
func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	// Check ReadTimeout
	if config.ReadTimeout != 5*time.Second {
		t.Errorf("DefaultConfig().ReadTimeout = %v, want %v", config.ReadTimeout, 5*time.Second)
	}

	// Check WriteTimeout
	if config.WriteTimeout != 10*time.Second {
		t.Errorf("DefaultConfig().WriteTimeout = %v, want %v", config.WriteTimeout, 10*time.Second)
	}

	// Check IdleTimeout
	if config.IdleTimeout != 15*time.Second {
		t.Errorf("DefaultConfig().IdleTimeout = %v, want %v", config.IdleTimeout, 15*time.Second)
	}

	// Check DisableStartupMessage
	if config.DisableStartupMessage != false {
		t.Errorf("DefaultConfig().DisableStartupMessage = %v, want %v", config.DisableStartupMessage, false)
	}

	// Check ErrorHandler is not nil
	if config.ErrorHandler == nil {
		t.Error("DefaultConfig().ErrorHandler is nil, want non-nil")
	}
}

// TestConfigZeroValues tests that a zero-value Config has zero values for all fields
func TestConfigZeroValues(t *testing.T) {
	var config Config

	// Check ReadTimeout
	if config.ReadTimeout != 0 {
		t.Errorf("Zero-value Config.ReadTimeout = %v, want 0", config.ReadTimeout)
	}

	// Check WriteTimeout
	if config.WriteTimeout != 0 {
		t.Errorf("Zero-value Config.WriteTimeout = %v, want 0", config.WriteTimeout)
	}

	// Check IdleTimeout
	if config.IdleTimeout != 0 {
		t.Errorf("Zero-value Config.IdleTimeout = %v, want 0", config.IdleTimeout)
	}

	// Check DisableStartupMessage
	if config.DisableStartupMessage != false {
		t.Errorf("Zero-value Config.DisableStartupMessage = %v, want false", config.DisableStartupMessage)
	}

	// Check ErrorHandler
	if config.ErrorHandler != nil {
		t.Errorf("Zero-value Config.ErrorHandler = %v, want nil", config.ErrorHandler)
	}
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

	// Check ReadTimeout
	if config.ReadTimeout != 30*time.Second {
		t.Errorf("Custom Config.ReadTimeout = %v, want %v", config.ReadTimeout, 30*time.Second)
	}

	// Check WriteTimeout
	if config.WriteTimeout != 45*time.Second {
		t.Errorf("Custom Config.WriteTimeout = %v, want %v", config.WriteTimeout, 45*time.Second)
	}

	// Check IdleTimeout
	if config.IdleTimeout != 60*time.Second {
		t.Errorf("Custom Config.IdleTimeout = %v, want %v", config.IdleTimeout, 60*time.Second)
	}

	// Check DisableStartupMessage
	if config.DisableStartupMessage != true {
		t.Errorf("Custom Config.DisableStartupMessage = %v, want %v", config.DisableStartupMessage, true)
	}

	// Check ErrorHandler is the custom handler
	if config.ErrorHandler == nil {
		t.Error("Custom Config.ErrorHandler is nil, want non-nil")
	}

	// We can't directly compare function values, so we just check it's not nil
	// The fact that we set it to customHandler is sufficient for this test
}
