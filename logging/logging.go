package logging

import (
	"os"

	"go.uber.org/zap"
)

// TODO: make a real, overrideable, global logging interface

// Debug prints messages about all internal changes in the SDK.
func Debug(msg string, keysAndValues ...interface{}) {
	if os.Getenv("STANZA_DEBUG") != "" || os.Getenv("STANZA_DEBUG_LOGGING") != "" {
		zap.S().Debugw(msg, keysAndValues...)
	}
}

// Info prints messages about the general state of the SDK.
func Info(msg string, keysAndValues ...interface{}) {
	zap.S().Infow(msg, keysAndValues...)
}

// Warn prints messages about sub-critical states of the SDK.
func Warn(msg string, keysAndValues ...interface{}) {
	zap.S().Warnw(msg, keysAndValues...)
}

// Error prints messages about exceptional states of the SDK.
func Error(err error, keysAndValues ...interface{}) {
	// TODO: Add "error", err.Error() to keysAndValues
	zap.S().Errorw(err.Error(), keysAndValues...)
}
