//go:generate mockgen -source=$GOFILE -package=$GOPACKAGE -destination=../../../test/mock/$GOPACKAGE/$GOFILE
package log

import (
	"io/ioutil"
	stdlog "log"
	"os"
)

// BaseLogger is a generic logger interface with no level mechanism.
//
// Since BaseLogger's methods are a subset of log.Logger in the standard library,
// you can use log.New() to create a BaseLogger, or you can implement a custom BaseLogger.
type BaseLogger interface {
	// Print outputs a message.
	//
	// Print is equivalent to log.Logger.Print in the standard library.
	Print(values ...interface{})

	// Printf outputs a message, applying a format string.
	//
	// Printf is equivalent to log.Logger.Printf in the standard library.
	Printf(format string, values ...interface{})
}

var (
	flags = stdlog.LstdFlags

	defaultDebugLogger = stdlog.New(os.Stdout, "[DEBUG] ", flags)
	discardDebugLogger = stdlog.New(ioutil.Discard, "[DEBUG] ", flags)

	// DefaultErrorLogger is a default logger for Bucketeer SDK error logs.
	// For example, DefaultErrorLogger outputs,
	//   [ERROR] 2021/01/01 10:00:00 message
	//
	// DefaultErrorLoger implements BaseLogger interface.
	DefaultErrorLogger = stdlog.New(os.Stderr, "[ERROR] ", flags)

	// DiscardErrorLogger discards all Bucketeer SDK error logs.
	//
	// DiscardErrorLoger implements BaseLogger interface.
	DiscardErrorLogger = stdlog.New(ioutil.Discard, "[ERROR] ", flags)
)

// Loggers is a logging compornent used in the Bucketeer SDK.
//
// Debug logs are for Bucketeer SDK developers.
// Error logs are for Bucketeer SDK users.
type Loggers struct {
	debugLogger BaseLogger
	errorLogger BaseLogger
}

// LoggersConfig is the config for Loggers.
type LoggersConfig struct {
	// EnableDebugLog enables debug logs if true.
	EnableDebugLog bool

	// ErrorLogger is used to output error logs.
	ErrorLogger BaseLogger
}

// NewLoggers creates a new Loggers.
func NewLoggers(conf *LoggersConfig) *Loggers {
	dbgLogger := discardDebugLogger
	if conf.EnableDebugLog {
		dbgLogger = defaultDebugLogger
	}
	errLogger := conf.ErrorLogger
	return &Loggers{
		debugLogger: dbgLogger,
		errorLogger: errLogger,
	}
}

// Debug outputs a debug log.
func (l *Loggers) Debug(values ...interface{}) {
	l.debugLogger.Print(values...)
}

// Debugf outputs a formatted debug log.
func (l *Loggers) Debugf(format string, values ...interface{}) {
	l.debugLogger.Printf(format, values...)
}

// Error outputs a error log.
func (l *Loggers) Error(values ...interface{}) {
	l.errorLogger.Print(values...)
}

// Errorf outputs a formatted error log.
func (l *Loggers) Errorf(format string, values ...interface{}) {
	l.errorLogger.Printf(format, values...)
}
