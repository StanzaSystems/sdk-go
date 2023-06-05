package otel

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"time"

	hubv1 "github.com/StanzaSystems/sdk-go/proto/stanza/hub/v1"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials"
)

func initDebugTracer(resource *resource.Resource, config *hubv1.TraceConfig) (*trace.TracerProvider, error) {
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, fmt.Errorf("creating stdout trace exporter: %w", err)
	}

	// ParentBased will enable sampling if the Parent sampled, otherwise use *default*
	// ratio of 1/10 of a percent (can be changed via Hub or STANZA_DEFAULT_TRACE_RATIO)
	// TODO: Handle trace sample rate overrides
	sampleRate := float64(config.GetSampleRateDefault())

	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.ParentBased(trace.TraceIDRatioBased(sampleRate))),
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

func initGrpcTracer(ctx context.Context, resource *resource.Resource, config *hubv1.TraceConfig, token string) (*trace.TracerProvider, error) {
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithRetry(otlptracegrpc.RetryConfig{
			Enabled:         true,
			InitialInterval: 5 * time.Second,
			MaxInterval:     30 * time.Second,
			MaxElapsedTime:  2 * time.Minute,
		}),
		otlptracegrpc.WithHeaders(map[string]string{
			"Authorization": "Bearer " + token,
		}),
	}
	if os.Getenv("STANZA_OTEL_NO_TLS") != "" { // disable TLS for local OTEL development
		opts = append(opts,
			otlptracegrpc.WithInsecure(),
		)
	} else {
		opts = append(opts,
			otlptracegrpc.WithTLSCredentials(credentials.NewTLS(&tls.Config{})),
		)
	}
	exporter, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("creating OTLP trace exporter: %w", err)
	}

	// ParentBased will enable sampling if the Parent sampled, otherwise use *default*
	// ratio of 1/10 of a percent (can be changed via Hub or STANZA_DEFAULT_TRACE_RATIO)
	// TODO: Handle trace sample rate overrides
	sampleRate := float64(config.GetSampleRateDefault())

	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.ParentBased(trace.TraceIDRatioBased(sampleRate))),
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
