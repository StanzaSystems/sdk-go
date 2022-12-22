package otel

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

func initDebugMeter(res *resource.Resource) (*metric.MeterProvider, error) {
	exp, err := stdoutmetric.New()
	if err != nil {
		return nil, err
	}
	mp := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(exp)))
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
		// otlpmetricgrpc.WithEndpoint("a631a047755c747a59fe9a6be9491922-155001334.us-east-2.elb.amazonaws.com:4317"),
		// otlpmetricgrpc.WithEndpoint("localhost:4317"),
	)
	if err != nil {
		return nil, err
	}
	mp := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(exp)))
	global.SetMeterProvider(mp)
	return mp, nil
}
