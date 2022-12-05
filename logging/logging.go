package logging

import (
	"log"
	"os"
	"sync"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
)

// globalLogger is the logging interface used within the SDK.
//
// The default logger uses stdr which is backed by the standard `log.Logger`
// interface. This logger will only show messages at the Error Level.
var globalLogger logr.Logger = stdr.New(log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile))
var globalLoggerLock = &sync.RWMutex{}

// SetLogger overrides the globalLogger with l.
//
// To see Info messages use a logger with `l.V(1).Enabled() == true`
// To see Debug messages use a logger with `l.V(5).Enabled() == true`.
func SetLogger(l logr.Logger) {
	globalLoggerLock.Lock()
	defer globalLoggerLock.Unlock()
	globalLogger = l
}

// Info prints messages about the general state of the SDK.
func Info(msg string, keysAndValues ...interface{}) {
	globalLoggerLock.RLock()
	defer globalLoggerLock.RUnlock()
	globalLogger.V(0).Info(msg, keysAndValues...)
}

// Error prints messages about exceptional states of the SDK.
func Error(err error, msg string, keysAndValues ...interface{}) {
	globalLoggerLock.RLock()
	defer globalLoggerLock.RUnlock()
	globalLogger.Error(err, msg, keysAndValues...)
}

// Debug prints messages about all internal changes in the SDK.
func Debug(msg string, keysAndValues ...interface{}) {
	globalLoggerLock.RLock()
	defer globalLoggerLock.RUnlock()
	globalLogger.V(1).Info(msg, keysAndValues...)
}
