package sentinel

import (
	"fmt"

	"github.com/StanzaSystems/sdk-go/logging"
)

// A simple adapater which implements the Sentinel logger interface but actually logs
// using the Stanza global logger.
type loggerAdapter struct{}

func (l *loggerAdapter) Debug(msg string, keysAndValues ...interface{}) {
	logging.Debug(msg, keysAndValues...)
}

func (l *loggerAdapter) DebugEnabled() bool {
	return true // todo: lookup logging level
}

func (l *loggerAdapter) Info(msg string, keysAndValues ...interface{}) {
	logging.Info(msg, keysAndValues...)
}

func (l *loggerAdapter) InfoEnabled() bool {
	return true // todo: lookup logging level
}

func (l *loggerAdapter) Warn(msg string, keysAndValues ...interface{}) {
	logging.Warn(msg, keysAndValues...)
}

func (l *loggerAdapter) WarnEnabled() bool {
	return true // todo: lookup logging level
}

func (l *loggerAdapter) Error(err error, msg string, keysAndValues ...interface{}) {
	if msg != "" {
		logging.Error(fmt.Errorf("%v: %v", msg, err), keysAndValues...)
	} else {
		logging.Error(err, keysAndValues...)
	}
}

func (l *loggerAdapter) ErrorEnabled() bool {
	return true // todo: lookup logging level
}
