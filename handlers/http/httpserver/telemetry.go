package httpserver

import (
	"github.com/StanzaSystems/sdk-go/handlers"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

const (
	// Standard HTTP Server Metrics:
	// https://opentelemetry.io/docs/specs/otel/metrics/semantic_conventions/http-metrics/#http-server
	httpServerDuration       = "http.server.duration"        // histogram
	httpServerRequestSize    = "http.server.request.size"    // histogram
	httpServerResponseSize   = "http.server.response.size"   // histogram
	httpServerActiveRequests = "http.server.active_requests" // counter
)

type HttpMeter struct {
	ServerDuration       metric.Float64Histogram
	ServerRequestSize    metric.Int64Histogram
	ServerResponseSize   metric.Int64Histogram
	ServerActiveRequests metric.Int64UpDownCounter
}

// var (
// 	httpRequestMethodKey = attribute.Key("http.request.method")
// 	httpResponseCodeKey  = attribute.Key("http.response.status_code")
// )

func NewHttpMeter() (*HttpMeter, error) {
	name := handlers.GetInstrumentationName()
	version := handlers.GetInstrumentationVersion()
	meter := otel.Meter(name, metric.WithInstrumentationVersion(version))

	var err error
	var m HttpMeter

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

	return &m, nil
}
