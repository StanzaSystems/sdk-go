package otel

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"google.golang.org/grpc/credentials"
)

func newMeterProvider(ctx context.Context, res *resource.Resource, headers map[string]string, url string) (*metric.MeterProvider, error) {
	if os.Getenv("STANZA_OTEL_DEBUG") != "" {
		return initDebugMeter(res)
	} else {
		return initGrpcMeter(ctx, res, url, headers)
	}
}

func initDebugMeter(res *resource.Resource) (*metric.MeterProvider, error) {
	exporter, err := stdoutmetric.New()
	if err != nil {
		return nil, fmt.Errorf("creating stdout meter exporter: %w", err)
	}
	mp := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(exporter)))
	return mp, nil
}

func initGrpcMeter(ctx context.Context, res *resource.Resource, url string, headers map[string]string) (*metric.MeterProvider, error) {
	opts := []otlpmetricgrpc.Option{
		// WithRetry sets the retry policy for transient retryable errors that are
		//   returned by the target collector endpoint.
		//   Defaults to 5 seconds after receiving a retryable error and increase
		//   exponentially after each error for no more than a total time of 1 minute.
		// otlpmetricgrpc.WithRetry(otlpmetricgrpc.RetryConfig{
		// 	Enabled:         true,
		// 	InitialInterval: 5 * time.Second,
		// 	MaxInterval:     30 * time.Second,
		// 	MaxElapsedTime:  time.Minute,
		// }),
		//
		// WithTimeout sets the max amount of time an Exporter will attempt an export.
		//   This takes precedence over the retry settings defined above. Once this time
		//   limit has been reached the export is abandoned and the metric data is dropped.
		//   Defaults to 10 seconds.
		// otlpmetricgrpc.WithTimeout(30 * time.Second),
		//
		// WithReconnectionPeriod sets the minimum amount of time between connection
		//   attempts to the target endpoint.
		// otlpmetricgrpc.WithReconnectionPeriod(1 * time.Minute),
		//
		otlpmetricgrpc.WithEndpoint(url),
		otlpmetricgrpc.WithHeaders(headers),
	}
	if os.Getenv("STANZA_OTEL_NO_TLS") != "" { // disable TLS for local OTEL development
		opts = append(opts,
			otlpmetricgrpc.WithInsecure(),
		)
	} else {
		opts = append(opts,
			otlpmetricgrpc.WithTLSCredentials(credentials.NewTLS(&tls.Config{})),
		)
	}
	exp, err := otlpmetricgrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("creating OTLP gRPC meter exporter: %w", err)
	}
	mp := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(
			metric.NewPeriodicReader(exp,
				metric.WithInterval(10*time.Second))),
	)
	return mp, nil
}
