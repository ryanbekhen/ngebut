package log

// ColoredString returns a colored string if color is enabled
func ColoredString(s string, color string, noColor bool) string {
	if noColor {
		return s
	}
	return color + s + ColorReset
}

// ColoredLevel returns a colored level string based on the level
func ColoredLevel(level Level, noColor bool) string {
	if noColor {
		switch level {
		case DebugLevel:
			return "| DEBUG |"
		case InfoLevel:
			return "| INFO  |"
		case WarnLevel:
			return "| WARN  |"
		case ErrorLevel:
			return "| ERROR |"
		case FatalLevel:
			return "| FATAL |"
		default:
			return "| UNKN  |"
		}
	}

	// With color
	switch level {
	case DebugLevel:
		return ColorBlue + "| DEBUG |" + ColorReset
	case InfoLevel:
		return ColorGreen + "| INFO  |" + ColorReset
	case WarnLevel:
		return ColorYellow + "| WARN  |" + ColorReset
	case ErrorLevel:
		return ColorRed + "| ERROR |" + ColorReset
	case FatalLevel:
		return ColorRed + ColorBold + "| FATAL |" + ColorReset
	default:
		return "| UNKN  |"
	}
}
