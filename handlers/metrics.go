package handlers

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	clientIdKey    = attribute.Key("client_id")
	customerIdKey  = attribute.Key("customer_id")
	decoratorKey   = attribute.Key("decorator")
	environmentKey = attribute.Key("environment")
	featureKey     = attribute.Key("feature")
	serviceKey     = attribute.Key("service")
	reasonKey      = attribute.Key("reason")
)

type Meter struct {
	AllowedCount   metric.Int64Counter
	BlockedCount   metric.Int64Counter
	FailedCount    metric.Int64Counter
	SucceededCount metric.Int64Counter

	ClientDuration     metric.Float64Histogram
	ClientRequestSize  metric.Int64Histogram
	ClientResponseSize metric.Int64Histogram

	ServerDuration       metric.Float64Histogram
	ServerRequestSize    metric.Int64Histogram
	ServerResponseSize   metric.Int64Histogram
	ServerActiveRequests metric.Int64UpDownCounter
}
