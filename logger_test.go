package ngebut

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/ryanbekhen/ngebut/log"
)

// TestInitLogger tests the initLogger function
func TestInitLogger(t *testing.T) {
	// Save the original logger to restore it later
	originalLogger := logger
	// We can't directly access the default logger's output and level,
	// so we'll create a new writer to capture output
	var buf bytes.Buffer
	defer func() {
		logger = originalLogger
		// Restore default output
		log.SetOutput(os.Stdout)
	}()

	// Test with different log levels
	testCases := []struct {
		level    log.Level
		expected log.Level
	}{
		{log.DebugLevel, log.DebugLevel},
		{log.InfoLevel, log.InfoLevel},
		{log.WarnLevel, log.WarnLevel},
		{log.ErrorLevel, log.ErrorLevel},
		{log.Level(99), log.InfoLevel}, // Invalid level should default to InfoLevel
	}

	for _, tc := range testCases {
		// Initialize the logger with the test level
		initLogger(tc.level)

		// Check that the logger level is set correctly
		if logger.GetLevel() != tc.expected {
			t.Errorf("initLogger(%v) set logger level to %v, want %v", tc.level, logger.GetLevel(), tc.expected)
		}

		// We can't directly check the global log level, but we can test that
		// logging works at the expected level by writing to our buffer
		log.SetOutput(&buf)

		// Write a test message at each level
		log.Debug().Msg("Debug message")
		log.Info().Msg("Info message")
		log.Warn().Msg("Warn message")
		log.Error().Msg("Error message")

		output := buf.String()
		buf.Reset()

		// Check that messages at or above the expected level are logged
		hasDebug := strings.Contains(output, "Debug message")
		hasInfo := strings.Contains(output, "Info message")
		hasWarn := strings.Contains(output, "Warn message")
		hasError := strings.Contains(output, "Error message")

		switch tc.expected {
		case log.DebugLevel:
			if !hasDebug || !hasInfo || !hasWarn || !hasError {
				t.Errorf("initLogger(%v) didn't log all messages at DebugLevel", tc.level)
			}
		case log.InfoLevel:
			if hasDebug || !hasInfo || !hasWarn || !hasError {
				t.Errorf("initLogger(%v) didn't log correctly at InfoLevel", tc.level)
			}
		case log.WarnLevel:
			if hasDebug || hasInfo || !hasWarn || !hasError {
				t.Errorf("initLogger(%v) didn't log correctly at WarnLevel", tc.level)
			}
		case log.ErrorLevel:
			if hasDebug || hasInfo || hasWarn || !hasError {
				t.Errorf("initLogger(%v) didn't log correctly at ErrorLevel", tc.level)
			}
		}
	}
}

// TestDisplayStartupMessage tests the displayStartupMessage function
func TestDisplayStartupMessage(t *testing.T) {
	// Save the original logger to restore it later
	originalLogger := logger
	defer func() {
		logger = originalLogger
	}()

	// Create a buffer to capture the output
	var buf bytes.Buffer
	console := log.DefaultConsoleWriter()
	console.Out = &buf
	logger = log.New(console, log.InfoLevel)

	// Call the function with a test address
	addr := ":8080"
	displayStartupMessage(addr)

	// Check that the output contains the expected messages
	output := buf.String()

	// Check for the ASCII art logo (first line)
	if !strings.Contains(output, "_   _            _           _") {
		t.Errorf("displayStartupMessage() output doesn't contain the logo")
	}

	// Check for the server address
	if !strings.Contains(output, "Server is running on :8080") {
		t.Errorf("displayStartupMessage() output doesn't contain the server address")
	}

	// Check for the stop message
	if !strings.Contains(output, "Press Ctrl+C to stop the server") {
		t.Errorf("displayStartupMessage() output doesn't contain the stop message")
	}
}

// TestLoggerIntegration tests the integration of the logger with the rest of the system
func TestLoggerIntegration(t *testing.T) {
	// Save the original stdout to restore it later
	originalStdout := os.Stdout
	defer func() {
		os.Stdout = originalStdout
	}()

	// Create a pipe to capture stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = w

	// Initialize the logger with debug level
	initLogger(log.DebugLevel)

	// Write some log messages
	logger.Debug().Msg("This is a debug message")
	logger.Info().Msg("This is an info message")
	logger.Warn().Msg("This is a warning message")
	logger.Error().Msg("This is an error message")

	// Close the write end of the pipe to flush the buffer
	w.Close()

	// Read the output from the pipe
	var buf bytes.Buffer
	_, err = buf.ReadFrom(r)
	if err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}
	output := buf.String()

	// Check that the output contains the expected messages
	if !strings.Contains(output, "This is a debug message") {
		t.Errorf("Logger output doesn't contain the debug message")
	}
	if !strings.Contains(output, "This is an info message") {
		t.Errorf("Logger output doesn't contain the info message")
	}
	if !strings.Contains(output, "This is a warning message") {
		t.Errorf("Logger output doesn't contain the warning message")
	}
	if !strings.Contains(output, "This is an error message") {
		t.Errorf("Logger output doesn't contain the error message")
	}
}
