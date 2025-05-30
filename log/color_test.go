package log

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestColoredString tests the ColoredString function
func TestColoredString(t *testing.T) {
	// Test with color enabled
	result := ColoredString("test", ColorRed, false)
	expected := ColorRed + "test" + ColorReset
	assert.Equal(t, expected, result, "ColoredString dengan warna merah seharusnya menambahkan kode warna")

	// Test with color disabled
	result = ColoredString("test", ColorRed, true)
	expected = "test"
	assert.Equal(t, expected, result, "ColoredString dengan warna dinonaktifkan seharusnya tidak menambahkan kode warna")

	// Test with different colors
	colors := []string{ColorRed, ColorGreen, ColorBlue, ColorYellow, ColorPurple, ColorCyan, ColorWhite, ColorBold}
	for _, color := range colors {
		result = ColoredString("test", color, false)
		expected = color + "test" + ColorReset
		assert.Equal(t, expected, result, "ColoredString dengan warna %s seharusnya bekerja dengan benar", color)
	}
}

// TestColoredLevel tests the ColoredLevel function
func TestColoredLevel(t *testing.T) {
	// Test with color disabled
	levels := []struct {
		level    Level
		expected string
	}{
		{DebugLevel, "| DEBUG |"},
		{InfoLevel, "| INFO  |"},
		{WarnLevel, "| WARN  |"},
		{ErrorLevel, "| ERROR |"},
		{FatalLevel, "| FATAL |"},
		{Level(99), "| UNKN  |"}, // Unknown level
	}

	for _, test := range levels {
		result := ColoredLevel(test.level, true)
		assert.Equal(t, test.expected, result, "ColoredLevel(%v, true) seharusnya mengembalikan string level tanpa warna", test.level)
	}

	// Test with color enabled
	coloredLevels := []struct {
		level    Level
		expected string
	}{
		{DebugLevel, ColorBlue + "| DEBUG |" + ColorReset},
		{InfoLevel, ColorGreen + "| INFO  |" + ColorReset},
		{WarnLevel, ColorYellow + "| WARN  |" + ColorReset},
		{ErrorLevel, ColorRed + "| ERROR |" + ColorReset},
		{FatalLevel, ColorRed + ColorBold + "| FATAL |" + ColorReset},
		{Level(99), "| UNKN  |"}, // Unknown level
	}

	for _, test := range coloredLevels {
		result := ColoredLevel(test.level, false)
		assert.Equal(t, test.expected, result, "ColoredLevel(%v, false) seharusnya mengembalikan string level dengan warna", test.level)
	}
}
