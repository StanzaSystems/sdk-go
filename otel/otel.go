package otel

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

func Init(ctx context.Context, name, rel, env string) error {
	// TODO: connect to stanza-hub and get an otel config (at least a sample rate)?

	res, err := resource.New(ctx,
		resource.WithHost(),
		resource.WithFromEnv(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(name),
			semconv.ServiceVersionKey.String(rel),
			semconv.DeploymentEnvironmentKey.String(env),
		),
	)
	if err != nil {
		return fmt.Errorf("creating opentelemetry resource: %w", err)
	}

	if os.Getenv("STANZA_DEBUG") != "" {
		if _, err := initDebugMeter(res); err != nil {
			panic(err)
		}
		if _, err := initDebugTracer(res); err != nil {
			panic(err)
		}
	} else {
		if _, err := initGrpcMeter(ctx, res); err != nil {
			panic(err) // TODO: don't panic here
		}
		if _, err := initGrpcTracer(ctx, res); err != nil {
			panic(err) // TODO: don't panic here
		}
	}
	// Handle shutdown to ensure all sub processes are closed correctly and telemetry is exported
	//
	// TODO: add something like the below (but NOT just deferred from here)
	// defer func() {
	// 	_ = exp.Shutdown(ctx)
	// 	_ = tp.Shutdown(ctx)
	// }()
	return nil
}
