package log

import "fmt"

type outputFunc func(string)

// Logger is a logging compornent used in the Bucketeer SDK.
type Logger struct {
	warnFunc  outputFunc
	errorFunc outputFunc
}

// LoggerConfig is the config for Logger.
//
// Each output func is optional, and they are used in the logging method corresponding to their log level.
type LoggerConfig struct {
	// WarnFunc is a function to output warn log
	WarnFunc outputFunc

	// ErrorFunc is a function to output error log
	ErrorFunc outputFunc
}

// NewLogger creates a new Logger.
func NewLogger(conf *LoggerConfig) *Logger {
	return &Logger{
		warnFunc:  conf.WarnFunc,
		errorFunc: conf.ErrorFunc,
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

func (l *Logger) output(f outputFunc, msg string) {
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

func (l *Logger) outputf(f outputFunc, format string, args ...interface{}) {
	if f == nil {
		return
	}
	f(fmt.Sprintf(format, args...))
}
