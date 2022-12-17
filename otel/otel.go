package otel

import (
	"context"

	"github.com/StanzaSystems/sdk-go/global"

	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/semconv/v1.9.0"
)

func Init(ctx context.Context) error {
	// TODO: connect to an otel collector here?
	// TODO: connect to stanza-hub and get an otel config?

	res, err := resource.New(ctx,
		resource.WithFromEnv(), // pull attributes from OTEL_RESOURCE_ATTRIBUTES and OTEL_SERVICE_NAME environment variables
		resource.WithAttributes(
			semconv.ServiceNameKey.String(global.Name()),
			semconv.ServiceVersionKey.String(global.Release()),
			semconv.DeploymentEnvironmentKey.String(global.Environment()),
		),
	)
	if err != nil {
		return err
	}

	initMetricsGrpc(ctx, res)

	// if err := sg.SetOtelConfig(meterProvider, otel.GetTextMapPropagator()); err != nil {
	// 	return err
	// }
	return nil
}
