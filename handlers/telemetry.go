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
	stanzaAllowed         = "stanza.decorator.allowed"          // counter
	stanzaAllowedSuccess  = "stanza.decorator.allowed.success"  // counter
	stanzaAllowedFailure  = "stanza.decorator.allowed.failure"  // counter
	stanzaAllowedDuration = "stanza.decorator.allowed.duration" // histogram (milliseconds)
	stanzaBlocked         = "stanza.decorator.blocked"          // counter
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
	AllowedCount        metric.Int64Counter
	AllowedFailureCount metric.Int64Counter
	AllowedSuccessCount metric.Int64Counter
	AllowedDuration     metric.Float64Histogram
	BlockedCount        metric.Int64Counter
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
		stanzaAllowed,
		metric.WithUnit("1"),
		metric.WithDescription("measures the number of requests that were allowed"))
	if err != nil {
		return nil, err
	}
	m.BlockedCount, err = meter.Int64Counter(
		stanzaBlocked,
		metric.WithUnit("1"),
		metric.WithDescription("measures the number of requests that were backpressured"))
	if err != nil {
		return nil, err
	}
	m.AllowedSuccessCount, err = meter.Int64Counter(
		stanzaAllowedSuccess,
		metric.WithUnit("1"),
		metric.WithDescription("measures the number of requests that succeeded"))
	if err != nil {
		return nil, err
	}
	m.AllowedFailureCount, err = meter.Int64Counter(
		stanzaAllowedFailure,
		metric.WithUnit("1"),
		metric.WithDescription("measures the number of requests that failed"))
	if err != nil {
		return nil, err
	}
	m.AllowedDuration, err = meter.Float64Histogram(
		stanzaAllowedDuration,
		metric.WithUnit("ms"),
		metric.WithDescription("measures the total execution time of decorated requests"))
	if err != nil {
		return nil, err
	}

	stanzaMeter = &m
	return stanzaMeter, nil
}
