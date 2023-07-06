package otel

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"

	hubv1 "github.com/StanzaSystems/sdk-go/proto/stanza/hub/v1"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials"
)

func initDebugTracer(resource *resource.Resource, config *hubv1.TraceConfig) (*trace.TracerProvider, error) {
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, fmt.Errorf("creating stdout trace exporter: %w", err)
	}

	// ParentBased will enable sampling if the Parent sampled, otherwise use the
	// default sample rate given by Hub (which is 1/10th of 1% of requests).
	// TODO: Handle trace sample rate overrides
	sampleRate := float64(config.GetSampleRateDefault())

	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.ParentBased(trace.TraceIDRatioBased(sampleRate))),
		trace.WithBatcher(exporter),
		trace.WithResource(resource),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}

func initGrpcTracer(ctx context.Context, resource *resource.Resource, config *hubv1.TraceConfig, token string) (*trace.TracerProvider, error) {
	opts := []otlptracegrpc.Option{
		// WithRetry sets the retry policy for transient retryable errors that may be
		//   returned by the target collector endpoint when exporting a batch of spans.
		//   Defaults to 5 seconds after receiving a retryable error and increase
		//   exponentially after each error for no more than a total time of 1 minute.
		// otlptracegrpc.WithRetry(otlptracegrpc.RetryConfig{
		// 	Enabled:         true,
		// 	InitialInterval: 5 * time.Second,
		// 	MaxInterval:     30 * time.Second,
		// 	MaxElapsedTime:  time.Minute,
		// }),
		//
		// WithTimeout sets the max amount of time a client will attempt to export a
		//   batch of spans. This takes precedence over any retry settings defined above,
		//   once this time limit has been reached the export is abandoned and the batch
		//   of spans is dropped. Defaults to 10 seconds.
		// otlptracegrpc.WithTimeout(30 * time.Second),
		//
		// WithReconnectionPeriod sets the minimum amount of time between connection
		//   attempts to the target endpoint.
		// otlptracegrpc.WithReconnectionPeriod(30 * time.Second),
		//
		otlptracegrpc.WithEndpoint(config.GetCollectorUrl()),
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

	// ParentBased will enable sampling if the Parent sampled, otherwise use the
	// default sample rate given by Hub (which is 1/10th of 1% of requests).
	// TODO: Handle trace sample rate overrides
	sampleRate := float64(config.GetSampleRateDefault())

	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.ParentBased(trace.TraceIDRatioBased(sampleRate))),
		trace.WithBatcher(exporter),
		trace.WithResource(resource),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}
