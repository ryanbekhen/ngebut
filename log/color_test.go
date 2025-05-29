package log

import (
	"testing"
)

// TestColoredString tests the ColoredString function
func TestColoredString(t *testing.T) {
	// Test with color enabled
	result := ColoredString("test", ColorRed, false)
	expected := ColorRed + "test" + ColorReset
	if result != expected {
		t.Errorf("ColoredString(\"test\", ColorRed, false) = %q, expected %q", result, expected)
	}

	// Test with color disabled
	result = ColoredString("test", ColorRed, true)
	expected = "test"
	if result != expected {
		t.Errorf("ColoredString(\"test\", ColorRed, true) = %q, expected %q", result, expected)
	}

	// Test with different colors
	colors := []string{ColorRed, ColorGreen, ColorBlue, ColorYellow, ColorPurple, ColorCyan, ColorWhite, ColorBold}
	for _, color := range colors {
		result = ColoredString("test", color, false)
		expected = color + "test" + ColorReset
		if result != expected {
			t.Errorf("ColoredString(\"test\", %q, false) = %q, expected %q", color, result, expected)
		}
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
		if result != test.expected {
			t.Errorf("ColoredLevel(%v, true) = %q, expected %q", test.level, result, test.expected)
		}
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
		if result != test.expected {
			t.Errorf("ColoredLevel(%v, false) = %q, expected %q", test.level, result, test.expected)
		}
	}
}
