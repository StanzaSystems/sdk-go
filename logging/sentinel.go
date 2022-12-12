package logging

import (
	"github.com/alibaba/sentinel-golang/logging"
)

// A simple adapator which implements the Sentinel logger interface but actually logs
// using the Stanza global logger.
type SentinelAdaptor struct{}

func (l *SentinelAdaptor) Debug(msg string, keysAndValues ...interface{}) {
	s := logging.AssembleMsg(logging.GlobalCallerDepth, "DEBUG", msg, nil, keysAndValues...)
	Debug(s)
}

func (l *SentinelAdaptor) DebugEnabled() bool {
	return true // TODO: make this real based on verbosity level?
}

func (l *SentinelAdaptor) Info(msg string, keysAndValues ...interface{}) {
	s := logging.AssembleMsg(logging.GlobalCallerDepth, "INFO", msg, nil, keysAndValues...)
	Info(s)
}

func (l *SentinelAdaptor) InfoEnabled() bool {
	return true // TODO: make this real based on verbosity level?
}

func (l *SentinelAdaptor) Warn(msg string, keysAndValues ...interface{}) {
	s := logging.AssembleMsg(logging.GlobalCallerDepth, "WARNING", msg, nil, keysAndValues...)
	Info(s) // TODO: add a proper warning level?
}

func (l *SentinelAdaptor) WarnEnabled() bool {
	return true // TODO: make this real based on verbosity level?
}

func (l *SentinelAdaptor) Error(err error, msg string, keysAndValues ...interface{}) {
	s := logging.AssembleMsg(logging.GlobalCallerDepth, "ERROR", msg, nil, keysAndValues...)
	Error(err, s)
}

func (l *SentinelAdaptor) ErrorEnabled() bool {
	return true
}
