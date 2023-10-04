package otel

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// GetTracePropagator is a passthrough helper function
func GetTracerProvider() trace.TracerProvider {
	return otel.GetTracerProvider()
}

// GetTextMapPropagator is a passthrough helper function
func GetTextMapPropagator() propagation.TextMapPropagator {
	return otel.GetTextMapPropagator()
}

// InitTextMapPropagator is a passthrough helper function
func InitTextMapPropagator(propagator propagation.TextMapPropagator) {
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
			propagator))
}
