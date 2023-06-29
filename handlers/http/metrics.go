package http

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Stanza SDK Metrics:
// https://github.com/StanzaSystems/sdk-spec#telemetry-metrics
const (
	instrumentationName    = "github.com/StanzaSystems/sdk-go/handlers/http"
	instrumentationVersion = "0.0.1" // TODO: stanza sdk-go version/build number to go here

	stanzaRequestAllowed   = "stanza.request.allowed"   // counter
	stanzaRequestBlocked   = "stanza.request.blocked"   // counter
	stanzaRequestFailed    = "stanza.request.failed"    // counter
	stanzaRequestSucceeded = "stanza.request.succeeded" // counter
	stanzaRequestTotal     = "stanza.request.total"     // counter
	stanzaRequestDuration  = "stanza.request.duration"  // histogram (milliseconds)
)

var (
	environmentKey = attribute.Key("stanza.environment")
	decoratorKey   = attribute.Key("stanza.decorator")
	featureKey     = attribute.Key("stanza.feature")
	clientIdKey    = attribute.Key("stanza.client_id")

	httpMeter *Meter = nil
)

type Meter struct {
	AllowedCount   metric.Int64Counter
	BlockedCount   metric.Int64Counter
	FailedCount    metric.Int64Counter
	SucceededCount metric.Int64Counter
	TotalCount     metric.Int64Counter
	Duration       metric.Float64Histogram
	RequestSize    metric.Int64Histogram
	ResponseSize   metric.Int64Histogram
}

func GetMeter() (*Meter, error) {
	if httpMeter != nil {
		return httpMeter, nil
	}

	meter := otel.Meter(
		instrumentationName,
		metric.WithInstrumentationVersion(instrumentationVersion),
	)

	var err error
	var m Meter
	m.AllowedCount, err = meter.Int64Counter(
		stanzaRequestAllowed,
		metric.WithUnit("1"),
		metric.WithDescription("measures the number of HTTP requests that were allowed"))
	if err != nil {
		return nil, err
	}
	m.BlockedCount, err = meter.Int64Counter(
		stanzaRequestBlocked,
		metric.WithUnit("1"),
		metric.WithDescription("measures the number of HTTP requests that were backpressured"))
	if err != nil {
		return nil, err
	}
	m.SucceededCount, err = meter.Int64Counter(
		stanzaRequestSucceeded,
		metric.WithUnit("1"),
		metric.WithDescription("measures the number of HTTP requests that succeeded"))
	if err != nil {
		return nil, err
	}
	m.FailedCount, err = meter.Int64Counter(
		stanzaRequestFailed,
		metric.WithUnit("1"),
		metric.WithDescription("measures the number of HTTP requests that failed"))
	if err != nil {
		return nil, err
	}
	m.TotalCount, err = meter.Int64Counter(
		stanzaRequestTotal,
		metric.WithUnit("1"),
		metric.WithDescription("measures the number of HTTP requests that were seen"))
	if err != nil {
		return nil, err
	}
	m.Duration, err = meter.Float64Histogram(
		stanzaRequestDuration,
		metric.WithUnit("ms"),
		metric.WithDescription("measures the duration inbound HTTP requests"))
	if err != nil {
		return nil, err
	}
	httpMeter = &m
	return httpMeter, nil
}
