package otel

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
)

func initDebugTracer(resource *resource.Resource) (*trace.TracerProvider, error) {
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, fmt.Errorf("creating stdout trace exporter: %w", err)
	}
	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithBatcher(exporter),
		trace.WithResource(resource),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{}))
	return tp, nil
}

func initGrpcTracer(ctx context.Context, resource *resource.Resource) (*trace.TracerProvider, error) {
	retryConfig := otlptracegrpc.RetryConfig{
		Enabled:         true,
		InitialInterval: 5 * time.Second,
		MaxInterval:     30 * time.Second,
		MaxElapsedTime:  2 * time.Minute,
	}
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithRetry(retryConfig),
		otlptracegrpc.WithInsecure(), // TODO: what else needs to be done for TLS?
		// otlptracegrpc.WithTLSCredentials(creds)
	)
	if err != nil {
		return nil, fmt.Errorf("creating OTLP trace exporter: %w", err)
	}
	tp := trace.NewTracerProvider(
		// TODO: make default trace sampler a tiny fractional (will be able to override from hub)
		// https://github.com/open-telemetry/opentelemetry-go/blob/main/sdk/trace/sampling.go
		// -- need to understand remote vs local and parent to be a good otel citizen
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithBatcher(exporter),
		trace.WithResource(resource),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{}))
	return tp, nil
}
