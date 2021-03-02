package log

import "fmt"

// OutputFunc is a output function to log a given string argument.
type OutputFunc func(string)

// Logger is a logging compornent used in the Bucketeer SDK.
type Logger struct {
	warnFunc  OutputFunc
	errorFunc OutputFunc
}

// NewLogger creates a new Logger.
//
// Each output func is optional, and they are used in the logging method corresponding to their log level.
func NewLogger(warnFunc OutputFunc, errorFunc OutputFunc) *Logger {
	return &Logger{
		warnFunc:  warnFunc,
		errorFunc: errorFunc,
	}
}

// Warn outputs msg if warnFunc is not nil, otherwise do nothing.
func (l *Logger) Warn(msg string) {
	l.output(l.warnFunc, msg)
}

// Error outputs msg if errorFunc is not nil, otherwise do nothing.
func (l *Logger) Error(msg string) {
	l.output(l.errorFunc, msg)
}

func (l *Logger) output(f OutputFunc, msg string) {
	if f == nil {
		return
	}
	f(msg)
}

// Warnf outputs formatted message if warnFunc is not nil, otherwise do nothing.
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.outputf(l.warnFunc, format, args...)
}

// Errorf outputs formatted message if errorFunc is not nil, otherwise do nothing.
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.outputf(l.errorFunc, format, args...)
}

func (l *Logger) outputf(f OutputFunc, format string, args ...interface{}) {
	if f == nil {
		return
	}
	f(fmt.Sprintf(format, args...))
}
