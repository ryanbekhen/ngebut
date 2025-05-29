package log

import (
	"io"
	"sync"
	"time"
)

// ANSI color codes
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorBold   = "\033[1m"
)

// ConsoleWriter is a writer that formats logs for console output
type ConsoleWriter struct {
	Out         io.Writer
	TimeFormat  string
	NoColor     bool
	FormatLevel func(level Level) string
	mu          sync.Mutex
	buf         []byte
}

// NewConsoleWriter creates a new ConsoleWriter
func NewConsoleWriter(out io.Writer) *ConsoleWriter {
	if out == nil {
		out = io.Discard
	}
	return &ConsoleWriter{
		Out:        out,
		TimeFormat: time.RFC3339,
		buf:        make([]byte, 0, 512),
	}
}

// Write implements io.Writer
func (w *ConsoleWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Reset buffer
	w.buf = w.buf[:0]

	// Parse the log line directly from bytes
	// Find the first and second " | " separator
	firstSep := findSeparator(p, 0)
	if firstSep == -1 {
		// If we can't parse it, just write it as is
		return w.Out.Write(p)
	}

	secondSep := findSeparator(p, firstSep+3)
	if secondSep == -1 {
		// If we can't parse it, just write it as is
		return w.Out.Write(p)
	}

	// Extract timestamp and level
	timestamp := p[:firstSep]
	level := p[firstSep+3 : secondSep]

	// Format timestamp
	if w.TimeFormat != "" {
		// Parse timestamp without allocations
		// Format is "2006-01-02 15:04:05"
		if len(timestamp) == 19 &&
			timestamp[4] == '-' && timestamp[7] == '-' && timestamp[10] == ' ' &&
			timestamp[13] == ':' && timestamp[16] == ':' {
			year := parseIntFromBytes(timestamp[0:4])
			month := time.Month(parseIntFromBytes(timestamp[5:7]))
			day := parseIntFromBytes(timestamp[8:10])
			hour := parseIntFromBytes(timestamp[11:13])
			min := parseIntFromBytes(timestamp[14:16])
			sec := parseIntFromBytes(timestamp[17:19])

			t := time.Date(year, month, day, hour, min, sec, 0, time.Local)

			// Format using the specified format
			if w.TimeFormat == time.RFC3339 {
				// Optimize for common case
				w.buf = t.AppendFormat(w.buf, time.RFC3339)
			} else {
				w.buf = t.AppendFormat(w.buf, w.TimeFormat)
			}
		} else {
			// Fallback to string conversion if format doesn't match
			w.buf = append(w.buf, timestamp...)
		}
	} else {
		w.buf = append(w.buf, timestamp...)
	}

	// Add timestamp with color if enabled
	if !w.NoColor {
		// Insert color codes without allocations
		// Save the current buffer
		oldBuf := w.buf

		// Reset the buffer and add color codes + original content
		w.buf = w.buf[:0]
		w.buf = append(w.buf, ColorCyan...)
		w.buf = append(w.buf, oldBuf...)
		w.buf = append(w.buf, ColorReset...)
	}

	w.buf = append(w.buf, ' ')

	// Format level
	if w.FormatLevel != nil {
		var lvl Level
		if bytesEqual(level, []byte("DEBUG")) {
			lvl = DebugLevel
		} else if bytesEqual(level, []byte("INFO")) {
			lvl = InfoLevel
		} else if bytesEqual(level, []byte("WARN")) {
			lvl = WarnLevel
		} else if bytesEqual(level, []byte("ERROR")) {
			lvl = ErrorLevel
		} else if bytesEqual(level, []byte("FATAL")) {
			lvl = FatalLevel
		}
		formattedLevel := w.FormatLevel(lvl)
		w.buf = append(w.buf, formattedLevel...)
	} else {
		w.buf = append(w.buf, level...)
	}
	w.buf = append(w.buf, ' ')

	// Add the rest of the message
	msgStart := secondSep + 3

	// Check if the message starts with "error: "
	if msgStart+7 <= len(p) &&
		p[msgStart] == 'e' && p[msgStart+1] == 'r' && p[msgStart+2] == 'r' &&
		p[msgStart+3] == 'o' && p[msgStart+4] == 'r' && p[msgStart+5] == ':' &&
		p[msgStart+6] == ' ' && !w.NoColor {

		// Find the next separator or end of message
		nextSep := findSeparator(p, msgStart)
		if nextSep == -1 {
			nextSep = len(p)
		}

		// Add error message with color
		w.buf = append(w.buf, ColorRed...)
		w.buf = append(w.buf, p[msgStart:nextSep]...)
		w.buf = append(w.buf, ColorReset...)

		// Add the rest of the message if any
		if nextSep < len(p) {
			w.buf = append(w.buf, ' ')
			w.buf = append(w.buf, p[nextSep+3:]...)
		}
	} else {
		// Add the whole message
		w.buf = append(w.buf, p[msgStart:]...)
	}

	w.buf = append(w.buf, '\n')

	// Write to output
	return w.Out.Write(w.buf)
}

// findSeparator finds the index of " | " in the byte slice starting from the given position
func findSeparator(p []byte, start int) int {
	for i := start; i <= len(p)-3; i++ {
		if p[i] == ' ' && p[i+1] == '|' && p[i+2] == ' ' {
			return i
		}
	}
	return -1
}

// parseIntFromBytes parses an integer from a byte slice without allocations
func parseIntFromBytes(b []byte) int {
	n := 0
	for _, c := range b {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return n
}

// bytesEqual compares two byte slices for equality
func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// formatLevelNoAlloc formats a level string without allocations
func formatLevelNoAlloc(buf []byte, level Level, noColor bool) []byte {
	if noColor {
		buf = append(buf, '|', ' ')
		switch level {
		case DebugLevel:
			buf = append(buf, 'D', 'E', 'B', 'U', 'G', ' ', ' ')
		case InfoLevel:
			buf = append(buf, 'I', 'N', 'F', 'O', ' ', ' ', ' ')
		case WarnLevel:
			buf = append(buf, 'W', 'A', 'R', 'N', ' ', ' ', ' ')
		case ErrorLevel:
			buf = append(buf, 'E', 'R', 'R', 'O', 'R', ' ', ' ')
		case FatalLevel:
			buf = append(buf, 'F', 'A', 'T', 'A', 'L', ' ', ' ')
		default:
			// Use level.String() but this will allocate
			levelStr := level.String()
			buf = append(buf, levelStr...)
			// Pad to 6 characters
			for i := len(levelStr); i < 6; i++ {
				buf = append(buf, ' ')
			}
			buf = append(buf, ' ')
		}
		buf = append(buf, '|')
	} else {
		// With color
		switch level {
		case DebugLevel:
			buf = append(buf, ColorBlue...)
			buf = append(buf, '|', ' ', 'D', 'E', 'B', 'U', 'G', ' ', ' ', '|')
			buf = append(buf, ColorReset...)
		case InfoLevel:
			buf = append(buf, ColorGreen...)
			buf = append(buf, '|', ' ', 'I', 'N', 'F', 'O', ' ', ' ', ' ', '|')
			buf = append(buf, ColorReset...)
		case WarnLevel:
			buf = append(buf, ColorYellow...)
			buf = append(buf, '|', ' ', 'W', 'A', 'R', 'N', ' ', ' ', ' ', '|')
			buf = append(buf, ColorReset...)
		case ErrorLevel:
			buf = append(buf, ColorRed...)
			buf = append(buf, '|', ' ', 'E', 'R', 'R', 'O', 'R', ' ', ' ', '|')
			buf = append(buf, ColorReset...)
		case FatalLevel:
			buf = append(buf, ColorRed...)
			buf = append(buf, ColorBold...)
			buf = append(buf, '|', ' ', 'F', 'A', 'T', 'A', 'L', ' ', ' ', '|')
			buf = append(buf, ColorReset...)
		default:
			// Use level.String() but this will allocate
			buf = append(buf, '|', ' ')
			levelStr := level.String()
			buf = append(buf, levelStr...)
			// Pad to 6 characters
			for i := len(levelStr); i < 6; i++ {
				buf = append(buf, ' ')
			}
			buf = append(buf, ' ', '|')
		}
	}
	return buf
}

// DefaultConsoleWriter returns a ConsoleWriter with default settings
func DefaultConsoleWriter() *ConsoleWriter {
	w := NewConsoleWriter(nil)

	// Use a buffer to avoid allocations in FormatLevel
	var levelBuf []byte

	w.FormatLevel = func(level Level) string {
		// Reset the buffer
		levelBuf = levelBuf[:0]

		// Format the level without allocations
		levelBuf = formatLevelNoAlloc(levelBuf, level, w.NoColor)

		// Convert to string (this allocates, but it's unavoidable due to the interface)
		return string(levelBuf)
	}

	return w
}
