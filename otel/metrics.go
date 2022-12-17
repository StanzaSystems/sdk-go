package otel

import (
	"context"
	"time"

	// "go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

func initMetricsGrpc(ctx context.Context, res *resource.Resource) {
	// https://github.com/open-telemetry/opentelemetry-go/blob/main/exporters/otlp/otlpmetric/otlpmetricgrpc/config.go
	mo := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithTimeout(5 * time.Second), // TODO: be better than this...
		// otlpmetricgrpc.WithRetry(retryConfig)
		// otlpmetricgrpc.WithReconnectionPerid(10 * time.Second)
		otlpmetricgrpc.WithInsecure(), // TODO: what else needs to be done for TLS?
		// otlpmetricgrpc.WithTLSCredentials(creds)
		otlpmetricgrpc.WithEndpoint("a631a047755c747a59fe9a6be9491922-155001334.us-east-2.elb.amazonaws.com:4317"),
	}
	exp, err := otlpmetricgrpc.New(ctx, mo...)
	if err != nil {
		panic(err) // TODO: change to otel.Handle
	}
	mp := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(exp)),
	)
	defer func() {
		if err := mp.Shutdown(ctx); err != nil {
			panic(err) // TODO: change to otel.Handle?
		}
	}()
	global.SetMeterProvider(mp)
}
