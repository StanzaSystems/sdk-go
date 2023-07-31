package httphandler

import (
	"github.com/StanzaSystems/sdk-go/handlers"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	instrumentationName    = "github.com/StanzaSystems/sdk-go/handlers/http"
	instrumentationVersion = "0.0.1-beta"

	// Stanza SDK Metrics:
	// https://github.com/StanzaSystems/sdk-spec#telemetry-metrics
	stanzaRequestAllowed   = "stanza.request.allowed"   // counter
	stanzaRequestBlocked   = "stanza.request.blocked"   // counter
	stanzaRequestFailed    = "stanza.request.failed"    // counter
	stanzaRequestSucceeded = "stanza.request.succeeded" // counter
	// stanzaRequestLatency   = "stanza.request.latency"   // histogram (milliseconds)

	// Standard HTTP Client Metrics:
	// https://opentelemetry.io/docs/specs/otel/metrics/semantic_conventions/http-metrics/#http-client
	httpClientDuration     = "http.client.duration"      // histogram
	httpClientRequestSize  = "http.client.request.size"  // histogram
	httpClientResponseSize = "http.client.response.size" // histogram

	// Standard HTTP Server Metrics:
	// https://opentelemetry.io/docs/specs/otel/metrics/semantic_conventions/http-metrics/#http-server
	httpServerDuration       = "http.server.duration"        // histogram
	httpServerRequestSize    = "http.server.request.size"    // histogram
	httpServerResponseSize   = "http.server.response.size"   // histogram
	httpServerActiveRequests = "http.server.active_requests" // counter
)

var (
	httpRequestMethodKey = attribute.Key("http.request.method")
	httpResponseCodeKey  = attribute.Key("http.response.status_code")

	httpMeter *handlers.Meter = nil
)

func GetInstrumentationName() string {
	return instrumentationName
}

func GetInstrumentationVersion() string {
	return instrumentationVersion
}

func GetMeter() (*handlers.Meter, error) {
	if httpMeter != nil {
		return httpMeter, nil
	}
	meter := otel.Meter(
		instrumentationName,
		metric.WithInstrumentationVersion(instrumentationVersion),
	)

	var err error
	var m handlers.Meter
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

	m.ClientDuration, err = meter.Float64Histogram(
		httpClientDuration,
		metric.WithUnit("ms"),
		metric.WithDescription("measures the execution time of HTTP client requests"))
	if err != nil {
		return nil, err
	}
	m.ClientRequestSize, err = meter.Int64Histogram(
		httpClientRequestSize,
		metric.WithUnit("By"),
		metric.WithDescription("measures the size of HTTP client request messages"))
	if err != nil {
		return nil, err
	}
	m.ClientResponseSize, err = meter.Int64Histogram(
		httpClientResponseSize,
		metric.WithUnit("By"),
		metric.WithDescription("measures the size of HTTP client response messages"))
	if err != nil {
		return nil, err
	}

	m.ServerDuration, err = meter.Float64Histogram(
		httpServerDuration,
		metric.WithUnit("ms"),
		metric.WithDescription("measures the execution time of HTTP server requests"))
	if err != nil {
		return nil, err
	}
	m.ServerRequestSize, err = meter.Int64Histogram(
		httpServerRequestSize,
		metric.WithUnit("By"),
		metric.WithDescription("measures the size of HTTP server request messages"))
	if err != nil {
		return nil, err
	}
	m.ServerResponseSize, err = meter.Int64Histogram(
		httpServerResponseSize,
		metric.WithUnit("By"),
		metric.WithDescription("measures the size of HTTP server response messages"))
	if err != nil {
		return nil, err
	}
	m.ServerActiveRequests, err = meter.Int64UpDownCounter(
		httpServerActiveRequests,
		metric.WithUnit("1"),
		metric.WithDescription("measures the number of concurrent HTTP server requests in-flight"))
	if err != nil {
		return nil, err
	}

	httpMeter = &m
	return httpMeter, nil
}
