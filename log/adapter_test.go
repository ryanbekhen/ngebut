package log

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

// MockEvent implements IEvent for testing
type MockEvent struct {
	errCalled  bool
	err        error
	msgCalled  bool
	msg        string
	msgfCalled bool
	format     string
	args       []interface{}
}

func (e *MockEvent) Err(err error) IEvent {
	e.errCalled = true
	e.err = err
	return e
}

func (e *MockEvent) Msg(msg string) {
	e.msgCalled = true
	e.msg = msg
}

func (e *MockEvent) Msgf(format string, v ...interface{}) {
	e.msgfCalled = true
	e.format = format
	e.args = v
}

// MockLogger implements ILogger for testing
type MockLogger struct {
	debugCalled    bool
	infoCalled     bool
	warnCalled     bool
	errorCalled    bool
	fatalCalled    bool
	level          Level
	setLevelCalled bool
	getLevelCalled bool
	mockEvent      *MockEvent
}

func (l *MockLogger) Debug() IEvent {
	l.debugCalled = true
	l.mockEvent = &MockEvent{}
	return l.mockEvent
}

func (l *MockLogger) Info() IEvent {
	l.infoCalled = true
	l.mockEvent = &MockEvent{}
	return l.mockEvent
}

func (l *MockLogger) Warn() IEvent {
	l.warnCalled = true
	l.mockEvent = &MockEvent{}
	return l.mockEvent
}

func (l *MockLogger) Error() IEvent {
	l.errorCalled = true
	l.mockEvent = &MockEvent{}
	return l.mockEvent
}

func (l *MockLogger) Fatal() IEvent {
	l.fatalCalled = true
	l.mockEvent = &MockEvent{}
	return l.mockEvent
}

func (l *MockLogger) SetLevel(level Level) {
	l.setLevelCalled = true
	l.level = level
}

func (l *MockLogger) GetLevel() Level {
	l.getLevelCalled = true
	return l.level
}

// TestGlobalLogger tests the global logger functions
func TestGlobalLogger(t *testing.T) {
	// Save the original global logger to restore it later
	originalLogger := globalLogger
	defer func() {
		globalLogger = originalLogger
	}()

	// Test GetLogger with no logger set
	globalLogger = nil
	logger := GetLogger()
	assert.NotNil(t, logger, "GetLogger() returned nil when no logger was set")
	assert.Equal(t, defaultLogger, logger, "GetLogger() did not return the default logger when no logger was set")

	// Test SetLogger and GetLogger
	mockLogger := &MockLogger{}
	SetLogger(mockLogger)
	logger = GetLogger()
	assert.Equal(t, mockLogger, logger, "GetLogger() did not return the logger set with SetLogger()")
}

// TestAdapterEvent tests the AdapterEvent type
func TestAdapterEvent(t *testing.T) {
	mockEvent := &MockEvent{}
	adapterEvent := &AdapterEvent{event: mockEvent}

	// Test Err
	testErr := errors.New("test error")
	adapterEvent.Err(testErr)
	assert.True(t, mockEvent.errCalled, "AdapterEvent.Err() did not call the underlying event's Err method")
	assert.Equal(t, testErr, mockEvent.err, "AdapterEvent.Err() passed incorrect error to the underlying event")

	// Test Msg
	adapterEvent.Msg("test message")
	assert.True(t, mockEvent.msgCalled, "AdapterEvent.Msg() did not call the underlying event's Msg method")
	assert.Equal(t, "test message", mockEvent.msg, "AdapterEvent.Msg() passed incorrect message to the underlying event")

	// Test Msgf
	adapterEvent.Msgf("test %s %d", "format", 42)
	assert.True(t, mockEvent.msgfCalled, "AdapterEvent.Msgf() did not call the underlying event's Msgf method")
	assert.Equal(t, "test %s %d", mockEvent.format, "AdapterEvent.Msgf() passed incorrect format to the underlying event")
	assert.Len(t, mockEvent.args, 2, "AdapterEvent.Msgf() passed incorrect number of arguments")
	assert.Equal(t, "format", mockEvent.args[0], "AdapterEvent.Msgf() passed incorrect first argument")
	assert.Equal(t, 42, mockEvent.args[1], "AdapterEvent.Msgf() passed incorrect second argument")
}

// TestAdapterLogger tests the AdapterLogger type
func TestAdapterLogger(t *testing.T) {
	mockLogger := &MockLogger{}
	adapterLogger := NewAdapterLogger(mockLogger)

	// Test Debug
	adapterLogger.Debug()
	assert.True(t, mockLogger.debugCalled, "AdapterLogger.Debug() did not call the underlying logger's Debug method")

	// Test Info
	adapterLogger.Info()
	assert.True(t, mockLogger.infoCalled, "AdapterLogger.Info() did not call the underlying logger's Info method")

	// Test Warn
	adapterLogger.Warn()
	assert.True(t, mockLogger.warnCalled, "AdapterLogger.Warn() did not call the underlying logger's Warn method")

	// Test Error
	adapterLogger.Error()
	assert.True(t, mockLogger.errorCalled, "AdapterLogger.Error() did not call the underlying logger's Error method")

	// Test Fatal
	adapterLogger.Fatal()
	assert.True(t, mockLogger.fatalCalled, "AdapterLogger.Fatal() did not call the underlying logger's Fatal method")

	// Test SetLevel
	adapterLogger.SetLevel(WarnLevel)
	assert.True(t, mockLogger.setLevelCalled, "AdapterLogger.SetLevel() did not call the underlying logger's SetLevel method")
	assert.Equal(t, WarnLevel, mockLogger.level, "AdapterLogger.SetLevel() passed incorrect level to the underlying logger")

	// Test GetLevel
	mockLogger.level = ErrorLevel
	level := adapterLogger.GetLevel()
	assert.True(t, mockLogger.getLevelCalled, "AdapterLogger.GetLevel() did not call the underlying logger's GetLevel method")
	assert.Equal(t, ErrorLevel, level, "AdapterLogger.GetLevel() returned incorrect level")
}

// TestNewAdapterLogger tests the NewAdapterLogger function
func TestNewAdapterLogger(t *testing.T) {
	mockLogger := &MockLogger{}
	adapterLogger := NewAdapterLogger(mockLogger)
	assert.NotNil(t, adapterLogger, "NewAdapterLogger() returned nil")
	assert.Equal(t, mockLogger, adapterLogger.logger, "NewAdapterLogger() did not set the logger field correctly")
}
