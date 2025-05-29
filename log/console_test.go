package log

import (
	"bytes"
	"testing"
	"time"
)

// TestNewConsoleWriter tests the NewConsoleWriter function
func TestNewConsoleWriter(t *testing.T) {
	// Test with nil writer
	cw := NewConsoleWriter(nil)
	if cw == nil {
		t.Fatal("NewConsoleWriter(nil) returned nil")
	}

	// Test with custom writer
	buf := &bytes.Buffer{}
	cw = NewConsoleWriter(buf)
	if cw.Out != buf {
		t.Error("NewConsoleWriter did not set the output writer correctly")
	}
	if cw.TimeFormat != time.RFC3339 {
		t.Errorf("NewConsoleWriter set TimeFormat to %q, expected %q", cw.TimeFormat, time.RFC3339)
	}
	if cw.NoColor {
		t.Error("NewConsoleWriter set NoColor to true, expected false")
	}
	if cw.buf == nil {
		t.Error("NewConsoleWriter did not initialize the buffer")
	}
}

// TestConsoleWriterWrite tests the Write method of ConsoleWriter
func TestConsoleWriterWrite(t *testing.T) {
	buf := &bytes.Buffer{}
	cw := NewConsoleWriter(buf)
	cw.NoColor = true // Disable color for easier testing

	// Test writing a simple log line
	logLine := []byte("2023-01-01 12:34:56 | INFO | Test message")
	n, err := cw.Write(logLine)
	if err != nil {
		t.Errorf("ConsoleWriter.Write returned error: %v", err)
	}
	if n == 0 {
		t.Error("ConsoleWriter.Write returned 0 bytes written")
	}

	output := buf.String()
	if output == "" {
		t.Error("ConsoleWriter.Write did not write anything to the buffer")
	}

	// Test writing a malformed log line (no separators)
	buf.Reset()
	logLine = []byte("Malformed log line")
	_, _ = cw.Write(logLine)
	if buf.String() != "Malformed log line" {
		t.Errorf("ConsoleWriter.Write did not pass through malformed log line, got: %q", buf.String())
	}

	// Test writing a log line with error
	buf.Reset()
	cw.NoColor = true
	logLine = []byte("2023-01-01 12:34:56 | ERROR | error: Something went wrong")
	_, _ = cw.Write(logLine)
	if !bytes.Contains(buf.Bytes(), []byte("error: Something went wrong")) {
		t.Errorf("ConsoleWriter.Write did not format error message correctly, got: %q", buf.String())
	}
}

// TestDefaultConsoleWriter tests the DefaultConsoleWriter function
func TestDefaultConsoleWriter(t *testing.T) {
	cw := DefaultConsoleWriter()
	if cw == nil {
		t.Fatal("DefaultConsoleWriter() returned nil")
	}
	if cw.FormatLevel == nil {
		t.Error("DefaultConsoleWriter() did not set FormatLevel function")
	}

	// Test the FormatLevel function
	levels := []Level{DebugLevel, InfoLevel, WarnLevel, ErrorLevel, FatalLevel, Level(99)}
	for _, level := range levels {
		formatted := cw.FormatLevel(level)
		if formatted == "" {
			t.Errorf("FormatLevel(%v) returned empty string", level)
		}
	}
}

// TestFormatLevelNoAlloc tests the formatLevelNoAlloc function
func TestFormatLevelNoAlloc(t *testing.T) {
	// Test with color disabled
	buf := make([]byte, 0, 32)
	levels := []Level{DebugLevel, InfoLevel, WarnLevel, ErrorLevel, FatalLevel, Level(99)}

	for _, level := range levels {
		result := formatLevelNoAlloc(buf[:0], level, true)
		if len(result) == 0 {
			t.Errorf("formatLevelNoAlloc(buf, %v, true) returned empty result", level)
		}
	}

	// Test with color enabled
	for _, level := range levels {
		result := formatLevelNoAlloc(buf[:0], level, false)
		if len(result) == 0 {
			t.Errorf("formatLevelNoAlloc(buf, %v, false) returned empty result", level)
		}
	}
}

// TestHelperFunctions tests the helper functions in console.go
func TestHelperFunctions(t *testing.T) {
	// Test findSeparator
	testData := []byte("part1 | part2 | part3")
	pos := findSeparator(testData, 0)
	if pos != 5 {
		t.Errorf("findSeparator returned %d, expected 5", pos)
	}

	pos = findSeparator(testData, 6)
	if pos != 13 {
		t.Errorf("findSeparator returned %d, expected 13", pos)
	}

	pos = findSeparator(testData, 14)
	if pos != -1 {
		t.Errorf("findSeparator returned %d, expected -1", pos)
	}

	// Test parseIntFromBytes
	intBytes := []byte("12345")
	n := parseIntFromBytes(intBytes)
	if n != 12345 {
		t.Errorf("parseIntFromBytes returned %d, expected 12345", n)
	}

	// Test bytesEqual
	a := []byte("test")
	b := []byte("test")
	c := []byte("different")

	if !bytesEqual(a, b) {
		t.Error("bytesEqual returned false for equal byte slices")
	}

	if bytesEqual(a, c) {
		t.Error("bytesEqual returned true for different byte slices")
	}

	if bytesEqual(a, a[:2]) {
		t.Error("bytesEqual returned true for slices of different length")
	}
}
