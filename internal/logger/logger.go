package logger

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

var log = logrus.New()

// Initialize sets up the logger with the appropriate configuration
// and log level from environment variables
func InitializeAndConfigure() {

	// Set default formatter
	log.SetFormatter(&logrus.JSONFormatter{})

	// Set output to stdout
	log.SetOutput(os.Stdout)

	// Try to read log level from environment
	configureLogLevel()

}

func configureLogLevel() {
	log.SetLevel(logrus.InfoLevel)

	levelStr := os.Getenv("LOG_LEVEL")
	if levelStr == "" {
		log.Warnf("Invalid log level '%s', defaulting to 'info'", levelStr)
		// Defaults to InfoLevel set above
		return
	}

	// Parse the log level string
	level, err := logrus.ParseLevel(strings.ToLower(levelStr))
	if err != nil {
		// If parsing fails, log a warning and keep the default
		log.Warnf("Invalid log level '%s', defaulting to 'info'", levelStr)
		return
	}

	log.SetLevel(level)
	log.Infof("Log level set to '%s'", level)

}

//Logrus has seven logging levels: Trace, Debug, Info, Warning, Error, Fatal and Panic.

// Trace logs a message at the Trace level
func Trace(args ...interface{}) {
	log.Trace(args...)
}

// Debug logs a message at the debug level
func Debug(args ...interface{}) {
	log.Debug(args...)
}

// Info logs a message at the Info level
func Info(args ...interface{}) {
	log.Info(args...)
}

// Warn logs a message at the Warn level
func Warn(args ...interface{}) {
	log.Warn(args...)
}

// Error logs a message at the Error level
func Error(args ...interface{}) {
	log.Error(args...)
}

// Fatal logs a message at the Fatal level
func Fatal(args ...interface{}) {
	log.Fatal(args...)
}

// Panic logs a message at the Panic level
func Panic(args ...interface{}) {
	log.Panic(args...)
}

// Formatted Logs
//

// Tracef logs a message at the TraceF level
func Tracef(format string, args ...interface{}) {
	log.Tracef(format, args...)
}

// Debugf logs a message at the Debugf level
func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

// Infof logs a message at the Infof level
func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

// Warnf logs a message at the Warnf level
func Warnf(format string, args ...interface{}) {
	log.Warnf(format, args...)
}

// Errorf logs a message at the Errorf level
func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

// Fatalf logs a message at the Fatalf level
func Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}

// Panicf logs a message at the Panicf level
func Panicf(format string, args ...interface{}) {
	log.Panicf(format, args...)
}

// Log levels with fields

// InfoWithFields logs a message at the info level with additional fields
func InfoWithFields(msg string, fields map[string]interface{}) {
	log.WithFields(logrus.Fields(fields)).Info(msg)
}

// DebugWithFields logs a message at the debug level with additional fields
func DebugWithFields(msg string, fields map[string]interface{}) {
	log.WithFields(logrus.Fields(fields)).Debug(msg)
}

// WarnWithFields logs a message at the warn level with additional fields
func WarnWithFields(msg string, fields map[string]interface{}) {
	log.WithFields(logrus.Fields(fields)).Warn(msg)
}

// ErrorWithFields logs a message at the error level with additional fields
func ErrorWithFields(msg string, fields map[string]interface{}) {
	log.WithFields(logrus.Fields(fields)).Error(msg)
}

// FatalWithFields logs a message at the fatal level with additional fields
func FatalWithFields(msg string, fields map[string]interface{}) {
	log.WithFields(logrus.Fields(fields)).Fatal(msg)
}

// PanicWithFields logs a message at the panic level with additional fields
func PanicWithFields(msg string, fields map[string]interface{}) {
	log.WithFields(logrus.Fields(fields)).Panic(msg)
}

// TraceWithFields logs a message at the trace level with additional fields
func TraceWithFields(msg string, fields map[string]interface{}) {
	log.WithFields(logrus.Fields(fields)).Trace(msg)
}
