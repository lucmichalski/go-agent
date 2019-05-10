// Implementation of simple logging interfaces efficient in production
// environments, aiming at being as fast as possible when disabled. The trick
// consists in changing the underlying implementation pointer with a disabled
// logger which does nothing when called. The call when disabled costs the
// underlying interface call indirection, equivalent to 2 method calls.

package plog

import (
	"fmt"
	"io"
	"log"
	"strings"
)

// LogLevel represents the log level. Higher levels include lowers.
type LogLevel int

const (
	// Disabled value.
	Disabled LogLevel = iota
	// Error logs.
	Error
	// Info to Error logs.
	Info
	// Debug to Error logs.
	Debug
)

const (
	// Error logs.
	ErrorString = "error"
	// Info to Error logs.
	InfoString = "info"
	// Debug to Error logs.
	DebugString = "debug"
)

// LogLevel type stringer.
func (l LogLevel) String() string {
	switch l {
	case Error:
		return "error"
	case Info:
		return "info"
	case Debug:
		return "debug"
	}
	return ""
}

// Logger structure wrapping logger interfaces, one per level.
type Logger struct {
	ErrorLogger
	InfoLogger
	DebugLogger

	output    io.Writer
	namespace string
	cache     struct {
		error, info, debug logger
	}
}

type ErrorLogger interface {
	Error(v ...interface{})
	Errorf(format string, v ...interface{})
	OutputSetter
}

type InfoLogger interface {
	Info(v ...interface{})
	Infof(format string, v ...interface{})
	OutputSetter
}

type DebugLogger interface {
	Debug(v ...interface{})
	Debugf(format string, v ...interface{})
	OutputSetter
}

type OutputSetter interface {
	SetOutput(output io.Writer)
}

// NewLogger returns a Logger instance wrapping one logger instance per level.
// They can thus be individually enabled or disabled.
func NewLogger(namespace string, parent *Logger) *Logger {
	logger := &Logger{
		ErrorLogger: disabledLogger{},
		InfoLogger:  disabledLogger{},
		DebugLogger: disabledLogger{},
		namespace:   namespace,
	}

	if parent == nil {
		namespace = "sqreen/" + namespace
	} else {
		namespace = parent.namespace + "/" + namespace
		logger.SetLevel(parent.getLevel())
		logger.SetOutput(parent.output)
	}
	return logger
}

// SetOutput sets the output of the logger. When `nil`, the logger is disabled
// and equivalent to `SetLevel(Disabled)`.
func (l *Logger) SetOutput(output io.Writer) {
	l.output = output
	if output == nil {
		l.SetLevel(Disabled)
		return
	}
	l.ErrorLogger.SetOutput(output)
	l.InfoLogger.SetOutput(output)
	l.DebugLogger.SetOutput(output)
}

// ParseLogLevel returns the logger level corresponding to the string
// representation `level`. The returned LogLevel is Disabled when none matches.
func ParseLogLevel(level string) LogLevel {
	switch strings.TrimSpace(strings.ToLower(level)) {
	case DebugString:
		return Debug
	case InfoString:
		return Info
	case ErrorString:
		return Error
	default:
		return Disabled
	}
}

// SetLevel changes the level of the logger to `level`, possibly disabling it
// when `Disabled` is passed.
func (l *Logger) getLevel() LogLevel {
	if _, disabled := l.DebugLogger.(disabledLogger); !disabled {
		return Debug
	}

	if _, disabled := l.InfoLogger.(disabledLogger); !disabled {
		return Info
	}

	if _, disabled := l.ErrorLogger.(disabledLogger); !disabled {
		return Error
	}

	return Disabled
}

// SetLevel changes the level of the logger to `level`, possibly disabling it
// when `Disabled` is passed.
func (l *Logger) SetLevel(level LogLevel) {
	switch level {
	case Disabled:
		l.DebugLogger = disabledLogger{}
		l.InfoLogger = disabledLogger{}
		l.ErrorLogger = disabledLogger{}
		break
	case Debug:
		l.DebugLogger = l.getLogger(Debug)
		l.InfoLogger = l.getLogger(Info)
		l.ErrorLogger = l.getLogger(Error)
		break
	case Info:
		l.DebugLogger = disabledLogger{}
		l.InfoLogger = l.getLogger(Info)
		l.ErrorLogger = l.getLogger(Error)
		break
	case Error:
		l.DebugLogger = disabledLogger{}
		l.InfoLogger = disabledLogger{}
		l.ErrorLogger = l.getLogger(Error)
		break
	}
}

// *log.Logger format prefix: UTC time, date of the day and time with microseconds.
const flags = log.LUTC | log.Ldate | log.Lmicroseconds

// Enabled logger instance.
type logger struct {
	logger *log.Logger
}

// Return the logger instance. Create a new one if first time.
func (l *Logger) getLogger(level LogLevel) logger {
	cache := l.getCachedLogger(level)
	if cache.logger == nil {
		*cache = newLogger(l.namespace, level, l.output)
	} else {
		cache.SetOutput(l.output)
	}
	return *cache
}

// Return a new logger.
func newLogger(namespace string, level LogLevel, output io.Writer) logger {
	return logger{log.New(output, namespace+":"+level.String()+": ", flags)}
}

// Return the pointer to the cache entry.
func (l *Logger) getCachedLogger(level LogLevel) *logger {
	switch level {
	case Debug:
		return &l.cache.debug
	case Info:
		return &l.cache.info
	case Error:
		return &l.cache.error
	}
	return nil
}

func (l logger) Debug(v ...interface{}) {
	l.logger.Output(3, fmt.Sprint(v...))
}

func (l logger) Debugf(format string, v ...interface{}) {
	l.logger.Output(3, fmt.Sprintf(format, v...))
}

func (l logger) Info(v ...interface{}) {
	l.logger.Output(3, fmt.Sprint(v...))
}

func (l logger) Infof(format string, v ...interface{}) {
	l.logger.Output(3, fmt.Sprintf(format, v...))
}

func (l logger) Error(v ...interface{}) {
	l.logger.Output(3, fmt.Sprint(v...))
}

func (l logger) Errorf(format string, v ...interface{}) {
	l.logger.Output(3, fmt.Sprintf(format, v...))
}

func (l logger) SetOutput(output io.Writer) {
	l.logger.SetOutput(output)
}

type disabledLogger struct {
}

func (_ disabledLogger) Error(_ ...interface{}) {
}
func (_ disabledLogger) Errorf(_ string, _ ...interface{}) {
}
func (_ disabledLogger) Info(_ ...interface{}) {
}
func (_ disabledLogger) Infof(_ string, _ ...interface{}) {
}
func (_ disabledLogger) Debug(_ ...interface{}) {
}
func (_ disabledLogger) Debugf(_ string, _ ...interface{}) {
}
func (_ disabledLogger) SetOutput(_ io.Writer) {
}
