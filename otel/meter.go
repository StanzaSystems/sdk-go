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
	exp, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithTimeout(5*time.Second), // TODO: be better than this...
		// otlpmetricgrpc.WithRetry(retryConfig)
		// otlpmetricgrpc.WithReconnectionPerid(10 * time.Second)
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
