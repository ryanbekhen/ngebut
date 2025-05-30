package log

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewConsoleWriter tests the NewConsoleWriter function
func TestNewConsoleWriter(t *testing.T) {
	// Test with nil writer
	cw := NewConsoleWriter(nil)
	require.NotNil(t, cw, "NewConsoleWriter(nil) returned nil")

	// Test with custom writer
	buf := &bytes.Buffer{}
	cw = NewConsoleWriter(buf)
	assert.Equal(t, buf, cw.Out, "NewConsoleWriter did not set the output writer correctly")
	assert.Equal(t, time.RFC3339, cw.TimeFormat, "NewConsoleWriter set TimeFormat to %q, expected %q", cw.TimeFormat, time.RFC3339)
	assert.False(t, cw.NoColor, "NewConsoleWriter set NoColor to true, expected false")
	assert.NotNil(t, cw.buf, "NewConsoleWriter did not initialize the buffer")
}

// TestConsoleWriterWrite tests the Write method of ConsoleWriter
func TestConsoleWriterWrite(t *testing.T) {
	buf := &bytes.Buffer{}
	cw := NewConsoleWriter(buf)
	cw.NoColor = true // Disable color for easier testing

	// Test writing a simple log line
	logLine := []byte("2023-01-01 12:34:56 | INFO | Test message")
	n, err := cw.Write(logLine)
	assert.NoError(t, err, "ConsoleWriter.Write returned error: %v", err)
	assert.NotZero(t, n, "ConsoleWriter.Write returned 0 bytes written")

	output := buf.String()
	assert.NotEmpty(t, output, "ConsoleWriter.Write did not write anything to the buffer")

	// Test writing a malformed log line (no separators)
	buf.Reset()
	logLine = []byte("Malformed log line")
	_, _ = cw.Write(logLine)
	assert.Equal(t, "Malformed log line", buf.String(), "ConsoleWriter.Write did not pass through malformed log line")

	// Test writing a log line with error
	buf.Reset()
	cw.NoColor = true
	logLine = []byte("2023-01-01 12:34:56 | ERROR | error: Something went wrong")
	_, _ = cw.Write(logLine)
	assert.Contains(t, buf.String(), "error: Something went wrong", "ConsoleWriter.Write did not format error message correctly")
}

// TestDefaultConsoleWriter tests the DefaultConsoleWriter function
func TestDefaultConsoleWriter(t *testing.T) {
	cw := DefaultConsoleWriter()
	require.NotNil(t, cw, "DefaultConsoleWriter() returned nil")
	require.NotNil(t, cw.FormatLevel, "DefaultConsoleWriter() did not set FormatLevel function")

	// Test the FormatLevel function
	levels := []Level{DebugLevel, InfoLevel, WarnLevel, ErrorLevel, FatalLevel, Level(99)}
	for _, level := range levels {
		formatted := cw.FormatLevel(level)
		assert.NotEmpty(t, formatted, "FormatLevel(%v) returned empty string", level)
	}
}

// TestFormatLevelNoAlloc tests the formatLevelNoAlloc function
func TestFormatLevelNoAlloc(t *testing.T) {
	// Test with color disabled
	buf := make([]byte, 0, 32)
	levels := []Level{DebugLevel, InfoLevel, WarnLevel, ErrorLevel, FatalLevel, Level(99)}

	for _, level := range levels {
		result := formatLevelNoAlloc(buf[:0], level, true)
		assert.NotEmpty(t, result, "formatLevelNoAlloc(buf, %v, true) returned empty result", level)
	}

	// Test with color enabled
	for _, level := range levels {
		result := formatLevelNoAlloc(buf[:0], level, false)
		assert.NotEmpty(t, result, "formatLevelNoAlloc(buf, %v, false) returned empty result", level)
	}
}

// TestHelperFunctions tests the helper functions in console.go
func TestHelperFunctions(t *testing.T) {
	// Test findSeparator
	testData := []byte("part1 | part2 | part3")
	pos := findSeparator(testData, 0)
	assert.Equal(t, 5, pos, "findSeparator returned %d, expected 5", pos)

	pos = findSeparator(testData, 6)
	assert.Equal(t, 13, pos, "findSeparator returned %d, expected 13", pos)

	pos = findSeparator(testData, 14)
	assert.Equal(t, -1, pos, "findSeparator returned %d, expected -1", pos)

	// Test parseIntFromBytes
	intBytes := []byte("12345")
	n := parseIntFromBytes(intBytes)
	assert.Equal(t, 12345, n, "parseIntFromBytes returned %d, expected 12345", n)

	// Test bytesEqual
	a := []byte("test")
	b := []byte("test")
	c := []byte("different")

	assert.True(t, bytesEqual(a, b), "bytesEqual returned false for equal byte slices")
	assert.False(t, bytesEqual(a, c), "bytesEqual returned true for different byte slices")
	assert.False(t, bytesEqual(a, a[:2]), "bytesEqual returned true for slices of different length")
}
