package ngebut

type logger struct{}

func (l *logger) Debugf(format string, args ...interface{}) {}
func (l *logger) Infof(format string, args ...interface{})  {}
func (l *logger) Warnf(format string, args ...interface{})  {}
func (l *logger) Errorf(format string, args ...interface{}) {}
func (l *logger) Fatalf(format string, args ...interface{}) {}
