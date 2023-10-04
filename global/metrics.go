package global

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const (
	// Stanza SDK Metrics:
	// https://github.com/StanzaSystems/sdk-spec#telemetry-metrics
	stanzaAllowed         = "stanza.guard.allowed"          // counter
	stanzaAllowedSuccess  = "stanza.guard.allowed.success"  // counter
	stanzaAllowedFailure  = "stanza.guard.allowed.failure"  // counter
	stanzaAllowedUnknown  = "stanza.guard.allowed.unknown"  // counter
	stanzaAllowedDuration = "stanza.guard.allowed.duration" // histogram (milliseconds)
	stanzaBlocked         = "stanza.guard.blocked"          // counter
)

type StanzaMeter struct {
	AllowedCount        metric.Int64Counter
	AllowedSuccessCount metric.Int64Counter
	AllowedFailureCount metric.Int64Counter
	AllowedUnknownCount metric.Int64Counter
	AllowedDuration     metric.Float64Histogram
	BlockedCount        metric.Int64Counter
}

func NewStanzaTracer() *trace.Tracer {
	t := otel.GetTracerProvider().Tracer(
		InstrumentationName(),
		InstrumentationTraceVersion(),
	)
	return &t
}

func NewStanzaMeter() (*StanzaMeter, error) {
	om := otel.GetMeterProvider().Meter(
		InstrumentationName(),
		InstrumentationMetricVersion(),
	)

	var m StanzaMeter
	var err error
	m.AllowedCount, err = om.Int64Counter(
		stanzaAllowed,
		metric.WithUnit("1"),
		metric.WithDescription("measures the number of executions that were allowed"))
	if err != nil {
		return nil, err
	}
	m.BlockedCount, err = om.Int64Counter(
		stanzaBlocked,
		metric.WithUnit("1"),
		metric.WithDescription("measures the number of executions that were backpressured"))
	if err != nil {
		return nil, err
	}
	m.AllowedSuccessCount, err = om.Int64Counter(
		stanzaAllowedSuccess,
		metric.WithUnit("1"),
		metric.WithDescription("measures the number of executions that succeeded"))
	if err != nil {
		return nil, err
	}
	m.AllowedFailureCount, err = om.Int64Counter(
		stanzaAllowedFailure,
		metric.WithUnit("1"),
		metric.WithDescription("measures the number of executions that failed"))
	if err != nil {
		return nil, err
	}
	m.AllowedUnknownCount, err = om.Int64Counter(
		stanzaAllowedUnknown,
		metric.WithUnit("1"),
		metric.WithDescription("measures the number of executions where the success (or failure) was unknown"))
	if err != nil {
		return nil, err
	}
	m.AllowedDuration, err = om.Float64Histogram(
		stanzaAllowedDuration,
		metric.WithUnit("ms"),
		metric.WithDescription("measures the total executions time of guarded requests"))
	if err != nil {
		return nil, err
	}

	return &m, nil
}
