package log

import (
	"errors"
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
	if logger == nil {
		t.Error("GetLogger() returned nil when no logger was set")
	}
	if logger != defaultLogger {
		t.Error("GetLogger() did not return the default logger when no logger was set")
	}

	// Test SetLogger and GetLogger
	mockLogger := &MockLogger{}
	SetLogger(mockLogger)
	logger = GetLogger()
	if logger != mockLogger {
		t.Error("GetLogger() did not return the logger set with SetLogger()")
	}
}

// TestAdapterEvent tests the AdapterEvent type
func TestAdapterEvent(t *testing.T) {
	mockEvent := &MockEvent{}
	adapterEvent := &AdapterEvent{event: mockEvent}

	// Test Err
	testErr := errors.New("test error")
	adapterEvent.Err(testErr)
	if !mockEvent.errCalled {
		t.Error("AdapterEvent.Err() did not call the underlying event's Err method")
	}
	if mockEvent.err != testErr {
		t.Errorf("AdapterEvent.Err() passed %v to the underlying event, expected %v", mockEvent.err, testErr)
	}

	// Test Msg
	adapterEvent.Msg("test message")
	if !mockEvent.msgCalled {
		t.Error("AdapterEvent.Msg() did not call the underlying event's Msg method")
	}
	if mockEvent.msg != "test message" {
		t.Errorf("AdapterEvent.Msg() passed %s to the underlying event, expected 'test message'", mockEvent.msg)
	}

	// Test Msgf
	adapterEvent.Msgf("test %s %d", "format", 42)
	if !mockEvent.msgfCalled {
		t.Error("AdapterEvent.Msgf() did not call the underlying event's Msgf method")
	}
	if mockEvent.format != "test %s %d" {
		t.Errorf("AdapterEvent.Msgf() passed format %s to the underlying event, expected 'test %%s %%d'", mockEvent.format)
	}
	if len(mockEvent.args) != 2 || mockEvent.args[0] != "format" || mockEvent.args[1] != 42 {
		t.Errorf("AdapterEvent.Msgf() passed args %v to the underlying event, expected ['format', 42]", mockEvent.args)
	}
}

// TestAdapterLogger tests the AdapterLogger type
func TestAdapterLogger(t *testing.T) {
	mockLogger := &MockLogger{}
	adapterLogger := NewAdapterLogger(mockLogger)

	// Test Debug
	adapterLogger.Debug()
	if !mockLogger.debugCalled {
		t.Error("AdapterLogger.Debug() did not call the underlying logger's Debug method")
	}

	// Test Info
	adapterLogger.Info()
	if !mockLogger.infoCalled {
		t.Error("AdapterLogger.Info() did not call the underlying logger's Info method")
	}

	// Test Warn
	adapterLogger.Warn()
	if !mockLogger.warnCalled {
		t.Error("AdapterLogger.Warn() did not call the underlying logger's Warn method")
	}

	// Test Error
	adapterLogger.Error()
	if !mockLogger.errorCalled {
		t.Error("AdapterLogger.Error() did not call the underlying logger's Error method")
	}

	// Test Fatal
	adapterLogger.Fatal()
	if !mockLogger.fatalCalled {
		t.Error("AdapterLogger.Fatal() did not call the underlying logger's Fatal method")
	}

	// Test SetLevel
	adapterLogger.SetLevel(WarnLevel)
	if !mockLogger.setLevelCalled {
		t.Error("AdapterLogger.SetLevel() did not call the underlying logger's SetLevel method")
	}
	if mockLogger.level != WarnLevel {
		t.Errorf("AdapterLogger.SetLevel() passed %v to the underlying logger, expected %v", mockLogger.level, WarnLevel)
	}

	// Test GetLevel
	mockLogger.level = ErrorLevel
	level := adapterLogger.GetLevel()
	if !mockLogger.getLevelCalled {
		t.Error("AdapterLogger.GetLevel() did not call the underlying logger's GetLevel method")
	}
	if level != ErrorLevel {
		t.Errorf("AdapterLogger.GetLevel() returned %v, expected %v", level, ErrorLevel)
	}
}

// TestNewAdapterLogger tests the NewAdapterLogger function
func TestNewAdapterLogger(t *testing.T) {
	mockLogger := &MockLogger{}
	adapterLogger := NewAdapterLogger(mockLogger)
	if adapterLogger == nil {
		t.Error("NewAdapterLogger() returned nil")
	}
	if adapterLogger.logger != mockLogger {
		t.Error("NewAdapterLogger() did not set the logger field correctly")
	}
}
