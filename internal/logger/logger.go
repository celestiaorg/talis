package logger

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

var talisLog = logrus.New()

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
