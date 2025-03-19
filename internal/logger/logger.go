package logger

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

var talisLog = logrus.New()

// Initialize sets up the logger with the appropriate configuration
// and log level from environment variables
func Initialize() {
	var log = logrus.New()

	log.SetFormatter(&logrus.JSONFormatter{})
	configureLogLevel()

}

func configureLogLevel() {
	talisLog.SetLevel(logrus.InfoLevel)

	levelStr := os.Getenv("LOG_LEVEL")
	if levelStr == "" {
		talisLog.Warnf("Invalid log level '%s', defaulting to 'info'", levelStr)
		return
	}

	level, err := logrus.ParseLevel(strings.ToLower(levelStr))
	if err != nil {
		// If parsing fails, log a warning and keep the default
		talisLog.Warnf("Invalid log level '%s', defaulting to 'info'", levelStr)
		return
	}

	talisLog.SetLevel(level)
	talisLog.Infof("Log level set to '%s'", level)

}

// GetLogger returns the configured logger instance
func GetLogger() *logrus.Logger {
	return talisLog
}

//Logrus has seven logging levels: Trace, Debug, Info, Warning, Error, Fatal and Panic.

// Trace logs a message at the Trace level
func Trace(args ...interface{}) {
	talisLog.Trace(args...)
}

// Debug logs a message at the debug level
func Debug(args ...interface{}) {
	talisLog.Debug(args...)
}

// Info logs a message at the Info level
func Info(args ...interface{}) {
	talisLog.Info(args...)
}

// Warn logs a message at the Warn level
func Warn(args ...interface{}) {
	talisLog.Warn(args...)
}

// Error logs a message at the Error level
func Error(args ...interface{}) {
	talisLog.Error(args...)
}

// Fatal logs a message at the Fatal level
func Fatal(args ...interface{}) {
	talisLog.Fatal(args...)
}

// Panic logs a message at the Panic level
func Panic(args ...interface{}) {
	talisLog.Panic(args...)
}

// Formatted Logs
//

// Tracef logs a message at the TraceF level
func Tracef(format string, args ...interface{}) {
	talisLog.Tracef(format, args...)
}

// Debugf logs a message at the Debugf level
func Debugf(format string, args ...interface{}) {
	talisLog.Debugf(format, args...)
}

// Infof logs a message at the Infof level
func Infof(format string, args ...interface{}) {
	talisLog.Infof(format, args...)
}

// Warnf logs a message at the Warnf level
func Warnf(format string, args ...interface{}) {
	talisLog.Warnf(format, args...)
}

// Errorf logs a message at the Errorf level
func Errorf(format string, args ...interface{}) {
	talisLog.Errorf(format, args...)
}

// Fatalf logs a message at the Fatalf level
func Fatalf(format string, args ...interface{}) {
	talisLog.Fatalf(format, args...)
}

// Panicf logs a message at the Panicf level
func Panicf(format string, args ...interface{}) {
	talisLog.Panicf(format, args...)
}
