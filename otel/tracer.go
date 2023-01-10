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

	// ParentBased will enable sampling if the Parent sampled, otherwise use *default*
	// raito of 1/10 of a percent (can be changed via Hub or STANZA_DEFAULT_TRACE_RATIO)
	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.ParentBased(trace.TraceIDRatioBased(config.traceRatio))),
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

	// ParentBased will enable sampling if the Parent sampled, otherwise use *default*
	// raito of 1/10 of a percent (can be changed via Hub or STANZA_DEFAULT_TRACE_RATIO)
	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.ParentBased(trace.TraceIDRatioBased(config.traceRatio))),
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
