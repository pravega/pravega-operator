package logging

import (
	"github.com/sirupsen/logrus"
	"sync"
)

const loggerNameField = "name"

var globalLogLevel = logrus.InfoLevel
var globalLogLevelLock = sync.Mutex{}

// Create a new logger with the given name
func GetLogger(name string) *logrus.Logger {
	logger := logrus.New()

	logger.Level = globalLogLevel

	logger.Formatter = &logrus.TextFormatter{
		FullTimestamp: true,
	}

	if name != "" {
		return logger.WithField(loggerNameField, name).Logger
	} else {
		return logger.WithFields(make(logrus.Fields)).Logger
	}
}

// Set the log level. Note: this ONLY affects loggers created using the GetLogger function AFTER this function has been
// called. Therefore, you need to call this as early in the life of your CLI app as possible!
func SetGlobalLogLevel(level logrus.Level) {
	// We need to lock here as this function may be called from multiple threads concurrently (e.g. especially at
	// test time)
	defer globalLogLevelLock.Unlock()
	globalLogLevelLock.Lock()

	globalLogLevel = level
}
