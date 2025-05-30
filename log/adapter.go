package log

// globalLogger is the global logger instance that can be replaced by the user
var globalLogger ILogger

// SetLogger sets the global logger instance.
// This allows developers to use their own logger implementation.
func SetLogger(l ILogger) {
	globalLogger = l
}

// GetLogger returns the global logger instance.
func GetLogger() ILogger {
	if globalLogger == nil {
		// Initialize with the default logger
		globalLogger = defaultLogger
	}
	return globalLogger
}

// AdapterEvent is an adapter that wraps an IEvent to make it compatible with other logger interfaces
type AdapterEvent struct {
	event IEvent
}

// Err adds an error to the event
func (e *AdapterEvent) Err(err error) IEvent {
	return e.event.Err(err)
}

// Msg logs a message
func (e *AdapterEvent) Msg(msg string) {
	e.event.Msg(msg)
}

// Msgf logs a formatted message
func (e *AdapterEvent) Msgf(format string, v ...interface{}) {
	e.event.Msgf(format, v...)
}

// AdapterLogger is an adapter that wraps an ILogger to make it compatible with other logger interfaces
type AdapterLogger struct {
	logger ILogger
}

// NewAdapterLogger creates a new AdapterLogger
func NewAdapterLogger(logger ILogger) *AdapterLogger {
	return &AdapterLogger{
		logger: logger,
	}
}

// Debug returns a debug level event
func (l *AdapterLogger) Debug() IEvent {
	return l.logger.Debug()
}

// Info returns an info level event
func (l *AdapterLogger) Info() IEvent {
	return l.logger.Info()
}

// Warn returns a warn level event
func (l *AdapterLogger) Warn() IEvent {
	return l.logger.Warn()
}

// Error returns an error level event
func (l *AdapterLogger) Error() IEvent {
	return l.logger.Error()
}

// Fatal returns a fatal level event
func (l *AdapterLogger) Fatal() IEvent {
	return l.logger.Fatal()
}

// SetLevel sets the log level
func (l *AdapterLogger) SetLevel(level Level) {
	l.logger.SetLevel(level)
}

// GetLevel returns the current log level
func (l *AdapterLogger) GetLevel() Level {
	return l.logger.GetLevel()
}
