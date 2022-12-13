package logging

// A simple adapator which implements the Sentinel logger interface but actually logs
// using the Stanza global logger.
type SentinelAdaptor struct{}

func (l *SentinelAdaptor) Debug(msg string, keysAndValues ...interface{}) {
	Debug(msg, keysAndValues...)
}

func (l *SentinelAdaptor) DebugEnabled() bool {
	return true // todo: lookup logging level
}

func (l *SentinelAdaptor) Info(msg string, keysAndValues ...interface{}) {
	Info(msg, keysAndValues...)
}

func (l *SentinelAdaptor) InfoEnabled() bool {
	return true // todo: lookup logging level
}

func (l *SentinelAdaptor) Warn(msg string, keysAndValues ...interface{}) {
	Warn(msg, keysAndValues...)
}

func (l *SentinelAdaptor) WarnEnabled() bool {
	return true // todo: lookup logging level
}

func (l *SentinelAdaptor) Error(err error, msg string, keysAndValues ...interface{}) {
	Error(err, msg, keysAndValues...)
}

func (l *SentinelAdaptor) ErrorEnabled() bool {
	return true // todo: lookup logging level
}