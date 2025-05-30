package log

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLevelString tests the String method of Level
func TestLevelString(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{DebugLevel, "DEBUG"},
		{InfoLevel, "INFO"},
		{WarnLevel, "WARN"},
		{ErrorLevel, "ERROR"},
		{FatalLevel, "FATAL"},
		{Level(99), "LEVEL(99)"}, // Unknown level
	}

	for _, test := range tests {
		got := test.level.String()
		assert.Equal(t, test.expected, got, "Level(%d).String() should match expected value", test.level)
	}
}

// TestLoggerCreation tests the creation of loggers
func TestLoggerCreation(t *testing.T) {
	// Test New with nil writer
	logger := New(nil, InfoLevel)
	assert.NotNil(t, logger, "New(nil, InfoLevel) should not return nil")
	assert.NotNil(t, logger.writer, "New(nil, InfoLevel) should set writer to os.Stdout")
	assert.Equal(t, InfoLevel, logger.level, "New(nil, InfoLevel) should set correct level")

	// Test New with custom writer
	buf := &bytes.Buffer{}
	logger = New(buf, DebugLevel)
	assert.Equal(t, buf, logger.writer, "New(buf, DebugLevel) should set the correct writer")
	assert.Equal(t, DebugLevel, logger.level, "New(buf, DebugLevel) should set the correct level")

	// Test NewWithConfig
	config := DefaultLoggerConfig()
	config.Writer = buf
	config.Level = WarnLevel
	config.TimeFormat = "custom-format"
	config.NoColor = true
	logger = NewWithConfig(config)
	assert.Equal(t, buf, logger.writer, "NewWithConfig should set the correct writer")
	assert.Equal(t, WarnLevel, logger.level, "NewWithConfig should set the correct level")
	assert.Equal(t, "custom-format", logger.timeFormat, "NewWithConfig should set the correct timeFormat")
	assert.True(t, logger.noColor, "NewWithConfig should set noColor to true")
}

// TestLoggerLevelMethods tests the level methods of Logger
func TestLoggerLevelMethods(t *testing.T) {
	logger := New(nil, InfoLevel)

	// Debug should return nil because level is InfoLevel
	event := logger.Debug()
	assert.Nil(t, event, "Debug() should return nil when level is InfoLevel")

	// Info should return an event
	event = logger.Info()
	assert.NotNil(t, event, "Info() should return an event when level is InfoLevel")
	assert.Equal(t, InfoLevel, event.(*Event).level, "Info() should return event with correct level")

	// Warn should return an event
	event = logger.Warn()
	assert.NotNil(t, event, "Warn() should return an event when level is InfoLevel")
	assert.Equal(t, WarnLevel, event.(*Event).level, "Warn() should return event with correct level")

	// Error should return an event
	event = logger.Error()
	assert.NotNil(t, event, "Error() should return an event when level is InfoLevel")
	assert.Equal(t, ErrorLevel, event.(*Event).level, "Error() should return event with correct level")

	// Fatal should always return an event
	event = logger.Fatal()
	assert.NotNil(t, event, "Fatal() should always return an event")
	assert.Equal(t, FatalLevel, event.(*Event).level, "Fatal() should return event with correct level")

	// Test SetLevel and GetLevel
	logger.SetLevel(DebugLevel)
	assert.Equal(t, DebugLevel, logger.GetLevel(), "GetLevel() should return correct level after SetLevel()")
}

// TestEventMethods tests the methods of Event
func TestEventMethods(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(buf, DebugLevel)

	// Test Err method indirectly
	testErr := errors.New("test error")
	// We can't access the err field directly, but we can verify the Err method returns the event
	event := logger.Debug().Err(testErr)
	assert.NotNil(t, event, "Err() should return the event")

	// Test Msg method
	buf.Reset()
	logger.Debug().Msg("test message")
	output := buf.String()
	assert.Contains(t, output, "DEBUG", "Msg() output should contain level")
	assert.Contains(t, output, "test message", "Msg() output should contain message")

	// Test Msgf method
	buf.Reset()
	logger.Info().Msgf("formatted %s %d", "message", 42)
	output = buf.String()
	assert.Contains(t, output, "INFO", "Msgf() output should contain level")
	assert.Contains(t, output, "formatted message 42", "Msgf() output should contain formatted message")

	// Test nil event handling - we're just making sure it doesn't panic
	var nilEvent *Event
	assert.NotPanics(t, func() {
		nilEvent.Msg("should not panic")
		nilEvent.Msgf("should not %s", "panic")
		nilEvent.Err(testErr)
	}, "Nil events should not panic")
}

// TestDefaultLogger tests the default logger functions
func TestDefaultLogger(t *testing.T) {
	// Save the original writer to restore it later
	originalWriter := defaultLogger.writer
	defer func() {
		defaultLogger.writer = originalWriter
		defaultLogger.level = InfoLevel
	}()

	buf := &bytes.Buffer{}
	SetOutput(buf)
	SetLevel(DebugLevel)

	// Test Debug
	buf.Reset()
	Debug().Msg("debug message")
	assert.Contains(t, buf.String(), "DEBUG", "Debug() output should contain level")
	assert.Contains(t, buf.String(), "debug message", "Debug() output should contain message")

	// Test Info
	buf.Reset()
	Info().Msg("info message")
	assert.Contains(t, buf.String(), "INFO", "Info() output should contain level")
	assert.Contains(t, buf.String(), "info message", "Info() output should contain message")

	// Test Warn
	buf.Reset()
	Warn().Msg("warn message")
	assert.Contains(t, buf.String(), "WARN", "Warn() output should contain level")
	assert.Contains(t, buf.String(), "warn message", "Warn() output should contain message")

	// Test Error
	buf.Reset()
	Error().Msg("error message")
	assert.Contains(t, buf.String(), "ERROR", "Error() output should contain level")
	assert.Contains(t, buf.String(), "error message", "Error() output should contain message")

	// Test Fatal
	buf.Reset()
	Fatal().Msg("fatal message")
	assert.Contains(t, buf.String(), "FATAL", "Fatal() output should contain level")
	assert.Contains(t, buf.String(), "fatal message", "Fatal() output should contain message")

	// Test level filtering
	SetLevel(ErrorLevel)
	buf.Reset()
	Debug().Msg("should not appear")
	Info().Msg("should not appear")
	Warn().Msg("should not appear")
	assert.Empty(t, buf.String(), "Messages below ErrorLevel should be filtered out")

	Error().Msg("should appear")
	assert.Contains(t, buf.String(), "should appear", "Error message should appear when level is ErrorLevel")
}

// TestAppendInt tests the appendInt function
func TestAppendInt(t *testing.T) {
	tests := []struct {
		n        int64
		expected string
	}{
		{0, "0"},
		{123, "123"},
		{-123, "-123"},
		{9223372036854775807, "9223372036854775807"},   // Max int64
		{-9223372036854775808, "-9223372036854775808"}, // Min int64
	}

	for _, test := range tests {
		buf := make([]byte, 0, 32)
		buf = appendInt(buf, test.n)
		assert.Equal(t, test.expected, string(buf), "appendInt(%d) should produce correct string", test.n)
	}
}
