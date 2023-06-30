package http

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	instrumentationName    = "github.com/StanzaSystems/sdk-go/handlers/http"
	instrumentationVersion = "0.0.1" // TODO: stanza sdk-go version/build number to go here

	// Stanza SDK Metrics:
	// https://github.com/StanzaSystems/sdk-spec#telemetry-metrics
	stanzaRequestAllowed   = "stanza.request.allowed"   // counter
	stanzaRequestBlocked   = "stanza.request.blocked"   // counter
	stanzaRequestFailed    = "stanza.request.failed"    // counter
	stanzaRequestSucceeded = "stanza.request.succeeded" // counter
	stanzaRequestDuration  = "stanza.request.duration"  // histogram (milliseconds)

	// Standard HTTP Client Metrics:
	// https://opentelemetry.io/docs/specs/otel/metrics/semantic_conventions/http-metrics/#http-client
	// httpClientDuration     = "http.client.duration"      // histogram
	// httpClientRequestSize  = "http.client.request.size"  // histogram
	// httpClientResponseSize = "http.client.response.size" // histogram

	// Standard HTTP Server Metrics:
	// https://opentelemetry.io/docs/specs/otel/metrics/semantic_conventions/http-metrics/#http-server
	// httpServerDuration       = "http.server.duration"        // histogram
	// httpServerRequestSize    = "http.server.request.size"    // histogram
	// httpServerResponseSize   = "http.server.response.size"   // histogram
	// httpServerActiveRequests = "http.server.active_requests" // counter
)

var (
	clientIdKey    = attribute.Key("client_id")
	customerIdKey  = attribute.Key("customer_id")
	decoratorKey   = attribute.Key("decorator")
	environmentKey = attribute.Key("environment")
	featureKey     = attribute.Key("feature")
	serviceKey     = attribute.Key("service")

	httpMeter *Meter = nil
)

type Meter struct {
	AllowedCount   metric.Int64Counter
	BlockedCount   metric.Int64Counter
	FailedCount    metric.Int64Counter
	SucceededCount metric.Int64Counter
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
	m.Duration, err = meter.Float64Histogram(
		stanzaRequestDuration,
		metric.WithUnit("ms"),
		metric.WithDescription("measures the duration of HTTP requests"))
	if err != nil {
		return nil, err
	}
	httpMeter = &m
	return httpMeter, nil
}
