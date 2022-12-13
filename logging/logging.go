package logging

import (
	"go.uber.org/zap"
)

// TODO: make an overrideable interface

// Debug prints messages about all internal changes in the SDK.
func Debug(msg string, keysAndValues ...interface{}) {
	zap.S().Debugw(msg, keysAndValues...)
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
func Error(err error, msg string, keysAndValues ...interface{}) {
	// TODO: Add "error", err.Error() to keysAndValues
	zap.S().Errorw(msg, keysAndValues...)
}
