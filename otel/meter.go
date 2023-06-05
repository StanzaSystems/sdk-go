package otel

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"time"

	hubv1 "github.com/StanzaSystems/sdk-go/proto/stanza/hub/v1"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"google.golang.org/grpc/credentials"
)

func initDebugMeter(res *resource.Resource) (*metric.MeterProvider, error) {
	exporter, err := stdoutmetric.New()
	if err != nil {
		return nil, fmt.Errorf("creating stdout meter exporter: %w", err)
	}
	mp := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(exporter)))
	otel.SetMeterProvider(mp)
	return mp, nil
}

func initGrpcMeter(ctx context.Context, res *resource.Resource, config *hubv1.MetricConfig, token string) (*metric.MeterProvider, error) {
	opts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithRetry(otlpmetricgrpc.RetryConfig{
			Enabled:         true,
			InitialInterval: 5 * time.Second,
			MaxInterval:     30 * time.Second,
			MaxElapsedTime:  time.Minute,
		}),
		otlpmetricgrpc.WithHeaders(map[string]string{
			"Authorization": "Bearer " + token,
		}),
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
		metric.WithReader(metric.NewPeriodicReader(exp)))
	otel.SetMeterProvider(mp)
	return mp, nil
}
