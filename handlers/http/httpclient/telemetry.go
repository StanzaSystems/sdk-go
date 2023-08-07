package httpclient

import (
	"github.com/StanzaSystems/sdk-go/handlers"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	// Standard HTTP Client Metrics:
	// https://opentelemetry.io/docs/specs/otel/metrics/semantic_conventions/http-metrics/#http-client
	httpClientDuration     = "http.client.duration"      // histogram
	httpClientRequestSize  = "http.client.request.size"  // histogram
	httpClientResponseSize = "http.client.response.size" // histogram
)

type HttpMeter struct {
	ClientDuration     metric.Float64Histogram
	ClientRequestSize  metric.Int64Histogram
	ClientResponseSize metric.Int64Histogram
}

var (
	httpRequestMethodKey = attribute.Key("http.request.method")
	httpResponseCodeKey  = attribute.Key("http.response.status_code")
)

func NewHttpMeter() (*HttpMeter, error) {
	meter := otel.Meter(
		handlers.GetInstrumentationName(),
		metric.WithInstrumentationVersion(handlers.GetInstrumentationVersion()))

	var err error
	var m HttpMeter

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

	return &m, nil
}
