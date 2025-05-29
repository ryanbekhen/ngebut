package log

import (
	"bytes"
	"errors"
	"strings"
	"testing"
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
		if got := test.level.String(); got != test.expected {
			t.Errorf("Level(%d).String() = %s, expected %s", test.level, got, test.expected)
		}
	}
}

// TestLoggerCreation tests the creation of loggers
func TestLoggerCreation(t *testing.T) {
	// Test New with nil writer
	logger := New(nil, InfoLevel)
	if logger == nil {
		t.Fatal("New(nil, InfoLevel) returned nil")
	}
	if logger.writer == nil {
		t.Error("New(nil, InfoLevel) should set writer to os.Stdout")
	}
	if logger.level != InfoLevel {
		t.Errorf("New(nil, InfoLevel) set level to %v, expected %v", logger.level, InfoLevel)
	}

	// Test New with custom writer
	buf := &bytes.Buffer{}
	logger = New(buf, DebugLevel)
	if logger.writer != buf {
		t.Error("New(buf, DebugLevel) did not set the correct writer")
	}
	if logger.level != DebugLevel {
		t.Errorf("New(buf, DebugLevel) set level to %v, expected %v", logger.level, DebugLevel)
	}

	// Test NewWithConfig
	config := DefaultLoggerConfig()
	config.Writer = buf
	config.Level = WarnLevel
	config.TimeFormat = "custom-format"
	config.NoColor = true
	logger = NewWithConfig(config)
	if logger.writer != buf {
		t.Error("NewWithConfig did not set the correct writer")
	}
	if logger.level != WarnLevel {
		t.Errorf("NewWithConfig set level to %v, expected %v", logger.level, WarnLevel)
	}
	if logger.timeFormat != "custom-format" {
		t.Errorf("NewWithConfig set timeFormat to %v, expected custom-format", logger.timeFormat)
	}
	if !logger.noColor {
		t.Error("NewWithConfig did not set noColor to true")
	}
}

// TestLoggerLevelMethods tests the level methods of Logger
func TestLoggerLevelMethods(t *testing.T) {
	logger := New(nil, InfoLevel)

	// Debug should return nil because level is InfoLevel
	if event := logger.Debug(); event != nil {
		t.Error("Debug() should return nil when level is InfoLevel")
	}

	// Info should return an event
	if event := logger.Info(); event == nil {
		t.Error("Info() should return an event when level is InfoLevel")
	} else if event.(*Event).level != InfoLevel {
		t.Errorf("Info() returned event with level %v, expected %v", event.(*Event).level, InfoLevel)
	}

	// Warn should return an event
	if event := logger.Warn(); event == nil {
		t.Error("Warn() should return an event when level is InfoLevel")
	} else if event.(*Event).level != WarnLevel {
		t.Errorf("Warn() returned event with level %v, expected %v", event.(*Event).level, WarnLevel)
	}

	// Error should return an event
	if event := logger.Error(); event == nil {
		t.Error("Error() should return an event when level is InfoLevel")
	} else if event.(*Event).level != ErrorLevel {
		t.Errorf("Error() returned event with level %v, expected %v", event.(*Event).level, ErrorLevel)
	}

	// Fatal should always return an event
	if event := logger.Fatal(); event == nil {
		t.Error("Fatal() should always return an event")
	} else if event.(*Event).level != FatalLevel {
		t.Errorf("Fatal() returned event with level %v, expected %v", event.(*Event).level, FatalLevel)
	}

	// Test SetLevel and GetLevel
	logger.SetLevel(DebugLevel)
	if level := logger.GetLevel(); level != DebugLevel {
		t.Errorf("GetLevel() returned %v after SetLevel(DebugLevel), expected %v", level, DebugLevel)
	}
}

// TestEventMethods tests the methods of Event
func TestEventMethods(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(buf, DebugLevel)

	// Test Err method indirectly
	testErr := errors.New("test error")
	// We can't access the err field directly, but we can verify the Err method returns the event
	event := logger.Debug().Err(testErr)
	if event == nil {
		t.Error("Err() should return the event")
	}

	// Test Msg method
	buf.Reset()
	logger.Debug().Msg("test message")
	output := buf.String()
	if !strings.Contains(output, "DEBUG") {
		t.Errorf("Msg() output does not contain level: %s", output)
	}
	if !strings.Contains(output, "test message") {
		t.Errorf("Msg() output does not contain message: %s", output)
	}

	// Test Msgf method
	buf.Reset()
	logger.Info().Msgf("formatted %s %d", "message", 42)
	output = buf.String()
	if !strings.Contains(output, "INFO") {
		t.Errorf("Msgf() output does not contain level: %s", output)
	}
	if !strings.Contains(output, "formatted message 42") {
		t.Errorf("Msgf() output does not contain formatted message: %s", output)
	}

	// Test nil event handling
	var nilEvent *Event
	nilEvent.Msg("should not panic")
	nilEvent.Msgf("should not %s", "panic")
	nilEvent.Err(testErr)
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
	if !strings.Contains(buf.String(), "DEBUG") || !strings.Contains(buf.String(), "debug message") {
		t.Errorf("Debug() output incorrect: %s", buf.String())
	}

	// Test Info
	buf.Reset()
	Info().Msg("info message")
	if !strings.Contains(buf.String(), "INFO") || !strings.Contains(buf.String(), "info message") {
		t.Errorf("Info() output incorrect: %s", buf.String())
	}

	// Test Warn
	buf.Reset()
	Warn().Msg("warn message")
	if !strings.Contains(buf.String(), "WARN") || !strings.Contains(buf.String(), "warn message") {
		t.Errorf("Warn() output incorrect: %s", buf.String())
	}

	// Test Error
	buf.Reset()
	Error().Msg("error message")
	if !strings.Contains(buf.String(), "ERROR") || !strings.Contains(buf.String(), "error message") {
		t.Errorf("Error() output incorrect: %s", buf.String())
	}

	// Test Fatal
	buf.Reset()
	Fatal().Msg("fatal message")
	if !strings.Contains(buf.String(), "FATAL") || !strings.Contains(buf.String(), "fatal message") {
		t.Errorf("Fatal() output incorrect: %s", buf.String())
	}

	// Test level filtering
	SetLevel(ErrorLevel)
	buf.Reset()
	Debug().Msg("should not appear")
	Info().Msg("should not appear")
	Warn().Msg("should not appear")
	if buf.Len() > 0 {
		t.Errorf("Messages below ErrorLevel should be filtered out, but got: %s", buf.String())
	}

	Error().Msg("should appear")
	if !strings.Contains(buf.String(), "should appear") {
		t.Errorf("Error message should appear when level is ErrorLevel, but got: %s", buf.String())
	}
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
		if got := string(buf); got != test.expected {
			t.Errorf("appendInt(%d) = %s, expected %s", test.n, got, test.expected)
		}
	}
}
