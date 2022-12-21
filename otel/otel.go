package otel

import (
	"context"
	"os"

	"github.com/StanzaSystems/sdk-go/global"

	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/semconv/v1.12.0"
)

func Init(ctx context.Context) error {
	// TODO: connect to stanza-hub and get an otel config (at least a sample rate)?

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(global.Name()),
			semconv.ServiceVersionKey.String(global.Release()),
			semconv.DeploymentEnvironmentKey.String(global.Environment()),
		),
	)
	if err != nil {
		return err
	}

	if os.Getenv("STANZA_DEBUG") != "" {
		if _, err := initDebugMeter(res); err != nil {
			panic(err)
		}
		if _, err := initDebugTracer(res); err != nil {
			panic(err)
		}
		// TODO: add metrics and tracer provider shutdowns
	} else {
		if _, err := initGrpcMeter(ctx, res); err != nil {
			panic(err) // TODO: don't panic here
		}
		if _, err := initDebugTracer(res); err != nil { // TODO: initGrpcTracer
			panic(err) // TODO: don't panic here
		}
		// TODO: add metrics and tracer provider shutdowns
	}
	return nil
}
