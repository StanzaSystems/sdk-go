package otel

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

func initDebugMeter(res *resource.Resource) (*metric.MeterProvider, error) {
	exporter, err := stdoutmetric.New()
	if err != nil {
		return nil, fmt.Errorf("creating stdout meter exporter: %w", err)
	}
	mp := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(exporter)))
	global.SetMeterProvider(mp)
	return mp, nil
}

func initGrpcMeter(ctx context.Context, res *resource.Resource) (*metric.MeterProvider, error) {
	retryConfig := otlpmetricgrpc.RetryConfig{
		Enabled:         true,
		InitialInterval: 5 * time.Second,
		MaxInterval:     30 * time.Second,
		MaxElapsedTime:  time.Minute,
	}
	exp, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithRetry(retryConfig),
		otlpmetricgrpc.WithInsecure(), // TODO: what else needs to be done for TLS?
		// otlpmetricgrpc.WithTLSCredentials(creds)
	)
	if err != nil {
		return nil, fmt.Errorf("creating OTLP gRPC meter exporter: %w", err)
	}
	mp := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(exp)))
	global.SetMeterProvider(mp)
	return mp, nil
}
