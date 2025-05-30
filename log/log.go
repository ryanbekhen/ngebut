package log

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// ILogger is the interface that wraps the basic logging methods.
type ILogger interface {
	// Debug returns a debug level event
	Debug() IEvent
	// Info returns an info level event
	Info() IEvent
	// Warn returns a warn level event
	Warn() IEvent
	// Error returns an error level event
	Error() IEvent
	// Fatal returns a fatal level event
	Fatal() IEvent
	// SetLevel sets the log level
	SetLevel(level Level)
	// GetLevel returns the current log level
	GetLevel() Level
}

// IEvent is the interface that wraps the basic event methods.
type IEvent interface {
	// Err adds an error to the event
	Err(err error) IEvent
	// Msg logs a message
	Msg(msg string)
	// Msgf logs a formatted message
	Msgf(format string, v ...interface{})
}

// LoggerConfig represents the configuration for a logger.
type LoggerConfig struct {
	// Writer is the output writer
	Writer io.Writer
	// Level is the log level
	Level Level
	// TimeFormat is the format for timestamps
	TimeFormat string
	// NoColor disables colored output
	NoColor bool
}

// DefaultLoggerConfig returns the default configuration for a logger.
func DefaultLoggerConfig() LoggerConfig {
	return LoggerConfig{
		Writer:     nil, // Will be set to os.Stdout in New
		Level:      InfoLevel,
		TimeFormat: "2006-01-02 15:04:05",
		NoColor:    false,
	}
}

// Level represents the log level
type Level int8

const (
	// DebugLevel defines debug log level
	DebugLevel Level = iota
	// InfoLevel defines info log level
	InfoLevel
	// WarnLevel defines warn log level
	WarnLevel
	// ErrorLevel defines error log level
	ErrorLevel
	// FatalLevel defines fatal log level
	FatalLevel
)

var levelNames = map[Level]string{
	DebugLevel: "DEBUG",
	InfoLevel:  "INFO",
	WarnLevel:  "WARN",
	ErrorLevel: "ERROR",
	FatalLevel: "FATAL",
}

// String returns the string representation of the log level
func (l Level) String() string {
	if name, ok := levelNames[l]; ok {
		return name
	}
	return fmt.Sprintf("LEVEL(%d)", l)
}

// Logger represents a logger instance
type Logger struct {
	writer     io.Writer
	level      Level
	mu         sync.Mutex
	buf        []byte
	timeFormat string
	noColor    bool
}

// SetLevel sets the log level
func (l *Logger) SetLevel(level Level) {
	l.level = level
}

// GetLevel returns the current log level
func (l *Logger) GetLevel() Level {
	return l.level
}

// Event represents a log event
type Event struct {
	logger *Logger
	level  Level
	err    error
}

// New creates a new logger with the given writer and level
func New(writer io.Writer, level Level) *Logger {
	if writer == nil {
		writer = os.Stdout
	}
	return &Logger{
		writer:     writer,
		level:      level,
		buf:        make([]byte, 0, 512),
		timeFormat: "2006-01-02 15:04:05",
		noColor:    false,
	}
}

// NewWithConfig creates a new logger with the given configuration
func NewWithConfig(config LoggerConfig) *Logger {
	if config.Writer == nil {
		config.Writer = os.Stdout
	}
	return &Logger{
		writer:     config.Writer,
		level:      config.Level,
		buf:        make([]byte, 0, 512),
		timeFormat: config.TimeFormat,
		noColor:    config.NoColor,
	}
}

// Debug returns a debug level event
func (l *Logger) Debug() IEvent {
	if l.level > DebugLevel {
		return nil
	}
	return &Event{logger: l, level: DebugLevel}
}

// Info returns an info level event
func (l *Logger) Info() IEvent {
	if l.level > InfoLevel {
		return nil
	}
	return &Event{logger: l, level: InfoLevel}
}

// Warn returns a warn level event
func (l *Logger) Warn() IEvent {
	if l.level > WarnLevel {
		return nil
	}
	return &Event{logger: l, level: WarnLevel}
}

// Error returns an error level event
func (l *Logger) Error() IEvent {
	if l.level > ErrorLevel {
		return nil
	}
	return &Event{logger: l, level: ErrorLevel}
}

// Fatal returns a fatal level event
func (l *Logger) Fatal() IEvent {
	return &Event{logger: l, level: FatalLevel}
}

// Err adds an error to the event
func (e *Event) Err(err error) IEvent {
	if e == nil {
		return nil
	}
	e.err = err
	return e
}

// Msg logs a message
func (e *Event) Msg(msg string) {
	if e == nil {
		return
	}

	l := e.logger
	l.mu.Lock()
	defer l.mu.Unlock()

	// Reset buffer
	l.buf = l.buf[:0]

	// Add timestamp - use a pre-allocated buffer for formatting
	now := time.Now()
	year, month, day := now.Date()
	hour, min, sec := now.Clock()

	// Format: 2006-01-02 15:04:05
	l.buf = append(l.buf, '2', '0')
	if year >= 1000 {
		l.buf = append(l.buf, byte('0'+year/1000%10), byte('0'+year/100%10), byte('0'+year/10%10), byte('0'+year%10))
	} else {
		l.buf = append(l.buf, byte('0'+year/100%10), byte('0'+year/10%10), byte('0'+year%10))
	}
	l.buf = append(l.buf, '-')
	if month < 10 {
		l.buf = append(l.buf, '0', byte('0'+month))
	} else {
		l.buf = append(l.buf, byte('0'+month/10), byte('0'+month%10))
	}
	l.buf = append(l.buf, '-')
	if day < 10 {
		l.buf = append(l.buf, '0', byte('0'+day))
	} else {
		l.buf = append(l.buf, byte('0'+day/10), byte('0'+day%10))
	}
	l.buf = append(l.buf, ' ')
	if hour < 10 {
		l.buf = append(l.buf, '0', byte('0'+hour))
	} else {
		l.buf = append(l.buf, byte('0'+hour/10), byte('0'+hour%10))
	}
	l.buf = append(l.buf, ':')
	if min < 10 {
		l.buf = append(l.buf, '0', byte('0'+min))
	} else {
		l.buf = append(l.buf, byte('0'+min/10), byte('0'+min%10))
	}
	l.buf = append(l.buf, ':')
	if sec < 10 {
		l.buf = append(l.buf, '0', byte('0'+sec))
	} else {
		l.buf = append(l.buf, byte('0'+sec/10), byte('0'+sec%10))
	}

	l.buf = append(l.buf, " | "...)

	// Add level
	l.buf = append(l.buf, e.level.String()...)
	l.buf = append(l.buf, " | "...)

	// We don't add the error here anymore, it will be added by the accesslog middleware
	// This prevents duplicate error messages in the log

	// Add message
	l.buf = append(l.buf, msg...)

	// Write to output
	l.writer.Write(l.buf)
}

// Msgf logs a formatted message
func (e *Event) Msgf(format string, v ...interface{}) {
	if e == nil {
		return
	}

	// Get the logger and lock it
	l := e.logger
	l.mu.Lock()
	defer l.mu.Unlock()

	// Reset buffer
	l.buf = l.buf[:0]

	// Add timestamp - use a pre-allocated buffer for formatting
	now := time.Now()
	year, month, day := now.Date()
	hour, min, sec := now.Clock()

	// Format: 2006-01-02 15:04:05
	l.buf = append(l.buf, '2', '0')
	if year >= 1000 {
		l.buf = append(l.buf, byte('0'+year/1000%10), byte('0'+year/100%10), byte('0'+year/10%10), byte('0'+year%10))
	} else {
		l.buf = append(l.buf, byte('0'+year/100%10), byte('0'+year/10%10), byte('0'+year%10))
	}
	l.buf = append(l.buf, '-')
	if month < 10 {
		l.buf = append(l.buf, '0', byte('0'+month))
	} else {
		l.buf = append(l.buf, byte('0'+month/10), byte('0'+month%10))
	}
	l.buf = append(l.buf, '-')
	if day < 10 {
		l.buf = append(l.buf, '0', byte('0'+day))
	} else {
		l.buf = append(l.buf, byte('0'+day/10), byte('0'+day%10))
	}
	l.buf = append(l.buf, ' ')
	if hour < 10 {
		l.buf = append(l.buf, '0', byte('0'+hour))
	} else {
		l.buf = append(l.buf, byte('0'+hour/10), byte('0'+hour%10))
	}
	l.buf = append(l.buf, ':')
	if min < 10 {
		l.buf = append(l.buf, '0', byte('0'+min))
	} else {
		l.buf = append(l.buf, byte('0'+min/10), byte('0'+min%10))
	}
	l.buf = append(l.buf, ':')
	if sec < 10 {
		l.buf = append(l.buf, '0', byte('0'+sec))
	} else {
		l.buf = append(l.buf, byte('0'+sec/10), byte('0'+sec%10))
	}

	l.buf = append(l.buf, " | "...)

	// Add level
	l.buf = append(l.buf, e.level.String()...)
	l.buf = append(l.buf, " | "...)

	// We don't add the error here anymore, it will be added by the accesslog middleware
	// This prevents duplicate error messages in the log

	// Format the message directly into the buffer
	// This is a simplified version that handles %s, %d, %v
	// For more complex formatting, you would need to implement more format specifiers
	var argIndex int
	for i := 0; i < len(format); i++ {
		if format[i] == '%' && i+1 < len(format) {
			if argIndex >= len(v) {
				// Not enough arguments, just append the % and continue
				l.buf = append(l.buf, '%')
				continue
			}

			switch format[i+1] {
			case 's':
				// String
				if str, ok := v[argIndex].(string); ok {
					l.buf = append(l.buf, str...)
				} else {
					l.buf = append(l.buf, fmt.Sprint(v[argIndex])...)
				}
				argIndex++
				i++ // Skip the format specifier
			case 'd':
				// Integer
				if n, ok := v[argIndex].(int); ok {
					l.buf = appendInt(l.buf, int64(n))
				} else if n, ok := v[argIndex].(int64); ok {
					l.buf = appendInt(l.buf, n)
				} else {
					l.buf = append(l.buf, fmt.Sprint(v[argIndex])...)
				}
				argIndex++
				i++ // Skip the format specifier
			case 'v':
				// Any value
				l.buf = append(l.buf, fmt.Sprint(v[argIndex])...)
				argIndex++
				i++ // Skip the format specifier
			default:
				// Unknown format specifier, just append it
				l.buf = append(l.buf, '%', format[i+1])
				i++ // Skip the format specifier
			}
		} else {
			// Regular character, just append it
			l.buf = append(l.buf, format[i])
		}
	}

	// Write to output
	l.writer.Write(l.buf)
}

// appendInt appends an integer to the buffer without allocations
func appendInt(buf []byte, n int64) []byte {
	// Handle the special case of minimum int64 value
	if n == -9223372036854775808 {
		return append(buf, "-9223372036854775808"...)
	}

	if n < 0 {
		buf = append(buf, '-')
		n = -n
	}

	// Handle 0 specially
	if n == 0 {
		return append(buf, '0')
	}

	// Convert the number to a string in reverse order
	var temp [20]byte // Max int64 is 19 digits
	i := len(temp)
	for n > 0 {
		i--
		temp[i] = byte('0' + n%10)
		n /= 10
	}

	// Append the digits in the correct order
	return append(buf, temp[i:]...)
}

// Default logger
var defaultLogger = New(os.Stdout, InfoLevel)

// Debug returns a debug level event from the default logger
func Debug() *Event {
	if event := defaultLogger.Debug(); event != nil {
		return event.(*Event)
	}
	return nil
}

// Info returns an info level event from the default logger
func Info() *Event {
	if event := defaultLogger.Info(); event != nil {
		return event.(*Event)
	}
	return nil
}

// Warn returns a warn level event from the default logger
func Warn() *Event {
	if event := defaultLogger.Warn(); event != nil {
		return event.(*Event)
	}
	return nil
}

// Error returns an error level event from the default logger
func Error() *Event {
	if event := defaultLogger.Error(); event != nil {
		return event.(*Event)
	}
	return nil
}

// Fatal returns a fatal level event from the default logger
func Fatal() *Event {
	if event := defaultLogger.Fatal(); event != nil {
		return event.(*Event)
	}
	return nil
}

// SetLevel sets the log level for the default logger
func SetLevel(level Level) {
	defaultLogger.level = level
}

// SetOutput sets the output writer for the default logger
func SetOutput(w io.Writer) {
	defaultLogger.writer = w
}
