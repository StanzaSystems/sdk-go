package handlers

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	instrumentationName    = "github.com/StanzaSystems/sdk-go"
	instrumentationVersion = "0.0.1-beta"

	// Stanza SDK Metrics:
	// https://github.com/StanzaSystems/sdk-spec#telemetry-metrics
	stanzaRequestAllowed   = "stanza.request.allowed"   // counter
	stanzaRequestBlocked   = "stanza.request.blocked"   // counter
	stanzaRequestFailed    = "stanza.request.failed"    // counter
	stanzaRequestSucceeded = "stanza.request.succeeded" // counter
	stanzaRequestDuration  = "stanza.request.duration"  // histogram (milliseconds)
)

var (
	clientIdKey    = attribute.Key("client_id")
	customerIdKey  = attribute.Key("customer_id")
	decoratorKey   = attribute.Key("decorator")
	environmentKey = attribute.Key("environment")
	featureKey     = attribute.Key("feature")
	serviceKey     = attribute.Key("service")
	reasonKey      = attribute.Key("reason")

	stanzaMeter *StanzaMeter
)

type StanzaMeter struct {
	AllowedCount   metric.Int64Counter
	BlockedCount   metric.Int64Counter
	FailedCount    metric.Int64Counter
	SucceededCount metric.Int64Counter
	Duration       metric.Float64Histogram
}

func GetInstrumentationName() string {
	return instrumentationName
}

func GetInstrumentationVersion() string {
	return instrumentationVersion
}

func GetStanzaMeter() (*StanzaMeter, error) {
	if stanzaMeter != nil {
		return stanzaMeter, nil
	}
	meter := otel.Meter(
		instrumentationName,
		metric.WithInstrumentationVersion(instrumentationVersion))

	var err error
	var m StanzaMeter
	m.AllowedCount, err = meter.Int64Counter(
		stanzaRequestAllowed,
		metric.WithUnit("1"),
		metric.WithDescription("measures the number of requests that were allowed"))
	if err != nil {
		return nil, err
	}
	m.BlockedCount, err = meter.Int64Counter(
		stanzaRequestBlocked,
		metric.WithUnit("1"),
		metric.WithDescription("measures the number of requests that were backpressured"))
	if err != nil {
		return nil, err
	}
	m.SucceededCount, err = meter.Int64Counter(
		stanzaRequestSucceeded,
		metric.WithUnit("1"),
		metric.WithDescription("measures the number of requests that succeeded"))
	if err != nil {
		return nil, err
	}
	m.FailedCount, err = meter.Int64Counter(
		stanzaRequestFailed,
		metric.WithUnit("1"),
		metric.WithDescription("measures the number of requests that failed"))
	if err != nil {
		return nil, err
	}
	m.Duration, err = meter.Float64Histogram(
		stanzaRequestDuration,
		metric.WithUnit("ms"),
		metric.WithDescription("measures the total execution time of decorated requests"))
	if err != nil {
		return nil, err
	}

	stanzaMeter = &m
	return stanzaMeter, nil
}
